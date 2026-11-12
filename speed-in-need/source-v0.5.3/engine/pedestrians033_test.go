package engine

import (
	"math"
	"testing"
)

func TestPedestrian033PlantedFootAndIdle(t *testing.T) {
	p := Pedestrian{Moving: true, Phase: .3, Speed: 1.3}
	a := p.walkingPose()
	distance := .025
	p.Pos.Z += distance
	p.Phase += distance * 5.5
	b := p.walkingPose()
	if a[11].Sub(b[11]).Length() > .001 {
		t.Fatal("stance shoe slides", a[11], b[11])
	}
	for i := 0; i < 100; i++ {
		p.Phase = float64(i) * math.Pi / 50
		pose := p.walkingPose()
		for _, j := range []int{11, 14} {
			if pose[j].Y < .089 || pose[j].Y > .201 {
				t.Fatal("shoe penetrates pavement", pose[j])
			}
		}
	}
	p.Moving = false
	p.Pos = Vec{}
	p.Phase = .7
	a = p.walkingPose()
	p.Phase = 2.8
	b = p.walkingPose()
	if a != b || math.Abs(a[11].Y-.09) > .0001 {
		t.Fatal("stationary pedestrian marches or hovers")
	}
}
