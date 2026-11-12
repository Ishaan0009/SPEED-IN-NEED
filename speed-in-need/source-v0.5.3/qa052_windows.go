package main

import (
	"math"
	"speedinneed/engine"
)

func stageQA052() {
	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	beta.carSelect = 3
	beta.paint = 1
	garage050.Setups[garageKey050()] = garageSetup050{Model: 1, Tyre: 2, Livery: 4}
	weather050.Rain = false
	weather050.Amount = 0
	remaster.Mode = 0
	remaster.Time = 0
	remaster.Opponents = 0
	startRemasteredDrive()
	beta.cfg.Headlights = false
	app.world.Traffic = nil
	app.world.Pedestrians = nil
	if qa.stage == 229 || qa.stage == 230 {
		remaster.Time = 2
		t := app.world.Remaster.Def.Tunnels[0]
		f := engine.HeadingVec(t.Heading)
		along := -t.Length*.5 + 34.
		if qa.stage == 230 {
			along += 300
		}
		remasterQAPlace(t.Center.Add(f.Mul(along)), t.Heading)
	}
	if qa.stage == 231 {
		weather050.Rain = true
		weather050.Amount = 1
		remaster.Time = 0
	}
	if qa.stage == 232 || qa.stage == 233 {
		enterWorkshop051()
		app.world.ToggleWalker051(&walker051)
		walker051.Ped.Moving = qa.stage == 233
		walker051.Ped.Speed = 0
		if qa.stage == 233 {
			walker051.Ped.Speed = 2
			walker051.Ped.Phase = math.Pi * .5
		}
		loadWalk052()
		qa.checks["052-original-rig-clip-loaded"] = walk052.Frames == 33
	}
	if qa.stage == 234 {
		b := app.world.Remaster.Def.BridgeSpans[0]
		remasterQAPlace(b.Start.Add(b.End.Sub(b.Start).Mul(.15)), math.Atan2(b.End.X-b.Start.X, b.End.Z-b.Start.Z))
	}
	if qa.stage == 235 {
		for _, o := range app.world.Remaster.Def.Objects {
			if o.Asset == "b052_pink_flower" {
				remasterQAPlace(engine.Vec{X: o.Pos[0] + 1.8, Z: o.Pos[2] + 2.8}, math.Pi)
				break
			}
		}
	}
}
func viewQA052(c engine.Car, target *pos3) {
	defer viewQA053(c, target)
	if !qa.enabled {
		return
	}
	if qa.stage == 229 || qa.stage == 230 {
		t := app.world.Remaster.Def.Tunnels[0]
		f := engine.HeadingVec(t.Heading)
		p := t.Center.Add(f.Mul(-t.Length*.5 - 8))
		q := t.Center.Add(f.Mul(-t.Length*.5 + 100))
		app.cam = pos3{p.X, t.StartY + 1.5, p.Z}
		*target = pos3{q.X, t.StartY + 2, q.Z}
	}
	if qa.stage == 231 {
		p := app.world.Car.Pos
		app.cam = pos3{p.X - 5, c.Body.Height + 2.7, p.Z - 6}
		*target = pos3{p.X + 20, c.Body.Height + 15, p.Z + 80}
	}
	if qa.stage == 232 || qa.stage == 233 {
		p := walker051.Ped
		out := p.Pos.Sub(app.world.Car.Pos).Unit()
		app.cam = pos3{p.Pos.X + out.X*3.5 + out.Z*.6, p.Height + 1.15, p.Pos.Z + out.Z*3.5 - out.X*.6}
		*target = pos3{p.Pos.X, p.Height + .9, p.Pos.Z}
	}
}
