package server

import (
	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/zworld/server/chain"
)

func chainUpdateHandler(json *easyjson.Object) easyjson.Object {
	if err := chain.Load(string(json.MustStringAt("chain"))); err != nil {
		return easyjson.Object{"ok": false, "reason": err}
	}
	return easyjson.Object{"signature": chain.Signature()}
}

func chainQueryHandler(json *easyjson.Object) easyjson.Object {
	return easyjson.Object{"blocks": chain.Query(string(json.MustStringAt("server")))}
}
