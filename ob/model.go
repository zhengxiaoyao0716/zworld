package ob

// Model .
type Model struct {
	gene    Gene
	gathers Gathers
	samples Samples
}

// NewModel .
func NewModel() *Model { return &Model{gene, newGathers(gene, gatherN), Samples(sampleN)} }

// Place .
type Place struct {
	x, y, z  float64 // 坐标
	nearest  int     // 距离最近的样点
	distance float64 // 到最近点的距离
	// 坐标投影公式
	projector func(x, y, z float64) (float64, float64, bool)
	// 地貌计算公式
	terrain func(x, y, z float64) (float64, bool)
}

// Place create a `Place` object at the point.
func (m *Model) Place(x, y, z float64) Place {
	nearest, distance := m.samples.near(x, y, z)
	p := Place{
		x, y, z,
		nearest, distance,
		m.samples.projector(nearest),
		m.samples.terrain(nearest, m.gathers),
	}
	// si, _ := m.samples.near(x, y, z)
	// sx, sy, sz := m.samples.point(sn)
	// gi, _ := m.gathers.near(sx, sy, sz)
	// g := m.gathers[gn][3]
	return p
}
