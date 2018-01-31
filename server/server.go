package server

import (
	"log"

	"github.com/zhengxiaoyao0716/util/console"
	"github.com/zhengxiaoyao0716/util/cout"

	"github.com/zhengxiaoyao0716/zcli/connect"
	"github.com/zhengxiaoyao0716/zcli/server"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/info"
)

// Run .
func Run() {
	name := info.Name()
	// gene := config.GetString("gene")
	addr := config.GetString("addr")

	server.Cmds["ping"] = server.Command{
		Mode:  connect.ModeAll,
		Usage: "ping.",
		Handler: func(c connect.Conn) (string, server.Handler) {
			return "pong\n" + server.In, nil
		},
	}
	if err := server.Start(name, addr); err != nil {
		log.Fatalln(err)
	}
	console.Log("Service start, use `%s cli -addr %s` to connect it.", name, cout.Info(addr))
}
