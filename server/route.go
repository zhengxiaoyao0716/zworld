package server

import (
	"log"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/gin-gonic/gin"
	"github.com/zhengxiaoyao0716/util/requests"
	"github.com/zhengxiaoyao0716/zmodule/config"
)

var routeSet *hashset.Set

// Route .
func Route(c *gin.Context) {
	var json gin.H
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(400, gin.H{"reason": "invalid json param"})
		return
	}
	addr := json["addr"].(string)
	routeSet.Add(addr)
	c.JSON(200, gin.H{"route": routeSet.Values()})
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
		routeSet.Add(json["route"].([]interface{})...)
		return
	}
	log.Println("all routes synchronize failed.")
}
