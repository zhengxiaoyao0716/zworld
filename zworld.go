package main

import (
	"github.com/kardianos/service"

	"github.com/zhengxiaoyao0716/util/console"
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

func initArgs() {
	zmodule.Args["gene"] = zmodule.Argument{
		Type:    "string",
		Default: "❤",
		Usage:   "A random key for the world.",
	}
	zmodule.Args["addr"] = zmodule.Argument{
		Type:    "string",
		Default: "127.0.0.1:2017",
		Usage:   "Service address witch to listening.",
	}
}

func initCmds() {
	zmodule.Cmds["ping"] = zmodule.Command{
		Usage:   "ping",
		Handler: func(parsed string, args []string) { console.Yes("pong") },
	}
	zmodule.Cmds["cli"] = zmodule.Command{
		Usage:   "Enter the command line",
		Handler: func(parsed string, args []string) { client.Start(args) },
	}
}

func init() {
	initInfo()
	initArgs()
	initCmds()
}
