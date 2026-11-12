package main

import (
	"math"
	"speedinneed/engine"
)

var setupSpawnRefine struct {
	Selected bool
	Point    engine.Vec
	Heading  float64
	Safe     bool
}

func pickSetupSpawnRefine(x, y int) bool {
	if beta.screen != screenMaps || ui020.Stage != 1 || remaster.Mode != 0 {
		return false
	}
	w, h := float64(app.width), float64(app.height)
	left, top, size := w*.48, h*.215, h*.56
	if float64(x) < left || float64(x) > left+size || float64(y) < top || float64(y) > top+size {
		return false
	}
	d := remasterDefinition(0)
	if d == nil {
		return true
	}
	request := engine.Vec{X: engine.Lerp(d.Bounds[0].X, d.Bounds[1].X, (float64(x)-left)/size), Z: engine.Lerp(d.Bounds[0].Z, d.Bounds[1].Z, (float64(y)-top)/size)}
	setupSpawnRefine.Point, setupSpawnRefine.Heading, setupSpawnRefine.Safe = engine.SetupSpawnPoint(d, request, beta.carSelect)
	setupSpawnRefine.Selected = true
	return true
}
func drawSetupSpawnRefine(w, h float64) {
	if remaster.Mode != 0 {
		return
	}
	message := "CLICK THE MAP TO CHOOSE YOUR START"
	if setupSpawnRefine.Selected {
		d := remasterDefinition(0)
		if d == nil {
			return
		}
		size := h * .56
		x := w*.48 + (setupSpawnRefine.Point.X-d.Bounds[0].X)/(d.Bounds[1].X-d.Bounds[0].X)*size
		y := h*.215 + (setupSpawnRefine.Point.Z-d.Bounds[0].Z)/(d.Bounds[1].Z-d.Bounds[0].Z)*size
		pulse := .5 + .5*math.Sin(beta.menuTime*3)
		circle033(x, y, 12+pulse*5, [4]float64{.8, .96, .55, .15})
		circle033(x, y, 7, [4]float64{.04, .06, .06, 1})
		circle033(x, y, 4, cyan)
		message = "START LOCATION SELECTED / CLICK TO CHANGE"
	}
	remasterText(message, w*.078, h*.68, h*.011, cyan)
}
func applySetupSpawnRefine() {
	if !setupSpawnRefine.Selected || remaster.Mode != 0 {
		return
	}
	app.world.ApplySetupSpawn(setupSpawnRefine.Point)
	app.prevCar, app.renderCar = app.world.Car, app.world.Car
	app.renderAlpha = 1
	beta.cameraHeading = app.world.Car.Heading
	app.camYaw, app.camPitch = 0, 0
	remaster.CameraReset = true
	nav033.Active = false
}
