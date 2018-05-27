package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/zhengxiaoyao0716/util/cout"
	"github.com/zhengxiaoyao0716/util/easyjson"
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
	cout.Printf("Manager service start, use `%s` to connect it.\n", cout.Log("%s cli -addr %s", name, cout.Info(addr)))
}

var engine = gin.New()

func startServer() {
	initLogger()

	// html
	engine.LoadHTMLGlob(file.AbsPath("./browser/*.html"))
	engine.Static("/static", file.AbsPath("./browser/static"))
	regPage("index.html")
	regPage("dashboard.html")
	engine.GET("/", func(c *gin.Context) { c.Redirect(http.StatusTemporaryRedirect, "/index.html") })

	// api
	engine.GET("/ws", wsHandler)
	regHandle("/route", routeHandler)
	regHandle("/api/dashboard", dashboardHandler)
	regHandle("/chain/update", chainUpdateHandler)
	regHandle("/chain/query", chainQueryHandler)

	go engine.Run(config.GetString("server"))
	go p2pRun()
}

func initLogger() {
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
}

func regPage(page string) {
	engine.GET("/"+page, func(c *gin.Context) { c.HTML(http.StatusOK, page, nil) })
}
func regHandle(path string, rawHandler func(*easyjson.Object) easyjson.Object) {
	handler := func(json easyjson.Object) (resp easyjson.Object) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			code := 500
			reason := fmt.Sprint(err)
			switch err := err.(type) {
			case *easyjson.ValueNotFoundError:
				if err.IsRef(&json) {
					code = 400
					reason = "missing argument, " + err.Error()
				}
			case *easyjson.ValueTypeNotMatchError:
				if err.IsRef(&json) {
					code = 400
					reason = "invalid argument, " + err.Error()
				}
			default:
				log.Println(err)
				stacks := bytes.SplitN(debug.Stack(), []byte("\n"), 8)
				log.Output(2, string(stacks[7]))
			}
			resp = easyjson.Object{"ok": false, "code": code, "reason": reason}
		}()
		resp = rawHandler(&json)
		_, ok := resp["ok"]
		if !ok {
			resp["ok"] = true
		}
		return
	}
	engine.GET(path, func(c *gin.Context) {
		json := easyjson.Object{}
		for key, values := range c.Request.URL.Query() {
			if len(values) == 1 {
				json[key] = values[0]
			} else {
				json[key] = values
			}
		}
		resp := handler(json)
		c.JSON(int(resp.MustNumberAt("code", 200)), resp)
	})
	engine.POST(path, func(c *gin.Context) {
		var json, resp easyjson.Object
		if err := c.ShouldBind(&json); err != nil {
			resp = handler(nil)
		}
		resp = handler(json)
		c.JSON(int(resp.MustNumberAt("code", 200)), resp)
	})
	wsHandlers[path] = func(json map[string]interface{}, conn *Conn) {
		id, ok := json["_messageId"]
		delete(json, "_messageId")
		resp := handler(json)
		resp["action"] = path
		if ok {
			resp["_messageId"] = id
		}
		conn.send <- resp
	}
}
