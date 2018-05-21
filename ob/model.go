package ob

import (
	"crypto/sha256"
	"fmt"
	"math"
)

// Model .
type Model struct {
	gene    Gene
	gathers Gathers
	samples Samples
}

// NewModel .
func NewModel() *Model { return &Model{gene, newGathers(gene, gatherN), Samples(sampleN)} }

// Chunk 地图区块
type Chunk struct {
	i       int     // 区块的样点序号
	x, y, z float64 // 区块的样点坐标
	// 坐标投影公式
	projector func(x, y, z float64) (float64, float64, bool)
	// 地貌计算公式
	terrain func(x, y, z float64) (float64, bool)
}

// Place .
type Place struct {
	x, y, z  float64 // 坐标
	chunk    *Chunk  // 距离最近的区块
	distance float64 // 到最近点的距离
}

// Place create a `Place` object at the point.
func (m *Model) Place(x, y, z float64) *Place {
	ni, distance := m.samples.near(x, y, z)
	nx, ny, nz := m.samples.coord(ni)
	c := &Chunk{
		ni, nx, ny, nz,
		m.samples.projector(ni),
		m.terrain(ni, nx, ny, nz),
	}
	p := &Place{
		x, y, z, c, distance,
	}
	return p
}

// altitude 计算给定样点（及所管辖区块）的平均海拔
func (m *Model) altitude(index int, x, y, z float64) float64 {
	gi, gd := m.gathers.near(x, y, z)
	gl := m.gathers.level(gi)

	// TODO 确定聚合力度与区块平均等高线，当前为临时方案
	gene := append(m.gene, fmt.Sprintf("altitude%d", index)...)
	rand := gene.rand()
	wave := circumProportion(gd) / m.gathers.strength(gi) // 距离与聚合力度之比正比于波动力度
	wave = 1 + math.Pow(wave-rand.Float64(), 3)           // 对波动做随机消减并调整锐度与相位
	if wave < 1 {
		wave = 1
	}
	level := gl / wave                         // 样点（及所管辖区块）的平均海拔等级
	level = level + rand.NormFloat64()*level/8 // 对平均海拔等级做一定比例的随机浮动

	return hightFn(level)
}

// terrain 生成给定样点的地貌函数，地貌函数返回某坐标的海拔、是否属于区块等
func (m *Model) terrain(index int, x, y, z float64) func(x, y, z float64) (float64, bool) {
	altitude := m.altitude(index, x, y, z)
	return func(x, y, z float64) (float64, bool) {
		near, _ := m.samples.near(x, y, z)
		// TODO 确定坐标海拔高度，当前为临时方案
		return altitude, near != index
	}
}

// Hash 生成一致性校验用的签名
func (m *Model) Hash() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprint(m)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
