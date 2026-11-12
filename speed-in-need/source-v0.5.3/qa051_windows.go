package main

import "speedinneed/engine"

func registerQA051() {
	defer registerQA053()
	remasterQANames = append(remasterQANames, "051-workshop", "051-live-customize", "051-walking", "051-compact", "051-coupe", "051-hatchback", "051-offroad", "051-pickup", "051-sedan", "051-suv", "051-wagon", "051-rx7", "051R-car-selection", "051R-customization", "051R-time-rain", "051R-clear-sky", "051R-rain-sky", "051R-new-car-fpp", "051R-nitro", "051R-skyline-r34", "051R-4x4-jeep", "051R-trueno-custom", "052-tunnel-lights-player-near", "052-tunnel-lights-player-far", "052-overcast-rain", "052-walker-idle", "052-walker-stride", "052-bridge-rails", "052-supplied-flowers")
}
func stageQA051() {
	if qa.stage >= 236 {
		stageQA053()
		return
	}
	if qa.stage >= 229 {
		stageQA052()
		return
	}
	weather050.Rain = false
	weather050.Amount = 0
	if qa.stage >= 219 {
		garage050.Loaded = true
		garage050.Setups = map[string]garageSetup050{}
		beta.carSelect = 3
		garage050.Setups[garageKey050()] = garageSetup050{Model: 1, Tyre: 1, Wheel: 2, Livery: 1}
		if qa.stage <= 221 {
			ui020.Stage = 2
			customize051 = qa.stage == 220
			if qa.stage == 221 {
				ui020.Stage = 3
				weather050.Rain = true
			}
			selectScreen(screenMaps)
			qa.checks["053-nine-cars-selectable"] = len(playerVariants052) == 9
			qa.checks["051R-new-car-selected"] = selectedCar051() == 0
			return
		}
		startRemasteredDrive()
		if qa.stage == 223 {
			weather050.Rain = true
			weather050.Amount = 1
			remaster.Time = 0
		}
		if qa.stage == 222 {
			weather050.Rain = false
			weather050.Amount = 0
			remaster.Time = 0
		}
		if qa.stage == 228 {
			garage050.Setups[garageKey050()] = garageSetup050{Model: 0, Tyre: 1, Wheel: 2, Livery: 1}
			beta.paint = 1
			applyGarage050()
		}
		if qa.stage == 227 {
			garage050.Setups[garageKey050()] = garageSetup050{Model: 14, Tyre: 3}
			applyGarage050()
		}
		if qa.stage == 226 {
			garage050.Setups[garageKey050()] = garageSetup050{Model: 13, Tyre: 1, Wheel: 1}
			applyGarage050()
		}
		if qa.stage >= 224 {
			enterWorkshop051()
			app.boolFPP = qa.stage == 224
			if qa.stage == 225 {
				app.world.Car.Boosting = true
				app.world.Car.Speed = 25
			}
		}
		return
	}

	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	model := 2
	if qa.stage >= 210 {
		model = qa.stage - 206
	}
	garage050.Setups[garageKey050()] = garageSetup050{Model: model, Tyre: 1}
	enterWorkshop051()
	app.camYaw = .65
	app.camDist = 6
	app.camPitch = .12
	if qa.stage == 208 {
		garage050.Open = true
	}
	if qa.stage == 209 {
		app.world.ToggleWalker051(&walker051)
		app.world.StepWalker051(&walker051, engine.Vec{Z: 1}, false, .03)
		qa.checks["051-player-exits-car"] = walker051.Active
		qa.checks["051-supplied-walker-renderable"] = characters032.Models[4] != nil
	}
	qa.checks["051-workshop-is-test-session"] = workshop051.Active && app.world.Remaster.Mode == 0 && len(app.world.Remaster.Racers) == 0
	qa.checks["051-supplied-garage-loaded"] = pbrAssets["b051_basement"] != nil
}

func observeMenuRemastered051() {
	if !qa.enabled {
		return
	}
	has := func(id string) bool {
		for _, b := range ui020.Buttons {
			if b.ID == id {
				return true
			}
		}
		return false
	}
	if qa.stage == 0 {
		qa.checks["051R-no-test-garage-on-title"] = !has("playGarage051") && has("start")
	}
	if qa.stage == 219 {
		qa.checks["053-nine-car-buttons-visible"] = has("vehicle051:0") && has("vehicle051:8") && !has("vehicle051:9") && has("customize051")
	}
	if qa.stage == 220 {
		qa.checks["051R-modification-controls-visible"] = has("modify051:1:1") && has("modify051:2:1") && has("modify051:3:1") && has("paint:0") && has("playGarage051")
	}
	if qa.stage == 221 {
		qa.checks["051R-rain-toggle-on-time-screen"] = has("rain051") && has("day") && has("night")
	}
}
