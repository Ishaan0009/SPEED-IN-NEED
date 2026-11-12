package engine

import (
	"math"
	"testing"
)

func TestHandling036LateralDriftSeamDoesNotLaunch(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := flat22(class, false)
		w.Remaster.Def.Surfaces = []SurfaceTriangle{{A: [3]float64{-300, .05, 0}, B: [3]float64{300, .05, 0}, C: [3]float64{300, .05, 300}, Road: true}, {A: [3]float64{-300, .05, 0}, B: [3]float64{300, .05, 300}, C: [3]float64{-300, .05, 300}, Road: true}}
		w.Remaster.buildSpatial(w)
		w.Car.Pos = Vec{0, -.4}
		w.Car.Heading = math.Pi / 2
		w.Car.Speed = 22
		w.Car.Lateral = -8
		w.Car.Body.Initialized = false
		peak, vertical := 0., 0.
		for tick := 0; tick < 300; tick++ {
			w.Step(Input{Throttle: .3, Steer: .45, Handbrake: tick < 90}, 1./120)
			peak = math.Max(peak, w.Car.Body.Height-w.GroundHeight(w.Car.Pos))
			vertical = math.Max(vertical, w.Car.Body.VerticalSpeed)
		}
		if peak > 1.3 || vertical > 3 {
			t.Fatalf("car%d lateral shoulder launched: clearance%.2f vertical%.2f", class, peak, vertical)
		}
	}
}
