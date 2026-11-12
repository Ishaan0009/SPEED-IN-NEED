package engine

import (
	"math"
	"testing"
)

func TestBeta014TurningAlonePreservesWorldMomentum(t *testing.T) {
	for _, yaw := range []float64{-1.8, .7, 1.8} {
		speed, lateral, heading := 30., 2., .63
		f := HeadingVec(heading)
		r := Vec{f.Z, -f.X}
		expected := f.Mul(speed).Add(r.Mul(lateral))
		for i := 0; i < 1200; i++ {
			speed, lateral, heading = AdvancePlanarMomentum(speed, lateral, heading, yaw, 0, 0, 1./120)
		}
		f = HeadingVec(heading)
		r = Vec{f.Z, -f.X}
		actual := f.Mul(speed).Add(r.Mul(lateral))
		if actual.Dist(expected) > 1e-8 {
			t.Fatalf("orientation created a force: expected=%v actual=%v", expected, actual)
		}
	}
}

func TestBeta014SteeringMirrorsAcrossAllCars(t *testing.T) {
	for class := 0; class < 4; class++ {
		angles := [2]float64{}
		for j, steer := range []float64{-.45, .45} {
			w := remasteredFixture(0, class)
			w.Remaster.Racers = nil
			w.Car.Pos = Vec{0, -300}
			w.Car.Heading = 0
			w.Car.Speed = 17
			for i := 0; i < 120; i++ {
				w.StepRemastered(Input{Throttle: .25, Steer: steer}, 1./120)
			}
			angles[j] = w.Car.Heading
			if w.Car.DriftActive || math.IsNaN(angles[j]) || math.Abs(w.Car.Rig.Roll) > .19 {
				t.Fatalf("unstable ordinary turn class=%d car=%+v", class, w.Car)
			}
		}
		if angles[0] >= -.02 || angles[1] <= .02 || math.Abs(angles[0]+angles[1]) > .04 {
			t.Fatalf("left/right mismatch class=%d angles=%v", class, angles)
		}
	}
}

func TestBeta014MeshWallStopsAHighSpeedCar(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Remaster.Racers = nil
	w.Remaster.Def.CollisionTriangles = []SurfaceTriangle{
		{A: [3]float64{-50, 0, 0}, B: [3]float64{50, 0, 0}, C: [3]float64{-50, 5, 0}},
		{A: [3]float64{50, 0, 0}, B: [3]float64{50, 5, 0}, C: [3]float64{-50, 5, 0}},
	}
	w.Remaster.buildMeshContacts()
	w.Car.Pos = Vec{0, -8}
	w.Car.Heading = 0
	w.Car.Speed = 70
	for i := 0; i < 120; i++ {
		w.StepRemastered(Input{}, 1./120)
		if w.Car.Pos.Z > -.5 {
			t.Fatalf("crossed visible wall at frame %d: %v", i, w.Car.Pos)
		}
	}
}
