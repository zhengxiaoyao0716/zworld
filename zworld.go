package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/kardianos/service"

	"github.com/zhengxiaoyao0716/util/address"
	"github.com/zhengxiaoyao0716/util/console"
	"github.com/zhengxiaoyao0716/util/cout"
	"github.com/zhengxiaoyao0716/zcli/client"
	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zworld/server"
)

func main() {
	zmodule.Main("zworld",
		&service.Config{
			Name:        "ZhengWorldService",
			DisplayName: "Virtual world creator",
			Description: "Daemon service for zworld.",
		}, server.Run)
}

// In this way that override those values,
// you can use `main` as the module name, instead of `github.com/zhengxiaoyao0716/zmodule`.
var (
	Version   string // `git describe --tags`
	Built     string // `date +%FT%T%z`
	GitCommit string // `git rev-parse --short HEAD`
	GoVersion string // `go version`
)

func initInfo() {
	zmodule.Author = "zhengxiaoyao0716"
	zmodule.Homepage = "https://zhengxiaoyao0716.github.io/zworld"
	zmodule.Repository = "https://github.com/zhengxiaoyao0716/zworld"
	zmodule.License = "${Repository}/blob/master/LICENSE"

	zmodule.Version = Version
	zmodule.Built = Built
	zmodule.GitCommit = GitCommit
	zmodule.GoVersion = GoVersion
}

func initArgs() { // 初始化运行参数
	zmodule.Args["gene"] = zmodule.Argument{
		Default: "❤",
		Usage:   "A random key for the world.",
	}
	zmodule.Args["server"] = zmodule.Argument{
		Default: "127.0.0.1:2017",
		Usage:   "Main service address.",
	}
	zmodule.Args["manager"] = zmodule.Argument{
		Default: "127.0.0.1:2018",
		Usage:   "Remote cli manager service address.",
	}
	zmodule.Args["route"] = zmodule.Argument{
		Default: "127.0.0.1:2017",
		Usage:   "Routes to broadcast the service.",
	}
}

func initCmds() { // 初始化本地命令行指令
	zmodule.Cmds["scan"] = zmodule.Command{
		Usage: "Scan available hosts and ports",
		Handler: func(parsed string, args []string) {
			// 扫描可用网段
			netMap, err := address.ScanNets()
			if err != nil {
				log.Fatalln(err)
			}
			for name, nets := range netMap {
				console.Log("[%s]", name)
				for _, net := range nets {
					fmt.Println(net.String())
				}
				fmt.Println()
			}

			// 解析地址参数
			addr := flag.String("addr", "127.0.0.1:2017", "Scan ports from the given port with given host.")
			flag.CommandLine.Parse(args)

			// 扫描可用端口
			host, _port, err := net.SplitHostPort(*addr)
			if err != nil {
				log.Fatalln(err)
			}
			port, err := strconv.Atoi(_port)
			if err != nil {
				log.Fatalln(err)
			}
			console.Log("[Available ports of %s]", cout.Info(host))
			address.FindPorts("["+host+"]", port, func(port int, ok bool) bool {
				if ok {
					fmt.Print(port, " ")
				}
				return false
			})
		},
	}
	zmodule.Cmds["cli"] = zmodule.Command{
		Usage:   "Enter the client of remote cli manager",
		Handler: func(parsed string, args []string) { client.Start(args) },
	}
}

func init() {
	initInfo()
	initArgs()
	initCmds()
}
