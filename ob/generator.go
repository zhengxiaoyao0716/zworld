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

// Genesis .
func Genesis() Gene { return gene }

func init() {
	zmodule.Args["gene"] = zmodule.Argument{
		Default: string(gene),
		Usage:   "A random key for the world.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			gene = Gene(config.GetString("gene"))
			return nil
		})
		return nil
	})
}
