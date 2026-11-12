package main

import "testing"

func TestControls036ResetRequiresHoldAndRelease(t *testing.T) {
	old := camera036
	defer func() { camera036 = old }()
	camera036.ResetHold = 0
	camera036.ResetFired = false
	for i := 0; i < 9; i++ {
		if resetHold036(true, .1) {
			t.Fatal("reset before one second")
		}
	}
	fired := 0
	for i := 0; i < 20; i++ {
		if resetHold036(true, .1) {
			fired++
		}
	}
	if fired != 1 {
		t.Fatalf("reset repeated: %d", fired)
	}
	resetHold036(false, .1)
	if camera036.ResetHold != 0 || camera036.ResetFired {
		t.Fatal("release did not clear hold")
	}
}
func TestControls036ZoomLimitsAndFirstPerson(t *testing.T) {
	oldDist, oldMode, oldFPP, oldCamera := app.camDist, beta.cfg.CameraMode, app.boolFPP, camera036
	defer func() {
		app.camDist = oldDist
		beta.cfg.CameraMode = oldMode
		app.boolFPP = oldFPP
		camera036 = oldCamera
	}()
	beta.cfg.CameraMode = 3
	app.boolFPP = false
	app.camDist = 5.8
	zoomCamera036(-1)
	if app.camDist != 4.8 {
		t.Fatal("zoom does not change camera distance")
	}
	zoomCamera036(100)
	if app.camDist != 17 {
		t.Fatal("far bound")
	}
	zoomCamera036(-100)
	if app.camDist != 3.2 {
		t.Fatal("near bound")
	}
	app.boolFPP = true
	zoomCamera036(2)
	if app.camDist != 3.2 {
		t.Fatal("cockpit zoom moved camera through cabin")
	}
}

func TestControls036RightStickZoomDoesNotLook(t *testing.T) {
	oldCfg, oldCamera := beta.cfg, camera036
	oldDist, oldYaw, oldPitch, oldFPP := app.camDist, app.camYaw, app.camPitch, app.boolFPP
	defer func() {
		beta.cfg = oldCfg
		camera036 = oldCamera
		app.camDist = oldDist
		app.camYaw = oldYaw
		app.camPitch = oldPitch
		app.boolFPP = oldFPP
	}()
	beta.cfg = defaultCfg()
	beta.cfg.CameraMode = 3
	app.boolFPP = false
	app.camDist = 6
	app.camYaw = .2
	app.camPitch = .1
	s := padSample033{Buttons: 0x80, Axes: [6]float64{0, 0, .7, .8, 0, 0}}
	drivePad036(s, 0, .1)
	if app.camDist >= 6 || app.camYaw != .2 || app.camPitch != .1 {
		t.Fatal("R3 zoom changed look direction")
	}
	distance := app.camDist
	s.Buttons = 0
	drivePad036(s, 0, .1)
	if app.camDist != distance || app.camYaw == .2 || app.camPitch == .1 {
		t.Fatal("normal right stick did not restore look")
	}
}
