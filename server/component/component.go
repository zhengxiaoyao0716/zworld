// Package component used to manage the components (instances such as ob.Model, sub.World, etc.)
package component

import (
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

var wave = struct {
	scale     float64 // 缩放
	amplitude float64 // 振幅
}{5.0, 1.0 / 5}

// Terrain .
func Terrain() map[string]interface{} {
	width, depth := 256, 256
	size := width * depth
	data := make([]float64, size)
	quality := 2.0
	for j := 0; j < 4; j++ {
		for i := 0; i < size; i++ {
			x, y := i%width, int(i/width)
			data[i] += wave.amplitude * noise.Noise(float64(x)/quality, float64(y)/quality, 0) * quality
		}
		quality *= wave.scale
	}
	return map[string]interface{}{"width": width, "depth": depth, "data": data}
}

// InitArg .
type InitArg struct {
	Pubkey []byte
}

// Init .
func Init(arg InitArg) {
	model = ob.NewModel()
	place := model.Birth(arg.Pubkey)
	go sub.Service(place)
}
