package main

import (
	"math"
	"speedinneed/engine"
)

var workshop051 struct {
	Active                bool
	Random051             bool
	Original              *engine.WorldDefinition
	Mode, Time, Opponents int
}
var walker051 engine.Walker051

func enterWorkshop051() {
	if multiplayer050.Session != nil {
		multiplayer050.Message = "Leave the LAN room to enter your private workshop"
		multiplayer050.Open = true
		return
	}
	raw, err := assets.ReadFile("assets/b051_basement.json")
	if err != nil {
		remaster.Status = err.Error()
		return
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		remaster.Status = err.Error()
		return
	}
	workshop051.Original = remasterDefinition(0)
	workshop051.Mode, workshop051.Time, workshop051.Opponents = remaster.Mode, remaster.Time, remaster.Opponents
	workshop051.Random051 = beta.cfg.RandomTime040
	beta.cfg.RandomTime040 = false
	workshop051.Active = true
	walker051 = engine.Walker051{}
	remaster.Definitions[0] = d
	remaster.Loaded[0] = false
	remaster.BufferMode = -1
	remaster.Mode, remaster.Opponents = 0, 0
	activateRemasteredWorld(true)
	reloadEnvironment051()
	app.world.Traffic = nil
	app.world.Remaster.TrafficPool = nil
	app.world.Pedestrians = nil
	app.world.Car.Body.Initialized = false
	applyGarage050()
	app.prevCar = app.world.Car
	app.renderCar = app.world.Car
	app.started = true
	app.paused = false
	app.boolFPP = false
	app.camYaw = .55
	app.camPitch = .08
	remaster.CameraReset = true
	selectScreen(screenDrive)
	garage050.Open = false
}
func exitWorkshop051() {
	if !workshop051.Active {
		return
	}
	remaster.Definitions[0] = workshop051.Original
	remaster.Loaded[0] = false
	remaster.BufferMode = -1
	remaster.Mode, remaster.Time, remaster.Opponents = workshop051.Mode, workshop051.Time, workshop051.Opponents
	beta.cfg.RandomTime040 = workshop051.Random051
	workshop051.Active = false
	reloadEnvironment051()
	walker051 = engine.Walker051{}
	garage050.Open = false
}
func stepPlayer051(in engine.Input, dt float64) {
	if walker051.Active {
		move := engine.Vec{X: in.Steer, Z: in.Throttle - in.Brake}
		// The controller's movement is camera-relative on foot, vehicle-relative in cars.
		yaw := app.camYaw + app.world.Car.Heading
		c, s := math.Cos(yaw), math.Sin(yaw)
		move = engine.Vec{X: move.X*c + move.Z*s, Z: -move.X*s + move.Z*c}
		app.world.StepWalker051(&walker051, move, in.Boost, dt)
		in = engine.Input{Handbrake: true, EngineGated: true, EngineEnabled: false}
	}
	app.world.Step(in, dt)
	if workshop051.Active && math.Abs(app.world.Car.Pos.X) > 39 || workshop051.Active && math.Abs(app.world.Car.Pos.Z) > 30 {
		app.world.RecoverRemasteredCar()
		remaster.CameraReset = true
	}
}
func drawWalker051() {
	if !walker051.Active || characters032.Models[4] == nil {
		return
	}
	if drawWalkClip052() {
		return
	}
	b := &characters032.Batches[4]
	b.Vertices = b.Vertices[:0]
	b.Indices = b.Indices[:0]
	appendCharacter032(4, 0, walker051.Ped.Pose, false)
	flushCharacter032(4)
}
func drawWorkshopHUD051(w, h float64) {
	hudRect(w*.025, h*.035, w*.65, h*.09, .015, .025, .035, .86)
	remasterText("PLAY GARAGE", w*.04, h*.05, h*.023, white)
	headline035("G CUSTOMIZE   E ENTER / EXIT CAR   H LIGHTS   F7 RAIN   F6 DAY / NIGHT   ESC MENU", w*.04, h*.09, h*.010, w*.61, cyan)
	if walker051.Active {
		remasterText("WASD / LEFT STICK: WALK    NITRO BUTTON: RUN", w*.04, h*.92, h*.013, white)
	} else {
		drawSpeed033(app.renderCar)
	}
}

func reloadEnvironment051() {
	releaseSky040()
	sky040.Failed = nil
	for label := 0; label < 3; label++ {
		bootstrapSky040(label)
	}
}
