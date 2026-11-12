package engine

import (
	"math"
	"testing"
)

func TestV3HandlingTuningDistinct(t *testing.T) {
	tour := HandlingProfile(0, 0)
	sport := HandlingProfile(0, 1)
	drift := HandlingProfile(0, 2)
	classic := HandlingProfile(1, 1)
	if !(tour.SteerLock < sport.SteerLock && sport.SteerLock < drift.SteerLock) {
		t.Fatalf("mode steering must be ordered tour<sport<drift: %+v %+v %+v", tour, sport, drift)
	}
	if !(tour.LateralGrip > sport.LateralGrip && sport.LateralGrip > drift.LateralGrip) {
		t.Fatal("drift should have less grip than sport; tour should have more")
	}
	if !(classic.EngineAcceleration < sport.EngineAcceleration && classic.SpeedCap < sport.SpeedCap) {
		t.Fatal("classic car must have distinct acceleration and top speed")
	}
}

func TestV3CarTyresMatchRoadAndGrass(t *testing.T) {
	for _, car := range []struct {
		id            int
		tyreClearance float64
	}{{0, 0}, {1, .09}} {
		road := CarElevation(0, car.id) + car.tyreClearance
		grass := CarElevation(20, car.id) + car.tyreClearance
		if math.Abs(road-.235) > .045 {
			t.Fatalf("car %d wheel contact off road elevation %.3f", car.id, road)
		}
		if math.Abs(grass+.065) > .05 {
			t.Fatalf("car %d wheel contact off grass elevation %.3f", car.id, grass)
		}
		if !(CarElevation(9, car.id) < CarElevation(7.1, car.id)) {
			t.Fatal("edge transition must lower car smoothly")
		}
	}
}

func TestV3GPSHeadingUp(t *testing.T) {
	origin := Vec{1, 2}
	right, forward := MapRelative(origin, Vec{1, 20}, 0)
	if math.Abs(right) > 1e-6 || forward <= 0 {
		t.Fatalf("+Z should appear above player: %g %g", right, forward)
	}
	right, forward = MapRelative(origin, Vec{20, 2}, 0)
	if right <= 0 || math.Abs(forward) > 1e-6 {
		t.Fatalf("+X should appear right: %g %g", right, forward)
	}
	right, forward = MapRelative(origin, Vec{20, 2}, math.Pi/2)
	if forward <= 0 || math.Abs(right) > 1e-6 {
		t.Fatalf("+X should appear above after right turn: %g %g", right, forward)
	}
}

func TestV3ResetPreservesSelectedCarMode(t *testing.T) {
	w := NewWorld(2)
	w.CarClass, w.DriveMode = 1, 2
	w.Reset()
	if w.CarClass != 1 || w.DriveMode != 2 {
		t.Fatal("restarting a challenge must preserve player handling setup")
	}
}
