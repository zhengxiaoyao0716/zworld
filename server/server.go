package server

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zhengxiaoyao0716/util/cout"
	"github.com/zhengxiaoyao0716/util/safefile"
	"github.com/zhengxiaoyao0716/zcli/connect"
	"github.com/zhengxiaoyao0716/zcli/server"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/file"
	"github.com/zhengxiaoyao0716/zmodule/info"
)

var name string

// Run .
func Run() {
	name = info.Name()
	startManager()
	startServer()
}

func initCmds() { // 初始化远程管理服务命令行指令
}

func startManager() {
	server.Cmds["ping"] = server.Command{
		Mode:  connect.ModeAll,
		Usage: "ping.",
		Handler: func(c connect.Conn) (string, server.Handler) {
			return "pong\n" + server.In, nil
		},
	}

	addr := config.GetString("manager")
	if err := server.Start(name, addr); err != nil {
		log.Fatalln(err)
	}
	cout.Printf("Service start, use `%s` to connect it.\n", cout.Log("%s cli -addr %s", name, cout.Info(addr)))
}

func startServer() {
	engine := gin.New()
	dir := config.GetString("log")
	if dir == "" {
		engine.Use(gin.Logger(), gin.Recovery())
	} else {
		logPath := file.AbsPath(dir, fmt.Sprintf(".%s.http.log", info.Name()))
		logFile, err := safefile.Create(logPath)
		if err != nil {
			log.Fatalln(err)
		}
		errPath := file.AbsPath(dir, fmt.Sprintf(".%s.http-error.log", info.Name()))
		errFile, err := safefile.Create(errPath)
		if err != nil {
			log.Fatalln(err)
		}
		engine.Use(gin.LoggerWithWriter(logFile), gin.RecoveryWithWriter(errFile))
		gin.DisableConsoleColor()
	}
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	engine.GET("/route", Route)
	engine.POST("/route", Route)
	go engine.Run(config.GetString("server"))

	initRoute()
}
