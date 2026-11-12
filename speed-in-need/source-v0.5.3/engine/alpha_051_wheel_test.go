package engine

import (
	"math"
	"testing"
)

func TestAlpha051FrontWheelsSpinAndSteerIndependently(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.Poles = nil
	w.Pickups = nil
	for i := 0; i < 240; i++ {
		w.Step(Input{Throttle: 1}, 1.0/120)
	}
	initial := w.Car.Rig.Wheels
	for i := 0; i < 90; i++ {
		w.Step(Input{Throttle: 1, Steer: .7}, 1.0/120)
	}
	wheels := w.Car.Rig.Wheels
	for i := 0; i < 4; i++ {
		if math.Abs(wheels[i].RollAngle-initial[i].RollAngle) < .2 {
			t.Fatalf("wheel %d didn't spin", i)
		}
		if math.IsNaN(wheels[i].RollAngle) || math.IsInf(wheels[i].RollAngle, 0) {
			t.Fatalf("wheel %d became nonfinite", i)
		}
	}
	if wheels[0].SteerAngle <= 0 || wheels[1].SteerAngle <= 0 {
		t.Fatalf("front wheels failed to steer: %v, %v", wheels[0].SteerAngle, wheels[1].SteerAngle)
	}
	if math.Abs(wheels[0].SteerAngle-wheels[1].SteerAngle) < .001 {
		t.Fatalf("both front wheels locked to one angle instead of Ackermann: %.6f, %.6f", wheels[0].SteerAngle, wheels[1].SteerAngle)
	}
	if math.Abs(wheels[2].SteerAngle) > 1e-6 || math.Abs(wheels[3].SteerAngle) > 1e-6 {
		t.Fatal("rear axle should not steer")
	}
}
func TestAlpha051WheelRotationSnapshotInterpolation(t *testing.T) {
	a, b := Car{}, Car{}
	a.Rig.Wheels[0].RollAngle = 15.0
	b.Rig.Wheels[0].RollAngle = 18.0
	a.Rig.Wheels[0].SteerAngle = -.20
	b.Rig.Wheels[0].SteerAngle = .24
	a.WheelSpin = Tau - .04
	b.WheelSpin = .06
	c := InterpolateCar(a, b, .5)
	if math.Abs(c.Rig.Wheels[0].RollAngle-16.5) > 1e-6 {
		t.Fatalf("wheel angle incorrectly interpolated: %v", c.Rig.Wheels[0].RollAngle)
	}
	if math.Abs(c.Rig.Wheels[0].SteerAngle-.02) > 1e-6 {
		t.Fatalf("steering jump: %v", c.Rig.Wheels[0].SteerAngle)
	}
	if math.Abs(math.Remainder(c.WheelSpin, Tau)-.01) > 1e-6 {
		t.Fatalf("wrapped legacy wheel jumped: %v", c.WheelSpin)
	}
}
func TestAlpha051BurnoutTurnsDrivenTyresWithoutCarMoving(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.Poles = nil
	w.Pickups = nil
	origin := w.Car.Pos
	for i := 0; i < 180; i++ {
		w.Step(Input{Throttle: 1, Handbrake: true}, 1.0/120)
	}
	if w.Car.Pos.Dist(origin) > .1 || math.Abs(w.Car.Speed) > .1 {
		t.Fatalf("stationary burnout moved chassis: %+v speed %v", w.Car.Pos, w.Car.Speed)
	}
	if w.Car.Rig.Wheels[2].RollAngle < 1 || w.Car.Rig.Wheels[3].RollAngle < 1 {
		t.Fatal("rear wheels must show tire rotation during burnout")
	}
	if math.Abs(w.Car.Rig.Wheels[0].RollAngle) > .1 {
		t.Fatal("front wheels should not spin during a stationary rear-wheel burnout")
	}
}
func TestAlpha051AckermannBounds(t *testing.T) {
	for _, angle := range []float64{-.45, -.2, 0, .2, .45} {
		for _, side := range []float64{-1, 1} {
			got := AckermannSteer(angle, 2.5, 1.65, side)
			if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got) > .8 {
				t.Fatalf("invalid wheel steering: %.4f -> %f", angle, got)
			}
			if math.Abs(angle) > .01 && math.Signbit(got) != math.Signbit(angle) {
				t.Fatalf("wheel went opposite: %v => %v", angle, got)
			}
		}
	}
}
