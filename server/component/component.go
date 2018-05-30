// Package component used to manage the components (instances such as ob.Model, sub.World, etc.)
package component

import (
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
