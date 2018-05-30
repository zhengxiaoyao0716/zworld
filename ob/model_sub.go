package ob

import (
	"log"
	"time"
)

// Functions used for `sub` pakage.

// Birth return a random place where player birth at.
// the param `salt` can use the `pubkey`.
func (m *Model) Birth(salt []byte) *Place {
	timeout := time.AfterFunc(time.Minute, func() {
		log.Fatalln("failed to birth, timeout.")
	})
	rand := append(m.gene, salt...).rand()
	for {
		x, y, z := randPoint(rand)
		place := m.Place(x, y, z)
		if _, ok := place.chunk.terrain(x, y, z); ok {
			timeout.Stop()
			return place
		}
	}
}

// PlaceProjection return [coordOfPoint, coordOfChunk, ...coordsOfArea ] and their projection.
func (m *Model) PlaceProjection(p *Place) ([][3]float64, [][2]float64) {
	points := [][3]float64{[3]float64{p.chunk.x, p.chunk.y, p.chunk.z}, [3]float64{p.x, p.y, p.z}}
	for point := range m.samples.area(p.chunk.i) {
		points = append(points, point)
	}
	projections := [][2]float64{}
	for _, point := range points {
		u, v, _ := p.chunk.projector(point[0], point[1], point[2])
		projections = append(projections, [2]float64{u, v})
	}
	return points, projections
}
