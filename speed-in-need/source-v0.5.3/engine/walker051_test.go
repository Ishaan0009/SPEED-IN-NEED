package engine

import "testing"

func Test051WalkingStopsAtImportedMeshWall(t *testing.T) {
	d := &WorldDefinition{Name: "test floor", Grid: HeightGrid{W: 3, H: 3, MinX: -50, MinZ: -50, Step: 50, Values: make([]float32, 9)}, Roads: []RoadDefinition{{Name: "floor", Width: 80, Points: []Vec{{-40, 0}, {40, 0}}}}, Spawn: Vec{-15, 0}, WaterHeight: -100}
	d.CollisionTriangles = []SurfaceTriangle{{A: [3]float64{2, 0, -10}, B: [3]float64{2, 3, -10}, C: [3]float64{2, 0, 10}}, {A: [3]float64{2, 3, -10}, B: [3]float64{2, 3, 10}, C: [3]float64{2, 0, 10}}}
	w := NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	p := Walker051{Active: true}
	for i := 0; i < 600; i++ {
		w.StepWalker051(&p, Vec{1, 0}, false, 1./120)
	}
	if p.Ped.Pos.X > 1.72 || p.Ped.Pos.X < 1.5 || p.Ped.Height != 0 {
		t.Fatalf("walker penetrated wall or floor: %+v", p.Ped.Pos)
	}
}
