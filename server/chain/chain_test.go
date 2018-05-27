package chain

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestBlock(*testing.T) {
	data := Data{"http://127.0.0.1", []byte("public key.")}
	block, err := NewBlock(data)
	if err != nil {
		log.Fatalln(err)
	}
	if err := block.validate(); err != nil {
		log.Fatalln(err)
	}
	Insert(block)
	block = &Block{}
	*block = *Query(data.Server)[0]
	block.Nonce++
	fmt.Println(block.validate())
	bs, _ := dump(block.Data)
	block.Hash = blockHash(block.Index, block.Time, bs, block.Prev, block.Nonce)
	fmt.Println(block.validate())
}

func TestSerialize(*testing.T) {
	sign := Signature()
	text, _ := Dump()
	Load(text)
	if sign != Signature() {
		log.Fatalln("invalid signature, Dump & Load not reciprocal.")
	}
	fmt.Println(sign)
}

func BenchmarkMining(b *testing.B) {
	mining := func(d int) {
		target = newMiningTarget(d)
		b.Run(fmt.Sprintf("mining-%d", d), func(b *testing.B) {
			t := time.Now().Unix()
			bs, _ := time.Now().MarshalBinary()
			hash := blockHash(0, 0, nil, nil, 0)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hash, _ = target.mining(int64(i), t, bs, hash)
			}
		})
	}
	mining(8)
	mining(16)
}

func init() {
	Init()
}
