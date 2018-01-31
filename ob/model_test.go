package ob

import (
	"fmt"
	"testing"
)

func TestNewModel(t *testing.T) {
	m := NewModel()
	fmt.Println(m.gathers)
	fmt.Println(m.samples)
}
