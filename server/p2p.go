package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/util/requests"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zworld/server/chain"
	"github.com/zhengxiaoyao0716/zworld/server/component"
	"github.com/zhengxiaoyao0716/zworld/server/secret"
)

type routeSetType struct {
	set    map[int]*hashset.Set
	Add    func(int, string)
	extend func(...interface{})
	Has    func(int, string) bool
	Each   func(func(int, string))
	All    func() [][2]interface{}
	Move   func(int, int, string)
}

var routeSet routeSetType

func routeHandler(json *easyjson.Object) (resp easyjson.Object) {
	defer func() {
		text, err := chain.Dump()
		if err != nil {
			panic(err)
		}
		resp = easyjson.Object{"route": routeSet.All(), "chain": text}
	}()
	if json.IsEmpty() {
		return
	}
	addr := string(json.MustStringAt("addr"))
	if routeSet.Has(-1, addr) {
		resp = easyjson.Object{"ok": "false", "code": 403, "reason": "address already used, addr: " + addr}
	}
	// TODO 验证身份，交换公钥等
	routeSet.Add(-1, addr)
	return
}
func routeShiftChunkHandler(json *easyjson.Object) (resp easyjson.Object) {
	address := string(json.MustStringAt("addr"))
	chunkID := int(json.MustNumberAt("id"))
	old := int(json.MustNumberAt("old", -1))
	routeSet.Move(old, chunkID, address)

	return easyjson.Object{"cache": component.TodoSearchCacheDump(chunkID)}
}

func p2pRun() {
	startRouteSetService()
	chain.Init()

	addr := config.GetString("server")
	syncRoute(addr)
	joinChain(addr)

	component.Init(component.InitArg{
		Pubkey: secret.Pubkey(),
	})
	notifyChunkShift()
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
			routeSet.extend(json.MustArrayAt("route")...)
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
	block, err := chain.NewBlock(chain.Data{Server: addr, Pubkey: secret.Pubkey()})
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
	anyOk := false
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
		if err == nil {
			anyOk = true
		} else {
			log.Println("join chain failed,", err)
		}
	}
	if !anyOk {
		log.Fatalln("failed to join the chain.")
	}
}

var chunkID = -1

func notifyChunkShift() {
	id := component.TodoMyPlaceChunkIndex()
	addr := config.GetString("server")

	anyOk := false
	for err := range p2pBroadcase(
		"/route/chunk/shift",
		map[string]interface{}{"old": chunkID, "id": id, "addr": addr},
		func(json *easyjson.Object, route string) error {
			cache, err := json.StringAt("cache")
			if err != nil {
				return err
			}
			if err := component.TodoCacheLoad(string(cache)); err != nil {
				return err
			}
			return nil
		},
	) {
		if err == nil {
			anyOk = true
		} else {
			log.Println("chunk shift failed,", err)
		}
	}
	if !anyOk {
		log.Fatalf("failed to shift chunk: %d\n", id)
	}
	chunkID = id
}

func p2pBroadcase(path string, data map[string]interface{}, action func(*easyjson.Object, string) error) chan error {
	var routes []string
	if len(routeSet.set) == 0 {
		routes = strings.Fields(config.GetString("route"))
	} else {
		routeSet.Each(func(_ int, route string) {
			routes = append(routes, route)
		})
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
	run := make(chan func()) // routeSet的同步队列
	go func() {
		for {
			action := <-run
			action()
		}
	}()
	rs := map[int]*hashset.Set{}
	routeSet = routeSetType{
		set: rs,
		Add: func(chunkID int, address string) {
			wait := make(chan bool)
			run <- func() {
				set, ok := rs[chunkID]
				if !ok {
					set = hashset.New()
					rs[chunkID] = set
				}
				set.Add(address)
				wait <- true
			}
			<-wait
		},
		extend: func(items ...interface{}) {
			wait := make(chan bool)
			run <- func() {
				for _, value := range items {
					route := value.([]interface{})
					chunkID := int(route[0].(float64))
					address := route[1].(string)
					set, ok := rs[chunkID]
					if !ok {
						set = hashset.New()
						rs[chunkID] = set
					}
					set.Add(address)
				}
				wait <- true
			}
			<-wait
		},
		Has: func(chunkID int, adddress string) bool {
			has := make(chan bool)
			run <- func() {
				if set, ok := rs[chunkID]; ok {
					has <- set.Contains(adddress)
					return
				}
				has <- false
			}
			return <-has
		},
		Each: func(each func(int, string)) {
			wait := make(chan bool)
			run <- func() {
				for chunkID, set := range rs {
					for _, value := range set.Values() {
						each(chunkID, value.(string))
					}
				}
				wait <- true
			}
			<-wait
		},
		All: func() (entries [][2]interface{}) {
			routeSet.Each(func(id int, addr string) {
				entries = append(entries, [2]interface{}{id, addr})
			})
			return entries
		},
		Move: func(old, chunkID int, address string) {
			wait := make(chan bool)
			run <- func() {
				if set, ok := rs[old]; ok {
					set.Remove(address)
				}
				set, ok := rs[chunkID]
				if !ok {
					set = hashset.New()
					rs[chunkID] = set
				}
				set.Add(address)
				wait <- true
			}
			<-wait
		},
	}
}
