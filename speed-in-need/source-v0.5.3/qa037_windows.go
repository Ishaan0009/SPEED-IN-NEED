package main

import (
	"math"
	"speedinneed/engine"
)

func registerQA037() {
	if len(remasterQANames) > 136 {
		return
	}
	remasterQANames = append(remasterQANames, "037-tunnel-south-road", "037-tunnel-north-road", "037-gt-grip-turn", "037-pine-crowns", "037-fallen-wood", "037-garage-orbit-near", "037-garage-orbit-far", "037-fullframe-effects", "037-scaled-effects", "037-extended-zoom")
	remasterQANames = append(remasterQANames, "038-graphics-menu", "038-nvidia-runtime-gate")
}
func stageQA037() {
	if qa.stage == 146 || qa.stage == 147 {
		remaster.SettingsTab = 8
		beta.cfg.RendererBackend = 1
		beta.cfg.Upscaling038 = 1
		beta.cfg.FrameGeneration038 = 1
		beta.cfg.RayShadows038 = 1
		beta.selection = 3
		if qa.stage == 147 {
			beta.selection = 5
		}
		selectScreen(screenSettings)
		return
	}
	if qa.stage < 136 {
		return
	}
	w := app.world
	switch qa.stage {
	case 136, 137:
		t := w.Remaster.Def.Tunnels[0]
		sign := -1.
		if qa.stage == 137 {
			sign = 1
		}
		f := engine.HeadingVec(t.Heading)
		remasterQAPlace(t.Center.Add(f.Mul(sign*(t.Length*.5+24))), t.Heading+math.Pi*float64(boolInt(sign > 0)))
	case 138:
		beta.carSelect = 2
		activateRemasteredWorld(true)
		w = app.world
		remasterQAPlace(w.Remaster.Def.Spawn, w.Remaster.Def.SpawnHeading)
		w.Car.VisualSteer = .65
	case 139, 140:
		for _, o := range w.Remaster.Def.Objects {
			want := o.Asset == "b038_b037_pine" && o.Pos[1] > 60 && o.Pos[1] < 140
			if qa.stage == 140 {
				want = o.Asset == "b035_fallen_branch" && o.Scale[0] > 3
			}
			if want {
				remasterQAPlace(engine.Vec{X: o.Pos[0] + 6, Z: o.Pos[2] + 8}, math.Pi)
				break
			}
		}
	case 141, 142:
		selectScreen(screenMaps)
		ui020.Stage = 2
		app.started = false
		garage037.Touched = true
		garage037.Yaw = 1.25
		garage037.Pitch = .12
		garage037.Zoom = .82
		if qa.stage == 142 {
			garage037.Yaw = 3.0
			garage037.Zoom = 1.30
		}
	case 143, 144:
		w.Car.Speed = 30
		beta.cfg.MotionBlur = true
		beta.cfg.BlurStrength = .75
		beta.cfg.Bloom = true
		beta.cfg.SSAO = true
		if qa.stage == 144 && temporalMode038() == 0 {
			beta.cfg.RenderScale = .75
		}
	case 145:
		app.camDist = 17
	}
}
func viewQA037(c engine.Car, target *pos3) {
	if !qa.enabled {
		return
	}
	if qa.stage == 136 || qa.stage == 137 {
		t := app.world.Remaster.Def.Tunnels[0]
		sign := -1.
		if qa.stage == 137 {
			sign = 1
		}
		f := engine.HeadingVec(t.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		p := t.Center.Add(f.Mul(sign * (t.Length*.5 + 38))).Add(r.Mul(25))
		app.cam = pos3{p.X, c.Body.Height + 10, p.Z}
		p = t.Center.Add(f.Mul(sign * (t.Length*.5 + 5)))
		*target = pos3{p.X, c.Body.Height + 1, p.Z}
	}
	if qa.stage == 138 {
		f := engine.HeadingVec(c.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		p := c.Pos.Add(f.Mul(4.3)).Sub(r.Mul(1.7))
		app.cam = pos3{p.X, c.Body.Height + .7, p.Z}
		*target = pos3{c.Pos.X, c.Body.Height + .12, c.Pos.Z}
	}
}
func injectRasterQA037() {
	if qa.enabled && (qa.stage == 143 || qa.stage == 144) {
		en(0x0C11)
		gl("glScissor", uintptr(app.width/4), uintptr(app.height/4), uintptr(app.width/2), uintptr(app.height/2))
	}
}

func checks037() {
	if qa.stage == 143 || qa.stage == 144 {
		qa.checks["037-fullframe-scissor-cleared"] = gl("glIsEnabled", 0x0C11) == 0
		qa.checks["037-enabled-bloom-and-effects"] = cinematic034.BloomActive && cinematic034.EffectsActive
		if qa.stage == 144 {
			if temporalMode038() == 0 {
				qa.checks["037-scaled-post-target"] = hdrScene.width == app.width*3/4 && hdrScene.height == app.height*3/4
			} else {
				qa.checks["038-reconstruction-overrides-manual-scale"] = postSceneWidth038() == app.width && postSceneHeight038() == app.height
			}
		}
	}
	if qa.stage == 138 {
		qa.checks["037-visible-steering-mesh"] = pbrAssets["b033_cabin_2"] != nil && math.Abs(app.renderCar.VisualSteer) > .05
	}
}
