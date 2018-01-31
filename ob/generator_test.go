package ob

import (
	"fmt"
	"testing"
)

func TestGene(*testing.T) {
	r := gene.rand()
	fmt.Println(r.Float64())
	fmt.Println(r.Float64())
	fmt.Println(r.Float64())
}
