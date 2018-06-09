// Package component used to manage the components (instances such as ob.Model, sub.World, etc.)
package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zhengxiaoyao0716/zworld/sub/cache"

	"github.com/zhengxiaoyao0716/util/easyjson"
	"github.com/zhengxiaoyao0716/util/terrain/improved-noise"
	"github.com/zhengxiaoyao0716/zworld/ob"
	"github.com/zhengxiaoyao0716/zworld/sub"
)

var model *ob.Model

// ModelSignature .
func ModelSignature() string { return model.Signature() }

// MyPlaceProjection .
func MyPlaceProjection() ([][3]float64, [][2]float64) {
	return model.PlaceProjection(sub.My.Place)
}

// TodoMyPlaceChunkIndex 时间不够了，临时方法.
func TodoMyPlaceChunkIndex() int {
	return sub.My.Place.TodoChunkIndex()
}

var world = struct {
	unit   [2]int // [width, depth]
	border [4]int // [startX, startZ, endX, endZ]
}{
	[2]int{256, 256},
	[4]int{-512, -512, 512, 512},
}

var wave = struct {
	scale     float64 // 缩放
	amplitude float64 // 振幅
}{5.0, 1.0 / 5}

// Terrain .
func Terrain(json *easyjson.Object) map[string]interface{} {
	width, depth := int(json.MustNumberAt("width", world.unit[0])), int(json.MustNumberAt("depth", world.unit[1]))
	offsetX, offsetZ := int(json.MustNumberAt("x", -width/2)), int(json.MustNumberAt("z", -depth/2))
	if offsetX < world.border[0] || offsetZ < world.border[1] ||
		offsetX >= world.border[2] || offsetZ >= world.border[2] {
		return map[string]interface{}{
			"x": offsetX, "z": offsetZ,
			"out":    true,
			"border": world.border,
		}
	}
	todoSeed := model.TodoPlaceSeed(sub.My.Place)

	// generate
	size := width * depth
	heights := make([]float64, size)
	quality := 2.0
	for j := 0; j < 4; j++ {
		for i := 0; i < size; i++ {
			x, y := offsetX+i%width, offsetZ+int(i/width)
			heights[i] += wave.amplitude * noise.Noise(float64(x)/quality, float64(y)/quality, todoSeed) * quality
		}
		quality *= wave.scale
	}
	// [height, blockId]
	data := make([][2]int, size)
	chunkIndex := sub.My.Place.TodoChunkIndex()
	for i, h := range heights {
		x, z := offsetX+i%width, offsetZ+int(i/width)
		if d, ok := sub.TodoGetData(chunkIndex, x, z); ok {
			data[i] = d
		} else {
			data[i][0] = int(h)
			data[i][1] = 0x01
		}
	}
	return map[string]interface{}{
		"x": offsetX, "z": offsetZ,
		"width": width, "depth": depth,
		"data": data,
	}
}

// Build .
func Build(json *easyjson.Object) map[string]interface{} {
	chunkIndex := sub.My.Place.TodoChunkIndex()
	for _, value := range json.MustArrayAt("block") {
		block := easyjson.MustObjectOf(value)
		x := int(block.MustNumberAt("x"))
		z := int(block.MustNumberAt("z"))
		h := int(block.MustNumberAt("h"))
		id := int(block.MustNumberAt("id"))
		sub.TodoPushData(chunkIndex, x, z, [2]int{h, id})
	}
	return map[string]interface{}{}
}

// ShiftChunk .
func ShiftChunk(json *easyjson.Object) easyjson.Object {
	id := int(json.MustNumberAt("id"))
	place := model.TodoChunkPlace(id)
	if place == nil {
		return easyjson.Object{
			"ok": false, "code": 403,
			"reason": fmt.Sprintf("chunk not found: %d", id),
		}
	}
	sub.My.Place = place
	return easyjson.Object{}
}

// TodoSearchCacheDump .
func TodoSearchCacheDump(i int) string {
	fields := []string{}
	for data := range cache.TodoSearch(i) {
		fields = append(fields, fmt.Sprint(data[0], data[1], data[2], data[3], data[4]))
	}
	return strings.Join(fields, " ")
}

// TodoCacheLoad .
func TodoCacheLoad(text string) (err error) {
	fields := strings.Fields(text)
	atoi := func(a string) int {
		i, err := strconv.Atoi(a)
		if err != nil {
			panic(err)
		}
		return i
	}
	defer func() {
		e := recover()
		if e == nil {
			err = nil
			return
		}
		if e, ok := e.(error); ok {
			err = e
			return
		}
		err = fmt.Errorf("%v", e)
	}()
	for i := 0; i < len(fields); i += 5 {
		sub.TodoPushData(
			atoi(fields[i+0]), atoi(fields[i+1]), atoi(fields[i+2]),
			[2]int{atoi(fields[i+3]), atoi(fields[i+4])},
		)
	}
	return nil
}

// InitArg .
type InitArg struct {
	Pubkey []byte
}

// Init .
func Init(arg InitArg) {
	model = ob.NewModel()
	place := model.Birth(arg.Pubkey)
	sub.Init(place)
}
