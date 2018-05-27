package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/event"
)

// Data .
type Data struct {
	Server string
	Pubkey []byte
}

// Block .
type Block struct {
	Index int64
	Time  int64
	Data  Data
	Prev  []byte // previous hashcode
	Hash  []byte // hashcode
	Nonce int64
}

var chain struct {
	blocks  []*Block            // 区块链本体
	servers map[string][]*Block // 链上所有数据按server分组整理
	run     chan func()         // 保证区块链重要操作同步执行
	append  func(...*Block) error
	set     func([]*Block) error
}

// Init .
func Init() {
	chain.run = make(chan func())
	go func() {
		for {
			action := <-chain.run
			action()
		}
	}()
	load := func(blocks []*Block) {
		for _, block := range blocks {
			blocks, ok := chain.servers[block.Data.Server]
			if !ok {
				blocks = []*Block{}
			}
			blocks = append(blocks, block)
			chain.servers[block.Data.Server] = blocks

		}
	}
	chain.append = func(blocks ...*Block) error {
		if len(blocks) == 0 {
			return errors.New("no blocks to append")
		}
		err := make(chan error)
		chain.run <- func() {
			if e := validate(chain.blocks[len(chain.blocks)-1], blocks[0]); e != nil {
				err <- fmt.Errorf("append chain failed, %s", e.Error())
			}
			chain.blocks = append(chain.blocks, blocks...)
			load(blocks)
			err <- nil
		}
		return <-err
	}
	chain.set = func(blocks []*Block) error {
		err := make(chan error)
		chain.run <- func() {
			newLen, oldLen := len(blocks), len(chain.blocks)
			if newLen == oldLen {
				eq := true
				blocks := blocks[:oldLen]
				for i, block := range chain.blocks {
					if !bytes.Equal(blocks[i].Hash, block.Hash) {
						eq = false
						break
					}
				}
				if eq {
					err <- nil
					return
				}
			}
			if newLen <= oldLen {
				err <- fmt.Errorf("invalid new chain, length: %d, should longer then: %d", newLen, oldLen)
			}
			chain.blocks = blocks
			chain.servers = map[string][]*Block{}
			load(blocks)
			err <- nil
		}
		return <-err
	}

	data := Data{}
	bs, err := dump(data)
	if err != nil {
		log.Fatalln(err)
	}
	hash, nonce := target.mining(0, 0, bs, nil)
	if err := chain.set([]*Block{&Block{0, 0, data, nil, hash, nonce}}); err != nil {
		log.Println(err)
	}
}

// NewBlock .
func NewBlock(data Data) (*Block, error) {
	prev := chain.blocks[len(chain.blocks)-1]
	index := 1 + prev.Index
	time := time.Now().Unix()
	bs, err := dump(data)
	if err != nil {
		return nil, err
	}
	hash, nonce := target.mining(index, time, bs, prev.Hash)
	return &Block{index, time, data, prev.Hash, hash, nonce}, nil
}

// Insert block.
func Insert(block *Block) error {
	if err := block.validate(); err != nil {
		return err
	}
	return chain.append(block)
}

// Query blocks.
func Query(server string) []*Block {
	blocks, ok := chain.servers[server]
	if !ok {
		return nil
	}
	return blocks
}

// Dump the chain to base64 string.
func Dump() (string, error) {
	bs, err := dump(chain.blocks)
	if err != nil {
		return "", err
	}
	text := base64.StdEncoding.EncodeToString(bs)
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
	if err := validate(c...); err != nil {
		return err
	}
	return chain.set(c)
}

// Signature of the chain.
func Signature() string {
	var hash []byte
	for _, block := range chain.blocks {
		hash = append(hash, block.Hash...)
	}
	h := sha256.New()
	h.Write(hash)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func blockHash(index int64, time int64, data []byte, prev []byte, nonce int64) []byte {
	text := fmt.Sprint(index, time, data, prev, nonce)
	h := sha256.New()
	h.Write([]byte(text))
	return h.Sum(nil)
}

func (b *Block) validate() error {
	bs, err := dump(b.Data)
	if err != nil {
		return err
	}
	hash := blockHash(b.Index, b.Time, bs, b.Prev, b.Nonce)
	if !bytes.Equal(b.Hash, hash) {
		return fmt.Errorf("invalid hash of block, index: %d, hash: %x, expected: %x", b.Index, b.Hash, hash)
	}
	if !target.validate(hash) {
		return fmt.Errorf(
			"invalid proof of work, index: %d, proof: %e, target: %e",
			b.Index,
			new(big.Float).SetInt(new(big.Int).SetBytes(b.Hash)),
			new(big.Float).SetInt(target.Int),
		)
	}
	return nil
}

func validate(chain ...*Block) error {
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
			return fmt.Errorf(
				"invlid prev of block, index: %d, previous hash: %x, expected: %x",
				block.Index, block.Prev, hash)
		}
		if err := block.validate(); err != nil {
			return err
		}
		hash = block.Hash
	}
	return nil
}

func dump(data interface{}) ([]byte, error) {
	buff := &bytes.Buffer{}
	enc := gob.NewEncoder(buff)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

type miningTarget struct {
	difficulty int
	*big.Int
}

func newMiningTarget(d int) *miningTarget {
	t := big.NewInt(1)
	t.Lsh(t, uint(256-d))
	return &miningTarget{d, t}
}
func (t *miningTarget) validate(hash []byte) bool {
	return t.Cmp(new(big.Int).SetBytes(hash)) == 1
}
func (t *miningTarget) mining(index int64, time int64, data []byte, prev []byte) ([]byte, int64) {
	for nonce := int64(0); nonce < math.MaxInt64; nonce++ {
		hash := blockHash(index, time, data, prev, nonce)
		if t.validate(hash) {
			return hash, nonce
		}
	}
	return nil, -1
}

var target = newMiningTarget(8)

func init() {
	zmodule.Args["difficulty"] = zmodule.Argument{
		Default: target.difficulty,
		Usage:   "Mining difficulty of the chain.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			target = newMiningTarget(config.GetInt("difficulty"))
			return nil
		})
		return nil
	})
}
