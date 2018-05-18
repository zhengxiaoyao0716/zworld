package server

import (
	"log"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/util/requests"
	"github.com/zhengxiaoyao0716/zmodule/config"
)

var routeSet *hashset.Set

func routeHandler(json *easyjson.Object) easyjson.Object {
	if json.IsEmpty() {
		return easyjson.Object{"route": routeSet.Values()}
	}
	routeSet.Add(string(json.MustStringAt("addr")))
	return easyjson.Object{"route": routeSet.Values()}
}

func initRoute() {
	addr := config.GetString("server")
	routeSet = hashset.New()

	routes := strings.Fields(config.GetString("route"))
	for _, route := range routes {
		r, err := requests.Post(route, map[string]interface{}{
			"addr": addr,
		})
		if err != nil {
			log.Println("synchronize route failed:", route)
			log.Println(err)
			continue
		}
		json := r.JSON()
		if json == nil {
			log.Println("invalid response of route: ", route)
			continue
		}
		if _, ok := json["route"]; !ok {
			log.Printf("invalid response of route: %s, %v.", route, json)
			continue
		}
		// TODO 这里需要校验身份
		routeSet.Add(json["route"].([]interface{})...)
		return
	}
	log.Println("all routes synchronize failed.")
}
