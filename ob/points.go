package ob

import (
	"github.com/emirpasic/gods/utils"
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/emirpasic/gods/sets/treeset"
)

// Points 点集
type Points interface {
	n() int
	index(z float64) int                     // 取某个坐标（附近）的索引
	coord(n int) (float64, float64, float64) // 取某个索引的（准确）坐标
}

// pointsNear 查找离某坐标或样点最近的样点
func pointsNear(s Points, payload interface{}, n int) chan [2]interface{} {
	var index int
	var ok bool
	var x, y, z float64
	var rs = treeset.NewWith(func(a, b interface{}) int {
		aAsserted := a.([2]interface{})
		bAsserted := b.([2]interface{})
		cmp := utils.Float64Comparator(aAsserted[1], bAsserted[1])
		if cmp !=0 {
			return cmp
		}
		return utils.IntComparator(aAsserted[0], bAsserted[0])
	})
	if index, ok = payload.(int); ok {
		x, y, z = s.coord(index)
		rs.Add([2]interface{}{math.NaN, 4.0})
	} else if xyz, ok := payload.([3]float64); ok {
		x, y, z = xyz[0], xyz[1], xyz[2]
		index = s.index(z)
		cx, cy, cz := s.coord(index)
		d := math.Pow(cx-x, 2) + math.Pow(cy-y, 2) + math.Pow(cz-z, 2)
		rs.Add([2]interface{}{index, d})
	} else {
		panic("invalid type of payload")
	}

	for _, incre := range [2]int{-1, 1} {
		i := index
		for {
			i += incre
			if i < 0 || i >= s.n() {
				break
			}
			xi, yi, zi := s.coord(i)

			dz := math.Pow(zi-z, 2)
			var rd float64
			if rs.Size() > n {
				rd = rs.Values()[n-1].([2]interface{})[1].(float64)
			} else {
				rd = rs.Values()[rs.Size()-1].([2]interface{})[1].(float64)
			}
			if dz > rd {
				break
			}

			dist := math.Pow(xi-x, 2) + math.Pow(yi-y, 2) + dz
			if dist > rd {
				continue
			}
			rs.Add([2]interface{}{i, dist})
		}
	}

	iter := make(chan [2]interface{})
	go func() {
		it := rs.Iterator()
		for i := 0; i < n && it.Next(); i++ {
			iter <- it.Value().([2]interface{})
		}
		close(iter)
	}()
	return iter
}

// pointsArea 查找离某样点最近的区域
func pointsArea(s Points, index int) chan [3]float64 {
	iter := make(chan [3]float64)
	go func() {
		x, y, z := s.coord(index)
		for near := range pointsNear(s, index, 8) {
			pi, _ := near[0].(int), near[1].(float64)
			xn, yn, zn := s.coord(pi)
			ca := [3]float64{(x + xn) / 2, (y + yn) / 2, (z + zn) / 2}
			_an := <-pointsNear(s, ca, 1)
			ai, _ := _an[0].(int), _an[1].(float64)
			if ai == pi || ai == index {
				iter <- ca
			}
		}
		close(iter)
	}()
	return iter
}

// pointsEach 遍历集合
func pointsEach(s Points) chan [3]float64 {
	iter := make(chan [3]float64)
	go func() {
		for n := 0; n < s.n(); n++ {
			x, y, z := s.coord(n)
			iter <- [3]float64{x, y, z}
		}
		close(iter)
	}()
	return iter
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
		gathers = append(gathers, [4]float64{x, y, z, (1 - level) * g})
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
func (s Gathers) coord(n int) (float64, float64, float64) {
	p := s[n]
	return p[0], p[1], p[2]
}

// inherit PointSet

func (s Gathers) near(x, y, z float64) (int, float64) {
	near := <-pointsNear(s, [3]float64{x, y, z}, 1)
	return near[0].(int), near[1].(float64)
}
func (s Gathers) area(index int) chan [3]float64 { return pointsArea(s, index) }
func (s Gathers) each() chan [3]float64          { return pointsEach(s) }

// level 计算某大陆核的海拔等级
func (s Gathers) level(index int) float64 {
	r := Gene(fmt.Sprintf("%v%d", s, index)).rand()
	l := (1+level)/2 + (1-level)/2*r.NormFloat64()/8 // 在平均陆地海拔上下浮动
	return math.Max(level, math.Min(l, 1))
}

// strength 聚合强度
func (s Gathers) strength(index int) float64 { return s[index][3] }

// Samples .
type Samples int

var incre = 2 * math.Pi * (math.Sqrt(5) - 1) / 2

// implement PointSet

func (s Samples) n() int              { return int(s) }
func (s Samples) index(z float64) int { return int(((z+1)*float64(s) - 1) / 2) }
func (s Samples) coord(n int) (float64, float64, float64) {
	z := float64(2*n+1)/float64(s) - 1
	rad := math.Sqrt(1 - math.Pow(z, 2))
	ang := float64(n) * incre
	x := rad * math.Cos(ang)
	y := rad * math.Sin(ang)
	return x, y, z
}

// inherit PointSet

func (s Samples) near(x, y, z float64) (int, float64) {
	near := <-pointsNear(s, [3]float64{x, y, z}, 1)
	return near[0].(int), near[1].(float64)
}
func (s Samples) area(index int) chan [3]float64 { return pointsArea(s, index) }
func (s Samples) each() chan [3]float64          { return pointsEach(s) }

// projector 生成给定样点的投影函数，投影函数接收三维坐标，返回二维坐标以及该坐标是否仍距样点最近
func (s Samples) projector(index int) func(x, y, z float64) (float64, float64, bool) {
	x, y, z := s.coord(index)
	lena := math.Sqrt(math.Pow(z, 2) + math.Pow(x, 2))
	cosa, sina := z/lena, -x/lena
	z = z*cosa - x*sina
	lenb := math.Sqrt(math.Pow(z, 2) + math.Pow(y, 2))
	cosb, sinb := z/lenb, -y/lenb
	return func(x, y, z float64) (float64, float64, bool) {
		w, u := z*cosa-x*sina, z*sina+x*cosa
		_, v := w*cosb-y*sinb, w*sinb+y*cosb
		near, _ := s.near(x, y, z)
		return u, v, near != index
	}
}

// utils

// randPoint 随机坐标点
func randPoint(r *rand.Rand) (float64, float64, float64) {
	u, v := r.Float64()*2*math.Pi, r.Float64()*math.Pi
	x := math.Cos(u) * math.Sin(v)
	y := math.Sin(u) * math.Sin(v)
	z := math.Cos(v)
	return x, y, z
}

// circumProportion 直线距离的圆周占比
func circumProportion(dist float64) float64 {
	if dist == 0 {
		return 0
	}
	if dist == 2 {
		return 1
	}
	sina := dist / 2
	a := math.Asin(sina)
	return a / math.Pi
}
