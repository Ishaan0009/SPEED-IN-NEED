package main

import (
	"math"
	"speedinneed/engine"
	"time"
)

var qa033PeakSpan float64

func beta033QAObserve() {
	if qa.stage == 116 {
		qa033PeakSpan = math.Max(qa033PeakSpan, nav033.Span)
	}
}

func beta033QAStage() {
	if qa.stage < 113 {
		return
	}
	w := app.world
	remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
	nav033.Active = false
	nav033.Span = 650
	nav033.Braked = time.Time{}
	beta.cfg.MapOrientation = 0
	switch qa.stage {
	case 113, 114, 115, 116:
		target := engine.Vec{X: 75, Z: -675}
		if qa.stage == 113 {
			target = engine.Vec{X: 300, Z: 500}
		}
		setWaypoint033(target)
		nav033.Cursor = target
		if qa.stage == 113 {
			remaster.FullMap = true
			x, y, size := navMapRect033()
			delta := target.Sub(nav033.Center).Mul(size / navMapSpan033())
			px, py := int(x+size*.5+delta.X), int(y+size*.5+delta.Z)
			packed := uintptr(uint32(uint16(px)) | uint32(uint16(py))<<16)
			navMapPointer033(WM_LBUTTONDOWN, 0, packed)
			navMapPointer033(0x202, 0, packed)
			qa.checks["033-mouse-sets-waypoint"] = nav033.Active && nav033.Target.Dist(target) < 5
			navMapPointer033(WM_RBUTTONDOWN, 0, packed)
			qa.checks["033-mouse-clears-waypoint"] = !nav033.Active
			nav033.Cursor = target
			navMapKey033(VK_RETURN)
			qa.checks["033-keyboard-sets-waypoint"] = nav033.Active && nav033.Target.Dist(target) < .01
		}
		if qa.stage == 115 {
			beta.cfg.MapOrientation = 1
		}
		if qa.stage == 116 {
			qa033PeakSpan = 650
			w.Car.Speed = 40
			w.Car.Braking = true
			nav033.Braked = time.Now()
		}
		qa.checks["033-road-route-reachable"] = nav033.Path.Reachable && len(nav033.Path.Points) > 5
		if qa.stage == 113 {
			qa.checks["033-offroad-target-snaps-to-road"] = nav033.Path.OffRoadMetres > 20 && nav033.Path.OffRoadMetres < 100
		}
	case 117:
		app.started = false
		selectScreen(screenTitle)
		beta020MenuUI()
		focus033.Ready = false
		menuKey033(0x28)
		qa.checks["033-title-arrow-down"] = focus033.ID == "settings"
		menuKey033(0x26)
		qa.checks["033-title-arrow-up"] = focus033.ID == "start"
		menuKey033(VK_RETURN)
		beta020MenuUI()
		menuKey033(0x27)
		qa.checks["033-mode-arrow"] = remaster.Mode == 1
		menuKey033(VK_RETURN)
		qa.checks["033-mode-confirm"] = ui020.Stage == 1
		beta020MenuUI()
		menuKey033(VK_RETURN)
		beta020MenuUI()
		before := beta.carSelect
		menuKey033(0x27)
		qa.checks["033-car-arrow"] = beta.carSelect == (before+1)%4
		selectScreen(screenMaps)
		ui020.Stage = 0
		remaster.Mode = 0
		focus033.Ready = false
	case 118, 119:
		beta.prevScreen = screenDrive
		selectScreen(screenSettings)
		remaster.SettingsTab = 7
		beta.selection = 8
		if qa.stage == 119 {
			remaster.SettingsTab = 6
			beta.selection = 0
		}
	case 120:
		remaster.Time = 1
		w.Remaster.Time = 1
		w.Car.Speed = 0
	case 121:
		app.boolFPP = true
		beta.cfg.CameraMode = 0
		w.Car.Speed = 35
		beta.cfg.MotionBlur = true
	case 122:
		setWaypoint033(engine.Vec{X: 75, Z: -675})
		nav033.Center = engine.Vec{X: -1000, Z: 600}
		nav033.Cursor = nav033.Target
		nav033.Zoom = 6
		remaster.FullMap = true
	case 123:
		best := 0
		for i, v := range w.Remaster.Def.Grid.Values {
			if v > w.Remaster.Def.Grid.Values[best] {
				best = i
			}
		}
		g := w.Remaster.Def.Grid
		p := engine.Vec{X: g.MinX + float64(best%g.W)*g.Step, Z: g.MinZ + float64(best/g.W)*g.Step}
		remasterQAPlace(p, 0)
		remaster.Time = 0
		w.Remaster.Time = 0
		w.Car.Speed = 0
	}
	app.prevCar = w.Car
	app.renderCar = w.Car
	remaster.CameraReset = true
}
func beta033QAChecks() {
	if qa.stage >= 113 && qa.stage <= 116 {
		qa.checks["033-textured-map-loaded"] = app.tex["b038_map_free.png"] != 0
		qa.checks["033-route-still-active"] = nav033.Active && nav033.Path.Reachable
		if qa.stage == 116 {
			qa.checks["033-braking-expands-minimap"] = qa033PeakSpan > 900
		}
	}
	if qa.stage == 118 {
		qa.checks["033-controller-settings-scroll"] = settingFirst033() <= beta.selection && settingFirst033()+9 > beta.selection
	}
	if qa.stage == 121 {
		qa.checks["033-fpp-finite-eye"] = !math.IsNaN(app.cam.y) && app.boolFPP
	}
}
