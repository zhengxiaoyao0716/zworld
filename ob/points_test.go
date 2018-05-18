package ob

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestGathers(*testing.T) {
	gathers := newGathers(gene, 10)
	for index, coord := range gathers {
		level := gathers.level(index)
		fmt.Printf(
			"[%d](%f, %f, %f), g: %f, level: %f, altitude: %f\n",
			index, coord[0], coord[1], coord[2],
			gathers.strength(index), level, hightFn(level),
		)
	}
	for _, z := range []float64{-1, -0.5, 0, +0.5, +1} {
		index := gathers.index(z) // z轴上距离该处最近的样点
		nearx, neary, nearz := gathers.coord(index)
		fmt.Printf("z: %f, nearest: [%d](%f, %f, %f)\n", z, index, nearx, neary, nearz)
	}
	fmt.Println()

	g := append(gene, "TestGathers"...)
	analy := [15]int{}
	total := 0.0
	for i := 0; i < 1000; i++ {
		gathers := newGathers(append(g, byte(i)), gatherN)
		analy[len(gathers)]++
		sum := 0.0
		for i := range gathers {
			sum += gathers.strength(i)
		}
		total += sum
	}
	fmt.Printf("average of total gathers: %f / 1000\n", total)
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

func TestProjector(*testing.T) {
	g := append(gene, "TestProjection"...)
	index := g.rand().Intn(100)
	samples := Samples(100)
	x, y, z := samples.coord(index)
	fmt.Printf("index: %d, coord: (%f, %f, %f)\n", index, x, y, z)

	xs, ys, zs := []float64{x}, []float64{y}, []float64{z} // 原坐标
	proj := samples.projector(index)
	u, v, near := proj(x, y, z)
	us, vs, ns := []float64{u}, []float64{v}, []bool{near} // 投影坐标

	for coord := range samples.area(index) {
		xs = append(xs, coord[0])
		ys = append(ys, coord[1])
		zs = append(zs, coord[2])

		u, v, near := proj(coord[0], coord[1], coord[2])
		us = append(us, u)
		vs = append(vs, v)
		ns = append(ns, near)
	}
	rus, rvs := projection(xs, ys, zs) // 参照坐标

	for i := range xs {
		x, y, z := xs[i], ys[i], zs[i]
		u, v, near := proj(x, y, z)
		ru, rv := rus[i], rvs[i]
		fmt.Println()
		fmt.Printf("raw coord: (%f, %f, %f)\n", x, y, z)
		fmt.Printf("projected: (%f, %f), near?: %t\n", u, v, near)
		fmt.Printf("reference: (%f, %f)\n", ru, rv)
	}
	fmt.Println()
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

func TestCircumProportion(*testing.T) {
	fmt.Println("dist=1       (1/6): ", circumProportion(1))
	fmt.Println("dist=sqrt(2) (1/4): ", circumProportion(math.Sqrt(2)))
}

// projection 三维坐标到二维投影
func projection(xs, ys, zs []float64) ([]float64, []float64) {
	rotate := func(us, vs []float64) ([]float64, []float64) {
		u, v := us[0], vs[0]
		l := math.Sqrt(math.Pow(u, 2) + math.Pow(v, 2))
		cosa, sina := u/l, -v/l
		ru, rv := []float64{}, []float64{}
		for i, u := range us {
			v := vs[i]
			ru = append(ru, u*cosa-v*sina)
			rv = append(rv, u*sina+v*cosa)
		}
		return ru, rv
	}
	zs, xs = rotate(zs, xs)
	zs, ys = rotate(zs, ys)
	return xs, ys
}
