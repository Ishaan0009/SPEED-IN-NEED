package engine

import (
	"math"
	"testing"
)

func TestBeta026TreadClearsInclinedGroundAtRenderedAxles(t *testing.T) {
	for _, grade := range [][2]float64{{0, 0}, {.12, .3}, {-.15, -.32}} {
		gx, gz := grade[0], grade[1]
		c := Car{}
		c.Body.Height = .74
		c.Rig.Pitch = -math.Atan(gz)
		c.Rig.Roll = math.Atan(gx * math.Cos(c.Rig.Pitch))
		ground := func(p Vec) float64 { return gx*p.X + gz*p.Z }
		for _, x := range []float64{-.8, .8} {
			for _, z := range []float64{-1.3, 1.25} {
				for _, steer := range []float64{-.4, 0, .4} {
					anchor := [3]float64{x, .34, z}
					hub := WheelHub026(c, anchor, .34, .15, steer, true, ground)
					minimum := math.Inf(1)
					for i := 0; i < 180; i++ {
						for _, side := range []float64{-.15, .15} {
							a := float64(i) * Tau / 180
							wx, wy, wz := side, .34*math.Cos(a), .34*math.Sin(a)
							wx, wz = math.Cos(steer)*wx+math.Sin(steer)*wz, -math.Sin(steer)*wx+math.Cos(steer)*wz
							wx += x
							wy += hub
							wz += z
							wx, wy = math.Cos(c.Rig.Roll)*wx-math.Sin(c.Rig.Roll)*wy, math.Sin(c.Rig.Roll)*wx+math.Cos(c.Rig.Roll)*wy
							wy, wz = math.Cos(c.Rig.Pitch)*wy-math.Sin(c.Rig.Pitch)*wz, math.Sin(c.Rig.Pitch)*wy+math.Cos(c.Rig.Pitch)*wz
							minimum = math.Min(minimum, wy-ground(Vec{wx, wz}))
						}
					}
					if minimum < .044 || minimum > .055 {
						t.Fatalf("tire cuts/floats above slope: grade=%v steer=%.2f clearance=%.5f", grade, steer, minimum)
					}
				}
			}
		}
	}
}

func TestBeta026SharpCrestsCannotDetachHubs(t *testing.T) {
	c := Car{}
	c.Body.Height = .74
	c.Rig.Pitch = -.9
	anchor := [3]float64{.8, .34, 1.3}
	hub := WheelHub026(c, anchor, .34, .15, .4, true, func(Vec) float64 { return 5 })
	if hub > anchor[1]+.231 || hub < anchor[1]-.151 {
		t.Fatalf("wheel detached from suspension: %.3f", hub)
	}
	if h := WheelHub026(c, anchor, .34, .15, 0, false, func(Vec) float64 { return 0 }); math.Abs(h-(anchor[1]-.1)) > 1e-8 {
		t.Fatal("airborne wheel did not use bounded droop")
	}
}

func TestBeta026TrafficHeadingFollowsRoundedLaneMotion(t *testing.T) {
	route := newRoute("right angle", []Vec{{0, -40}, {0, 0}, {40, 0}, {40, 40}})
	for _, direction := range []float64{-1, 1} {
		for s := 32.; s < 48; s += .08 {
			p, h := TrafficCornerPose(route, s, direction, 2.7, 12)
			q, _ := TrafficCornerPose(route, s+.02*direction, direction, 2.7, 12)
			movement := q.Sub(p).Unit()
			f := HeadingVec(h)
			if movement.X*f.X+movement.Z*f.Z < .995 {
				t.Fatalf("NPC slides sideways at knot %.2f", s)
			}
			stopped, _ := TrafficCornerPose(route, s, direction, 2.7, 0)
			if stopped.Dist(p) > .00001 {
				t.Fatal("braking moves the planned lane sideways")
			}
		}
	}
}

func TestBeta026DriftRetainsMomentumAndSettles(t *testing.T) {
	for car := 0; car < 4; car++ {
		for _, sign := range []float64{-1, 1} {
			w := flat22(car, true)
			w.Car.Speed = 27
			peak := 0.
			for tick := 0; tick < 84; tick++ {
				w.Step(Input{Throttle: .65, Steer: sign * .65, Handbrake: true}, 1./120)
				peak = math.Max(peak, math.Abs(w.Car.DriftAngle))
			}
			if peak < 7 || peak > 50 {
				t.Fatalf("car %d uncontrolled/no drift: %.2f", car, peak)
			}
			if math.Hypot(w.Car.Speed, w.Car.Lateral) < 16 {
				t.Fatalf("car %d sheds excessive speed during short drift", car)
			}
			for tick := 0; tick < 300; tick++ {
				w.Step(Input{Throttle: .5}, 1./120)
			}
			if math.Abs(w.Car.DriftAngle) > 5 || math.Abs(w.Car.YawRate) > .12 || w.Car.DriftActive {
				t.Fatalf("car %d keeps sliding after release: angle %.2f yaw %.2f", car, w.Car.DriftAngle, w.Car.YawRate)
			}
		}
	}
}

func TestBeta026RenderInterpolationIncludesGroundContact(t *testing.T) {
	a, b := Car{}, Car{}
	a.Rig.Wheels[0].GroundHeight = 2
	b.Rig.Wheels[0].GroundHeight = 4
	if c := InterpolateCar(a, b, .25); c.Rig.Wheels[0].GroundHeight != 2.5 {
		t.Fatal("wheel support still steps at physics ticks")
	}
}
