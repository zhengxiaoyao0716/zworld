package ob

import (
	"fmt"
	"testing"
)

func TestModel(t *testing.T) {
	m := NewModel()
	fmt.Println(m.gathers)
	fmt.Println(m.samples)

	rand := append(m.gene, "TesModel"...).rand()
	for i := 0; i < 100; i++ {
		place := m.Birth([]byte{byte(rand.Int())})
		chunk := place.chunk
		fmt.Println(m.altitude(chunk.i, chunk.x, chunk.y, chunk.z))
	}
}
