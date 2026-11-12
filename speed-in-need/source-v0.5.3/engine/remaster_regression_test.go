package engine

import (
	"math"
	"testing"
)

func TestRemasteredIndependentWheelLoads(t *testing.T) {
	w := NewWorld(2)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	w.Rocks = nil
	for i := 0; i < 120; i++ {
		w.Step(Input{Throttle: .3}, 1./120)
	}
	for i, wheel := range w.Car.Rig.Wheels {
		if math.IsNaN(wheel.NormalLoad) || math.IsInf(wheel.NormalLoad, 0) || wheel.NormalLoad < 0 {
			t.Fatalf("invalid wheel %d load: %+v", i, wheel)
		}
	}
	if w.Car.Gearbox.Gear < 1 || w.Car.Gearbox.RPM < 500 {
		t.Fatalf("gearbox not engaged: %+v", w.Car.Gearbox)
	}
}
func TestRemasteredReverseGearAndRelease(t *testing.T) {
	c := Car{}
	for i := 0; i < 120; i++ {
		a := DriveAcceleration(&c, Input{Brake: 1}, 1./120, 0)
		c.Speed = Clamp(c.Speed+a/120, -13, 60)
	}
	if c.Gearbox.Gear != -1 || c.Speed >= -1 {
		t.Fatalf("reverse gear does not engage: gear=%d speed=%g", c.Gearbox.Gear, c.Speed)
	}
	for i := 0; i < 220; i++ {
		a := DriveAcceleration(&c, Input{Throttle: 1}, 1./120, 0)
		c.Speed = Clamp(c.Speed+a/120, -13, 60)
	}
	if c.Gearbox.Gear < 1 || c.Speed < .2 {
		t.Fatalf("not back in forward: gear=%d speed=%g", c.Gearbox.Gear, c.Speed)
	}
}
func TestRemasteredPredictiveFollowing(t *testing.T) {
	slow := PredictiveFollowingSpeed(22, 8, 6)
	far := PredictiveFollowingSpeed(22, 8, 60)
	if slow > 8 || far < slow+8 {
		t.Fatalf("unsafe or stalled following: near=%g far=%g", slow, far)
	}
}
func TestRemasteredCarInterpolationWrap(t *testing.T) {
	a := Car{Pos: Vec{2, 1}, Heading: math.Pi - .05, Speed: 10}
	b := Car{Pos: Vec{4, 3}, Heading: -math.Pi + .05, Speed: 20}
	mid := InterpolateCar(a, b, .5)
	if mid.Pos.Dist(Vec{3, 2}) > .000001 || math.Abs(math.Abs(mid.Heading)-math.Pi) > .001 {
		t.Fatalf("incorrect wrap interpolation: %+v", mid)
	}
}
func TestRemasteredWaterPolygonBoundaries(t *testing.T) {
	water := WaterForMap(0)[0]
	if !water.Contains(Vec{-250, -285}) || water.Contains(Vec{-250, 600}) || water.Contains(Vec{0, 0}) {
		t.Fatal("water collider does not match coastal mesh")
	}
	dry := water.ClosestDry(Vec{-260, -285})
	if water.Contains(dry) || dry.X < CoastWaterMaxX {
		t.Fatalf("water recovery still wet: %v", dry)
	}
}
func TestRemasteredCollisionGrid(t *testing.T) {
	w := NewWorld(1)
	p := w.Trees[8].Pos
	indices := w.NearbyTreeIndices(p, 2)
	ok := false
	for _, index := range indices {
		if index == 8 {
			ok = true
			break
		}
	}
	if !ok {
		t.Fatal("near tree missing from spatial collision index")
	}
}
func TestRemasteredWorldFinite(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		w.Traffic = nil
		w.Poles = nil
		w.Trees = nil
		w.VisibleTrees = 0
		w.Rocks = nil
		w.Buildings = nil
		for step := 0; step < 1200; step++ {
			in := Input{Throttle: .75, Steer: math.Sin(float64(step)*.012) * .35}
			if step > 900 {
				in.Handbrake = true
			}
			w.Step(in, 1./120)
			nums := []float64{w.Car.Pos.X, w.Car.Pos.Z, w.Car.Speed, w.Car.Heading, w.Car.Rig.Heave, w.Car.Rig.Roll, w.Car.Rig.Pitch}
			for _, v := range nums {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					t.Fatalf("map %d simulation invalid at step %d", mapID, step)
				}
			}
		}
	}
}
