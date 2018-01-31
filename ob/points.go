package ob

import (
	"math"
	"math/rand"
	"sort"
)

func randPoint(r *rand.Rand) (float64, float64, float64) {
	u, v := r.Float64()*2*math.Pi, r.Float64()*math.Pi
	x := math.Cos(u) * math.Sin(v)
	y := math.Sin(u) * math.Sin(v)
	z := math.Cos(v)
	return x, y, z
}

// Points .
type Points interface {
	n() int
	index(z float64) int
	point(n int) (float64, float64, float64)
}

func pointsEach(s Points) chan [3]float64 {
	iter := make(chan [3]float64)
	go func() {
		for n := 0; n < s.n(); n++ {
			x, y, z := s.point(n)
			iter <- [3]float64{x, y, z}
		}
		close(iter)
	}()
	return iter
}

func pointsNear(s Points, x, y, z float64) (int, float64) {
	n := s.index(z)
	ri, rd := n, 4.0
	for _, incre := range [2]int{-1, 1} {
		ni := n
		for {
			if ni < 0 || ni >= s.n() {
				break
			}
			xi, yi, zi := s.point(ni)

			dz := math.Pow(zi-z, 2)
			if dz > rd {
				break
			}
			ni += incre

			dist := math.Pow(xi-x, 2) + math.Pow(yi-y, 2) + dz
			if dist > rd {
				continue
			}
			ri, rd = ni-incre, dist
		}
	}
	return ri, rd
}

// Gathers .
type Gathers [][4]float64 // [x, y, z, g], `g` is the gather strength.

func newGathers(g Gene, gatherN int) Gathers {
	g = append(g, "gather"...)
	n := float64(gatherN)/4*g.rand().NormFloat64() + float64(gatherN)
	n = math.Min(math.Max(n, 1), float64(gatherN*2-1)) // [1, 2n-1]
	var gathers Gathers
	for i := 0; i < int(math.Floor(n+0.5)); i++ { // math.Round
		r := append(g, byte(i)).rand()
		x, y, z := randPoint(r)
		g := gatherFn(r.Float64())
		gathers = append(gathers, [4]float64{x, y, z, g})
	}
	sort.Slice(gathers, func(i, j int) bool { return gathers[i][2] < gathers[j][2] })
	return gathers
}

// implement PointSet
func (s Gathers) n() int { return len(s) }
func (s Gathers) index(z float64) int {
	n := s.n()

	i := sort.Search(n, func(i int) bool { return s[i][2] >= z })
	if i == n { // not found
		return n - 1
	}

	li := i - 1
	if li < 0 { // no left point
		return 0
	}

	d, ld := s[i][2]-z, z-s[li][2]
	if ld < d { // near to left
		return li
	}

	return i
}
func (s Gathers) point(n int) (float64, float64, float64) {
	p := s[n]
	return p[0], p[1], p[2]
}
func (s Gathers) near(x, y, z float64) (int, float64) { return pointsNear(s, x, y, z) }
func (s Gathers) each() chan [3]float64               { return pointsEach(s) }

// Samples .
type Samples int

var incre = 2 * math.Pi * (math.Sqrt(5) - 1) / 2

// implement PointSet
func (s Samples) n() int              { return int(s) }
func (s Samples) index(z float64) int { return int(((z+1)*float64(s) - 1) / 2) }
func (s Samples) point(n int) (float64, float64, float64) {
	z := float64(2*n+1)/float64(s) - 1
	rad := math.Sqrt(1 - math.Pow(z, 2))
	ang := float64(n) * incre
	x := rad * math.Cos(ang)
	y := rad * math.Sin(ang)
	return x, y, z
}
func (s Samples) near(x, y, z float64) (int, float64) { return pointsNear(s, x, y, z) }
func (s Samples) each() chan [3]float64               { return pointsEach(s) }
