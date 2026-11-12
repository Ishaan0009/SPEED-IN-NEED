//go:build windows

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
	"strings"
	"time"
	"unsafe"
)

const (
	screenTitle    = 0
	screenMaps     = 1
	screenCars     = 2
	screenSettings = 3
	screenDrive    = 4
	screenPause    = 5
	screenResults  = 6
)

type nativeConfig struct {
	CrashEffects041                                                bool
	RandomTime040                                                  bool
	SkyDetail040                                                   int
	RaceDifficulty040                                              int
	InvertLookYRefine, CameraAutoReturnRefine                      bool
	RendererBackend                                                int `json:"rendererBackend"`
	Upscaling038, FrameGeneration038, RayShadows038, RayQuality038 int
	GPUTextureAuto                                                 bool    `json:"gpuTextureAuto"`
	MapOrientation                                                 int     `json:"mapOrientation"`
	CameraShake                                                    float64 `json:"cameraShake"`
	GamepadEnabled, GamepadVibration                               bool
	GamepadDeadzone, GamepadSensitivity                            float64
	GamepadSteerAxis, GamepadThrottleAxis, GamepadBrakeAxis        int
	GamepadButtons                                                 [8]uint32
	GamepadTriggerMode                                             int
	GamepadSteerInvert                                             bool
	PadMappingVersion                                              int
	MasterVolume, EffectsVolume, MusicVolume, AmbienceVolume       float64
	RadioStation                                                   int
	AudioMuted                                                     bool
	Exposure                                                       float64 `json:"exposure"`
	Saturation                                                     float64 `json:"saturation"`
	BlurStrength                                                   float64 `json:"blurStrength"`
	BloomStrength                                                  float64 `json:"bloomStrength"`
	TextureBudget                                                  int     `json:"textureBudget"`
	MeshDetail                                                     int     `json:"meshDetail"`
	GroundDetail                                                   int     `json:"groundDetail"`
	LightingDetail                                                 int     `json:"lightingDetail"`
	ShadowFilter                                                   int     `json:"shadowFilter"`
	FoliageDensity                                                 float64 `json:"foliageDensity"`
	Wind                                                           bool    `json:"wind"`
	TrafficDensity                                                 int     `json:"trafficDensity"`
	ColorGrade                                                     int     `json:"colorGrade"`
	CameraMode                                                     int     `json:"cameraMode"`
	RenderScale                                                    float64 `json:"renderScale"`
	AdaptiveQuality                                                bool    `json:"adaptiveQuality"`
	Lightweight                                                    bool    `json:"lightweight"`
	VSync                                                          bool    `json:"vsync"`
	FrameCap                                                       int     `json:"frameCap"`
	Resolution                                                     int     `json:"resolution"`
	Shadows                                                        int     `json:"shadows"`
	ViewDistance                                                   float64 `json:"viewDistance"`
	Vegetation                                                     float64 `json:"vegetation"`
	Anisotropy                                                     int     `json:"anisotropy"`
	TextureBias                                                    int     `json:"textureBias"`
	FXAA                                                           bool    `json:"fxaa"`
	Bloom                                                          bool    `json:"bloom"`
	MotionBlur                                                     bool    `json:"motionBlur"`
	SSAO                                                           bool    `json:"ssao"`

	GameMode          int       `json:"gameMode"`
	RaceLaps          int       `json:"raceLaps"`
	Opponents         int       `json:"opponents"`
	Headlights        bool      `json:"headlights"`
	TimeOfDay         int       `json:"timeOfDay"`
	MouseSteer        bool      `json:"mouseSteer"`
	Fullscreen        bool      `json:"fullscreen"`
	InvertMouseX      bool      `json:"invertMouseX"`
	Sensitivity       float64   `json:"steeringSensitivity"`
	CameraSensitivity float64   `json:"cameraSensitivity"`
	CameraDistance    float64   `json:"cameraDistance"`
	Quality           int       `json:"quality"`
	Bindings          [6]uint32 `json:"bindings"`
	Unlocks           int       `json:"unlocks"`
	Paint             int       `json:"paint"`
	SelectedMap       int       `json:"selectedMap"`
	SelectedCar       int       `json:"selectedCar"`
	UIScale           float64   `json:"uiScale"`
	DriveMode         int       `json:"driveMode"`
	Lighting          bool      `json:"lighting"`
	FirstPerson       bool      `json:"firstPerson"`
}

var beta struct {
	screen, prevScreen, selection, mapSelect, carSelect, mode, paint int
	menuTime, idleMouse, carTurn                                     float64
	bindWaiting                                                      int
	cfg                                                              nativeConfig
	mesh                                                             [8]uint32
	paintMesh                                                        [2][12]uint32
	trafficMesh                                                      [7]uint32
	wheelMesh                                                        [2][4]uint32
	wheelAnchor                                                      [2][4][3]float64
	cameraHeading, camVelocity                                       float64
	mouseLook                                                        bool
}
var paintPalette = [...][3]float64{{.96, .18, .09}, {.11, .42, .92}, {.92, .93, .96}, {.08, .09, .11}, {.72, .74, .78}, {.32, .34, .37}, {.99, .78, .12}, {1, .44, .08}, {.17, .74, .49}, {.64, .16, .86}, {.12, .72, .82}, {.65, .08, .16}, {.88, .62, .43}, {.32, .50, .26}, {.48, .33, .79}, {.91, .31, .58}, {.18, .23, .37}, {.82, .79, .64}, {.30, .025, .045}, {.08, .21, .48}, {.025, .13, .10}, {.88, .83, .72}, {.16, .17, .18}, {.43, .46, .50}, {.015, .07, .20}, {.045, .24, .22}, {.46, .23, .10}, {.94, .87, .12}, {.87, .39, .55}, {.29, .40, .53}, {.50, .12, .08}, {.52, .60, .30}, {.22, .09, .32}, {.80, .83, .86}, {.12, .13, .14}, {.69, .71, .74}}
var bindingLabels = []string{"ACCELERATE", "FOOT BRAKE", "STEER LEFT", "STEER RIGHT", "NITRO", "HANDBRAKE / DRIFT"}
var defaultKeys = [6]uint32{'W', 'S', 'A', 'D', VK_SHIFT, VK_SPACE}

func cfgPath() string {
	if executable, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(executable), "UserData", "settings.json")
	}
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = "."
	}
	return filepath.Join(base, "SpeedInNeedBeta030", "settings.json")
}
func defaultCfg() nativeConfig {
	c := nativeConfig{RaceDifficulty040: 1, RayQuality038: 2, GPUTextureAuto: true, CameraShake: .65, GamepadEnabled: true, GamepadVibration: true, GamepadDeadzone: .14, GamepadSensitivity: 1, GamepadSteerInvert: true, GamepadSteerAxis: 0, GamepadThrottleAxis: 5, GamepadBrakeAxis: 4, GamepadButtons: defaultPadButtons033, MasterVolume: .8, EffectsVolume: .9, MusicVolume: .7, AmbienceVolume: .65, RadioStation: 0, AdaptiveQuality: true, Exposure: 1, Saturation: 1, BlurStrength: .45, BloomStrength: 1.15, TextureBudget: 1, MeshDetail: 1, GroundDetail: 1, LightingDetail: 1, ShadowFilter: 1, FoliageDensity: 1, Wind: true, TrafficDensity: 2, ColorGrade: 1, CameraMode: 3, RenderScale: 1, VSync: true, FrameCap: 60, Resolution: 1, Shadows: 1, ViewDistance: 1.10, Vegetation: 1, Anisotropy: 16, FXAA: true, Bloom: true, MotionBlur: true, SSAO: true, RaceLaps: 3, Opponents: 4, Headlights: true, Sensitivity: 1, CameraSensitivity: 1, CameraDistance: 5.8, Quality: 1, Bindings: defaultKeys, Unlocks: 7, Paint: 0, UIScale: 1, DriveMode: 1, Lighting: true}
	configureQuality050(&c, 1)
	return c
}
func readCfg() {
	beta.cfg = defaultCfg()
	raw, err := os.ReadFile(cfgPath())
	newProfile023 := err != nil
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta025", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta024", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta023", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta22", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta022", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta021", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta020", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta013", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta012", "settings.json"))
	}
	if err != nil {
		raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaRemastered", "settings.json"))
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV0_3_1", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV3_0", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV2_3", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV2_2", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV2_1", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV1", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedAlphaV2", "settings.json"))
		}
		if err != nil {
			raw, err = os.ReadFile(filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta3", "settings.json"))
		}
	}
	if err == nil {
		c := defaultCfg()
		if json.Unmarshal(raw, &c) == nil {
			if c.CameraDistance > 0 && c.Sensitivity > 0 {
				beta.cfg = c
			}
		}
	}
	// Migrate old stock Space=nitro / Shift=drift defaults to new layout.
	// A user's explicitly remapped controls remain untouched.
	if beta.cfg.Bindings[4] == VK_SPACE && beta.cfg.Bindings[5] == VK_SHIFT {
		beta.cfg.Bindings[4], beta.cfg.Bindings[5] = VK_SHIFT, VK_SPACE
	}
	beta.cfg.Unlocks = 7 // Alpha release: all three maps available immediately
	beta.cfg.TimeOfDay = engine.ClampInt(beta.cfg.TimeOfDay, 0, 3)
	beta.cfg.UIScale = engine.Clamp(beta.cfg.UIScale, 1, 1.5)
	beta.cfg.CameraSensitivity = engine.Clamp(beta.cfg.CameraSensitivity, .25, 3)
	beta.cfg.DriveMode = engine.ClampInt(beta.cfg.DriveMode, 0, 2)
	for i, k := range beta.cfg.Bindings {
		if k == 0 {
			beta.cfg.Bindings[i] = defaultKeys[i]
		}
	}
	beta.paint = engine.ClampInt(beta.cfg.Paint, 0, len(paintPalette)-1)
	beta.mapSelect = engine.ClampInt(beta.cfg.SelectedMap, 0, 2)
	beta.carSelect = engine.ClampInt(beta.cfg.SelectedCar, 0, 3)
	app.quality = engine.ClampInt(beta.cfg.Quality, 0, 3)
	beta.cfg.RenderScale = engine.Clamp(beta.cfg.RenderScale, .5, 1)
	beta.cfg.CameraMode = engine.ClampInt(beta.cfg.CameraMode, 0, 3)
	beta.cfg.ViewDistance = engine.Clamp(beta.cfg.ViewDistance, .5, 1.5)
	beta.cfg.Vegetation = engine.Clamp(beta.cfg.Vegetation, .5, 1.5)
	beta.cfg.Shadows = engine.ClampInt(beta.cfg.Shadows, 0, 3)
	beta.cfg.Upscaling038 = engine.ClampInt(beta.cfg.Upscaling038, 0, 5)
	beta.cfg.FrameGeneration038 = engine.ClampInt(beta.cfg.FrameGeneration038, 0, 1)
	beta.cfg.RayShadows038 = engine.ClampInt(beta.cfg.RayShadows038, 0, 2)
	beta.cfg.RayQuality038 = engine.ClampInt(beta.cfg.RayQuality038, 0, 3)
	beta.cfg.Anisotropy = engine.ClampInt(beta.cfg.Anisotropy, 1, 16)
	beta.cfg.Exposure = engine.Clamp(beta.cfg.Exposure, .65, 1.4)
	beta.cfg.Saturation = engine.Clamp(beta.cfg.Saturation, .7, 1.3)
	beta.cfg.BlurStrength = engine.Clamp(beta.cfg.BlurStrength, 0, 1)
	beta.cfg.BloomStrength = engine.Clamp(beta.cfg.BloomStrength, 0, 1.5)
	beta.cfg.TextureBudget = engine.ClampInt(beta.cfg.TextureBudget, 0, 4)
	app.sensitivity = beta.cfg.Sensitivity
	app.camDist = beta.cfg.CameraDistance
	app.boolMouse = false
	beta.cfg.MouseSteer = false
	if beta.cfg.Bindings[2] == 'D' && beta.cfg.Bindings[3] == 'A' {
		beta.cfg.Bindings[2], beta.cfg.Bindings[3] = 'A', 'D'
	}
	app.invertMouseX = beta.cfg.InvertMouseX
	app.boolFPP = beta.cfg.FirstPerson
	if newProfile023 {
		applyAokawaPreset(app.quality)
		applyRTX3060Profile034()
	}
	migratePad034()
	migratePad036()
}
func saveCfg() {
	if qa.enabled {
		return
	}
	beta.cfg.MouseSteer = false
	app.boolMouse = false
	beta.cfg.Fullscreen = app.fullscreen
	beta.cfg.InvertMouseX = app.invertMouseX
	beta.cfg.Sensitivity = app.sensitivity
	beta.cfg.CameraDistance = app.camDist
	beta.cfg.Quality = app.quality
	beta.cfg.Paint = beta.paint
	beta.cfg.SelectedMap = beta.mapSelect
	beta.cfg.SelectedCar = beta.carSelect
	if app.world != nil {
		beta.cfg.DriveMode = app.world.DriveMode
	}
	beta.cfg.Lighting = app.lightingOn
	beta.cfg.FirstPerson = app.boolFPP
	os.MkdirAll(filepath.Dir(cfgPath()), 0755)
	raw, _ := json.MarshalIndent(beta.cfg, "", "  ")
	_ = os.WriteFile(cfgPath(), raw, 0644)
	saveProfileSettings040()
}
func bindingName(k uint32) string {
	if k == VK_SHIFT {
		return "SHIFT"
	}
	if k == VK_SPACE {
		return "SPACE"
	}
	if k >= 'A' && k <= 'Z' {
		return string(rune(k))
	}
	if k >= '0' && k <= '9' {
		return string(rune(k))
	}
	return fmt.Sprintf("KEY %d", k)
}
func pressed(i int) bool { return key(int(beta.cfg.Bindings[i])) }
func selectScreen(screen int) {
	if screen != screenDrive && photo041.Active {
		photo041.Active, photo041.Drag = false, false
		app.boolFPP = photo041.PreviousFPP
		invalidateTemporal038()
	}
	if screen != beta.screen {
		playAudio030("ui_confirm", .45)
	}
	beta.screen = screen
	if screen == screenSettings {
		settingsRefine.Category = !qa.enabled
	}
	beta.selection = 0
	beta.bindWaiting = -1
	app.mouseInit = false
	app.camHeld = false
	beta020SyncPointer()
	if screen == screenPause {
		remaster.FullMap = false
		win(user32, "ReleaseCapture")
	}
}
func betaInput(k uintptr) bool {
	if k == 0x75 && beta.screen == screenDrive {
		toggleBenchmark()
		return true
	}
	if k == 0x72 {
		app.debugCollision = !app.debugCollision
		return true
	} // F3
	if k == 0x73 {
		app.debugRoutes = !app.debugRoutes
		return true
	} // F4
	if k == 0x74 {
		app.debugWheels = !app.debugWheels
		return true
	} // F5
	if k == 0x71 { // F2 performance overlay
		app.showPerf = !app.showPerf
		return true
	}
	if k == 0x7a { // F11
		toggleFullscreen()
		saveCfg()
		return true
	}
	if beta.bindWaiting >= 0 {
		if k != VK_ESCAPE {
			beta.cfg.Bindings[beta.bindWaiting] = uint32(k)
			saveCfg()
		}
		beta.bindWaiting = -1
		return true
	}
	if beta.screen == screenDrive {
		switch k {
		case 'V':
			app.boolFPP = !app.boolFPP
			saveCfg()
		case 'L':
			app.lightingOn = !app.lightingOn
			saveCfg()
		case 'T':
			app.world.DriveMode = (app.world.DriveMode + 1) % 3
			saveCfg()
		case 'M':
			app.boolMouse = false
			saveCfg()
		case 'I':
			app.invertMouseX = !app.invertMouseX
			saveCfg()
		case VK_ESCAPE, 'P':
			selectScreen(screenPause)
		case 'R':
			app.world.Reset()
			app.prevCar = app.world.Car
			app.world.StartChallenge(beta.mode)
		case 'F':
			cycleRadio030()
		case 'H':
			beta.cfg.Headlights = !beta.cfg.Headlights
			saveCfg()
		case 0x70:
			app.showHelp = !app.showHelp
		}
		return true
	}
	if beta.screen == screenPause {
		if k == 0x26 {
			beta.selection = (beta.selection + 3) % 4
			return true
		}
		if k == 0x28 {
			beta.selection = (beta.selection + 1) % 4
			return true
		}
		if k == VK_RETURN {
			switch beta.selection {
			case 0:
				k = VK_ESCAPE
			case 1:
				k = 'S'
			case 2:
				k = 'R'
			case 3:
				k = 'Q'
			}
		}
		switch k {
		case VK_ESCAPE, 'P', VK_RETURN:
			selectScreen(screenDrive)
		case 'S':
			beta.prevScreen = screenPause
			selectScreen(screenSettings)
		case 'Q':
			selectScreen(screenTitle)
			app.started = false
		case 'R':
			app.world.Reset()
			app.prevCar = app.world.Car
			app.world.StartChallenge(beta.mode)
			selectScreen(screenDrive)
		}
		return true
	}
	if beta.screen == screenResults {
		if k == VK_RETURN || k == VK_ESCAPE {
			selectScreen(screenMaps)
			app.started = false
		}
		return true
	}
	if k == VK_ESCAPE {
		switch beta.screen {
		case screenMaps:
			selectScreen(screenTitle)
		case screenCars:
			selectScreen(screenMaps)
		case screenSettings:
			selectScreen(beta.prevScreen)
		default:
			selectScreen(screenTitle)
		}
		return true
	}
	switch beta.screen {
	case screenTitle:
		if k == 0x26 {
			beta.selection = (beta.selection + 3) % 4
		}
		if k == 0x28 {
			beta.selection = (beta.selection + 1) % 4
		}
		if k == VK_RETURN {
			switch beta.selection {
			case 0:
				selectScreen(screenMaps)
			case 1:
				selectScreen(screenCars)
			case 2:
				beta.prevScreen = screenTitle
				selectScreen(screenSettings)
			case 3:
				app.run = false
			}
		}
	case screenMaps:
		if k == 'A' || k == 0x25 {
			beta.mapSelect = (beta.mapSelect + 2) % 3
		}
		if k == 'D' || k == 0x27 {
			beta.mapSelect = (beta.mapSelect + 1) % 3
		}
		if k == 0x26 {
			beta.mapSelect = (beta.mapSelect + 2) % 3
		}
		if k == 0x28 {
			beta.mapSelect = (beta.mapSelect + 1) % 3
		}
		activateSelectedMap(false)
		if k == VK_RETURN {
			saveCfg()
			selectScreen(screenCars)
		}
	case screenCars:
		if k == 'A' || k == 0x25 {
			beta.carSelect = (beta.carSelect + 1) % 2
		}
		if k == 'D' || k == 0x27 {
			beta.carSelect = (beta.carSelect + 1) % 2
		}
		if k == 'Q' || k == 'E' {
			if k == 'Q' {
				beta.paint = (beta.paint + len(paintPalette) - 1) % len(paintPalette)
			} else {
				beta.paint = (beta.paint + 1) % len(paintPalette)
			}
			saveCfg()
			// Paint changes now only change a runtime tint. Recompiling 1M vertices on each keypress crashed some PCs.
		}
		if k == 'N' {
			beta.mode = (beta.mode + 1) % 4
		}
		if k == VK_RETURN {
			activateSelectedMap(true)
			saveCfg()
			app.boolFPP = beta.cfg.FirstPerson
			app.world.StartChallenge(beta.mode)
			invalidateWorldGeometry()
			app.camYaw = 0
			resetDriveCamera()
			app.started = true
			app.paused = false
			selectScreen(screenDrive)
		}
	case screenSettings:
		max := 15
		if k == 0x26 {
			beta.selection = (beta.selection + max - 1) % max
		}
		if k == 0x28 {
			beta.selection = (beta.selection + 1) % max
		}
		delta := 0.
		if k == 0x25 || k == 'A' {
			delta = -1
		}
		if k == 0x27 || k == 'D' {
			delta = 1
		}
		if k == VK_RETURN && beta.selection >= 7 && beta.selection <= 12 {
			beta.bindWaiting = beta.selection - 7
			return true
		}
		if k == VK_RETURN && (beta.selection == 5 || beta.selection == 6) {
			delta = 1
		}
		if k == VK_RETURN && beta.selection == 0 {
			delta = 1
		}
		if delta != 0 {
			switch beta.selection {
			case 0:
				app.boolMouse = false
			case 1:
				if delta != 0 {
					app.invertMouseX = !app.invertMouseX
				}
			case 2:
				app.sensitivity = engine.Clamp(app.sensitivity+delta*.15, .25, 3)
			case 3:
				beta.cfg.CameraSensitivity = engine.Clamp(beta.cfg.CameraSensitivity+delta*.15, .25, 3)
			case 4:
				app.camDist = engine.Clamp(app.camDist+delta*.3, 3.5, 12)
			case 5:
				app.quality = (app.quality + 3 + int(delta)) % 3
				invalidateWorldGeometry()
			case 6:
				toggleFullscreen()
			case 14:
				if remaster.Enabled {
					remaster.Time = (remaster.Time + 3 + int(delta)) % 3
					beta.cfg.TimeOfDay = remaster.Time + 1
					if app.world.Remaster != nil {
						app.world.Remaster.Time = remaster.Time
					}
				} else {
					beta.cfg.TimeOfDay = (beta.cfg.TimeOfDay + 4 + int(delta)) % 4
				}
			case 13:
				beta.cfg.UIScale = engine.Clamp(beta.cfg.UIScale+delta*.1, 1, 1.5)
			}
			saveCfg()
		}
	}
	return true
}
func legacyBetaClick(x, y int) {
	w, h := float64(app.width), float64(app.height)
	_ = h
	switch beta.screen {
	case screenTitle:
		if float64(x) >= w*.05 && float64(x) <= w*.43 {
			for i := 0; i < 4; i++ {
				top := h * (.375 + float64(i)*.105)
				if float64(y) >= top && float64(y) <= top+h*.082 {
					beta.selection = i
					betaInput(VK_RETURN)
					break
				}
			}
		}
	case screenMaps:
		if index, ok := nativeMenuHit("maps", x, y, app.width, app.height); ok {
			beta.mapSelect = index
			activateSelectedMap(false)
			saveCfg()
			return
		}
		if float64(x) >= w*.05 && float64(x) <= w*.45 && float64(y) >= h*.81 && float64(y) <= h*.91 {
			betaInput(VK_RETURN)
			return
		}
		if x < int(w*.18) && y > int(h*.88) {
			selectScreen(screenTitle)
		}
	case screenCars:
		if float64(y) >= h*.79 && float64(y) <= h*.91 && float64(x) < w*.46 {
			betaInput(VK_RETURN)
		} else if float64(x) < w*.26 && float64(y) > h*.49 && float64(y) < h*.56 {
			betaInput('A')
		} else if float64(x) > w*.27 && float64(x) < w*.45 && float64(y) > h*.49 && float64(y) < h*.56 {
			betaInput('D')
		}
	case screenPause:
		if float64(x) >= w*.07 && float64(x) <= w*.46 {
			for i := 0; i < 4; i++ {
				row := h * (.417 + float64(i)*.108)
				if float64(y) >= row-12 && float64(y) < row+h*.072 {
					beta.selection = i
					betaInput(VK_RETURN)
					break
				}
			}
		}
	case screenSettings:
		if y > int(h*.28) && y < int(h*.91) {
			row := int((float64(y) - h*.280) / (h * .041))
			beta.selection = engine.ClampInt(row, 0, 14)
			if beta.selection >= 7 && beta.selection <= 12 {
				betaInput(VK_RETURN)
			} else {
				betaInput(0x27)
			}
		}
	}
}
func modelFile(id int) string {
	switch id {
	case 0:
		return "challenge296.s3d"
	case 1:
		return "porsche911.s3d"
	case 2:
		return "mazda.s3d"
	case 3:
		return "truck.s3d"
	case 4:
		return "street_light.s3d"
	case 5:
		return "maple_tree.s3d"
	default:
		return "city_buildings.s3d"
	case 7:
		return "grass_chunks.s3d"
	}
}

// compileMesh creates immutable display lists once. The customizable body
// is stored separately in 12 lightness bands and tinted when drawn. This
// eliminates GPU display-list reconstruction when selecting paint in Garage.
func compileMesh(id int) {
	raw, err := assets.ReadFile("models/" + modelFile(id))
	if err != nil || len(raw) < 8 || string(raw[:4]) != "S3D1" {
		return
	}
	count := int(binary.LittleEndian.Uint32(raw[4:8]))
	if count < 3 || count > (len(raw)-8)/16 {
		return
	}
	if beta.mesh[id] != 0 {
		gl("glDeleteLists", uintptr(beta.mesh[id]), 1)
	}
	if id <= 1 {
		for i := range beta.paintMesh[id] {
			if beta.paintMesh[id][i] != 0 {
				gl("glDeleteLists", uintptr(beta.paintMesh[id][i]), 1)
			}
		}
	}
	var paintGeom [12][]float32
	beta.mesh[id] = uint32(gl("glGenLists", 1))
	gl("glNewList", uintptr(beta.mesh[id]), GL_COMPILE)
	dis(GL_TEXTURE_2D)
	begin(GL_TRIANGLES)
	for i := 0; i+2 < count; i += 3 {
		var pts [3][3]float64
		for j := 0; j < 3; j++ {
			p := 8 + (i+j)*16
			pts[j] = [3]float64{float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[p:]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[p+4:]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[p+8:])))}
		}
		e1, e2 := [3]float64{pts[1][0] - pts[0][0], pts[1][1] - pts[0][1], pts[1][2] - pts[0][2]}, [3]float64{pts[2][0] - pts[0][0], pts[2][1] - pts[0][1], pts[2][2] - pts[0][2]}
		nx, ny, nz := e1[1]*e2[2]-e1[2]*e2[1], e1[2]*e2[0]-e1[0]*e2[2], e1[0]*e2[1]-e1[1]*e2[0]
		norm := math.Sqrt(nx*nx+ny*ny+nz*nz) + 1e-9
		// Porsche import includes two large dark coplanar base triangles,
		// a fake floor/shadow that follows the vehicle everywhere.
		if id == 1 && math.Abs(ny)*.5 > 1.0 && math.Abs(ny)/norm > .985 &&
			pts[0][1] < .35 && pts[1][1] < .35 && pts[2][1] < .35 {
			avgColor := 0.0
			for j := 0; j < 3; j++ {
				off := 8 + (i+j)*16
				avgColor += float64(raw[off+12]) + float64(raw[off+13]) + float64(raw[off+14])
			}
			if avgColor/9 < .42*255 {
				continue
			}
		}
		meshFaceNormal(pts[0], pts[1], pts[2])
		shade := engine.Clamp(.69+.32*math.Abs(ny)/norm+.13*math.Abs(nz)/norm, .65, 1.16)
		if id <= 1 && raw[8+i*16+15] == 255 {
			band := engine.ClampInt(int(math.Round((shade-.65)/.51*11)), 0, 11)
			for j := 0; j < 3; j++ {
				paintGeom[band] = append(paintGeom[band], float32(pts[j][0]), float32(pts[j][1]), float32(pts[j][2]), float32(nx/norm), float32(ny/norm), float32(nz/norm))
			}
			continue
		}
		for j := 0; j < 3; j++ {
			p := 8 + (i+j)*16
			rr, gg, bb := float64(raw[p+12])/255, float64(raw[p+13])/255, float64(raw[p+14])/255
			if id >= 2 && id <= 3 {
				rr = math.Max(rr, .16)
				gg = math.Max(gg, .18)
				bb = math.Max(bb, .19)
			}
			col(engine.Clamp(rr*shade, 0, 1), engine.Clamp(gg*shade, 0, 1), engine.Clamp(bb*shade, 0, 1), 1)
			vtx(pts[j][0], pts[j][1], pts[j][2])
		}
	}
	end()
	gl("glEndList")
	if id <= 1 {
		for band, geom := range paintGeom {
			if len(geom) == 0 {
				continue
			}
			beta.paintMesh[id][band] = uint32(gl("glGenLists", 1))
			gl("glNewList", uintptr(beta.paintMesh[id][band]), GL_COMPILE)
			dis(GL_TEXTURE_2D)
			begin(GL_TRIANGLES)
			for i := 0; i < len(geom); i += 6 {
				normal := [3]float32{geom[i+3], geom[i+4], geom[i+5]}
				gl("glNormal3fv", uintptr(unsafe.Pointer(&normal[0])))
				vtx(float64(geom[i]), float64(geom[i+1]), float64(geom[i+2]))
			}
			end()
			gl("glEndList")
		}
	}
}

// Seven Mazda body finishes are compiled only once at startup. The source
// contains a restored red body rather than full GLB PBR materials, so this
// selectively retints its reddish hull while leaving glass/tyres dark.
func compileTrafficFinishes() {
	raw, err := assets.ReadFile("models/mazda.s3d")
	if err != nil || len(raw) < 8 || string(raw[:4]) != "S3D1" {
		return
	}
	count := int(binary.LittleEndian.Uint32(raw[4:8]))
	if count < 3 || count > (len(raw)-8)/16 {
		return
	}
	colors := [7][3]float64{{.92, .24, .16}, {.12, .47, .92}, {.92, .79, .19}, {.22, .85, .64}, {.87, .86, .81}, {.58, .27, .82}, {.96, .43, .22}}
	for k, c := range colors {
		beta.trafficMesh[k] = uint32(gl("glGenLists", 1))
		gl("glNewList", uintptr(beta.trafficMesh[k]), GL_COMPILE)
		dis(GL_TEXTURE_2D)
		begin(GL_TRIANGLES)
		for i := 0; i < count; i++ {
			if i%3 == 0 && i+2 < count {
				var pts [3][3]float64
				for q := 0; q < 3; q++ {
					off := 8 + (i+q)*16
					pts[q] = [3]float64{float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[off:]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[off+4:]))), float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[off+8:])))}
				}
				meshFaceNormal(pts[0], pts[1], pts[2])
			}
			j := 8 + i*16
			r, g, b := float64(raw[j+12])/255, float64(raw[j+13])/255, float64(raw[j+14])/255
			if r > .38 && r > g*1.9 && r > b*1.45 {
				luminance := engine.Clamp(.54+r*.63, .45, 1.13)
				col(engine.Clamp(c[0]*luminance, 0, 1), engine.Clamp(c[1]*luminance, 0, 1), engine.Clamp(c[2]*luminance, 0, 1), 1)
			} else {
				col(engine.Clamp(r, 0, 1), engine.Clamp(g, 0, 1), engine.Clamp(b, 0, 1), 1)
			}
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[j:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[j+4:])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[j+8:])))
			vtx(x, y, z)
		}
		end()
		gl("glEndList")
	}
}

// Model orientation is +Z forward after GLB hierarchy conversion. The old
// unconditional 180 degree rotation made the Ferrari point backward.
func drawImportedParts(id int, wheelSpin, steer float64) {
	if id < 2 && drawPBRCar(id, wheelSpin, steer) {
		return
	}
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.whiteTexture))
	if !drawBufferedCar(id) {
		gl("glCallList", uintptr(beta.mesh[id]))
	}
	if id <= 1 && !bufferedCars[id].available {
		// Dynamic paint for both player cars; body is separate from glass and tyres.
		color := paintPalette[beta.paint]
		dis(GL_TEXTURE_2D)
		for band, list := range beta.paintMesh[id] {
			if list == 0 {
				continue
			}
			brightness := .65 + float64(band)*.51/11
			col(engine.Clamp(color[0]*brightness, 0, 1), engine.Clamp(color[1]*brightness, 0, 1), engine.Clamp(color[2]*brightness, 0, 1), 1)
			gl("glCallList", uintptr(list))
		}
	}
	if id >= 2 {
		return
	}
	for i := 0; i < 4; i++ {
		setPBRMaterial(true, .06, .86)
		if beta.wheelMesh[id][i] == 0 && bufferedWheels[id][i].handle == 0 {
			continue
		}
		p := beta.wheelAnchor[id][i]
		wheelAngle, axleSteer := wheelSpin, 0.0
		suspensionTravel := 0.0
		if beta.screen == screenDrive {
			wheel := app.renderCar.Rig.Wheels[i]
			wheelAngle = wheel.RollAngle
			axleSteer = wheel.SteerAngle
			// Extend in air and rise with compression beneath the sprung body.
			spec := engine.VehicleSpec(id)
			suspensionTravel = engine.Clamp(wheel.Compression-spec.Mass*9.81/(4*spec.SpringRate), -.18, .22)
		} else if i < 2 {
			axleSteer = steer
		}
		push()
		move(p[0], p[1]+suspensionTravel, p[2])
		if i < 2 {
			rotate(axleSteer*180/math.Pi, 0, 1, 0)
		}
		// Axle is local X, so this spins the rim/tread while yaw steers it.
		rotate(wheelAngle*180/math.Pi, 1, 0, 0)
		if !drawBufferedWheel(id, i) {
			gl("glCallList", uintptr(beta.wheelMesh[id][i]))
		}
		pop()
	}
}
func renderNativeCar(id int, p engine.Vec, heading, wheelSpin, steer float64, trafficVariant int) {
	if id < 0 || id > 3 || (beta.mesh[id] == 0 && !(id < 2 && bufferedCars[id].available)) {
		return
	}
	setVehicleGloss(true)
	push()
	elevation := engine.CarElevation(engine.DistRoad(p, app.world.Routes), id)
	bodyY := engine.Terrain(p.X, p.Z) + elevation
	if id <= 1 && beta.screen == screenDrive {
		bodyY = app.renderCar.Body.Height - .725
		if id == 1 {
			bodyY -= .1
		}
	}
	move(p.X, bodyY, p.Z)
	rotate(heading*180/math.Pi, 0, 1, 0)
	if id <= 1 && beta.screen == screenDrive {
		rotate(app.renderCar.Rig.Pitch*180/math.Pi, 1, 0, 0)
		rotate(app.renderCar.Rig.Roll*180/math.Pi, 0, 0, 1)
	}
	if id == 2 {
		// Mazda source car is X-long (4.49m along X, 2.2m along Z).
		// Its rear spoiler is at +X, so the nose points toward -X.
		// Rotate +90 degrees: -X nose -> local +Z driving direction.
		rotate(90, 0, 1, 0)
	}
	if id == 2 && trafficVariant >= 0 && beta.trafficMesh[trafficVariant%7] != 0 {
		gl("glCallList", uintptr(beta.trafficMesh[trafficVariant%7]))
	} else {
		drawImportedParts(id, wheelSpin, steer)
	}
	// Small painted truck stripe: visual differentiation without modifying glass.
	if id == 3 {
		colors := [7][3]float64{{.93, .36, .15}, {.13, .65, .88}, {.85, .78, .20}, {.28, .82, .51}, {.78, .45, .80}, {.88, .88, .87}, {.94, .43, .28}}
		c := colors[trafficVariant%7]
		dis(GL_TEXTURE_2D)
		col(c[0]*.83, c[1]*.83, c[2]*.83, 1)
		box(-.94, 1.48, -1.55, .94, 1.60, 1.7)
	}
	if id <= 1 && beta.screen == screenDrive {
		drawPlayerLamps(id)
	}
	pop()
	setPBRMaterial(false, 0, .8)
	setVehicleGloss(false)
}
func compileWheelMeshes() {
	for car, model := range []string{"challenge296", "porsche911"} {
		metadata, err := assets.ReadFile("models/" + model + "_wheels.json")
		if err != nil {
			continue
		}
		var wheelData struct {
			Wheels map[string][3]float64 `json:"wheels"`
		}
		if json.Unmarshal(metadata, &wheelData) != nil {
			continue
		}
		for slot, key := range []string{"FL", "FR", "RL", "RR"} {
			beta.wheelAnchor[car][slot] = wheelData.Wheels[key]
			if bufferedWheels[car][slot].handle != 0 {
				continue
			}
			filename := "models/" + model + "_wheel_" + key + ".s3d"
			raw, err := assets.ReadFile(filename)
			if err != nil || len(raw) < 8 || string(raw[:4]) != "S3D1" {
				continue
			}
			count := int(binary.LittleEndian.Uint32(raw[4:8]))
			if count <= 0 || count > (len(raw)-8)/16 {
				continue
			}
			// v0.5.3: GLB wheel submesh positions were not reliably local.
			// In particular the Porsche's front and rear wheel vertices had
			// sizeable offsets, making the tire orbit instead of spinning.
			pivot, ok := s3dWheelPivot(raw)
			if !ok {
				continue
			}
			beta.wheelMesh[car][slot] = uint32(gl("glGenLists", 1))
			gl("glNewList", uintptr(beta.wheelMesh[car][slot]), GL_COMPILE)
			dis(GL_TEXTURE_2D)
			begin(GL_TRIANGLES)
			for i := 0; i+2 < count; i += 3 {
				var tri [3][3]float64
				for j := 0; j < 3; j++ {
					idx := 8 + (i+j)*16
					for axis := 0; axis < 3; axis++ {
						tri[j][axis] = float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[idx+4*axis:]))) - pivot[axis]
					}
				}
				meshFaceNormal(tri[0], tri[1], tri[2])
				for j := 0; j < 3; j++ {
					idx := 8 + (i+j)*16
					col(float64(raw[idx+12])/255, float64(raw[idx+13])/255, float64(raw[idx+14])/255, 1)
					vtx(tri[j][0], tri[j][1], tri[j][2])
				}
			}
			end()
			gl("glEndList")
		}
	}
}
func drawBetaHUD() {
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	w, h := float64(app.width), float64(app.height)
	m := [16]float32{float32(2 / w), 0, 0, 0, 0, float32(-2 / h), 0, 0, 0, 0, -1, 0, -1, 1, 0, 1}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&m[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	en(GL_BLEND)
	hudRect(18, 18, 370, 95, .025, .05, .075, .84)
	hudText("SPEED IN NEED", 35, 28, 29, white)
	hudText("BETA v0.1", 36, 70, 17, cyan)
	drawMinimap()
	hudRect(18, 120, 295, 30, .025, .055, .07, .84)
	hudText("HANDLING: "+[]string{"TOUR", "SPORT", "DRIFT"}[app.world.DriveMode]+"  [T]", 34, 124, 17, cyan)
	c := app.world.Car
	// Semi-circular speed and nitro dial built natively from geometry.
	cx, cy := w-155, h-125.
	hudRect(w-330, h-266, 311, 244, .018, .034, .055, .84)
	dis(GL_TEXTURE_2D)
	for ring := 0; ring < 2; ring++ {
		if ring == 0 {
			col(.16, .28, .31, 1)
		} else {
			col(.98, .44, .13, 1)
		}
		begin(GL_LINE_STRIP)
		start := 2.45
		length := math.Pi * 1.65
		points := 84
		if ring == 1 {
			points = int(engine.Clamp(math.Abs(c.Speed)*3.6/260., 0, 1)*84) + 1
		}
		for i := 0; i < points; i++ {
			a := start + length*float64(i)/84
			r := 97.
			vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
		}
		end()
	}
	hudText(fmt.Sprintf("%03.0f", math.Abs(c.Speed)*3.6), cx-75, cy-45, 56, white)
	hudText("KM/H", cx-27, cy+14, 18, cyan)
	if c.Speed < -.05 {
		hudText("R", cx+48, cy+16, 21, [4]float64{1, .52, .18, 1})
	}
	hudText(fmt.Sprintf("NITRO %02.0f%%", c.Nitro), w-294, h-88, 19, white)
	hudRect(w-292, h-49, 257, 13, .12, .22, .26, .9)
	hudRect(w-292, h-49, 2.57*c.Nitro, 13, .10, .89, 1, 1)
	hudRect(17, h-163, 395, 142, .02, .04, .065, .81)
	title := []string{"FREE DRIVE", "SPRINT RACE", "DRIFT CHALLENGE", "TRAFFIC RUSH"}[beta.mode]
	hudText(title, 34, h-151, 23, white)
	if beta.mode == 1 {
		hudText(fmt.Sprintf("CHECKPOINT %d / %d", app.world.Checkpoint, len(app.world.Checkpoints)), 34, h-115, 22, cyan)
	} else if beta.mode == 2 {
		hudText(fmt.Sprintf("DRIFT  %06.0f", c.DriftScore), 34, h-115, 23, cyan)
	} else {
		hudText(fmt.Sprintf("NEAR MISSES  %d", app.world.NearMisses), 34, h-115, 20, cyan)
	}
	hudText(fmt.Sprintf("CREDITS  %d    DRIFT %.0f", app.world.Credits, c.DriftScore), 34, h-75, 17, white)
	hudText(fmt.Sprintf("FPS %.0f  |  %s  |  %s", app.floatFPS, func() string {
		if app.boolMouse {
			return "MOUSE"
		}
		return "WASD"
	}(), func() string {
		if app.boolFPP {
			return "FPP"
		}
		return "TPP"
	}()), 34, h-49, 17, white)
	hudText(fmt.Sprintf("GEAR %s   RPM %04.0f", func() string {
		if c.Gearbox.Gear < 0 {
			return "R"
		}
		return fmt.Sprint(c.Gearbox.Gear)
	}(), c.Gearbox.RPM), w-295, h-245, 16, cyan)
	if app.showPerf {
		hudRect(w-420, 145, 400, 235, .014, .026, .035, .94)
		hudText("REMASTER PERFORMANCE  F2", w-347, 155, 17, cyan)
		hudText(fmt.Sprintf("FRAME %5.2f MS  FPS %3.0f", app.frameMS, app.floatFPS), w-347, 187, 16, white)
		hudText(fmt.Sprintf("CPU SIM %5.2f MS  RENDER %5.2f MS", app.simMS, app.renderMS), w-347, 211, 16, white)
		hudText(fmt.Sprintf("TRAFFIC %d   TREES %d", len(app.world.Traffic), app.world.VisibleTrees), w-347, 235, 16, white)
		gpuLine := "GPU TIME N/A (queries unsupported)"
		if app.gpuTiming {
			gpuLine = fmt.Sprintf("GPU TIME %.2f MS (async query)", app.gpuMS)
		}
		hudText(gpuLine, w-347, 260, 14, cyan)
		hudText(fmt.Sprintf("VSYNC %t  MAP %d", app.vsync, app.mapID+1), w-347, 282, 15, white)
		hudText(fmt.Sprintf("1%% LOW %.1f FPS  P99 %.2f MS", profiler.summary.LowOnePercentFPS, profiler.summary.P99MS), w-405, 308, 15, cyan)
		hudText(fmt.Sprintf("P95 %.2f MS  WORLD DRAWS %d", profiler.summary.P95MS, gpuWorld.draws), w-405, 332, 15, white)
		hudText("F6 - RECORD 60 SECOND BENCHMARK", w-405, 355, 14, white)
	}
	if profiler.active {
		hudText(fmt.Sprintf("BENCHMARK %.0f / 60 SEC", profiler.seconds), w*.38, h*.10, 18, cyan)
	}
	if !profiler.active && profiler.message != "" {
		hudText(profiler.message, w*.31, h*.10, 15, cyan)
	}
	if !c.Body.Grounded {
		hudText(fmt.Sprintf("AIRTIME %.1f S  JUMP %.0f M", c.Body.AirTime, c.Body.JumpDistance), w*.37, h*.20, 20, cyan)
	}
	if c.Boosting {
		hudText("NITRO!", w*.43, h*.15, 35, cyan)
	}
	if app.showHelp {
		hudRect(w*.20, h*.16, w*.60, h*.64, .012, .025, .035, .96)
		msg := []string{"CONTROLS    F1 TO HIDE", "W ACCEL  /  S BRAKE THEN REVERSE", "A / D   TURN LEFT / RIGHT", "SPACE   HANDBRAKE & SMOKE", "SHIFT   LIMITED NITRO", "V  CHANGE CAMERA / F RADIO", "M  TOGGLE MOUSE STEERING", "T  CYCLE HANDLING", "RIGHT MOUSE DRAG  ORBIT", "F2 FPS  F3 COLLISIONS  F4 ROUTES  F5 WHEELS", "H  HEADLIGHTS ON / OFF", "ESC  PAUSE"}
		for i, s := range msg {
			hudText(s, w*.25, h*.20+float64(i)*39, 21, white)
		}
	}
	en(GL_DEPTH_TEST)
}
func drawBetaMenuScene() {
	if app.width <= 0 || app.height <= 0 {
		return
	}
	if beta.screen == screenMaps {
		drawMapPreview()
		return
	}

	beginHDRScene()
	gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	perspective(43)
	var projection [16]float32
	gl("glGetFloatv", 0x0BA7, uintptr(unsafe.Pointer(&projection[0])))
	projection[8] = -.48
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&projection[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	cam := pos3{4.5, 2.7, 7.4}
	applyCamera(cam, pos3{0, .85, 0})
	gl("glDepthMask", 0)
	oldMap := app.mapID
	if beta.screen == screenMaps {
		app.mapID = beta.mapSelect
	} else {
		app.mapID = 1
	}
	drawAtmosphere(cam)
	app.mapID = oldMap
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
	// Keep the native showroom in legacy mode; gameplay fog values are only
	// configured while driving, so GLSL is activated for gameplay scenes only.
	useLighting(false)
	useLighting(app.lightingOn)
	updateSceneLights(app.world.Car)
	// Native showroom shares the vehicle material renderer.
	en(GL_TEXTURE_2D)
	bind("asphalt.jpg")
	col(.65, .71, .78, 1)
	begin(GL_QUADS)
	tq(pos3{-100, -.22, -100}, pos3{100, -.22, -100}, pos3{100, -.22, 100}, pos3{-100, -.22, 100}, 0, 0, 16, 16)
	end()
	// Stationary neon skyline for the title screen only (no live world simulation).
	dis(GL_TEXTURE_2D)
	for i := 0; i < 22; i++ {
		x := float64((i*17)%85 - 42)
		z := -22 - float64((i*13)%32)
		height := float64(8 + (i*11)%28)
		push()
		move(x, 0, z)
		col(.035, .061, .115, 1)
		box(-2.8, 0, -3.2, 2.8, height, 3.2)
		for floor := 2.; floor < height; floor += 3.4 {
			if (int(floor)+i)%3 == 0 {
				col(.16, .70, .96, 1)
			} else {
				col(.92, .32, .68, 1)
			}
			for xx := -2.; xx < 2.4; xx += 1.55 {
				box(xx, floor, 3.21, xx+.64, floor+.8, 3.26)
			}
		}
		pop()
	}

	push()
	move(0, 0, 0)
	useLighting(app.lightingOn)
	rotate(beta.carTurn*13, 0, 1, 0)
	if beta.mesh[beta.carSelect] != 0 || bufferedCars[beta.carSelect].available {
		drawImportedParts(beta.carSelect, beta.carTurn*.55, 0)
	}
	pop()
	useLighting(false)
	finishHDRScene()
	drawBetaMenuUI()
}
func drawLegacyBetaMenuUI() {
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	w, h := float64(app.width), float64(app.height)
	m := [16]float32{float32(2 / w), 0, 0, 0, 0, float32(-2 / h), 0, 0, 0, 0, -1, 0, -1, 1, 0, 1}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&m[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	en(GL_BLEND)
	hudRect(0, 0, w, h, .012, .020, .040, .26)
	hudRect(0, 0, w*.49, h, .014, .028, .046, .91)
	hudRect(0, 0, 6, h, .08, .88, 1, 1)
	hudText("SPEED IN NEED", w*.055, h*.08, 41, white)
	hudText("BETA v0.1", w*.057, h*.17, 20, cyan)
	switch beta.screen {
	case screenTitle:
		hudText("CHOOSE YOUR NEXT RUN", w*.057, h*.30, 23, white)
		buttons := []string{"START", "GARAGE", "SETTINGS", "QUIT"}
		for i, b := range buttons {
			y := h * (.39 + float64(i)*.105)
			active := i == beta.selection
			if active {
				hudRect(w*.05, y-.015*h, w*.37, h*.082, .08, .31, .41, .92)
			} else {
				hudRect(w*.05, y-.015*h, w*.37, h*.082, .04, .08, .13, .80)
			}
			hudText(b, w*.078, y, 28, func() [4]float64 {
				if active {
					return cyan
				}
				return white
			}())
		}
		hudText("ARROWS / ENTER  -  OR CLICK", w*.055, h*.82, 18, white)
	case screenMaps:
		hudText("SELECT YOUR MAP", w*.057, h*.285, 34, white)
		for i, n := range []string{"COASTAL SUNSET", "NEON CITY", "CANYON ROADS"} {
			y := h * (.385 + float64(i)*.135)
			if beta.mapSelect == i {
				hudRect(w*.053, y-h*.016, w*.39, h*.10, .05, .27, .35, .90)
			} else {
				hudRect(w*.053, y-h*.016, w*.39, h*.10, .05, .075, .11, .81)
			}
			hudText(fmt.Sprintf("0%d  %s", i+1, n), w*.066, y, 25, white)
			state := "AVAILABLE"
			if beta.cfg.Unlocks&(1<<i) != 0 {
				state = "AVAILABLE"
			}
			hudText(state, w*.078, y+35, 15, cyan)
		}
		hudRect(w*.053, h*.815, w*.39, h*.085, .05, .30, .38, .96)
		hudText("ENTER / CLICK TO CONTINUE", w*.067, h*.83, 22, white)
		hudText("ARROWS SELECT  -  ESC BACK", w*.058, h*.925, 16, white)
	case screenCars:
		hudText("YOUR GARAGE", w*.057, h*.29, 35, white)
		name := []string{"296 CHALLENGE", "911 TURBO CLASSIC"}[beta.carSelect]
		hudText(name, w*.06, h*.41, 30, cyan)
		hudText("CAR HANDLING: TRACK / CLASSIC", w*.06, h*.465, 15, white)
		hudText("A / D  CHANGE VEHICLE", w*.06, h*.51, 20, white)
		hudText("Q / E  CHANGE BODY PAINT", w*.06, h*.57, 20, white)
		mode := []string{"FREE DRIVE", "SPRINT", "DRIFT TRIAL", "TRAFFIC RUSH"}[beta.mode]
		hudText("N  MODE: "+mode, w*.06, h*.67, 22, cyan)
		hudRect(w*.052, h*.80, w*.37, h*.09, .06, .27, .34, .91)
		hudText("ENTER   START DRIVING", w*.062, h*.816, 26, white)
	case screenSettings:
		hudText("SETTINGS", w*.057, h*.235, 34, white)
		steering := "WASD"
		if app.boolMouse {
			steering = "MOUSE"
		}
		labels := []string{"STEERING MODE", "INVERT MOUSE X", "STEERING SENS", "CAMERA SENS", "CAMERA DISTANCE", "GRAPHICS", "FULLSCREEN F11", "BIND ACCEL", "BIND BRAKE", "BIND LEFT", "BIND RIGHT", "BIND NITRO", "BIND DRIFT", "UI SCALE", "TIME OF DAY"}
		vals := []string{steering, func() string {
			if app.invertMouseX {
				return "ON"
			}
			return "OFF"
		}(), fmt.Sprintf("%.2f", app.sensitivity), fmt.Sprintf("%.2f", beta.cfg.CameraSensitivity), fmt.Sprintf("%.1f M", app.camDist), []string{"LOW", "BALANCED", "HIGH"}[app.quality], func() string {
			if app.fullscreen {
				return "ON"
			}
			return "OFF"
		}(), bindingName(beta.cfg.Bindings[0]), bindingName(beta.cfg.Bindings[1]), bindingName(beta.cfg.Bindings[2]), bindingName(beta.cfg.Bindings[3]), bindingName(beta.cfg.Bindings[4]), bindingName(beta.cfg.Bindings[5]), fmt.Sprintf("%.0f%%", beta.cfg.UIScale*100), []string{"MAP DEFAULT", "DAY", "SUNSET", "NIGHT"}[beta.cfg.TimeOfDay%4]}
		for i, l := range labels {
			y := h*.280 + float64(i)*h*.041
			if i == beta.selection {
				hudRect(w*.049, y-3, w*.405, h*.048, .04, .21, .29, .92)
			}
			hudText(l, w*.058, y, 19, white)
			hudText(vals[i], w*.30, y, 19, cyan)
		}
		if beta.bindWaiting >= 0 {
			hudText("PRESS NEW KEY ...", w*.058, h*.925, 22, cyan)
		} else {
			hudText("ARROWS ADJUST   F11 FULLSCREEN   ESC BACK", w*.056, h*.925, 16, white)
		}
	case screenPause:
		// Native UI: translucent charcoal panel with orange/cyan neon accents.
		hudRect(w*.043, h*.205, w*.435, h*.687, .012, .027, .050, .96)
		hudRect(w*.043, h*.205, 5, h*.687, .98, .40, .12, 1)
		hudRect(w*.052, h*.213, w*.415, 3, .10, .82, 1, .95)
		hudText("GAME PAUSED", w*.073, h*.252, 46, white)
		hudText("SPEED IN NEED  /  BETA v0.1", w*.077, h*.346, 19, cyan)
		items := []string{"RESUME RACE", "SETTINGS", "RESTART DRIVE", "MAIN MENU"}
		cmds := []string{"ENTER / ESC", "S", "R", "Q"}
		for i, item := range items {
			y := h * (.417 + float64(i)*.108)
			active := i == beta.selection
			if active {
				hudRect(w*.071, y-10, w*.375, h*.078, .08, .34, .43, .94)
				hudRect(w*.071, y-10, 4, h*.078, .97, .40, .16, 1)
			} else {
				hudRect(w*.071, y-10, w*.375, h*.078, .045, .086, .13, .92)
			}
			hudText(item, w*.09, y, 26, func() [4]float64 {
				if active {
					return cyan
				}
				return white
			}())
			hudText(cmds[i], w*.344, y+5, 14, [4]float64{.98, .65, .33, 1})
		}
		hudText("UP/DOWN + ENTER OR CLICK", w*.079, h*.86, 17, white)

	case screenResults:
		hudText("RACE COMPLETE", w*.059, h*.32, 39, cyan)
		hudText(fmt.Sprintf("TIME  %.1f SEC", app.world.ChallengeTime), w*.06, h*.46, 23, white)
		hudText(fmt.Sprintf("REWARD  %d CREDITS", app.world.Credits), w*.06, h*.54, 24, white)
		hudText("ENTER - MAP SELECT", w*.06, h*.73, 24, cyan)
	}
	hudText("ESC BACK  |  BETA v0.1", w*.057, h*.955, 15, white)
	en(GL_DEPTH_TEST)
}
func checkFinish() {
	if remaster.Enabled {
		if app.world.Remaster != nil && app.world.Remaster.Finished && beta.screen == screenDrive {
			selectScreen(screenResults)
			app.started = false
		}
		return
	}
	if !app.world.Finished || beta.screen != screenDrive {
		return
	}
	if beta.mapSelect == 0 && beta.mode == 1 {
		beta.cfg.Unlocks |= 6
		saveCfg()
	}
	selectScreen(screenResults)
	app.started = false
}
func localClock() { _ = strings.TrimSpace(time.Now().String()) }
