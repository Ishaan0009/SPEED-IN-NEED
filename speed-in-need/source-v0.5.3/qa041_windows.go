//go:build windows

package main

func registerQA041() {
	remasterQANames = append(remasterQANames, "041-settings-category-picker", "041-photo-mode")
}
func stageQA041() {
	if qa.stage == 196 {
		selectScreen(screenSettings)
		settingsRefine.Category = true
	}
	if qa.stage == 197 {
		remaster.Mode = 0
		beta.cfg.TrafficDensity = 0
		startRemasteredDrive()
		app.prevCar, app.renderCar = app.world.Car, app.world.Car
		app.renderAlpha = 1
		if !photo041.Active {
			togglePhoto041()
		}
	}
}
func checksQA041() {
	if qa.stage == 196 {
		qa.checks["041-category-picker-visible"] = beta.screen == screenSettings && settingsRefine.Category
	}
	if qa.stage == 197 {
		qa.checks["041-photo-active"] = photo041.Active && beta.screen == screenDrive
		qa.checks["041-photo-physics-paused"] = app.world.Elapsed == 0
		qa.checks["041-photo-orbit-valid"] = photo041.Distance >= 2.5 && photo041.Distance <= 25 && photo041.FOV >= 25 && photo041.FOV <= 85
	}
}
