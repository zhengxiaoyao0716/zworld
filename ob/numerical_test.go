package ob

import (
	"fmt"
	"testing"
)

func TestHight(t *testing.T) {
	fmt.Println("hight:", hightFn(0.7), hightFn(0.85), hightFn(1))
}
func TestDepth(t *testing.T) {
	fmt.Println("depth:", depthFn(0), depthFn(0.35), depthFn(0.7))
}

func init() {
	initNumerical()
}
