package main

import (
	"math"
	"speedinneed/engine"
)

func registerQA053() {
	remasterQANames = append(remasterQANames, "053-M4-driver-lamps", "053-EB110-driver-lamps", "053-LP670-driver-lamps", "053-R34-driver-lamps", "053-Jeep-driver-lamps", "053-Ferrari-restored", "053-Porsche-restored", "053-McLaren-restored", "053-Trueno-restored", "053-walking-forward", "053-walking-backward", "053-grass-night", "053-rain-puddle-variation")
}
func stageQA053() {
	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	beta.carSelect = 3
	beta.paint = 1
	weather050.Rain = false
	weather050.Amount = 0
	remaster.Mode = 0
	remaster.Time = 2
	remaster.Opponents = 0
	variant := 1
	if qa.stage <= 244 {
		variant = playerVariants052[qa.stage-236]
	}
	garage050.Setups[garageKey050()] = garageSetup050{Model: variant, Tyre: 0, Livery: 0}
	startRemasteredDrive()
	app.world.Traffic = nil
	app.world.Pedestrians = nil
	beta.cfg.Headlights = true
	app.world.Car.Braking = true
	if qa.stage == 245 || qa.stage == 246 {
		app.world.ToggleWalker051(&walker051)
	}
	if qa.stage == 247 {
		for _, o := range app.world.Remaster.Def.Objects {
			if o.Kind == "tree" && o.Pos[1] > 60 && o.Pos[1] < 100 {
				remasterQAPlace(engine.Vec{X: o.Pos[0] + 6, Z: o.Pos[2] + 6}, math.Pi)
				break
			}
		}
	}
	if qa.stage == 248 {
		remaster.Time = 0
		weather050.Rain = true
		weather050.Amount = 1
	}
}
func viewQA053(c engine.Car, target *pos3) {
	if !qa.enabled || qa.stage < 236 || qa.stage > 248 {
		return
	}
	if qa.stage <= 244 {
		f := engine.HeadingVec(c.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		p := c.Pos.Add(f.Mul(4.8)).Add(r.Mul(-2.3))
		app.cam = pos3{p.X, c.Body.Height + .8, p.Z}
		*target = pos3{c.Pos.X, c.Body.Height + .1, c.Pos.Z}
	}
}
