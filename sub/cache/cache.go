package cache

import (
	"log"
	"time"
	"unsafe"

	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
	"github.com/zhengxiaoyao0716/zmodule"
	"github.com/zhengxiaoyao0716/zmodule/config"
	"github.com/zhengxiaoyao0716/zmodule/event"
)

// TileData .
type TileData struct {
	Nano  int64
	Coord [3]float64 // tite coord, [x, y, z].
	Data  []byte
}

var cache = map[[3]float64]*TileData{}
var retrieve = treeset.NewWith(func(a, b interface{}) int {
	aAsserted, bAsserted := a.(*TileData), b.(*TileData)
	if aAsserted == bAsserted { // pointer to same address
		return 0
	}
	// 除指针相等以外绝对不相等！

	if cmp := utils.Int64Comparator(aAsserted.Nano, bAsserted.Nano); cmp != 0 {
		return -cmp // 按时间反序
	}
	if cmp := utils.Float64Comparator(aAsserted.Coord[2], bAsserted.Coord[2]); cmp != 0 {
		return cmp
	}
	if uintptr(unsafe.Pointer(aAsserted)) < uintptr(unsafe.Pointer(bAsserted)) {
		return 1
	}
	return -1
})
var run = make(chan func()) // // all the read and write of cache and retrieve should delegate to this channel.

func (tile *TileData) refresh() {
	retrieve.Remove(tile)
	*tile = TileData{time.Now().UnixNano(), tile.Coord, tile.Data}
	retrieve.Add(tile)
}

// Get .
func Get(coord [3]float64) []byte {
	ret := make(chan []byte)
	run <- func() {
		tile, ok := cache[coord]
		if !ok {
			ret <- nil
			return
		}
		tile.refresh()
		ret <- tile.Data
	}
	return <-ret
}

// Push return [true, nil] if ok else [false, exist].
func Push(coord [3]float64, data []byte) (bool, []byte) {
	tile := &TileData{time.Now().UnixNano(), coord, data}
	okChan := make(chan bool)
	exist := make(chan []byte)
	run <- func() {
		if tile, ok := cache[coord]; ok {
			tile.refresh()
			okChan <- false
			exist <- tile.Data
			return
		}
		cache[coord] = tile
		retrieve.Add(tile)
		okChan <- true
		exist <- nil
	}
	return <-okChan, <-exist
}

// Remove .
func Remove(coord [3]float64) []byte {
	ret := make(chan []byte)
	run <- func() {
		tile, ok := cache[coord]
		if !ok {
			ret <- nil
			return
		}
		delete(cache, coord)
		retrieve.Remove(tile)
		ret <- tile.Data
	}
	return <-ret
}

// MustPush .
func MustPush(coord [3]float64, data []byte) []byte {
	tile := &TileData{time.Now().UnixNano(), coord, data}
	ret := make(chan []byte)
	run <- func() {
		var old []byte
		defer func() { ret <- old }()
		if tile, ok := cache[coord]; ok {
			old = tile.Data
			delete(cache, coord)
			retrieve.Remove(tile)
		} else {
			old = nil
		}
		cache[coord] = tile
	}
	return <-ret
}

type cachethresType struct {
	hold int
	max  int
}

func newCachethres(hold int) *cachethresType {
	return &cachethresType{hold, hold + hold>>1}
}

var cachethres = newCachethres(1024)

func service() {
	for {
		(<-run)()
		// check threshold
		if retrieve.Size() < cachethres.max {
			continue
		}
		items := retrieve.Values()[cachethres.hold:]
		log.Printf("[cache] clean started, arrived max size: %d.\n", cachethres.max)
		retrieve.Remove(items...)
		for _, item := range items {
			tile := item.(*TileData)
			delete(cache, tile.Coord)
		}
		nano := time.Unix(0, items[0].(*TileData).Nano)
		log.Printf("[cache] clean finished, %d items earlier than %v were removed.\n", len(items), nano)
	}
}

func init() {
	go service()
	zmodule.Args["cache-threshold"] = zmodule.Argument{
		Default: cachethres.hold,
		Usage:   "Threshold of the cache pool.",
	}
	event.OnInit(func(event.Event) error {
		event.On(event.KeyStart, func(event.Event) error {
			cachethres = newCachethres(config.GetInt("cache-threshold"))
			return nil
		})
		return nil
	})
}
