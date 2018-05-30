package sub

import (
	"github.com/zhengxiaoyao0716/zworld/ob"
)

// My .
var My struct {
	*ob.Place
}

// We .
var We struct{}

// Service .
func Service(myPlace *ob.Place) {
	My.Place = myPlace
}
