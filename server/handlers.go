package server

import (
	"fmt"

	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zworld/ob"
)

func dashboardHandler(json *easyjson.Object) easyjson.Object {
	baseArgs := [][3]interface{}{}
	cfg := easyjson.Object(*config.Config())
	for name, arg := range zmodule.Args {
		baseArgs = append(baseArgs, [3]interface{}{name, arg.Usage, cfg.MustValueAt(name)})
	}
	checkArgs := [][3]interface{}{}
	// TODO 构建checkArgs，比如关键模型的hash
	m := ob.NewModel()
	checkArgs = append(checkArgs, [3]interface{}{
		"model", "model", fmt.Sprint(m),
	})
	return easyjson.Object{
		"baseArgs":  baseArgs,
		"checkArgs": checkArgs,
	}
}
