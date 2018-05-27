package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zhengxiaoyao0716/zworld/server/secret"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/util/requests"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zworld/server/chain"
)

var routeSet struct {
	*hashset.Set
	Add func(...interface{})
}

func routeHandler(json *easyjson.Object) (resp easyjson.Object) {
	defer func() {
		text, err := chain.Dump()
		if err != nil {
			panic(err)
		}
		resp = easyjson.Object{"route": routeSet.Values(), "chain": text}
	}()
	if json.IsEmpty() {
		return
	}
	addr := string(json.MustStringAt("addr"))
	if routeSet.Contains(addr) {
		resp = easyjson.Object{"ok": "false", "code": 403, "reason": "address already used, addr: " + addr}
	}
	// TODO 验证身份，交换公钥等
	routeSet.Add(addr)
	return
}

func p2pRun() {
	startRouteSetService()
	chain.Init()

	addr := config.GetString("server")
	syncRoute(addr)
	joinChain(addr)
}

func syncRoute(addr string) {
	anyOk := false
	for err := range p2pBroadcase(
		"/route",
		map[string]interface{}{"addr": addr},
		func(json *easyjson.Object, route string) error {
			_, routeErr := json.ArrayAt("route")
			_, chainErr := json.StringAt("chain")
			if routeErr != nil || chainErr != nil {
				return fmt.Errorf("invalid response: %v", json)
			}
			// TODO 验证身份，交换公钥等
			if err := chain.Load(string(json.MustStringAt("chain"))); err != nil {
				return err
			}
			routeSet.Add(json.MustArrayAt("route")...)
			return nil
		},
	) {
		if err == nil {
			anyOk = true
		} else {
			log.Println("sync route failed,", err)
		}
	}
	if !anyOk {
		log.Fatalln("all routes synchronize failed.")
	}
}

func joinChain(addr string) {
	block, err := chain.NewBlock(chain.Data{addr, secret.Pubkey()})
	if err != nil {
		log.Fatalln(err)
	}
	if err := chain.Insert(block); err != nil {
		log.Fatalln(err)
	}
	bytes, err := chain.Dump()
	if err != nil {
		log.Fatalln(err)
	}
	sign := chain.Signature()
	anyFail := false
	for err := range p2pBroadcase("/chain/update", map[string]interface{}{"chain": bytes}, func(json *easyjson.Object, route string) error {
		signature, err := json.StringAt("signature")
		if err != nil {
			return fmt.Errorf("invalid response, missing field 'signature': %v", json)
		}
		if string(signature) != sign {
			return fmt.Errorf("signature not match, signature: %s, expected: %s", string(signature), sign)
		}
		return nil
	}) {
		if err != nil {
			log.Println("join chain failed,", err)
			anyFail = true
		}
	}
	if anyFail {
		log.Fatalln("failed to join the chain.")
	}
}

func p2pBroadcase(path string, data map[string]interface{}, action func(*easyjson.Object, string) error) chan error {
	var routes []string
	if routeSet.Set.Size() == 0 {
		routes = strings.Fields(config.GetString("route"))
	} else {
		for _, route := range routeSet.Values() {
			routes = append(routes, route.(string))
		}
	}
	works := make(chan error, len(routes))
	req := requests.New(&http.Client{Timeout: 60 * time.Second})
	for _, route := range routes {
		go func(route string) {
			defer func() {
				err := recover()
				if err != nil {
					works <- fmt.Errorf("broadcast route (%s) failed, %v", route, err)
				}
				works <- nil
			}()

			r, err := req.Post(route, data)
			if err != nil {
				panic(err)
			}
			json := easyjson.Object(r.JSON())
			if json == nil {
				panic(errors.New("unexpected response data type, expect json"))
			}

			if err := action(&json, route); err != nil {
				panic(err)
			}
		}("http://" + route + path) // TODO https?
	}

	errs := make(chan error)
	go func() {
		for i := 0; i < len(routes); i++ {
			errs <- <-works
		}
		close(works)
		close(errs)
	}()
	return errs
}

func startRouteSetService() {
	rs := hashset.New()
	addQueue := make(chan []interface{}) // routeSet的同步Add队列
	go func() {
		for {
			items := <-addQueue
			rs.Add(items...)
		}
	}()
	routeSet = struct {
		*hashset.Set
		Add func(...interface{})
	}{rs, func(items ...interface{}) { addQueue <- items }}
}
