package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"unsafe"
)

var garage033 bool

func init()                     { cyan = [4]float64{.79, .94, .57, 1}; amber = [4]float64{.87, .88, .79, 1} }
func drawTitle033(w, h float64) { drawTitle038(w, h) }
func drawModeChoice033Legacy(w, h float64) {
	for i, item := range [][3]string{{"OPEN WORLD", "CITY / COAST / FOREST / HILLS", "FIND YOUR OWN ROAD"}, {"CIRCUIT RACING", "LAPS / RIVAL AI / NITRO", "RACE THE COAST"}} {
		x := w * (.07 + float64(i)*.365)
		roundRect032(x, h*.27, w*.34, h*.43, 9, [4]float64{.028, .033, .038, .94})
		beta021Icon(i, x+w*.17, h*.40, h*.10)
		remasterText(fmt.Sprintf("0%d", i+1), x+w*.025, h*.30, h*.018, amber)
		beta020Button(fmt.Sprintf("mode%d", i), item[0], x+w*.018, h*.49, w*.304, h*.072, remaster.Mode == i)
		remasterText(item[1], x+w*.025, h*.59, h*.012, white)
		remasterText(item[2], x+w*.025, h*.64, h*.011, amber)
	}
	remasterText("MULTIPLAYER  /  LOCKED", w*.08, h*.755, h*.011, [4]float64{.58, .60, .61, 1})
}
func drawMapChoice033(w, h float64) { drawMapChoice036(w, h); drawSetupSpawnRefine(w, h) }

func drawCarChoice033(w, h float64) { drawCarChoice036(w, h) }

func italicSpeed033(label string, x, y, size float64, c [4]float64) {
	push()
	move(x+size*.1, y, 0)
	matrix := [16]float32{1, 0, 0, 0, -.13, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	gl("glMultMatrixf", uintptr(unsafe.Pointer(&matrix[0])))
	remasterText(label, 0, 0, size, c)
	pop()
}
func drawSpeed033Legacy(c engine.Car) {
	w, h := float64(app.width), float64(app.height)
	if autopilot034 {
		roundRect032(w*.40, h*.06, w*.20, h*.04, 5, [4]float64{.015, .08, .07, .9})
		remasterText("AP ACTIVE  /  A TO RELEASE", w*.415, h*.07, h*.012, cyan)
	}
	cx, cy, r := w*.90, h*.80, h*.117
	circle033(cx, cy, r*1.03, [4]float64{.014, .018, .023, .31})
	angle := func(v float64) float64 { return (142 + v*256) * math.Pi / 180 }
	gl("glLineWidth", b32(2))
	col(1, .60, .20, .82)
	begin(0x0003)
	for i := 0; i <= 90; i++ {
		a := angle(float64(i) / 90)
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	for i := 0; i <= 45; i++ {
		a := angle(float64(i) / 45)
		inner := .91
		if i%5 == 0 {
			inner = .83
		}
		color := [4]float64{1, 1, 1, .85}
		if i >= 35 {
			color = amber
		}
		col(color[0], color[1], color[2], color[3])
		begin(0x0001)
		vtx(cx+math.Cos(a)*r*inner, cy+math.Sin(a)*r*inner, 0)
		vtx(cx+math.Cos(a)*r*.97, cy+math.Sin(a)*r*.97, 0)
		end()
		if i%5 == 0 {
			label := fmt.Sprint(i / 5)
			remasterText(label, cx+math.Cos(a)*r*.68-h*.006, cy+math.Sin(a)*r*.68-h*.007, h*.012, white)
		}
	}
	rpm := engine.Clamp(c.Gearbox.RPM/9000, 0, 1)
	a := angle(rpm)
	col(1, .60, .20, .95)
	gl("glLineWidth", b32(2))
	begin(0x0001)
	vtx(cx+math.Cos(a)*r*.16, cy+math.Sin(a)*r*.16, 0)
	vtx(cx+math.Cos(a)*r*.78, cy+math.Sin(a)*r*.78, 0)
	end()
	circle033(cx, cy, 4, amber)
	label := fmt.Sprintf("%d", int(math.Abs(c.Speed)*3.6))
	size := h * .058
	italicSpeed033(label, cx-remasterTextWidth(label, size)*.5, cy+r*.12, size, white)
	remasterText("KM/H", cx-r*.20, cy+r*.73, h*.010, white)
	gear := "N"
	if c.Gearbox.Gear < 0 {
		gear = "R"
	} else if c.Gearbox.Gear > 0 {
		gear = fmt.Sprint(c.Gearbox.Gear)
	}
	remasterText(gear, cx-r*.79, cy+r*.28, h*.024, amber)
	hudRect(cx-r*.72, cy+r*1.13, r*1.44, 3, .18, .18, .19, .7)
	hudRect(cx-r*.72, cy+r*1.13, r*1.44*engine.Clamp(c.Nitro/100, 0, 1), 3, 1, .58, .16, .95)
	gl("glLineWidth", b32(1))
}
func drawGarage033(dt float64) {
	normalizePlayer052()
	garage033 = true
	defer func() { garage033 = false }()
	ensureSkyContext051()
	beginHDRScene()
	perspective(42)
	gl("glViewport", 0, 0, uintptr(hdrRenderWidth()), uintptr(hdrRenderHeight()))
	gl("glClearColor", b32(.034), b32(.035), b32(.039), b32(1))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	gl("glLoadIdentity")
	eye, target := beta014PreviewCamera(dt)
	applyCamera(eye, target)
	app.cam = eye
	useLighting(true)
	updateSceneLights(engine.Car{})
	sceneInt("uFogEnabled", 0)
	sceneInt("uShadowEnabled", 0)
	sceneFloat("uNight", .12)
	sceneFloat("uSunset", .15)
	if pbrAssets["b051_workshop"] == nil {
		loadPBRAsset("b051_workshop")
		loadPBRAsset("b051_workshop_floor")
	}
	drawPBRAsset("b051_workshop_floor", 0)
	drawPBRAsset("b051_workshop", 0)
	beta013PreviewContactShadow()
	c := engine.Car{}
	c.Body.Height = .74
	if garageView037() && garage037.Touched {
		c.Heading = garage037.Yaw
	}
	drawRemasterCar(c, beta.carSelect, true)
	useLighting(false)
	finishHDRScene()
}
