package main

import (
	"math"
	"speedinneed/engine"
)

func beta030QAStage() {
	if qa.stage < 94 {
		return
	}
	if app.world.Remaster.Mode != 0 {
		remaster.Mode = 0
		beta.mode = 0
		beta.mapSelect = 0
		activateRemasteredWorld(true)
	}
	w := app.world
	if qa.stage != 97 && qa.stage != 99 {
		w.Traffic = nil
	}
	beta.screen = screenDrive
	app.started = true
	beta.cfg.CameraMode = 1
	app.boolFPP = false
	remaster.Time = 0
	w.Remaster.Time = 0
	beta.cfg.Headlights = true
	tunnel := w.Remaster.Def.Tunnels[0]
	f := engine.HeadingVec(tunnel.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	pixel := func(x, y float64) engine.Vec {
		return engine.Vec{X: (x - 625.5) * 3000 / 1251, Z: (y - 626) * 3000 / 1251}
	}
	switch qa.stage {
	case 94:
		remasterQAPlace(tunnel.Center.Add(r.Mul(85)).Add(f.Mul(-35)), math.Pi*.35)
		beta.cfg.CameraMode = 2
	case 95:
		remasterQAPlace(pixel(410, 330), math.Pi*.15)
	case 96:
		remasterQAPlace(tunnel.Center.Add(f.Mul(tunnel.Length*.5+22)), tunnel.Heading+math.Pi)
	case 97:
		remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
		remaster.Time = 1
		w.Remaster.Time = 1
	case 98:
		remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
		beta.prevScreen = screenDrive
		beta.screen = screenSettings
		remaster.SettingsTab = 5
	case 100:
		remasterQAPlace(tunnel.Center, tunnel.Heading)
		remaster.Time = 2
		w.Remaster.Time = 2
	case 99:
		remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
		remaster.Time = 2
		w.Remaster.Time = 2
		beta.paint = 4
	case 101:
		b := w.Remaster.Def.BridgeSpans[0]
		remasterQAPlace(b.Start.Add(b.End).Mul(.5), math.Atan2(b.End.X-b.Start.X, b.End.Z-b.Start.Z))
	case 102:
		remasterQAPlace(pixel(1040, 180), math.Pi*.7)
		beta.cfg.CameraMode = 2
	case 103:
		remasterQAPlace(pixel(510, 780), math.Pi*.3)
	case 104:
		b := w.Remaster.Def.BridgeSpans[0]
		remasterQAPlace(b.Start, math.Atan2(b.End.X-b.Start.X, b.End.Z-b.Start.Z))
	}
	qa.checks["remastered-reference-name"] = w.Remaster.Def.Name == "SPEED IN NEED"
	qa.checks["remastered-central-long-tunnel"] = tunnel.Length > 490 && tunnel.Length < 520 && tunnel.Center.Z > -800
	oldCamera, oldStation, oldQuality, oldScreen := beta.cfg.CameraMode, beta.cfg.RadioStation, app.quality, beta.screen
	beta.screen = screenDrive
	remasterKey('V')
	qa.checks["v-changes-camera-without-quality"] = beta.cfg.CameraMode == (oldCamera+1)%4 && app.quality == oldQuality
	remasterKey('F')
	qa.checks["f-changes-radio-without-quality"] = beta.cfg.RadioStation == (oldStation+1)%len(radioNames030) && app.quality == oldQuality
	beta.cfg.CameraMode, beta.cfg.RadioStation, beta.screen = oldCamera, oldStation, oldScreen
	app.boolFPP = false
	remaster.CameraReset = true
	app.prevCar = w.Car
	app.renderCar = w.Car
}
