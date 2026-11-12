//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
	"strings"
)

// Append after registerQAProfiles040. The pre-existing indices remain stable.
const qaWorldFinalFirst040 = 191

var qaWorldFinalNames040 = [...]string{
	"040-race-closed-loop-start",
	"040-race-realigned-cliff-corner",
	"040-race-floating-deck-gap",
	"040-race-coastal-deck-connection",
	"040-npc-driver-close",
}

// These are the four authored review coordinates in RACE_WORLD_FINAL040.json.
// Placement is projected onto the unchanged authoritative route, not terrain.
var qaWorldFinalPoints040 = [...]engine.Vec{
	{X: -388.39351, Z: 264.84349},
	{X: 162.07505, Z: -319.39116},
	{X: 403.00019, Z: -52.439},
	{X: 542.91381, Z: 315.44843},
}

var qaWorldFinal040 struct {
	Stage, SupportSamples, ClearanceSamples, NPCVariant int
	Ready, Supported, Clear                            bool
	Point, Forward, NPCPoint                           engine.Vec
	NPCHeading, MaximumShift, MaximumSupportError       float64
	Definition                                        *engine.WorldDefinition
}

func registerQAWorldFinal040() {
	if len(remasterQANames) != qaWorldFinalFirst040 {
		return
	}
	remasterQANames = append(remasterQANames, qaWorldFinalNames040[:]...)
}

func qaWorldFinalActive040() bool {
	return qa.enabled && qa.stage >= qaWorldFinalFirst040 && qa.stage < qaWorldFinalFirst040+len(qaWorldFinalNames040)
}

// Call after all existing stage hooks. Several older hooks accept every later
// index and place the player in free roam; this final hook must restore its own
// map, buffers, car and placement before the first frame is rendered.
func stageQAWorldFinal040() {
	if !qaWorldFinalActive040() {
		return
	}
	qaWorldFinal040.Ready = false
	qaWorldFinal040.Stage = qa.stage
	qaWorldFinal040.SupportSamples, qaWorldFinal040.ClearanceSamples = 0, 0
	qaWorldFinal040.MaximumShift, qaWorldFinal040.MaximumSupportError = 0, 0
	qaWorldFinal040.NPCVariant = -1
	if profiles040.Store != nil {
		profiles040.Store.Select040("ishaan")
		profiles040.Page = 0
	}
	beta.carSelect, beta.paint = 0, 0
	beta.cfg.RandomTime040 = false
	beta.cfg.TrafficDensity = 0
	beta.cfg.CameraAutoReturnRefine = false
	remaster.Time, remaster.Opponents = 0, 0
	remaster.Mode = 1
	if qa.stage == qaWorldFinalFirst040+4 {
		remaster.Mode = 0
	}
	startRemasteredDrive()
	w := app.world
	if w == nil || w.Remaster == nil || len(w.Routes) == 0 || w.Map != remaster.Mode {
		qa.checks[fmt.Sprintf("040-world-%d-created", qa.stage)] = false
		return
	}
	w.Remaster.Time = 0
	w.Remaster.Racers = nil
	w.Remaster.Opponents = 0
	w.Remaster.Countdown = 0
	w.Remaster.Started = false
	w.Traffic = nil
	w.Pedestrians = nil
	qaWorldFinal040.Definition = w.Remaster.Def
	if remaster.Mode == 1 {
		point := qaWorldFinalPoints040[qa.stage-qaWorldFinalFirst040]
		distance, _ := w.Routes[0].ClosestDistance(point, -1)
		point, forward := w.Routes[0].AtDistance(distance)
		qaWorldFinal040.Point, qaWorldFinal040.Forward = point, forward
		remasterQAPlace(point, math.Atan2(forward.X, forward.Z))
		qaWorldFinal040.Supported, qaWorldFinal040.Clear = probeQARaceWorldFinal040(w, distance)
	} else {
		// Route zero at one quarter is the existing bounded bus/NPC QA road.
		// A stopped genuine-window b011 body remains deterministic even while a
		// high-detail sky upload extends the capture beyond its normal frame count.
		route := w.Routes[0]
		at := route.Len * .25
		point, heading := engine.TrafficCornerPose(route, at, 1, 2.7, 0)
		t := engine.Traffic{Path: 0, T: at, Direction: 1, Lane: 2.7, Variant: 4,
			Pos: point, Heading: heading, Speed: 0, CruiseSpeed: 0}
		w.Traffic = []engine.Traffic{t}
		w.Remaster.TrafficPool = []engine.Traffic{t}
		w.Remaster.TrafficSnapshot = []engine.Traffic{t}
		behind, forward := route.AtDistance(at - 16)
		behind = behind.Add(engine.Vec{X: forward.Z, Z: -forward.X}.Mul(2.7))
		remasterQAPlace(behind, math.Atan2(forward.X, forward.Z))
		qaWorldFinal040.Point, qaWorldFinal040.Forward = behind, forward
		qaWorldFinal040.NPCPoint, qaWorldFinal040.NPCHeading = point, heading
		qaWorldFinal040.NPCVariant = t.Variant
		trafficVisual040.Definition = nil
	}
	app.prevCar, app.renderCar = w.Car, w.Car
	app.renderAlpha = 1
	app.boolFPP = false
	app.camYaw, app.camPitch = 0, 0
	app.camDist = 8.5
	beta.cameraHeading = w.Car.Heading
	beta.cfg.CameraMode = 1
	remaster.FullMap = false
	remaster.CameraReset = true
	remasterQAPath, remasterQAHint = 0, -1
	qaWorldFinal040.Ready = true
}

// This tests the actual decoded definition used by the renderer. It complements
// the existing full-course engine sweeps; it never substitutes for PNG review.
func probeQARaceWorldFinal040(w *engine.World, distance float64) (bool, bool) {
	supported, clear := true, true
	for _, along := range []float64{-12, -6, 0, 6, 12} {
		centre, forward := w.Routes[0].AtDistance(distance + along)
		right := engine.Vec{X: forward.Z, Z: -forward.X}
		for _, side := range []float64{-8, -4, 0, 4, 8} {
			point := centre.Add(right.Mul(side))
			ground := w.GroundHeight(point)
			best := math.Inf(1)
			for _, triangle := range w.Remaster.Def.Surfaces {
				if triangle.Road && triangle.Support {
					if height, ok := qaTriangleHeightWorldFinal040(triangle, point); ok {
						best = math.Min(best, math.Abs(height-ground))
					}
				}
			}
			qaWorldFinal040.SupportSamples++
			if math.IsNaN(ground) || math.IsInf(ground, 0) || math.IsInf(best, 0) || best > .05 {
				supported = false
			} else {
				qaWorldFinal040.MaximumSupportError = math.Max(qaWorldFinal040.MaximumSupportError, best)
			}
		}
	}
	probe := engine.NewRemasteredWorld(w.Remaster.Def, 1, 0, 0, 1, 0)
	probe.Remaster.Countdown, probe.Remaster.Started = 0, false
	probe.Remaster.Racers, probe.Traffic, probe.Pedestrians = nil, nil, nil
	for class := 0; class < 4; class++ {
		probe.CarClass = class
		for _, along := range []float64{-6, 0, 6} {
			centre, forward := w.Routes[0].AtDistance(distance + along)
			right := engine.Vec{X: forward.Z, Z: -forward.X}
			for _, side := range []float64{-6.4, 0, 6.4} {
				point := centre.Add(right.Mul(side))
				probe.Car = engine.Car{Pos: point, Heading: math.Atan2(forward.X, forward.Z), Nitro: 100, DriftCombo: 1, EngineTemperature: 80, BrakeTemperature: 25}
				probe.InitializeChassis()
				for step := 0; step < 8; step++ {
					probe.Step(engine.Input{Brake: 1, Handbrake: true}, 1./120)
				}
				shift := probe.Car.Pos.Dist(point)
				qaWorldFinal040.ClearanceSamples++
				qaWorldFinal040.MaximumShift = math.Max(qaWorldFinal040.MaximumShift, shift)
				if math.IsNaN(shift) || shift > .05 || math.Abs(probe.Car.Body.Height-probe.GroundHeight(probe.Car.Pos)-.74) > .20 {
					clear = false
				}
			}
		}
	}
	return supported, clear
}

func qaTriangleHeightWorldFinal040(t engine.SurfaceTriangle, p engine.Vec) (float64, bool) {
	ax, az, bx, bz, cx, cz := t.A[0], t.A[2], t.B[0], t.B[2], t.C[0], t.C[2]
	denominator := (bz-cz)*(ax-cx) + (cx-bx)*(az-cz)
	if math.Abs(denominator) < 1e-9 {
		return 0, false
	}
	a := ((bz-cz)*(p.X-cx) + (cx-bx)*(p.Z-cz)) / denominator
	b := ((cz-az)*(p.X-cx) + (ax-cx)*(p.Z-cz)) / denominator
	c := 1 - a - b
	if a < -.0001 || b < -.0001 || c < -.0001 {
		return 0, false
	}
	return a*t.A[1] + b*t.B[1] + c*t.C[1], true
}

// Call after viewQA036/viewQA037, immediately before the normal camera upload.
// The race view frames the deck, bank and neighboring slope together. The NPC
// view is a front-quarter windshield view, not an occluded chase view.
func viewQAWorldFinal040(c engine.Car, target *pos3) {
	if !qaWorldFinalActive040() || !qaWorldFinal040.Ready || qaWorldFinal040.Stage != qa.stage {
		return
	}
	w := app.world
	if qa.stage == qaWorldFinalFirst040+4 {
		if len(w.Traffic) != 1 {
			return
		}
		t := w.Traffic[0]
		forward := engine.HeadingVec(t.Heading)
		right := engine.Vec{X: forward.Z, Z: -forward.X}
		point := t.Pos.Add(forward.Mul(5.2)).Add(right.Mul(2.0))
		floor := w.GroundHeight(t.Pos)
		app.cam = pos3{point.X, floor + 2.10, point.Z}
		*target = pos3{t.Pos.X + forward.X*.15, floor + 1.28, t.Pos.Z + forward.Z*.15}
		return
	}
	forward := qaWorldFinal040.Forward
	right := engine.Vec{X: forward.Z, Z: -forward.X}
	centre := qaWorldFinal040.Point
	point := centre.Sub(forward.Mul(10)).Add(right.Mul(7.2))
	floor := w.GroundHeight(centre)
	app.cam = pos3{point.X, math.Max(floor+4.8, w.GroundHeight(point)+1.0), point.Z}
	look := centre.Add(forward.Mul(9))
	*target = pos3{look.X, w.GroundHeight(look) + .5, look.Z}
}

func checksWorldFinal040() {
	if !qaWorldFinalActive040() {
		return
	}
	name := qaWorldFinalNames040[qa.stage-qaWorldFinalFirst040]
	check := func(suffix string, value bool) { qa.checks[name+"-"+suffix] = value }
	w := app.world
	ready := qaWorldFinal040.Ready && qaWorldFinal040.Stage == qa.stage && w != nil && w.Remaster != nil && w.Remaster.Def == qaWorldFinal040.Definition
	check("fixture-ready", ready)
	if !ready {
		return
	}
	check("world-rendered", gpuWorld.ready && gpuWorld.draws > 0 && hdrScene.fbo != 0 && !hdrScene.failed)
	check("player-holds-position", w.Car.Pos.Dist(qaWorldFinal040.Point) < .15 && math.Abs(w.Car.Speed) < .15)
	proof := map[string]any{"fixture": name, "stage": qa.stage, "mode": w.Map, "car": w.CarClass,
		"world_point": qaWorldFinal040.Point, "camera": app.cam, "draw_calls": gpuWorld.draws,
		"png": name + ".png", "visual_review_required": true}
	if qa.stage < qaWorldFinalFirst040+4 {
		check("race-mode-and-car", remaster.Mode == 1 && w.Map == 1 && w.Remaster.Mode == 1 && w.CarClass == 0)
		check("full-width-support", qaWorldFinal040.Supported && qaWorldFinal040.SupportSamples == 25)
		check("four-car-lane-clearance", qaWorldFinal040.Clear && qaWorldFinal040.ClearanceSamples == 36)
		loaded := true
		for _, key := range []string{"b041_race_scenery", "b040_race_final_deck", "b040_race_final_banks", "b040_race_final_substrate", "b041_race_guardrails"} {
			a := pbrAssets[key]
			if a == nil || len(a.Batches) == 0 {
				loaded = false
				continue
			}
			for _, batch := range a.Batches {
				loaded = loaded && batch.vbo != 0 && batch.ebo != 0
			}
		}
		check("final-repair-assets-uploaded", loaded)
		proof["support_samples"], proof["clearance_samples"] = qaWorldFinal040.SupportSamples, qaWorldFinal040.ClearanceSamples
		proof["max_support_error_m"], proof["max_clearance_shift_m"] = qaWorldFinal040.MaximumSupportError, qaWorldFinal040.MaximumShift
	} else {
		check("free-road-genuine-window-npc", w.Map == 0 && len(w.Traffic) == 1 && w.Traffic[0].Variant == 4 && trafficAssetKey040(4) == "b011_npc0")
		glass := false
		if a := pbrAssets["b011_npc0"]; a != nil {
			for _, material := range a.Materials {
				if strings.HasPrefix(strings.ToLower(material.Name), "glass") {
					glass = glass || material.Glass && material.Alpha == "BLEND" && material.Metal == 0 && material.Base[3] > 0 && material.Base[3] < .4
				}
			}
		}
		check("glass-material-active", glass)
		bodyUploaded, glovesUploaded := false, false
		if len(w.Traffic) == 1 {
			bin, _ := trafficSteerBin040(w.Traffic[0].SteerAngle)
			if mesh := trafficDriverMeshes040[[2]int{0, bin}]; mesh != nil {
				bodyUploaded = mesh.Ready && mesh.Safe && mesh.Body.VBO != 0 && mesh.Body.EBO != 0
				glovesUploaded = mesh.Hands.VBO != 0 && mesh.Hands.EBO != 0
			}
			proof["npc_position"] = w.Traffic[0].Pos
			proof["npc_distance_m"] = w.Traffic[0].Pos.Dist(w.Car.Pos)
		}
		check("occupant-body-and-gloves-uploaded", bodyUploaded && glovesUploaded)
		check("occupant-draw-submitted", trafficVisual040.Drivers == 1)
		proof["npc_asset"], proof["driver_draws"] = "b011_npc0", trafficVisual040.Drivers
	}
	raw, err := json.MarshalIndent(proof, "", "  ")
	if err == nil {
		err = os.WriteFile(filepath.Join(qa.dir, name+".fixture040.json"), raw, 0644)
	}
	check("proof-report-written", err == nil)
}
