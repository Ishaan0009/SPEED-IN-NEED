package engine

import "math"

// Static collision broadphase: obstacles are indexed once per map, independent
// of their rendering LOD. Only nearby cells undergo precise contact testing.
const obstacleCellMetres = 28.0

type gridKey struct{ X, Z int }
type obstacleCell struct{ Trees, Rocks, Buildings []int }

func gridOf(p Vec) gridKey {
	return gridKey{int(math.Floor(p.X / obstacleCellMetres)), int(math.Floor(p.Z / obstacleCellMetres))}
}
func (w *World) RebuildCollisionGrid() {
	w.CollisionCells = make(map[gridKey]*obstacleCell)
	cell := func(p Vec) *obstacleCell {
		k := gridOf(p)
		c := w.CollisionCells[k]
		if c == nil {
			c = &obstacleCell{}
			w.CollisionCells[k] = c
		}
		return c
	}
	for i, t := range w.Trees {
		cell(t.Pos).Trees = append(cell(t.Pos).Trees, i)
	}
	for i, r := range w.Rocks {
		cell(r.Pos).Rocks = append(cell(r.Pos).Rocks, i)
	}
	for i, b := range w.Buildings {
		cell(b).Buildings = append(cell(b).Buildings, i)
	}
	w.gridTreeCount = len(w.Trees)
	w.gridRockCount = len(w.Rocks)
	w.gridBuildingCount = len(w.Buildings)
}
func (w *World) nearby(p Vec, r float64, kind int) []int {
	// Externally edited worlds (including unit-test fixtures) are allowed.
	if w.CollisionCells == nil || w.gridTreeCount != len(w.Trees) || w.gridRockCount != len(w.Rocks) || w.gridBuildingCount != len(w.Buildings) {
		w.RebuildCollisionGrid()
	}
	low := gridOf(Vec{p.X - r - obstacleCellMetres, p.Z - r - obstacleCellMetres})
	high := gridOf(Vec{p.X + r + obstacleCellMetres, p.Z + r + obstacleCellMetres})
	var indices []int
	for x := low.X; x <= high.X; x++ {
		for z := low.Z; z <= high.Z; z++ {
			c := w.CollisionCells[gridKey{x, z}]
			if c == nil {
				continue
			}
			switch kind {
			case 0:
				indices = append(indices, c.Trees...)
			case 1:
				indices = append(indices, c.Rocks...)
			case 2:
				indices = append(indices, c.Buildings...)
			}
		}
	}
	return indices
}
func (w *World) NearbyTreeIndices(p Vec, radius float64) []int     { return w.nearby(p, radius, 0) }
func (w *World) NearbyRockIndices(p Vec, radius float64) []int     { return w.nearby(p, radius, 1) }
func (w *World) NearbyBuildingIndices(p Vec, radius float64) []int { return w.nearby(p, radius, 2) }

// PredictiveFollowingSpeed produces a smooth desired speed based on a desired
// 1.25-second gap and relative closing speed. No hard brake thresholds.
func PredictiveFollowingSpeed(followerSpeed, leaderSpeed, bumperGap float64) float64 {
	if bumperGap <= .6 {
		return 0
	}
	leaderSpeed = math.Max(0, leaderSpeed)
	safe := 2.4 + math.Max(0, followerSpeed)*1.1
	pace := math.Max(0, leaderSpeed+.55*(bumperGap-safe))
	stopping := math.Sqrt(leaderSpeed*leaderSpeed + 7*math.Max(0, bumperGap-2.4))
	return math.Min(pace, stopping)
}
