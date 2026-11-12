package engine

import (
	"math"
	"os"
	"testing"
)

// Opt-in long simulation: six car/map combinations, each 30 simulated minutes.
// This is not a substitute for an interactive GPU soak.
func TestBetaThirtyMinuteSimulation(t *testing.T) {
	if os.Getenv("SPEED_IN_NEED_LONG_TEST") != "1" {
		t.Skip("set SPEED_IN_NEED_LONG_TEST=1")
	}
	for mapID := 0; mapID < 3; mapID++ {
		for car := 0; car < 2; car++ {
			w := NewWorld(mapID)
			w.CarClass = car
			w.InitializeChassis()
			for tick := 0; tick < 1800*120; tick++ {
				phase := float64(tick) / 120
				in := Input{Throttle: .75, Steer: math.Sin(phase*.27) * .35, Boost: tick%1800 < 120, Handbrake: tick%1200 > 1150}
				if tick%2400 > 2200 {
					in.Throttle = 0
					in.Brake = 1
				}
				w.Step(in, 1./120)
				if tick%120 == 0 {
					for _, v := range []float64{w.Car.Pos.X, w.Car.Pos.Z, w.Car.Speed, w.Car.Body.Height, w.Car.Body.VerticalSpeed, w.Car.Rig.Roll, w.Car.Rig.Pitch, w.Car.Gearbox.RPM} {
						if math.IsNaN(v) || math.IsInf(v, 0) {
							t.Fatalf("nonfinite map %d car %d tick %d", mapID, car, tick)
						}
					}
				}
				if tick > 0 && tick%12000 == 0 {
					w.Car.Pos, w.Car.Heading = NewWorld(mapID).Car.Pos, NewWorld(mapID).Car.Heading
					w.InitializeChassis()
				}
			}
			t.Logf("map=%d car=%d: 1800 simulated seconds / 216000 fixed steps passed", mapID, car)
		}
	}
}
