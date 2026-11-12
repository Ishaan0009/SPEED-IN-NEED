package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
)

func registerQA036() {
	if len(remasterQANames) > 124 {
		return
	}
	remasterQANames = append(remasterQANames, "036-driver-front-ferrari", "036-driver-front-porsche", "036-driver-front-gt", "036-driver-front-trueno", "036-day-metal", "036-night-blend", "036-select-map-free", "036-select-map-race", "036-loading", "036-zoom-near", "036-zoom-far", "036-circuit-detail")
}
func stageQA036() {
	if qa.stage < 124 {
		return
	}
	remaster.FullMap = false
	beta.screen = screenDrive
	app.started = true
	app.boolFPP = false
	beta.cfg.CameraMode = 1
	app.camDist = 5.8
	mode := 0
	id := 0
	if qa.stage >= 124 && qa.stage <= 127 {
		id = qa.stage - 124
	}
	if qa.stage == 131 || qa.stage == 135 {
		mode = 1
	}
	if app.world.CarClass != id || remaster.Mode != mode {
		remaster.Mode = mode
		beta.carSelect = id
		activateRemasteredWorld(true)
	}
	beta.carSelect = id
	w := app.world
	w.Traffic = nil
	remaster.Time = 0
	w.Remaster.Time = 0
	remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
	if qa.stage == 128 || qa.stage == 129 {
		beta.paint = 4
	}
	if qa.stage == 129 {
		remaster.Time = 2
		w.Remaster.Time = 2
	}
	if qa.stage == 130 || qa.stage == 131 {
		app.started = false
		beta.screen = screenMaps
		ui020.Stage = 1
	}
	if qa.stage == 133 {
		app.camDist = 3.5
	}
	if qa.stage == 134 {
		app.camDist = 14
	}
	if qa.stage == 135 {
		p, f := w.Routes[0].AtDistance(w.Routes[0].Len * .24)
		remasterQAPlace(p, math.Atan2(f.X, f.Z))
		w.Remaster.Countdown = 0
	}
	qa.checks["036-expected-map-mode"] = app.mapID == mode && w.Remaster.Mode == mode
}
func viewQA036(c engine.Car, target *pos3) {
	if !qa.enabled || qa.stage < 124 || qa.stage > 129 {
		return
	}
	f := engine.HeadingVec(c.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	p := c.Pos.Add(f.Mul(5.6)).Add(r.Mul(-2.5))
	app.cam = pos3{p.X, c.Body.Height + .75, p.Z}
	*target = pos3{c.Pos.X, c.Body.Height + .1, c.Pos.Z}
}
func paintQA036(index int) bool {
	for _, b := range ui020.Buttons {
		if b.ID == fmt.Sprintf("paint:%d", index) {
			beta020Click(int(b.X+b.W*.5), int(b.Y+b.H*.5))
			return beta.paint == index
		}
	}
	return false
}
func checks036() {
	if beta.screen == screenMaps && ui020.Stage == 2 {
		for _, b := range ui020.Buttons {
			if b.ID == "paint-" || b.ID == "paint+" {
				qa.checks["036-no-colour-cycle-buttons"] = false
				return
			}
		}
		qa.checks["036-no-colour-cycle-buttons"] = true
	}
	if qa.stage >= 124 && qa.stage <= 129 {
		qa.checks["036-close-driver-source-mesh"] = driverCache035.LOD == 0
		qa.checks["036-upright-driver-head"] = driverSkinBasis035(seatedDriverPose032(beta.carSelect, 0))[14].NormalAxis.Y > .99
	}
	if qa.stage == 133 || qa.stage == 134 {
		qa.checks["036-zoom-preset-effective"] = math.Abs(aokawaCameraDistance()-app.camDist) < .001
	}
}
