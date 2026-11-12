//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
	"speedinneed/engine"
	"strconv"
	"strings"
	"time"
)

var remasterQANames = []string{"title", "setup-ferrari", "setup-porsche", "setup-custom-gt", "setup-trueno", "free-day-ferrari", "free-afternoon-porsche", "free-night-custom-gt", "free-day-trueno", "cockpit-ferrari", "cockpit-porsche", "cockpit-custom-gt", "cockpit-trueno", "tunnel-approach", "tunnel-inside", "bridge", "forest", "night-lights-off", "night-lights-on", "reverse-lights", "drift-smoke", "pause", "settings", "race-setup", "race-grid", "race-driving", "race-cockpit", "race-results", "stunt-ramp", "stunt-air-nitro", "traffic-signal", "fallen-pole", "hill-road", "grass-close", "race-night", "race-seven-opponents", "setup-ui150", "full-map", "graphics-settings", "camera-far", "camera-near", "cave-reverse", "dual-exhaust-nitro", "signal-junction", "low-preset", "ultra-preset", "west-beach", "moving-water", "brake-skidmarks", "traffic-busy", "map-zoom", "moon-direction", "bus-scale", "npc-forward-lamps", "fallen-signal", "loose-roadside-prop"}

func init() {
	for mode := 0; mode < 2; mode++ {
		for car := 0; car < 4; car++ {
			for light := 0; light < 3; light++ {
				remasterQANames = append(remasterQANames, fmt.Sprintf("matrix-mode%d-car%d-light%d", mode, car, light))
			}
		}
	}
	remasterQANames = append(remasterQANames, "motion-drive", "motion-cockpit", "image-settings")
	remasterQANames = append(remasterQANames, "hill-profile-day", "hill-profile-night", "acceleration-steering")
	remasterQANames = append(remasterQANames, "grass-wind-off", "grass-wind-phase-a", "grass-wind-phase-b", "npc-rounded-turn", "hill-wheel-support", "polished-car-metal", "sloped-road-paint", "drift-recovery-026")
	remasterQANames = append(remasterQANames, "remastered-hill", "remastered-grass", "remastered-central-tunnel", "remastered-city", "remastered-audio-settings", "remastered-night-reflections", "remastered-tunnel-interior")
	remasterQANames = append(remasterQANames, "remastered-bridge", "remastered-wooded-hills", "remastered-street-blocks", "remastered-water")
	remasterQANames = append(remasterQANames, "031-sidewalk-walkers", "031-pedestrian-impact", "031-npc-impact", "031-brakes-camera", "031-slope-materials", "031-tunnel-south-seal", "031-pavement-clearance")
	remasterQANames = append(remasterQANames, "032-supplied-character-skins", "033-waypoint-map", "033-heading-minimap", "033-north-minimap", "033-braking-map-zoom", "033-mode-menu", "033-controller-settings", "033-map-settings", "033-sunset", "033-fpp-motion")
	remasterQANames = append(remasterQANames, "034-map-detail-zoom", "034-snow-peak")
}

var remasterQAPath, remasterQAHint int
var remasterQAGPU, remasterQACPU []float64
var remasterQASequence []int
var remasterQAMeasureStart time.Time
var remasterCPUProfile *os.File

func remasterQAMenuChecks() {
	qa.checks["raw-mouse-device-registered"] = input020.Raw
	selectScreen(screenTitle)
	beta020MenuUI()
	beta020Click(1, 1)
	qa.checks["title-background-keeps-menu"] = beta.screen == screenTitle
	// Click the real title button at its rendered centre after layout changes.
	for _, button := range ui020.Buttons {
		if button.ID == "start" {
			beta020Click(int(button.X+button.W*.5), int(button.Y+button.H*.5))
			break
		}
	}
	qa.checks["title-start-click"] = beta.screen == screenMaps && ui020.Stage == 0
	beta.carSelect = 0
	ui020.Stage = 0
	remaster.Mode = 0
	beta020Action("mode1")
	qa.checks["mode-card-selects-racing"] = remaster.Mode == 1
	beta020Action("mode0")
	qa.checks["mode-card-selects-free-roam"] = remaster.Mode == 0
	beta020Advance()
	qa.checks["next-shows-map-stage"] = ui020.Stage == 1 && !app.started
	beta020Advance()
	qa.checks["next-shows-car-stage"] = ui020.Stage == 2 && !app.started
	beta020Action("car+")
	qa.checks["car-selection-changes"] = beta.carSelect == 1
	for car := 0; car < 3; car++ {
		beta.carSelect, beta.paint = car, 0
		ui020.Stage = 2
		beta020MenuUI()
		qa.checks[fmt.Sprintf("036-direct-palette-car%d", car)] = paintQA036(len(paintPalette) - 1)
	}

	beta.carSelect, beta.paint = 0, 0
	beta020MenuUI()
	paintQA036(2)
	qa.checks["033-colour-swatch-mouse-selects"] = beta.paint == 2
	beta.paint = 0
	beta.carSelect = 3
	before := beta.paint
	beta020Action("paint+")
	qa.checks["fixed-fourth-car-paint"] = beta.paint == before
	beta020Action("settings")
	qa.checks["gear-opens-settings"] = beta.screen == screenSettings
	remasterKey(VK_ESCAPE)
	qa.checks["settings-back-to-setup"] = beta.screen == screenMaps
	beta020Action("night")
	qa.checks["time-choice-night"] = remaster.Time == 2
	beta020Action("day")
	beta.carSelect = 0
	ui020.Stage = 0
	remaster.Row = 0
	selectScreen(screenDrive)
	old := beta.cfg.CameraMode
	beta.cfg.CameraMode = 1
	for _, expected := range []int{2, 3, 0, 1} {
		remasterKey('V')
		qa.checks["camera-cycle-"+fmt.Sprint(expected)] = beta.cfg.CameraMode == expected
	}
	beta.cfg.CameraMode = old
	app.boolFPP = false
	app.camYaw = .2
	beta020Look(20, 0)
	qa.checks["relative-look-does-not-steer"] = app.camYaw != .2 && app.mouseSteer == 0
	app.started = false
	app.camYaw = 0
	selectScreen(screenMaps)
}
func initRemasterQA() bool {
	registerQA036()
	registerQA037()
	registerQA040()
	registerQAProfiles040()
	registerQAWorldFinal040()
	registerQA041()
	registerQA050()
	registerQA051()
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "--verify-remaster=") {
			qa.enabled = true
			remaster.QA = true
			qa.dir = strings.TrimPrefix(a, "--verify-remaster=")
			qa.checks = map[string]bool{}
			qa.limit = 45
			if n, e := strconv.Atoi(os.Getenv("SPEED_QA_SINGLE_STAGE")); e == nil {
				qa.stage = engine.ClampInt(n, 0, len(remasterQANames)-1)
			}
			for _, part := range strings.Split(os.Getenv("SPEED_QA_SEQUENCE"), ",") {
				if n, e := strconv.Atoi(strings.TrimSpace(part)); e == nil && n >= 0 && n < len(remasterQANames) {
					remasterQASequence = append(remasterQASequence, n)
				}
			}
			if len(remasterQASequence) > 0 {
				qa.stage = remasterQASequence[0]
			}
			os.MkdirAll(qa.dir, 0755)
			return true
		}
	}
	return false
}
func remasterQAStage() {
	if workshop051.Active {
		exitWorkshop051()
	}
	photo041.Active, photo041.Drag = false, false
	// Each rendered fixture starts from the protected driver. The locked-driver
	// fixture creates its own selection later; it must not leak into the next.
	if profiles040.Store != nil {
		profiles040.Store.Select040("ishaan")
		profiles040.Page = 0
		progress040.Gate = ""
		progress040.GateUntil = time.Time{}
	}
	frozenPresentation038 = nil
	presentationBenchmark038.Start = time.Time{}
	adaptive026.Reset()
	initGLDebug023()
	remasterQACheckpoint023("loading")
	defer func() { remasterQACheckpoint023("ready-to-render") }()
	defer func() {
		if beta.screen == screenMaps {
			if qa.stage >= 219 && qa.stage <= 221 {
				return
			}
			if qa.stage >= 1 && qa.stage <= 4 || qa.stage == 36 || qa.stage == 141 || qa.stage == 142 || qa.stage >= 148 && qa.stage < 184 || qa.stage == 189 {
				ui020.Stage = 2
			} else if qa.stage == 23 {
				ui020.Stage = 4
			} else {
				ui020.Stage = 0
			}
			if n, e := strconv.Atoi(os.Getenv("SPEED_QA_UI_STAGE")); e == nil && n >= 0 && n <= 4 {
				ui020.Stage = n
			}
			if qa.stage == 130 || qa.stage == 131 || qa.stage == 184 {
				ui020.Stage = 1
			}
		}
	}()
	qa.limit = 45
	remasterQAGPU, remasterQACPU = nil, nil
	remasterQAMeasureStart = time.Time{}
	defer func() {
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_PRESET")); e == nil && n >= 0 && n <= 3 {
			applyAokawaPreset(n)
		}
		applyQATextureBudget024()
		beta.cfg.Lightweight = os.Getenv("SPEED_QA_LIGHTWEIGHT") == "1"
		if qaAdaptive026 {
			beta.cfg.FrameCap = 60
		}
		if qa.stage == 86 {
			beta.cfg.Wind = false
		}
		if qa.stage == 80 || qa.stage == 81 {
			beta.cfg.MotionBlur = true
			beta.cfg.BlurStrength = .65
		}
		if qa.stage == 82 {
			beta.cfg.Exposure = 1.1
			beta.cfg.Saturation = .95
		}
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_FRAMES")); e == nil && n >= 30 {
			qa.limit = engine.ClampInt(n, 30, 180000)
		}
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_SECONDS")); e == nil && n >= 1 && n <= 120 {
			qa.limit = 180000
		}
		if os.Getenv("SPEED_QA_RTX3060") == "1" {
			applyRTX3060Profile034()
		}
		if os.Getenv("SPEED_QA_CINEMATIC") == "1" {
			applyCinematicShaderProfile034()
		}
		switch os.Getenv("SPEED_QA_035_PROFILE") {
		case "high":
			applyHigh035()
		case "ultra":
			applyUltra035()
		case "integrated":
			applyAokawaPreset(0)
			beta.cfg.Lightweight = true
			beta.cfg.RenderScale = .6
			beta.cfg.VSync = false
			beta.cfg.FrameCap = 0
			beta.cfg.TrafficDensity = 1
		}
		if qa.stage == 144 {
			beta.cfg.RenderScale = .75
		}
		if qa.stage == 93 {
			qa.limit = 240
		}
		// Scenario values override the selected benchmark graphics profile.
		if qa.stage == 86 {
			beta.cfg.Wind = false
		}
		if qa.stage == 82 {
			beta.cfg.Exposure = 1.1
			beta.cfg.Saturation = .95
		}
		if qa.stage == 108 && qa.limit < 300 {
			qa.limit = 300
		}
		if qa.stage == 116 && qa.limit < 180 {
			qa.limit = 180
		}
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_FRAME_CAP")); e == nil && n >= 0 && n <= 240 {
			beta.cfg.FrameCap = n
		}
		applyQASkyDetail040()
	}()
	qa.frame = 0
	qa.samples = nil
	qa.last = time.Now()
	qa.stageStart = qa.last
	beta.cfg.Headlights = true
	beta.cfg.UIScale = 1
	app.boolFPP = false
	beta.cfg.CameraMode = 1
	remaster.FullMap = false
	app.started = false
	remaster.SceneReady = false
	initialPreset := 1
	if n, e := strconv.Atoi(os.Getenv("SPEED_QA_PRESET")); e == nil && n >= 0 && n <= 3 {
		initialPreset = n
	}
	applyAokawaPreset(initialPreset)
	applyQATextureBudget024()
	beta.cfg.Lightweight = os.Getenv("SPEED_QA_LIGHTWEIGHT") == "1"
	app.paused = false
	remaster.Mode = 0
	remaster.Time = 0
	beta.carSelect = 0
	beta.paint = 0
	if qa.stage >= 207 {
		stageQA051()
		return
	}
	if qa.stage == 0 {
		selectScreen(screenTitle)
		return
	}
	if qa.stage < 5 {
		beta.carSelect = qa.stage - 1
		selectScreen(screenMaps)
		if qa.stage == 1 {
			remasterQAMenuChecks()
		}
		return
	}
	if qa.stage == 23 {
		remaster.Mode = 1
		selectScreen(screenMaps)
		return
	}
	if qa.stage >= 24 && qa.stage <= 36 {
		remaster.Mode = 1
	}
	if qa.stage >= 28 && qa.stage <= 33 {
		remaster.Mode = 0
	}
	remaster.Opponents = 4
	if qa.stage == 35 {
		remaster.Opponents = 7
	}
	if qa.stage >= 5 && qa.stage <= 8 {
		beta.carSelect = qa.stage - 5
		remaster.Time = (qa.stage - 5) % 3
	}
	if qa.stage >= 9 && qa.stage <= 12 {
		beta.carSelect = qa.stage - 9
		app.boolFPP = true
	}
	if qa.stage == 7 || qa.stage >= 17 && qa.stage <= 19 {
		remaster.Time = 2
	}
	if qa.stage == 34 {
		remaster.Time = 2
	}
	if qa.stage == 84 {
		remaster.Time = 2
	}
	if qa.stage >= 56 && qa.stage < 80 {
		q := qa.stage - 56
		remaster.Mode = q / 12
		beta.carSelect = q % 12 / 3
		remaster.Time = q % 3
	}
	if qa.stage == 91 {
		beta.carSelect = 1
	}
	startRemasteredDrive()
	qa.limit = 45
	w := app.world
	qa.checks[fmt.Sprintf("stage-%d-correct-mode", qa.stage)] = w.Remaster != nil && w.Map == remaster.Mode && w.CarClass == beta.carSelect
	if qa.stage == 13 || qa.stage == 14 || qa.stage == 41 {
		t := w.Remaster.Def.Tunnels[0]
		f := engine.HeadingVec(t.Heading)
		p := t.Center
		if qa.stage == 13 {
			p = p.Sub(f.Mul(t.Length/2 + 12))
		}
		if qa.stage == 41 {
			p = p.Add(f.Mul(t.Length/2 + 12))
		}
		angle := t.Heading
		if qa.stage == 41 {
			angle += math.Pi
		}
		remasterQAPlace(p, angle)
		app.boolFPP = qa.stage == 14
	}
	if qa.stage == 15 {
		for _, road := range w.Routes {
			for dist := 0.; dist < road.Len; dist += 8 {
				p, f := road.AtDistance(dist)
				if _, ok := w.RemasteredWaterLevel(p); ok && w.GroundHeight(p) > 18 {
					remasterQAPlace(p, math.Atan2(f.X, f.Z))
					goto bridgePlaced
				}
			}
		}
	bridgePlaced:
	}
	if qa.stage == 16 {
		placed := false
		for ri, road := range w.Routes {
			if w.Remaster.Def.Roads[ri].Kind == "town" || w.Remaster.Def.Roads[ri].Kind == "stunt" {
				continue
			}
			for at := 0.; at < road.Len; at += 8 {
				p, f := road.AtDistance(at)
				if p.X > 950 && p.Z > -200 && p.Z < 600 {
					remasterQAPlace(p, math.Atan2(f.X, f.Z))
					placed = true
					break
				}
			}
			if placed {
				break
			}
		}
		qa.checks["forest-view-inside-east-biome"] = placed
	}
	if qa.stage == 17 {
		remasterKey('H')
		qa.checks["H-switch-toggles-headlights"] = !beta.cfg.Headlights
	}
	if qa.stage == 19 {
		w.Car.Speed = -3
		w.Car.Gearbox.Gear = -1
	}
	if qa.stage == 20 {
		w.Car.Speed = 22
		w.Car.Lateral = 5
		w.Car.DriftActive = true
		w.Car.DriftRecovery = 1
		qa.limit = 75
	}
	if qa.stage == 21 {
		selectScreen(screenPause)
	}
	if qa.stage == 22 {
		beta.prevScreen = screenMaps
		selectScreen(screenSettings)
		beta.cfg.UIScale = 1.5
	}
	if qa.stage == 25 || qa.stage == 26 {
		w.Remaster.Countdown = 0
		w.Remaster.Started = true
		w.Car.Speed = 18
		qa.limit = 90
		app.boolFPP = qa.stage == 26
		if os.Getenv("SPEED_QA_RACE_OBSTACLE") != "" {
			remasterQAPlace(engine.Vec{X: 427, Z: 452.5}, 2.35)
			w.Remaster.Racers = nil
			w.Car.Speed = 0
			qa.limit = 20
		}
	}
	if qa.stage == 27 {
		w.Remaster.Finished = true
		w.Remaster.FinishTime = 176.42
		w.Remaster.BestLap = 88.1
		selectScreen(screenResults)
	}
	if qa.stage == 28 {
		remasterQAPlace(engine.Vec{X: 1000, Z: 710}, 0)
	}
	if qa.stage == 29 {
		remasterQAPlace(engine.Vec{X: 1000, Z: 790}, 0)
		w.Car.Body.Height += 8
		w.Car.Body.Grounded = false
		w.Car.Body.VerticalSpeed = 2
		w.Car.Body.AirVelocity = engine.Vec{X: 0, Z: 30}
		w.Car.Speed = 30
		w.Car.Rig.Pitch = -.2
	}
	if qa.stage == 30 {
		sig := w.Remaster.Def.Signals[0]
		f := engine.HeadingVec(sig.Heading)
		remasterQAPlace(sig.Pos.Sub(f.Mul(13)), sig.Heading)
	}
	if qa.stage == 31 && len(remaster.Lamps) > 0 {
		id := remaster.Lamps[0]
		o := w.Remaster.Def.Objects[id]
		p := &engine.Pole{Pos: engine.Vec{X: o.Pos[0], Z: o.Pos[2]}, Heading: o.Heading}
		p.Hit(engine.Vec{X: 18, Z: 7})
		w.Remaster.FallenPoles[id] = p
		remasterQAPlace(p.Pos.Sub(engine.Vec{X: 5, Z: 9}), .4)
		qa.limit = 300
	}
	if qa.stage == 32 {
		placed := false
		for ri, road := range w.Routes {
			if w.Remaster.Def.Roads[ri].Kind == "tunnel" || w.Remaster.Def.Roads[ri].Kind == "stunt" {
				continue
			}
			for at := 0.; at < road.Len; at += 8 {
				p, f := road.AtDistance(at)
				if p.Z < -950 && math.Abs(p.X) < 800 {
					remasterQAPlace(p, math.Atan2(f.X, f.Z))
					placed = true
					break
				}
			}
			if placed {
				break
			}
		}
		qa.checks["hill-view-inside-north-biome"] = placed
	}
	if qa.stage == 33 {
		for _, o := range w.Remaster.Def.Objects {
			if o.Asset == "b022_fern" && o.Pos[0] > 950 && o.Pos[0] < 1350 {
				remasterQAPlace(engine.Vec{X: o.Pos[0], Z: o.Pos[2] + 5}, math.Pi)
				break
			}
		}
	}
	if qa.stage == 36 || qa.stage == 141 || qa.stage == 142 {
		app.started = false
		beta.carSelect = 2
		beta.cfg.UIScale = 1.5
		selectScreen(screenMaps)
	}
	if qa.stage == 37 {
		remasterKey('M')
		qa.checks["M-opens-full-map"] = remaster.FullMap
	}
	if qa.stage == 38 {
		remaster.SettingsTab = 1
		beta.prevScreen = screenPause
		selectScreen(screenSettings)
	}
	if qa.stage == 39 {
		beta.cfg.CameraMode = 2
	}
	if qa.stage == 40 {
		beta.cfg.CameraMode = 3
	}
	if qa.stage == 42 {
		w.Car.Speed = 30
		w.Car.Boosting = true
		qa.limit = 60
	}
	if qa.stage == 43 {
		sig := w.Remaster.Def.Signals[0]
		f := engine.HeadingVec(sig.Heading)
		remasterQAPlace(sig.Pos.Sub(f.Mul(12)), sig.Heading)
	}
	if qa.stage == 44 {
		applyAokawaPreset(0)
	}
	if qa.stage == 45 {
		applyAokawaPreset(3)
	}
	qa.checks["race-civilian-traffic-absent"] = remaster.Mode != 1 || len(w.Traffic) == 0
	_, _, remasterQAPath = w.RemasteredRoadAt(w.Car.Pos)
	remasterQAPath = engine.ClampInt(remasterQAPath, 0, len(w.Routes)-1)
	remasterQAHint = -1
	app.prevCar = w.Car
	app.renderCar = w.Car
	app.renderAlpha = 1
	remaster.CameraReset = true
	if qa.stage == 46 || qa.stage == 47 {
		remasterQAPlace(w.Remaster.Def.BeachSpawn, -math.Pi/2)
	}

	if qa.stage == 48 {
		w.Car.Speed = 25
		qa.limit = 90
	}
	if qa.stage == 49 {
		w.SetTrafficDensity(3)
		qa.checks["busy-density-is-bounded"] = len(w.Traffic) == 1056
	}
	if qa.stage == 50 {
		remaster.FullMap = false
		remasterKey('M')
	}
	if qa.stage == 52 {
		w.Traffic = nil
		p, f := w.Routes[0].AtDistance(w.Routes[0].Len * .25)
		heading := math.Atan2(f.X, f.Z)
		remasterQAPlace(p, heading)
		right := engine.Vec{X: -f.Z, Z: f.X}
		w.Traffic = []engine.Traffic{{Pos: p.Add(right.Mul(3.3)).Add(f.Mul(3)), Heading: heading, Variant: 0, Path: 0, Direction: 1, T: w.Routes[0].Len * .25, Lane: 3.3}}
		qa.checks["bus-real-world-body-dimensions"] = engine.TrafficBodyFootprint(0).Z > 5 && engine.TrafficBodyMass(0) > 10000
	}

	if qa.stage == 53 {
		w.Remaster.Time = 2
		remaster.Time = 2
		base := w.Routes[0].Len * .25
		p, f := w.Routes[0].AtDistance(base)
		h := math.Atan2(f.X, f.Z)
		remasterQAPlace(p, h)
		w.Traffic = nil
		for i, v := range []int{4, 5, 7, 10, 13} {
			at := base + 10 + float64(i/3)*6
			pos, tangent := w.Routes[0].AtDistance(at)
			r := engine.Vec{X: -tangent.Z, Z: tangent.X}
			side := float64(i%3-1) * 3.2
			w.Traffic = append(w.Traffic, engine.Traffic{Pos: pos.Add(r.Mul(side)), Heading: math.Atan2(tangent.X, tangent.Z) + math.Pi, Variant: v, Path: 0, Direction: -1, T: at, Lane: -side})
		}
	}

	if qa.stage == 54 || qa.stage == 55 {
		for i, o := range w.Remaster.Def.Objects {
			want := o.Kind == "signal"
			if qa.stage == 55 {
				want = o.Kind == "prop" && o.Collider >= 0
			}
			if !want {
				continue
			}
			p := engine.Vec{X: o.Pos[0], Z: o.Pos[2]}
			remasterQAPlace(p.Sub(engine.Vec{Z: 6}), 0)
			shape := engine.ColliderShape{MinY: o.Pos[1], MaxY: o.Pos[1] + 8, Half: engine.Vec{X: .2, Z: .2}}
			if o.Collider >= 0 && o.Collider < len(w.Remaster.Def.Colliders) {
				shape = w.Remaster.Def.Colliders[o.Collider]
			}
			if qa.stage == 54 {
				w.Remaster.FallenPoles[i] = &engine.Pole{Pos: p, FallDirection: engine.Vec{X: 1}, Falling: true, Settled: true, Angle: math.Pi / 2, Length: shape.MaxY - shape.MinY}
			}
			if qa.stage == 55 {
				w.Remaster.DynamicProps[i] = &engine.DynamicProp{Pos: p.Add(engine.Vec{X: 1.5}), Height: shape.MinY, Heading: o.Heading, Half: shape.Half, Mass: 65, Roll: .8}
			}
			qa.checks[fmt.Sprintf("stage-%d-fixture-transform", qa.stage)] = true
			break
		}
	}

	if qa.stage == 51 {
		w.Remaster.Time = 2
		remaster.Time = 2
		releaseRemasterSky()
		qa.checks["moon-model-loaded"] = pbrAssets["b011_moon"] != nil
	}

	if qa.stage == 80 || qa.stage == 81 {
		// A speed/effect check needs an unobstructed trajectory. The previous
		// city-start fixture ran into the newly populated streetfronts.
		t := w.Remaster.Def.Tunnels[0]
		remasterQAPlace(t.Center.Sub(engine.HeadingVec(t.Heading).Mul(200)), t.Heading)
		w.Traffic = nil
		w.Car.Speed = 45
		app.boolFPP = qa.stage == 81
	}
	if qa.stage == 82 {
		remaster.SettingsTab = 3
		selectScreen(screenSettings)
	}
	if qa.stage == 83 || qa.stage == 84 {
		t := w.Remaster.Def.Tunnels[0]
		f := engine.HeadingVec(t.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		p := t.Center.Sub(f.Mul(t.Length*.5 + 28)).Sub(r.Mul(65))
		delta := t.Center.Sub(p)
		remasterQAPlace(p, math.Atan2(delta.X, delta.Z))
		qa.checks["tunnel-has-physical-portals-and-aprons"] = t.PortalApron >= 10
	}
	if qa.stage == 85 {
		w.Car.Speed = 25
	}
	beta026QAStage()
	beta030QAStage()
	beta031QAStage()
	beta033QAStage()
	stageQA036()
	stageQA037()
	stageQA040()
	stageQAProfiles040()
	stageQAWorldFinal040()
	stageQA041()
	stageQA050()
}
func remasterQAPlace(p engine.Vec, heading float64) {
	w := app.world
	w.Car.Pos = p
	w.Car.Heading = heading
	w.Car.Body = engine.ChassisState{}
	w.Car.Rig = engine.SuspensionRig{}
	w.InitializeChassis()
	w.Car.Speed = 0
	app.prevCar = w.Car
}
func remasterQAInput() engine.Input {
	if qa.stage == 245 {
		return engine.Input{Throttle: 1}
	}
	if qa.stage == 246 {
		return engine.Input{Brake: 1}
	}
	// Braking holds a car still, but becomes backwards movement on foot.
	// Keep the standing fixture genuinely stationary; exercise walking separately.
	if qa.stage == 232 {
		return engine.Input{}
	}
	if qa.stage == 233 {
		return engine.Input{Throttle: 1}
	}
	if qaWorldFinalActive040() {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage == 121 {
		return engine.Input{Throttle: .5}
	}
	if qa.stage == 116 {
		return engine.Input{Brake: .65}
	}
	if qa.stage >= 105 {
		return beta031QAInput()
	}
	if qa.stage == 80 || qa.stage == 81 {
		return engine.Input{Throttle: .5}
	}
	if qa.stage >= 94 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage >= 86 && qa.stage <= 92 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage == 93 {
		if app.world.Elapsed < .7 {
			return engine.Input{Throttle: .65, Steer: .65, Handbrake: true}
		}
		return engine.Input{Throttle: .5}
	}
	if qa.stage == 83 || qa.stage == 84 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage == 85 {
		return engine.Input{Throttle: 1, Steer: .65}
	}
	if qa.stage == 48 {
		return engine.Input{Handbrake: true, Steer: .32}
	}
	if qa.stage == 42 {
		return engine.Input{Throttle: .8, Boost: true}
	}
	if qa.stage >= 37 && qa.stage < 56 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage == 29 {
		return engine.Input{Boost: true}
	}
	if qa.stage >= 28 && qa.stage <= 35 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	if qa.stage == 19 {
		return engine.Input{Brake: .3}
	}
	if qa.stage == 20 {
		return engine.Input{Throttle: .6, Handbrake: true, Steer: .7}
	}
	if qa.stage >= 13 && qa.stage <= 22 || qa.stage == 24 {
		return engine.Input{Brake: 1, Handbrake: true}
	}
	w := app.world
	if w.Remaster == nil {
		return engine.Input{}
	}
	route := w.Routes[remasterQAPath]
	distance, hint := route.ClosestDistance(w.Car.Pos, remasterQAHint)
	remasterQAHint = hint
	p, _ := route.AtDistance(distance + 14 + math.Abs(w.Car.Speed)*.4)
	d := p.Sub(w.Car.Pos)
	a := math.Atan2(d.X, d.Z) - w.Car.Heading
	a = math.Atan2(math.Sin(a), math.Cos(a))
	in := engine.Input{Throttle: .5, Steer: engine.Clamp(a*2.5, -1, 1)}
	if w.Car.Speed > 22 {
		in.Throttle = 0
		in.Brake = .2
	}
	return in
}
func remasterQAAfterFrame() {
	if !qa.started {
		return
	}
	observeDriftQA026()
	beta031QAObserve()
	beta033QAObserve()
	now := time.Now()
	ms := now.Sub(qa.last).Seconds() * 1000
	qa.last = now
	qa.frame++
	if qa.stage == 93 && app.world.Elapsed < 3.2 && qa.frame >= qa.limit {
		qa.limit = qa.frame + 1
	}
	warmup := 12
	if qa.limit > 300 {
		warmup = 120
	}
	if qa.frame == warmup {
		remasterQAMeasureStart = now
		if qa.limit > 300 {
			remasterQACheckpoint023("measuring")
		}
		if remasterCPUProfile == nil && os.Getenv("SPEED_QA_CPU_PROFILE") == "1" {
			if f, err := os.Create(filepath.Join(qa.dir, "render-cpu.pprof")); err == nil {
				if err := pprof.StartCPUProfile(f); err == nil {
					remasterCPUProfile = f
				} else {
					f.Close()
				}
			}
		}
		beginPresentationBenchmark038()
	}
	if qa.frame > warmup {
		qa.samples = append(qa.samples, ms)
		remasterQAGPU = append(remasterQAGPU, app.gpuMS)
		remasterQACPU = append(remasterQACPU, app.renderMS)
	}
	if qa.limit > 300 && qa.frame%600 == 0 {
		remasterQACheckpoint023("measuring")
	}
	if seconds, err := strconv.Atoi(os.Getenv("SPEED_QA_SECONDS")); err == nil && seconds >= 1 && seconds <= 120 && !remasterQAMeasureStart.IsZero() && now.Sub(remasterQAMeasureStart).Seconds() >= float64(seconds) {
		qa.limit = qa.frame
	}
	// This is a bounded handbrake/recovery maneuver. Extending a global FPS
	// measurement must not keep driving the cleared fixture into distant scenery.
	if qa.stage == 93 && app.world.Elapsed >= 4.0 {
		qa.limit = qa.frame
	}
	if qa.stage >= 56 && qa.stage < 80 {
		qa.checks[fmt.Sprintf("matrix-stage-%d-car-and-time", qa.stage)] = app.world.CarClass == beta.carSelect && app.world.Remaster.Time == remaster.Time
	}
	// Timed runs replace qa.limit on their final frame. Exercise photo capture
	// during warmup so it completes before either frame- or time-based limits.
	if qa.stage == 9 && qa.frame == warmup-2 {
		photo020Pending = true
	}
	if qa.stage == 9 && qa.frame == qa.limit {
		qa.checks["rendered-photo-save"] = strings.Contains(remaster.Status, "PHOTO SAVED")
	}
	if qa.frame < qa.limit {
		return
	}
	if waitQASkyDetail040() {
		return
	}
	endPresentationBenchmark038()
	if qa.stage == 48 {
		qa.checks["rear-braking-produces-road-marks"] = len(beta011Skids.Marks) > 0
	}
	if qa.stage == 50 {
		qa.checks["map-zoom-finishes-while-world-paused"] = beta011MapProgress > .98 && app.world.Elapsed == 0
	}
	if qa.stage == 47 {
		qa.checks["water-clock-advances"] = app.world.Elapsed > 0 && app.world.Remaster.Def.WaterGrid.W > 1
	}
	if qa.stage == 46 {
		qa.checks["beach-camera-outside-car"] = math.Hypot(app.cam.x-app.renderCar.Pos.X, app.cam.z-app.renderCar.Pos.Z) > 2.5
	}

	if qa.stage == 13 {
		qa.checks["portal-car-above-visible-terrain"] = app.world.Car.Body.Height >= app.world.Remaster.Def.Grid.Height(app.world.Car.Pos)+.69
	}
	if qa.stage == 21 {
		qa.checks["pause-freezes-world-time"] = app.world.Elapsed == 0
	}
	if qa.stage == 37 {
		qa.checks["map-freezes-world-time"] = app.world.Elapsed == 0
	}
	if qa.stage == 42 {
		qa.checks["nitro-camera-distance-bounded"] = math.Hypot(app.cam.x-app.renderCar.Pos.X, app.cam.z-app.renderCar.Pos.Z) <= aokawaCameraDistance()+.6
	}
	if qa.stage == 80 || qa.stage == 81 {
		qa.checks["motion-control-reaches-post-shader"] = postValue024("uMotion") > .001
	}
	if qa.stage == 82 {
		qa.checks["exposure-control-reaches-post-shader"] = math.Abs(float64(postValue024("uExposure"))-1.1) < .001
		qa.checks["settings-disable-speed-blur"] = postValue024("uMotion") == 0
	}
	beta026QAChecks()
	beta033QAChecks()
	checks035()
	if beta.screen == screenDrive && !remaster.FullMap && app.quality > 0 && !reducedEffects026() {
		qa.checks["034-cinematic-targets-valid"] = !cinematic034.Failed
		if beta.cfg.Bloom && beta.cfg.BloomStrength > .001 {
			qa.checks["034-four-level-hdr-bloom"] = cinematic034.BloomActive && postValue024("uBloomEnabled") == 1 && cinematic034.Bloom[3][0].FBO != 0
		}
		if beta.cfg.SSAO {
			qa.checks["034-half-resolution-depth-effects"] = cinematic034.EffectsActive && cinematic034.Effects.Width == postSceneWidth038()/2 && cinematic034.Effects.Height == postSceneHeight038()/2 && postValue024("uEffectsEnabled") == 1
		}
	}
	gl("glFinish")
	name := remasterQANames[qa.stage]
	saveFramebuffer(filepath.Join(qa.dir, name+".png"))
	captureQASkyProof040(name)
	captureMotion038(name)
	captureRayScene038(name)
	if os.Getenv("SPEED_QA_SECONDS") != "0" && os.Getenv("SPEED_QA_SECONDS") != "" && reconstructedReady038 {
		captureTemporalProof038(name, nil)
	}
	errors := []uint32{}
	for i := 0; i < 16; i++ {
		e := uint32(gl("glGetError"))
		if e == 0 {
			break
		}
		errors = append(errors, e)
	}
	gpu, cpu := summarizeFrames(remasterQAGPU), summarizeFrames(remasterQACPU)
	qa.results = append(qa.results, qaResult{ShaderPipeline034: cinematicReport034(), Optimization026: optimizationReport026(), TextureBudget: beta.cfg.TextureBudget, GrassTextureWidth: grassTextureWidth024(), PostExposure: postValue024("uExposure"), PostMotion: postValue024("uMotion"), TrafficCount: len(app.world.Traffic), Memory: memory023(), Quality: app.quality, RenderScale: beta.cfg.RenderScale, Name: name, Map: remaster.Mode, Car: beta.carSelect, Frames: qa.frame, FrameStatistics: summarizeFrames(qa.samples), GPUTimeMS: app.gpuMS, GPUMedianMS: gpu.MedianMS, GPUP95MS: gpu.P95MS, SubmissionMedianMS: cpu.MedianMS, HDR: hdrScene.fbo != 0 && !hdrScene.failed, WorldVBO: gpuWorld.ready, Instancing: gpuWorld.instancing, SunShadows: sunShadow.ready, GLErrors: errors})
	remasterQACheckpoint023("completed-stage")
	qa.stage++
	if len(qa.results) < len(remasterQASequence) {
		qa.stage = remasterQASequence[len(qa.results)]
		qa.started = false
		return
	}
	if len(remasterQASequence) == 0 && qa.stage < len(remasterQANames) && os.Getenv("SPEED_QA_SINGLE_STAGE") == "" {
		qa.started = false
		return
	}
	report := map[string]any{"brake_zoom_peak_metres": qa033PeakSpan, "sky_dimensions": sky020.Sizes, "controller_connected": pad033.State.Connected, "version": "Beta v0.5.3", "mode": "Automated hidden WGL renderer verification. Not an interactive playtest or certified hardware benchmark.", "renderer": glString(0x1F01), "vendor": glString(0x1F00), "opengl": glString(0x1F02), "width": app.width, "height": app.height, "quality": app.quality, "vsync": app.vsync, "gameplay_cap_fps": beta.cfg.FrameCap, "shaderError": lightPipeline.errorText, "assetErrors": pbrErrors, "optimization026": optimizationReport026(), "checks": qa.checks, "stages": qa.results, "utc": time.Now().UTC().Format(time.RFC3339)}
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(filepath.Join(qa.dir, "RENDER_VERIFICATION_BETA_002.json"), raw, 0644)
	if remasterCPUProfile != nil {
		pprof.StopCPUProfile()
		remasterCPUProfile.Close()
		remasterCPUProfile = nil
	}
	app.run = false
}

// Keep the last completed results if a renderer run ends during a transition.
// This checkpoint is deliberately separate from the final acceptance report.
func remasterQACheckpoint023(state string) {
	progress := map[string]any{"state": state, "stage": qa.stage, "frame": qa.frame, "utc": time.Now().UTC().Format(time.RFC3339), "memory": memory023(), "checks": qa.checks, "completed": qa.results}
	if raw, err := json.MarshalIndent(progress, "", "  "); err == nil {
		os.WriteFile(filepath.Join(qa.dir, "QA_PROGRESS_BETA_023.json"), raw, 0644)
	}
}
