package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/util/requests"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zworld/server/chain"
)

var routeSet struct {
	Values   func() []interface{}
	Add      func(...interface{})
	Contains func(...interface{}) bool
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

	routes := strings.Fields(config.GetString("route"))
	works := make(chan bool, len(routes))

	addr := config.GetString("server")
	req := requests.New(&http.Client{Timeout: 60 * time.Second})
	for _, route := range routes {
		go func(route string) {
			defer func() {
				err := recover()
				if err != nil {
					log.Println(err)
					works <- false
				}
				works <- true
			}()

			r, err := req.Post(route, map[string]interface{}{
				"addr": addr,
			})
			if err != nil {
				log.Println("synchronize route failed:", route)
				log.Panicln(err)
			}
			json := easyjson.Object(r.JSON())
			if json == nil {
				log.Panicln("invalid response of route: ", route)
			}
			_, routeErr := json.ArrayAt("route")
			_, chainErr := json.StringAt("chain")
			if routeErr != nil || chainErr != nil {
				log.Panicln("invalid response of route: %s, %v.", route, json)
			}
			// TODO 验证身份，交换公钥等
			if err := chain.Load(string(json.MustStringAt("chain"))); err != nil {
				log.Println("synchronize route failed:", route)
				log.Panicln(err)
			}
			routeSet.Add(json.MustArrayAt("route")...)
		}(route)
	}

	oks := 0
	for i := 0; i < len(routes); i++ {
		if <-works {
			oks++
		}
	}
	if oks == 0 {
		log.Fatalln("all routes synchronize failed.")
	}
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
		Values   func() []interface{}
		Add      func(...interface{})
		Contains func(...interface{}) bool
	}{rs.Values, func(items ...interface{}) { addQueue <- items }, rs.Contains}
}
