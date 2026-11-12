package engine

import (
	"math"
	"testing"
)

func TestBeta030HillContactsRemainFiniteAndFollowDescent(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, grade := range []float64{-.35, .35} {
			w := flat22(class, true)
			g := &w.Remaster.Def.Grid
			for z := 0; z < g.H; z++ {
				for x := 0; x < g.W; x++ {
					g.Values[z*g.W+x] = float32(200 + grade*(g.MinZ+float64(z)*g.Step))
				}
			}
			w.Car.Pos = Vec{}
			w.Car.Heading = 0
			w.Car.Body.Initialized = false
			w.InitializeChassis()
			w.Car.Body.Height = 200.74
			w.Car.Rig.Pitch = -math.Atan(grade)
			w.Car.Speed = 12
			maxError, maxPitch := 0., 0.
			grounded := 0
			for tick := 0; tick < 600; tick++ {
				w.Step(Input{Throttle: .6}, 1./120)
				c := w.Car
				if math.IsNaN(c.Body.Height) || math.IsInf(c.Rig.Pitch, 0) {
					t.Fatal("non-finite hill motion")
				}
				if tick > 120 {
					maxError = math.Max(maxError, math.Abs(c.Body.Height-w.remasteredSupport(&c, c.Pos)-.74))
					maxPitch = math.Max(maxPitch, math.Abs(c.Rig.Pitch+math.Atan(grade)))
					if c.Body.Grounded {
						grounded++
					}
				}
			}
			t.Logf("class %d grade %.2f gap %.3f grounded %d pitch error %.3f vy %.3f speed %.3f", class, grade, maxError, grounded, maxPitch, w.Car.Body.VerticalSpeed, w.Car.Speed)
			if maxError > .65 || grounded < 360 {
				t.Errorf("class %d grade %.2f loses hill support: gap %.3f grounded %d", class, grade, maxError, grounded)
			}
			if grade < 0 && w.Car.Speed > 4 && w.Car.Body.VerticalSpeed > -.4 {
				t.Fatalf("class %d descent still sticks/jitters: vy %.3f", class, w.Car.Body.VerticalSpeed)
			}
		}
	}
}

func TestBeta030SidehillRollFollowsGroundBank(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, grade := range []float64{-.35, .35} {
			w := flat22(class, true)
			g := &w.Remaster.Def.Grid
			for z := 0; z < g.H; z++ {
				for x := 0; x < g.W; x++ {
					g.Values[z*g.W+x] = float32(200 + grade*(g.MinX+float64(x)*g.Step))
				}
			}
			w.Car.Pos = Vec{}
			w.Car.Heading = 0
			w.Car.Body.Initialized = false
			w.InitializeChassis()
			w.Car.Body.Height = 200.74
			w.Car.Speed = 8
			grounded := 0
			maxRollError := 0.
			for tick := 0; tick < 480; tick++ {
				w.Step(Input{Throttle: .3}, 1./120)
				if tick > 120 {
					maxRollError = math.Max(maxRollError, math.Abs(w.Car.Rig.Roll-math.Atan(grade)))
					if w.Car.Body.Grounded {
						grounded++
					}
				}
			}
			t.Logf("class %d bank %.2f roll %.3f error %.3f grounded %d", class, grade, w.Car.Rig.Roll, maxRollError, grounded)
			if maxRollError > .14 || grounded < 300 {
				t.Errorf("sidehill suspension fails to follow bank")
			}
		}
	}
}
