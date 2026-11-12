package main

import (
	"math"
	"speedinneed/engine"
)

func registerQA050() {
	remasterQANames = append(remasterQANames, "050-bmw", "050-eb110", "050-lp670", "050-night-rain", "050-dense-meadow", "050-lan-lobby", "050-garage", "050-balanced-meadow", "050-high-meadow")
}
func stageQA050() {
	if qa.stage < 198 {
		return
	}
	garage050.Open = false
	multiplayer050.Open = false
	weather050.Rain = false
	remaster.Mode = 0
	remaster.Time = 0
	beta.carSelect = 0
	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	if qa.stage <= 200 {
		garage050.Setups[garageKey050()] = garageSetup050{Model: qa.stage - 197, Livery: 1, Wheel: 1, Tyre: 1}
	}
	if qa.stage == 201 {
		remaster.Time = 2
		weather050.Rain = true
		weather050.Amount = 1
	}
	startRemasteredDrive()
	app.camDist = 5.8
	app.boolFPP = false
	if qa.stage == 202 || qa.stage == 205 || qa.stage == 206 {
		preset := 3
		if qa.stage == 205 {
			preset = 1
		}
		if qa.stage == 206 {
			preset = 2
		}
		applyAokawaPreset(preset)
		for _, o := range app.world.Remaster.Def.Objects {
			if o.Kind == "tree" && o.Pos[1] > 60 && o.Pos[1] < 100 {
				remasterQAPlace(engine.Vec{X: o.Pos[0] + 6, Z: o.Pos[2] + 6}, math.Pi)
				break
			}
		}
	}
	if qa.stage == 203 {
		multiplayer050.Open = true
		multiplayer050.Address = "192.168.1.10"
		multiplayer050.Message = "CREATE A ROOM OR JOIN THE HOST / F8 TO CLOSE"
	}
	if qa.stage == 204 {
		garage050.Open = true
		garage050.Setups[garageKey050()] = garageSetup050{Model: 1, Livery: 1, Wheel: 1}
	}
	applyGarage050()
	app.prevCar, app.renderCar = app.world.Car, app.world.Car
	app.renderAlpha = 1
}
