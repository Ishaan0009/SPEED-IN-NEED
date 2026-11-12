package engine

import (
	"math"
	"testing"
)

// Exercises 90 seconds of FIXED-STEP simulation per map. This is numerical
// regression coverage, not a Windows performance/soak/GPU verification.
func TestAlpha051ThreeMapNumericalRun(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		w.Finished = false
		for i := 0; i < 90*120; i++ {
			throttle := .65
			steer := 0.0
			if i%1800 < 400 {
				steer = .18
			}
			if i%2600 > 2250 {
				throttle = .1
			}
			w.Step(Input{Throttle: throttle, Steer: steer}, 1.0/120)
			values := []float64{w.Car.Pos.X, w.Car.Pos.Z, w.Car.Speed, w.Car.Heading, w.Car.VisualSteer, w.Car.Rig.Heave}
			for k, v := range values {
				if math.IsInf(v, 0) || math.IsNaN(v) {
					t.Fatalf("map %d tick %d component %d is %v", mapID, i, k, v)
				}
			}
			for k, wheel := range w.Car.Rig.Wheels {
				if math.IsNaN(wheel.RollAngle) || math.IsInf(wheel.RollAngle, 0) || wheel.NormalLoad < 0 {
					t.Fatalf("map %d tick %d wheel %d: %+v", mapID, i, k, wheel)
				}
			}
		}
	}
}
