package engine

import "math"

func pedCell(pos Vec) [2]int { return [2]int{int(math.Floor(pos.X / 50)), int(math.Floor(pos.Z / 50))} }
func (w *World) rebuildPedestrianGrid() {
	w.pedestrianGridCount = len(w.Pedestrians)
	if w.pedestrianGrid == nil {
		w.pedestrianGrid = make(map[[2]int][]int)
	}
	for k := range w.pedestrianGrid {
		w.pedestrianGrid[k] = w.pedestrianGrid[k][:0]
	}
	for i := range w.Pedestrians {
		k := pedCell(w.Pedestrians[i].Pos)
		w.pedestrianGrid[k] = append(w.pedestrianGrid[k], i)
	}
}
func (w *World) nearbyPedestrians(pos Vec, radius float64) []int {
	if w.pedestrianGrid == nil || w.pedestrianGridCount != len(w.Pedestrians) {
		w.rebuildPedestrianGrid()
	}
	w.pedestrianScratch = w.pedestrianScratch[:0]
	k := pedCell(pos)
	n := int(math.Ceil(radius/50)) + 1
	for x := k[0] - n; x <= k[0]+n; x++ {
		for z := k[1] - n; z <= k[1]+n; z++ {
			w.pedestrianScratch = append(w.pedestrianScratch, w.pedestrianGrid[[2]int{x, z}]...)
		}
	}
	return w.pedestrianScratch
}
