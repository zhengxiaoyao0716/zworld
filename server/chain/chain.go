package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/zhengxiaoyao0716/zworld/ob"
)

// Data .
type Data struct {
	Salt []byte
}

// Block .
type Block struct {
	Index int64
	Time  int64
	Data  Data
	Prev  []byte // previous hashcode
	Hash  []byte // hashcode
}

func blockHash(index int64, time int64, data Data, prev []byte) []byte {
	text := fmt.Sprint(index, time, data, prev)
	h := sha256.New()
	h.Write([]byte(text))
	return h.Sum(nil)
}

var chain []*Block

// Init .
func Init() {
	d := Data{Salt: []byte(ob.Genesis())}
	chain = append(chain, &Block{0, 0, d, nil, blockHash(0, 0, d, nil)})
}

// NewBlock .
func NewBlock(data Data) *Block {
	prev := chain[len(chain)-1]
	index := 1 + prev.Index
	time := time.Now().Unix()
	hash := blockHash(index, time, data, prev.Hash)
	return &Block{index, time, data, prev.Hash, hash}
}

// Validate chain blocks.
func Validate(chain ...*Block) error {
	if len(chain) < 1 {
		return errors.New("no blocks to validate")
	}
	index := chain[0].Index
	hash := chain[0].Prev
	for _, block := range chain {
		if block.Index != index {
			return fmt.Errorf("invalid index of block, index: %d, expected: %d", block.Index, index)
		}
		index = 1 + block.Index

		if !bytes.Equal(block.Prev, hash) {
			return fmt.Errorf("invlid previous block, previous hash: %x, expected: %x", block.Prev, hash)
		}
		hash = blockHash(block.Index, block.Time, block.Data, block.Prev)

		if !bytes.Equal(block.Hash, hash) {
			return fmt.Errorf("invalid hash: %x, expected: %x", block.Hash, hash)
		}
	}
	return nil
}

// Dump the chain to base64 string.
func Dump() (string, error) {
	buff := &bytes.Buffer{}
	enc := gob.NewEncoder(buff)
	if err := enc.Encode(chain); err != nil {
		return "", err
	}
	text := base64.StdEncoding.EncodeToString(buff.Bytes())
	return text, nil
}

// Load the chain from base64 string
func Load(text string) error {
	buff, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewReader(buff))
	var c []*Block
	if err := dec.Decode(&c); err != nil {
		return err
	}
	if len(c) < len(chain) {
		return fmt.Errorf("invalid new chain, length: %d shorter then now: %d", len(c), len(chain))
	}
	if err := Validate(c...); err != nil {
		return err
	}
	chain = c
	return nil
}

// Signature of the chain.
func Signature() string {
	var hash []byte
	for _, block := range chain {
		hash = append(hash, block.Hash...)
	}
	h := sha256.New()
	h.Write(hash)
	return fmt.Sprintf("%x", h.Sum(nil))
}
