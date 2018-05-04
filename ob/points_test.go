package ob

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestGathers(*testing.T) {
	gathers := newGathers(gene, 10)
	for coord := range gathers.each() {
		fmt.Print(coord[2], " ")
	}
	fmt.Println()
	for _, z := range []float64{-1, -0.5, 0, +0.5, +1} {
		index := gathers.index(z) // z轴上距离该处最近的样点
		nearx, neary, nearz := gathers.coord(index)
		fmt.Printf("z: %f, index: %d, coord: (%f, %f, %f)\n", z, index, nearx, neary, nearz)
	}
	fmt.Println()

	g := append(gene, "TestGathers"...)
	analy := [20]int{}
	for i := 0; i < 1000; i++ {
		gathers := newGathers(append(g, byte(i)), 10)
		analy[len(gathers)]++
	}
	for len, num := range analy {
		fmt.Printf("number of len=%d: %d\n", len, num)
	}
	fmt.Println()
}

func TestSamples(*testing.T) {
	samples := Samples(100)
	for i := range samples.each() {
		fmt.Print(i, " ")
	}
	fmt.Println()
	g := append(gene, "TestSample"...)
	for i := 0; i < 10; i++ {
		x, y, z := randPoint(append(g, byte(i)).rand())
		index, dist := samples.near(x, y, z) // 三维坐标上距离该点最近的样点
		nearx, neary, nearz := samples.coord(index)
		fmt.Printf("[%f\t%f\t%f\t] | index: %02d, distance: %f, coord: (%f, %f, %f)\n", x, y, z, index, dist, nearx, neary, nearz)
	}
	fmt.Println()
}

func TestProjection(*testing.T) {
	g := append(gene, "TestProjection"...)
	index := g.rand().Intn(100)
	samples := Samples(100)
	x, y, z := samples.coord(index)
	xs, ys, zs := []float64{x}, []float64{y}, []float64{z}
	for coord := range samples.area(index) {
		xs = append(xs, coord[0])
		ys = append(ys, coord[1])
		zs = append(zs, coord[2])
	}
	fmt.Println("raw coord:", xs, ys, zs)
	us, vs := projection(xs, ys, zs)
	fmt.Println("projected:", us, vs)
}

// `power, sqrt` or `sin, cos` , witch better?
func BenchmarkRandPoint(b *testing.B) {
	randPointXYZ := func(r *rand.Rand) (float64, float64, float64) {
		rp := 1.0
		x := 2 * (r.Float64() - 0.5)
		rp -= math.Pow(x, 2)
		y := 2 * math.Sqrt(rp) * (r.Float64() - 0.5)
		rp -= math.Pow(y, 2)
		z := math.Sqrt(rp) * float64(r.Int()%2*2-1)
		return x, y, z
	}
	randPointUV := func(r *rand.Rand) (float64, float64, float64) {
		u, v := r.Float64()*2*math.Pi, r.Float64()*math.Pi
		x := math.Cos(u) * math.Sin(v)
		y := math.Sin(u) * math.Sin(v)
		z := math.Cos(v)
		return x, y, z
	}
	r := append(gene, "TestRandPoint"...).rand()

	b.Run("XYZ", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			randPointXYZ(r)
		}
	})
	b.Run("UV", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			randPointUV(r)
		}
	})
}
