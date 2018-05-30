package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCURD(*testing.T) {
	fmt.Print("Push: ")
	fmt.Println(Push([3]float64{0, 0, 0}, []byte("000")))
	fmt.Println(Get([3]float64{0, 0, 0}))
	fmt.Print("Push: ")
	fmt.Println(Push([3]float64{0, 0, 0}, []byte("111")))
	fmt.Println(Get([3]float64{0, 0, 0}))
	fmt.Print("MustPush: ")
	fmt.Println(MustPush([3]float64{0, 0, 0}, []byte("222")))
	fmt.Println(Get([3]float64{0, 0, 0}))
	fmt.Print("Remove: ")
	fmt.Println(Remove([3]float64{0, 0, 0}))
	fmt.Println(Get([3]float64{0, 0, 0}))
}

func TestThreshold(*testing.T) {
	cachethres = newCachethres(4)
	for i := 0; i < 1+cachethres.max; i++ {
		Push([3]float64{float64(i)}, nil)
		time.Sleep(1)
	}
	fmt.Printf("%v\n", cache)
	for _, item := range retrieve.Values() {
		tile := item.(*TileData)
		fmt.Printf("%p: %v\n", tile, tile)
	}
}
