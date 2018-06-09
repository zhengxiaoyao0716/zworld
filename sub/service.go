package sub

import (
	"github.com/zhengxiaoyao0716/zworld/ob"
	"github.com/zhengxiaoyao0716/zworld/sub/cache"
)

// My .
var My struct {
	*ob.Place
}

// We .
var We struct{}

// TodoPushData 时间不够了，应急方法
func TodoPushData(i, x, z int, data [2]int) {
	cache.MustPush([3]float64{float64(i), float64(x), float64(z)}, []byte{byte(data[0]), byte(data[1])})
}

// TodoGetData .
func TodoGetData(i, x, z int) ([2]int, bool) {
	data := cache.Get([3]float64{float64(i), float64(x), float64(z)})
	if len(data) == 2 {
		return [2]int{int(data[0]), int(data[1])}, true
	}
	return [2]int{}, false
}

// Init .
func Init(myPlace *ob.Place) {
	My.Place = myPlace
}
