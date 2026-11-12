//go:build windows

package main

// The verification runner exercises the shipped renderer on a hidden WGL
// window. It reads its own framebuffer, never the user's desktop.
import (
	"encoding/json"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
	"strings"
	"time"
	"unsafe"
)

type qaResult struct {
	ShaderPipeline034                         map[string]any
	Optimization026                           map[string]any
	TextureBudget, GrassTextureWidth          int
	PostExposure, PostMotion                  float32
	TrafficCount                              int
	Memory                                    memorySample023
	Quality                                   int
	RenderScale                               float64
	Name                                      string
	Map, Car                                  int
	Frames                                    int
	FrameStatistics                           frameSummary
	GPUTimeMS                                 float64
	GPUMedianMS, GPUP95MS, SubmissionMedianMS float64
	HDR, WorldVBO, Instancing, SunShadows     bool
	GLErrors                                  []uint32
}

var qa struct {
	enabled, longRun    bool
	dir                 string
	stage, frame, limit int
	stageStart          time.Time
	last                time.Time
	samples             []float64
	results             []qaResult
	started             bool
	checks              map[string]bool
}
var qaNames = []string{"title", "coast-map", "city-map", "canyon-map", "ferrari-garage", "porsche-garage", "coast-ferrari", "coast-porsche", "city-ferrari", "city-porsche", "canyon-ferrari", "canyon-porsche", "night-headlights-off", "night-headlights-on", "reverse-lights", "ramp-deck", "ramp-flight", "fallen-streetlight", "pause", "settings", "settings-ui150", "garage-blue"}

func initQA() {
	qaAdaptive026 = os.Getenv("SPEED_QA_ADAPTIVE") == "1"
	if initRemasterQA() {
		return
	}
	for _, arg := range os.Args[1:] {
		if arg == "--verify-long" {
			qa.longRun = true
		}
		if strings.HasPrefix(arg, "--verify=") {
			qa.enabled = true
			qa.dir = strings.TrimPrefix(arg, "--verify=")
		}
	}
	if qa.enabled {
		qa.limit = 180
		qa.checks = map[string]bool{}
		os.MkdirAll(qa.dir, 0755)
	}
}
func qaSelectStage() {
	if remaster.QA {
		remasterQAStage()
		return
	}
	qa.frame = 0
	qa.samples = nil
	qa.last = time.Now()
	qa.stageStart = time.Now()
	beta.cfg.Headlights = true
	beta.cfg.TimeOfDay = 0
	app.paused = false
	app.boolFPP = false
	if qa.stage == 0 {
		selectScreen(screenTitle)
		return
	}
	if qa.stage <= 3 {
		beta.mapSelect = qa.stage - 1
		selectScreen(screenMaps)
		activateSelectedMap(false)
		qa.checks[qaNames[qa.stage]+"-selection"] = app.mapID == beta.mapSelect && app.world.Map == beta.mapSelect
		return
	}
	if qa.stage <= 5 {
		beta.carSelect = qa.stage - 4
		selectScreen(screenCars)
		return
	}
	mapID, car := 0, 0
	if qa.stage < 12 {
		mapID = (qa.stage - 6) / 2
		car = (qa.stage - 6) % 2
	} else {
		mapID = 1
	}
	beta.mapSelect = mapID
	beta.carSelect = car
	selectScreen(screenDrive)
	activateSelectedMap(true)
	app.started = true
	app.renderAlpha = 1
	resetDriveCamera()
	if qa.stage >= 12 && qa.stage <= 14 {
		beta.cfg.TimeOfDay = 3
		qa.limit = 90
	} else {
		qa.limit = 180
	}
	if qa.stage == 12 {
		betaInput('H')
		qa.checks["H-turns-headlights-off"] = !beta.cfg.Headlights
	}
	if qa.stage == 14 {
		app.world.Car.Speed = -3
		app.world.Car.Gearbox.Gear = -1
	}
	if qa.stage == 15 || qa.stage == 16 {
		r := app.world.Ramps[0]
		a := r.Length * .5
		if qa.stage == 16 {
			a = r.Length + 12
		}
		p, f := r.Sample(a)
		app.world.Car.Pos = p
		app.world.Car.Heading = math.Atan2(f.X, f.Z)
		app.world.InitializeChassis()
		if qa.stage == 16 {
			app.world.Car.Body.Height += r.Height + 3
			app.world.Car.Body.Grounded = false
			app.world.Car.Body.VerticalSpeed = 5
			app.world.Car.Body.AirVelocity = f.Mul(35)
			app.world.Car.Rig.Pitch = -.3
		}
		resetDriveCamera()
		qa.limit = 12
	}
	if qa.stage == 17 && len(app.world.Poles) > 0 {
		p := &app.world.Poles[0]
		p.Hit(engine.Vec{X: 1, Z: 1})
		for i := 0; i < 1800; i++ {
			p.Step(1. / 120)
		}
		app.world.Car.Pos = p.Pos.Add(engine.Vec{X: 4, Z: -6})
		app.world.InitializeChassis()
		resetDriveCamera()
		qa.checks["pole-settled"] = p.Settled
		qa.limit = 20
	}
	if qa.stage == 18 {
		selectScreen(screenPause)
		qa.limit = 20
	}
	if qa.stage == 20 {
		selectScreen(screenSettings)
		beta.cfg.UIScale = 1.5
		qa.limit = 20
	}
	if qa.stage == 21 {
		selectScreen(screenCars)
		beta.cfg.UIScale = 1
		beta.paint = 1
		qa.limit = 20
	}
	if qa.stage == 19 {
		selectScreen(screenSettings)
		qa.limit = 20
	}
	if qa.longRun && qa.stage >= 6 && qa.stage < 12 {
		qa.limit = 5400
	}
}
func qaInput() engine.Input {
	if remaster.QA {
		return remasterQAInput()
	}
	if qa.stage == 14 {
		return engine.Input{Brake: .3}
	}
	if qa.stage < 6 || qa.stage >= 12 {
		return engine.Input{}
	}
	c := app.world.Car
	r := app.world.Routes[0]
	best, bestD := 0., math.Inf(1)
	for d := 0.; d < r.Len; d += 6 {
		p, _ := r.AtDistance(d)
		if dist := p.Dist(c.Pos); dist < bestD {
			bestD = dist
			best = d
		}
	}
	p, _ := r.AtDistance(best + 14 + math.Abs(c.Speed)*.4)
	delta := p.Sub(c.Pos)
	angle := math.Atan2(delta.X, delta.Z)
	errorYaw := math.Atan2(math.Sin(angle-c.Heading), math.Cos(angle-c.Heading))
	in := engine.Input{Throttle: .6, Steer: engine.Clamp(errorYaw*2.8, -1, 1)}
	if c.Speed > 20 {
		in.Throttle = 0
		in.Brake = .15
	}
	return in
}
func presentFrame() {
 drawGarage050()
 drawLANOverlay050()
	beta020SavePhoto()
	if qa.enabled {
		qaAfterFrame()
	}
	exportUnityFrame040()
	present038()
}
func qaAfterFrame() {
	if remaster.QA {
		remasterQAAfterFrame()
		return
	}
	if !qa.started {
		return
	}
	now := time.Now()
	ms := now.Sub(qa.last).Seconds() * 1000
	qa.last = now
	qa.frame++
	warmup := 12
	if qa.longRun && qa.stage >= 6 && qa.stage < 12 {
		warmup = 180
	}
	if qa.frame > warmup {
		qa.samples = append(qa.samples, ms)
	}
	if qa.frame == qa.limit {
		gl("glFinish")
		saveFramebuffer(filepath.Join(qa.dir, qaNames[qa.stage]+".png"))
		captureMotion038(qaNames[qa.stage])
		errors := []uint32{}
		for i := 0; i < 16; i++ {
			e := uint32(gl("glGetError"))
			if e == 0 {
				break
			}
			errors = append(errors, e)
		}
		qa.results = append(qa.results, qaResult{Name: qaNames[qa.stage], Map: app.mapID, Car: beta.carSelect, Frames: qa.frame, FrameStatistics: summarizeFrames(qa.samples), GPUTimeMS: app.gpuMS, HDR: hdrScene.fbo != 0 && !hdrScene.failed, WorldVBO: gpuWorld.ready, Instancing: gpuWorld.instancing, SunShadows: sunShadow.ready, GLErrors: errors})
		qa.stage++
		if qa.stage >= len(qaNames) {
			finishQA()
			app.run = false
		} else {
			qa.started = false
		}
	}
}
func saveFramebuffer(path string) {
	w, h := app.width, app.height
	pixels := make([]byte, w*h*4)
	gl("glReadPixels", 0, 0, uintptr(w), uintptr(h), GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&pixels[0])))
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		copy(im.Pix[y*im.Stride:(y+1)*im.Stride], pixels[(h-y-1)*w*4:(h-y)*w*4])
	}
	f, err := os.Create(path)
	if err == nil {
		png.Encode(f, im)
		f.Close()
	}
}
func glString(which uintptr) string {
	p := gl("glGetString", which)
	if p == 0 {
		return "unavailable"
	}
	n := win(kernel32, "lstrlenA", p)
	if n == 0 || n > 1<<20 {
		return "unavailable"
	}
	data := make([]byte, n)
	win(kernel32, "RtlMoveMemory", uintptr(unsafe.Pointer(&data[0])), p, n)
	return string(data)

}
func finishQA() {
	report := map[string]any{"version": "Beta 0.1", "mode": "Automated hidden-window renderer verification; not a manual playtest or 30-minute interactive soak", "renderer": glString(0x1F01), "vendor": glString(0x1F00), "opengl": glString(0x1F02), "width": app.width, "height": app.height, "shaderError": lightPipeline.errorText, "assetErrors": pbrErrors, "checks": qa.checks, "stages": qa.results, "utc": time.Now().UTC().Format(time.RFC3339)}
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(filepath.Join(qa.dir, "RENDER_VERIFICATION.json"), raw, 0644)
}
