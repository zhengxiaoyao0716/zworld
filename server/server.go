package server

import (
	"log"

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
	cout.Printf("Service start, use `%s` to connect it.\n", cout.Log("%s cli -addr %s", name, cout.Info(addr)))
}
