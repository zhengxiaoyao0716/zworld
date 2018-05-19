package server

import (
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
)

func dashboardHandler(json *easyjson.Object) easyjson.Object {
	baseArgs := []map[string]interface{}{}
	cfg := easyjson.Object(*config.Config())
	for name, arg := range zmodule.Args {
		baseArgs = append(baseArgs, map[string]interface{}{
			"name":  name,
			"usage": arg.Usage,
			"value": cfg.MustValueAt(name),
		})
	}
	return easyjson.Object{
		"baseArgs": baseArgs,
	}
}
