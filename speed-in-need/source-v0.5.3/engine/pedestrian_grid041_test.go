package engine

import "testing"

func TestPedestrianGridPopulationReplacement041(t *testing.T) {
	w := &World{Pedestrians: []Pedestrian{{Pos: Vec{X: 5, Z: 8}}}}
	w.rebuildPedestrianGrid()
	w.Pedestrians = nil
	if got := w.nearbyPedestrians(Vec{}, 20); len(got) != 0 {
		t.Fatalf("removed population still has grid references: %v", got)
	}
	w.Pedestrians = []Pedestrian{{Pos: Vec{X: 400, Z: 400}}, {Pos: Vec{X: 2, Z: 2}}}
	got := w.nearbyPedestrians(Vec{}, 20)
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("replacement population was not indexed: %v", got)
	}
}
