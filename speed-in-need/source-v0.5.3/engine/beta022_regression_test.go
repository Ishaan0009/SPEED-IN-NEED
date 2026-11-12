package engine

import (
	"math"
	"testing"
)

func TestBeta030ProgressiveSteeringHasBoundedHoldAndReturn(t *testing.T) {
	for _, sign := range []float64{-1, 1} {
		angle := 0.
		for tick := 0; tick < 24; tick++ {
			angle = ProgressiveSteer(angle, sign, .56, 0, 1./120)
		}
		if math.Abs(angle) < .50 || math.Abs(angle) >= .56 || angle*sign <= 0 {
			t.Fatalf("short key hold did not give a partial steering angle: %f", angle)
		}
		for tick := 0; tick < 240; tick++ {
			angle = ProgressiveSteer(angle, sign, .56, 0, 1./120)
		}
		if math.Abs(angle-sign*.56) > 1e-9 {
			t.Fatalf("rack missed its stop: %f", angle)
		}
		for tick := 0; tick < 60; tick++ {
			angle = ProgressiveSteer(angle, 0, .56, 0, 1./120)
		}
		if math.Abs(angle) > 1e-9 {
			t.Fatal("released steering did not return to centre")
		}
	}
}

func TestBeta022EveryCarWheelsRollAndSteerTogether(t *testing.T) {
	for car := 0; car < 4; car++ {
		w := remasteredFixture(0, car)
		w.Car.Speed = 12
		before := w.Car.Rig.Wheels[0].RollAngle
		for tick := 0; tick < 60; tick++ {
			w.Step(Input{Throttle: .4, Steer: 1}, 1./120)
		}
		if math.Abs(math.Remainder(w.Car.Rig.Wheels[0].RollAngle-before, Tau)) < .05 {
			t.Fatalf("car %d front tire did not roll", car)
		}
		if w.Car.Rig.Wheels[0].SteerAngle <= .02 || w.Car.Rig.Wheels[1].SteerAngle <= .02 || math.Abs(w.Car.Rig.Wheels[2].SteerAngle) > 1e-9 {
			t.Fatalf("car %d steering/spin axes were not independent", car)
		}
		if w.Car.DriftActive || w.Car.DriftScore != 0 {
			t.Fatal("ordinary steering opened the drift state")
		}
	}
}

func TestBeta022TrafficCornerPoseDoesNotJumpBetweenLaneNormals(t *testing.T) {
	route := newRoute("junction", []Vec{{0, 0}, {0, 30}, {30, 30}})
	a, ha := TrafficCornerPose(route, 29.99, 1, 2.7, 4)
	b, hb := TrafficCornerPose(route, 30.01, 1, 2.7, 4)
	if a.Dist(b) > .1 || math.Abs(math.Remainder(hb-ha, Tau)) > .03 {
		t.Fatalf("NPC snapped at a turn: distance=%f yaw=%f", a.Dist(b), hb-ha)
	}
}
