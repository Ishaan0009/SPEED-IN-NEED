package main

import (
	"math"
	"time"
)

var roadRecoveryNotice038 time.Time

func recoverRoad038() {
	if app.world == nil || app.world.Remaster == nil {
		return
	}
	app.world.RecoverRemasteredCar()
	app.camYaw, app.camPitch = 0, 0
	app.prevCar = app.world.Car
	remaster.CameraReset = true
	autopilot034 = false
	roadRecoveryNotice038 = time.Now().Add(2 * time.Second)
}

func drawControllerHint038() {
	if !pad033.State.Connected || beta.screen != screenDrive || app.world == nil {
		return
	}
	w, h := float64(app.width), float64(app.height)
	if time.Now().Before(roadRecoveryNotice038) {
		remasterCenter("RETURNED TO ROAD", h*.86, h*.015, cyan)
		return
	}
	if camera036.ResetHold > 0 && !camera036.ResetFired {
		remasterCenter("HOLD DOWN: RETURN TO ROAD", h*.86, h*.013, white)
		hudRect(w*.4, h*.9, w*.2, 4, .06, .08, .1, .9)
		hudRect(w*.4, h*.9, w*.2*math.Min(1, camera036.ResetHold), 4, 0.1, .85, 1, 1)
	} else if !app.world.Car.OnRoad || math.Abs(app.world.Car.Speed) < 1 {
		remasterCenter("HOLD D-PAD DOWN 1 SEC: RETURN TO ROAD", h*.90, h*.011, white)
	}
}
