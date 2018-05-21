package server

import (
	"flag"
	"regexp"
	"sort"

	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zworld/ob"
)

func dashboardHandler(json *easyjson.Object) easyjson.Object {
	var baseArgs, numerical [][3]interface{}
	_baseArgs, _numerical := ArgGroups()
	cfg := easyjson.Object(*config.Config())
	for _, f := range _baseArgs {
		baseArgs = append(baseArgs, [3]interface{}{f.Name, f.Usage, cfg.MustValueAt(f.Name)})
	}
	for _, f := range _numerical {
		numerical = append(numerical, [3]interface{}{f.Name, f.Usage, cfg.MustValueAt(f.Name)})
	}
	// TODO 构建checkArgs，比如关键模型的hash
	checkArgs := [][3]interface{}{}
	m := ob.NewModel()
	checkArgs = append(checkArgs, [3]interface{}{
		"hash", "hashcode for model.", m.Hash(),
	})
	return easyjson.Object{
		"baseArgs":  baseArgs,
		"numerical": numerical,
		"checkArgs": checkArgs,
	}
}

// ArgGroups .
var ArgGroups func() ([]*flag.Flag, []*flag.Flag)

func init() {
	ArgGroups = func() (baseArgs, numerical []*flag.Flag) {
		names := sort.StringSlice{}
		for name := range zmodule.Args {
			names = append(names, name)
		}
		sort.Sort(names)

		r := regexp.MustCompile(`^\[(\w+)\]\s(.*)$`)
		for _, name := range names {
			f := flag.Lookup(name)
			subs := r.FindStringSubmatch(zmodule.Args[name].Usage)
			var group string
			if len(subs) > 2 {
				group = subs[1]
			}
			switch group {
			case "numerical":
				f.Usage = subs[2]
				numerical = append(numerical, f)
			default:
				baseArgs = append(baseArgs, f)
			}
		}
		ArgGroups = func() ([]*flag.Flag, []*flag.Flag) {
			return baseArgs, numerical
		}
		return
	}
}
