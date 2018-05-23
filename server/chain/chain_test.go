package chain

import (
	"fmt"
	"log"
	"testing"
)

func TestChain(*testing.T) {
	b := NewBlock(Data{})
	fmt.Printf("%x\n", b.Hash)
}

func TestSerialize(t *testing.T) {
	sign := Signature()
	text, _ := Dump()
	Load(text)
	if sign != Signature() {
		log.Fatalln("invalid signature, Dump & Load not reciprocal.")
	}
	fmt.Println(sign)
}

func init() {
	Init()
}
