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
	x, y, z float64
	// sample  [4]float64 // [x, y, z, gn], `gn` is the
}

// Place create a `Place` object at the point.
func (m *Model) Place(x, y, z float64) Place {
	p := Place{x, y, z}
	// si, _ := m.samples.near(x, y, z)
	// sx, sy, sz := m.samples.point(sn)
	// gi, _ := m.gathers.near(sx, sy, sz)
	// g := m.gathers[gn][3]
	return p
}
