package ob

import (
	"hash/fnv"
	"math/rand"

	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/event"
)

// Gene .
type Gene []byte

func (g Gene) rand() *rand.Rand {
	h := fnv.New64a()
	h.Write(g)
	seed := h.Sum64()
	return rand.New(rand.NewSource(int64(seed)))
}

var gene = Gene("❤")

func init() {
	zmodule.Args["gene"] = zmodule.Argument{
		Type:    "string",
		Default: gene,
		Usage:   "Random seed for the generator.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			gene = Gene(config.GetString("gene"))
			return nil
		})
		return nil
	})
}
