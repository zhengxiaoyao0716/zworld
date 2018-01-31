package ob

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestGathers(*testing.T) {
	gathers := newGathers(gene, 10)
	// for _, g := range gathers {
	// 	fmt.Print(g[2])
	// }
	for g := range gathers.all() {
		fmt.Print(g[2], " ")
	}
	fmt.Printf("\n")
	fmt.Println(gathers.index(-1), gathers.index(-0.9), gathers.index(-0.8), gathers.index(+1))

	g := append(gene, "TestGathers"...)
	for i := 0; i < 10; i++ {
		fmt.Println(gathers.near(randPoint(append(g, byte(i)).rand())))
	}

	analy := [20]int{}
	for i := 0; i < 1000; i++ {
		gathers := newGathers(append(g, byte(i)), 10)
		analy[len(gathers)]++
	}
	fmt.Println(analy)
}

func TestSamples(*testing.T) {
	samples := Samples(1000)
	for i := range samples.all() {
		fmt.Println(i)
	}

	g := append(gene, "TestSample"...)
	for i := 0; i < 10; i++ {
		fmt.Println(samples.near(randPoint(append(g, byte(i)).rand())))
	}
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
