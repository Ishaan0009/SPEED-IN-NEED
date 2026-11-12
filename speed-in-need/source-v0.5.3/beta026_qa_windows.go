package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
)

var driftQA026 struct{ FixtureClearance, PeakContact, PeakSlip float64 }

func observeDriftQA026() {
	if qa.stage != 93 || app.world == nil {
		return
	}
	driftQA026.PeakContact = math.Max(driftQA026.PeakContact, app.world.Car.ContactImpulse)
	driftQA026.PeakSlip = math.Max(driftQA026.PeakSlip, math.Abs(app.world.Car.DriftAngle))
}

func beta026QAStage() {
	if qa.stage < 86 {
		return
	}
	w := app.world
	if qa.stage <= 88 {
		for _, o := range w.Remaster.Def.Objects {
			if o.Asset == "b022_fern" && o.Pos[0] > 950 && o.Pos[0] < 1350 {
				remasterQAPlace(engine.Vec{X: o.Pos[0], Z: o.Pos[2] + 5}, math.Pi)
				break
			}
		}
		w.Traffic = nil
		beta.cfg.Wind = qa.stage != 86
		beta.cfg.Headlights = false
		beta.cfg.MotionBlur = false
		w.Elapsed = 120
		if qa.stage == 88 {
			w.Elapsed = 123.0
		}
	}
	if qa.stage == 89 {
		ri, at := 0, 40.
		found := false
		for i, route := range w.Routes {
			if w.Remaster.Def.Roads[i].Kind == "stunt" || w.Remaster.Def.Roads[i].Kind == "tunnel" {
				continue
			}
			for j := 1; j < len(route.Points)-1; j++ {
				a := route.Points[j].Sub(route.Points[j-1]).Unit()
				b := route.Points[j+1].Sub(route.Points[j]).Unit()
				if a.X*b.X+a.Z*b.Z < .55 {
					ri = i
					at = route.Cumulative[j] - 3
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		p, h := engine.TrafficCornerPose(w.Routes[ri], at, 1, 2.7, 7)
		f := engine.HeadingVec(h)
		r := engine.Vec{X: f.Z, Z: -f.X}
		carPos := p.Sub(f.Mul(9)).Add(r.Mul(2.5))
		delta := p.Sub(carPos)
		remasterQAPlace(carPos, math.Atan2(delta.X, delta.Z))
		w.Traffic = []engine.Traffic{{Path: ri, T: at, Direction: 1, Lane: 2.7, Speed: 7, CruiseSpeed: 9, Variant: 12, Pos: p, Heading: h}}
		w.Remaster.TrafficSnapshot = nil
	}
	if qa.stage == 90 || qa.stage == 92 {
		best := 0.
		point := w.Car.Pos
		heading := w.Car.Heading
		for _, road := range w.Remaster.Def.Roads {
			if road.Kind == "stunt" || road.Kind == "tunnel" {
				continue
			}
			for i := 1; i < len(road.Points); i++ {
				a, b := road.Points[i-1], road.Points[i]
				length := b.Sub(a).Length()
				if length < 5 {
					continue
				}
				grade := math.Abs(w.GroundHeight(b)-w.GroundHeight(a)) / length
				if grade > .11 && grade < .29 && grade > best {
					best = grade
					point = a.Add(b).Mul(.5)
					d := b.Sub(a)
					heading = math.Atan2(d.X, d.Z)
				}
			}
		}
		remasterQAPlace(point, heading)
		w.Traffic = nil
		if qa.stage == 92 {
			beta.cfg.CameraMode = 0
			app.boolFPP = true
		}
		qa.checks["hill-fixture-has-real-grade"] = best > .10
	}
	if qa.stage == 91 {
		app.camYaw = .65
		beta.cfg.CameraDistance = 4.6
		w.Traffic = nil
	}
	if qa.stage == 93 {
		best := -1.
		point, heading := w.Car.Pos, w.Car.Heading
		for x := -1300.; x <= 1300; x += 100 {
			for z := -1300.; z <= 1300; z += 100 {
				p := engine.Vec{X: x, Z: z}
				if w.IsWater(p, 60) || w.InsideTunnel(p) {
					continue
				}
				level := w.GroundHeight(p)
				grade := 0.
				for _, offset := range []engine.Vec{{X: 40}, {X: -40}, {Z: 40}, {Z: -40}} {
					grade = math.Max(grade, math.Abs(w.GroundHeight(p.Add(offset))-level)/40)
				}
				if grade > .045 {
					continue
				}
				clearance := 100.
				for _, shape := range w.Remaster.Def.Colliders {
					clearance = math.Min(clearance, p.Dist(shape.Center)-math.Max(shape.Radius, math.Hypot(shape.Half.X, shape.Half.Z)))
					if clearance < best {
						break
					}
				}
				if clearance > best {
					best = clearance
					point = p.Sub(engine.Vec{Z: 15})
					heading = 0
				}
			}
		}
		remasterQAPlace(point, heading)
		// A car-sized starting margin is useful; an arbitrary 45 m empty
		// circle does not prove that the subsequent trajectory stays clear.
		driftQA026 = struct{ FixtureClearance, PeakContact, PeakSlip float64 }{FixtureClearance: best}
		qa.checks["drift-fixture-clear-of-solid-scenery"] = best > 12
		w.Car.Speed = 27
		w.Traffic = nil
	}
	app.prevCar = w.Car
	app.renderCar = w.Car
	app.renderAlpha = 1
	remaster.CameraReset = true
}

func beta026QAChecks() {
	if qa.stage >= 86 && qa.stage <= 88 {
		qa.checks[remasterQANames[qa.stage]+"-wind-switch"] = beta.cfg.Wind == (qa.stage != 86)
	}
	if qa.stage == 89 {
		a := pbrAssets["b011_npc8"]
		qa.checks["npc-four-independent-wheel-slots"] = a != nil && len(a.Anchors) == 4
		qa.checks["npc-steering-and-spin-animate"] = len(app.world.Traffic) == 1 && math.Abs(app.world.Traffic[0].WheelSpin) > .01 && math.Abs(app.world.Traffic[0].SteerAngle) > .01
	}
	if qa.stage == 90 {
		c := app.world.Car
		a := pbrAssets["ferrari"]
		ok := a != nil
		if a != nil {
			for i := 0; i < 4; i++ {
				anchor := a.Anchors[string(rune('1'+i))]
				wheel := c.Rig.Wheels[i]
				hub := app.world.VisualWheelHub026(c, anchor, a.WheelRadii[string(rune('1'+i))], .16, wheel.SteerAngle, wheel.Touching)
				ok = ok && hub >= anchor[1]-.151 && hub <= anchor[1]+.231
			}
		}
		qa.checks["hill-wheel-travel-bounded"] = ok
	}
	if qa.stage == 91 {
		m, response := beta021Surface(pbrMaterial{Name: "chrome", Rough: .95})
		qa.checks["chrome-retains-metal-reflection"] = m.Metal > .9 && m.Rough < .2 && response.x > .9
	}
	if qa.stage == 93 {
		c := app.world.Car
		data, _ := json.MarshalIndent(map[string]any{"elapsed": app.world.Elapsed, "speed": c.Speed, "slip_angle": c.DriftAngle, "yaw_rate": c.YawRate, "position": c.Pos, "grounded": c.Body.Grounded, "contact_impulse": c.ContactImpulse, "peak_contact_impulse": driftQA026.PeakContact, "peak_slip_angle": driftQA026.PeakSlip, "fixture_clearance": driftQA026.FixtureClearance, "offroad_blend": c.OffroadBlend}, "", "  ")
		os.WriteFile(filepath.Join(qa.dir, "DRIFT_STATE_026.json"), data, 0644)
		qa.checks["drift-releases-and-settles"] = !c.DriftActive && math.Abs(c.DriftAngle) < 8 && math.Abs(c.YawRate) < .2
		qa.checks["drift-recovery-does-not-use-a-wall-stop"] = driftQA026.PeakContact < 1 && c.Speed > 15
	}
}
