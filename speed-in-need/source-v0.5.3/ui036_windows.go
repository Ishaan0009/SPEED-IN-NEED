package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"time"
)

const GL_CULL_FACE = 0x0B44

func titleWord036(text string, x, y, size, width float64, color [4]float64) {
	if actual := remasterTextWidth(text, size); actual > width {
		size *= width / actual
	}
	for i := 9; i >= 1; i-- {
		d := float64(i) * size * .014
		remasterText(text, x+d, y+d, size, [4]float64{.015 + .004*float64(i), .10, .14, 1})
	}
	remasterText(text, x-size*.007, y-size*.007, size, [4]float64{.75, .98, 1, 1})
	remasterText(text, x, y, size, color)
}
func drawTitle036(w, h float64) {
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.006, .014, .027, .99)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.012, .03, .055, .04)
	vtx(w*.70, h, 0)
	vtx(w*.70, 0, 0)
	end()
	hudRect(w*.065, h*.10, w*.065, 3, .12, .86, 1, 1)
	remasterText("OPEN ROAD SERIES  /  036", w*.065, h*.14, h*.013, cyan)
	titleWord036("SPEED", w*.06, h*.225, h*.104, w*.40, [4]float64{.18, .87, 1, 1})
	titleWord036("IN NEED", w*.06, h*.35, h*.104, w*.40, [4]float64{1, .52, .22, 1})
	remasterText("CITY STREETS. COAST ROADS. HIGHER GROUND.", w*.065, h*.515, h*.012, white)
	for i, item := range [][2]string{{"start", "START DRIVING"}, {"settings", "SETTINGS"}, {"quit", "EXIT GAME"}} {
		beta020Button(item[0], item[1], w*.065, h*(.61+float64(i)*.092), w*.275, h*.065, i == 0)
	}
	remasterText("V0.5.3  /  NATIVE EDITION", w*.065, h*.93, h*.012, cyan)
	remasterText("MAKE EVERY CORNER YOURS.", w*.70, h*.93, h*.012, white)
}

func drawMenuBackdrop036() {
	glExt("glBindFramebuffer", 0x8D40, 0)
	hdrScene.active = false
	gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
	gl("glClearColor", b32(.009), b32(.018), b32(.032), b32(1))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.014, .028, .045, 1)
	vtx(0, 0, 0)
	vtx(w, 0, 0)
	col(.025, .065, .085, 1)
	vtx(w, h, 0)
	vtx(0, h, 0)
	end()
	col(.08, .22, .28, .32)
	begin(0x0001)
	for i := 0; i < 18; i++ {
		x := float64(i) * w / 12
		vtx(x, h, 0)
		vtx(w*.54+(x-w*.54)*.14, h*.19, 0)
	}
	for i := 0; i < 9; i++ {
		y := h * (.20 + float64(i*i)*.012)
		vtx(0, y, 0)
		vtx(w, y, 0)
	}
	end()
}
func drawMapChoice036(w, h float64) {
	race := remaster.Mode == 1
	title, description, distance := "OPEN ROAD REGION", "CITY / COAST / FOREST / HIGHLANDS", "9 KM2 OPEN WORLD"
	if race {
		title = "COASTAL CIRCUIT"
		description = "COASTAL STRAIGHTS / TECHNICAL CORNERS"
		distance = "3.73 KM / CLOSED CIRCUIT"
	}
	roundRect032(w*.055, h*.25, w*.34, h*.49, 8, [4]float64{.016, .033, .05, .98})
	headline035(title, w*.078, h*.29, h*.027, w*.30, white)
	headline035(description, w*.078, h*.38, h*.013, w*.29, cyan)
	headline035(distance, w*.078, h*.47, h*.022, w*.29, white)
	if race {
		remasterText("CHOOSE LAPS AND RIVALS NEXT", w*.078, h*.60, h*.012, amber)
	} else {
		remasterText("ROADS AND WAYPOINTS CONNECTED", w*.078, h*.60, h*.012, amber)
	}
	size := h * .56
	x, y := w*.48, h*.215
	roundRect032(x-8, y-8, size+16, size+16, 8, [4]float64{.12, .22, .27, 1})
	// Map preview is an opaque UI pass, independent of the orbiting car's states.
	glExt("glActiveTexture", 0x84C0)
	dis(GL_CULL_FACE)
	dis(GL_DEPTH_TEST)
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(mapTexture033()))
	col(1, 1, 1, 1)
	u0, v0, u1, v1 := 0., 0., 1., 1.
	if race {
		if d := remasterDefinition(1); d != nil && len(d.Roads) > 0 {
			minX, minZ, maxX, maxZ := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
			for _, road := range d.Roads {
				for _, p := range road.Points {
					minX = math.Min(minX, p.X)
					maxX = math.Max(maxX, p.X)
					minZ = math.Min(minZ, p.Z)
					maxZ = math.Max(maxZ, p.Z)
				}
			}
			span := math.Max(maxX-minX, maxZ-minZ) + 220
			cx, cz := (maxX+minX)*.5, (maxZ+minZ)*.5
			u0 = (cx - span*.5 - d.Bounds[0].X) / (d.Bounds[1].X - d.Bounds[0].X)
			u1 = (cx + span*.5 - d.Bounds[0].X) / (d.Bounds[1].X - d.Bounds[0].X)
			v0 = (cz - span*.5 - d.Bounds[0].Z) / (d.Bounds[1].Z - d.Bounds[0].Z)
			v1 = (cz + span*.5 - d.Bounds[0].Z) / (d.Bounds[1].Z - d.Bounds[0].Z)
		}
	}
	begin(GL_QUADS)
	tex(u0, v0)
	vtx(x, y, 0)
	tex(u1, v0)
	vtx(x+size, y, 0)
	tex(u1, v1)
	vtx(x+size, y+size, 0)
	tex(u0, v1)
	vtx(x, y+size, 0)
	end()
	dis(GL_TEXTURE_2D)
}
func drawCarChoice036Legacy(w, h float64) {
	spec := engine.RemasteredVehicleSpec(beta.carSelect)
	roundRect032(w*.045, h*.245, w*.28, h*.52, 8, [4]float64{.012, .026, .039, .97})
	headline035(spec.Name, w*.065, h*.275, h*.026, w*.24, white)
	labels := []string{"TOP SPEED", "POWER", "TORQUE", "POWER / WEIGHT", "ROAD GRIP"}
	values := []float64{spec.TopSpeed * 3.6, spec.PowerKW * 1.341, spec.P.PeakTorque, spec.PowerKW * 1.341 / spec.P.Mass * 1000, spec.Mu}
	maxima := []float64{0, 0, 0, 0, 0}
	for id := 0; id < 4; id++ {
		s := engine.RemasteredVehicleSpec(id)
		v := []float64{s.TopSpeed * 3.6, s.PowerKW * 1.341, s.P.PeakTorque, s.PowerKW * 1.341 / s.P.Mass * 1000, s.Mu}
		for j := range v {
			maxima[j] = math.Max(maxima[j], v[j])
		}
	}
	units := []string{"KM/H", "BHP", "NM", "BHP/T", "MU"}
	for i, label := range labels {
		y := h * (.345 + float64(i)*.070)
		remasterText(label, w*.065, y, h*.011, cyan)
		value := fmt.Sprintf("%.0f %s", values[i], units[i])
		if i == 4 {
			value = fmt.Sprintf("%.2f %s", values[i], units[i])
		}
		remasterText(value, w*.205, y, h*.011, white)
		hudRect(w*.065, y+h*.028, w*.235, h*.007, .10, .16, .20, 1)
		hudRect(w*.065, y+h*.028, w*.235*values[i]/maxima[i], h*.007, .14, .79, .86, 1)
	}
	drive := "RWD"
	if spec.AWD {
		drive = "AWD"
	}
	remasterText(fmt.Sprintf("%.0f KG  /  %s", spec.P.Mass, drive), w*.065, h*.71, h*.012, white)
	beta020Button("car-", "<", w*.335, h*.46, w*.04, h*.065, false)
	beta020Button("car+", ">", w*.945, h*.46, w*.04, h*.065, false)
	for i, label := range []string{"296 CHALLENGE", "911 TURBO", "CUSTOM GT", "TRUENO"} {
		beta020Button(fmt.Sprintf("car%d", i), label, w*(.35+float64(i)*.156), h*.755, w*.145, h*.057, beta.carSelect == i)
	}
	remasterText("DRAG TO ROTATE  /  WHEEL TO ZOOM", w*.38, h*.60, h*.011, white)
	if beta.carSelect == 3 {
		remasterText("ORIGINAL LIVERY", w*.58, h*.69, h*.012, amber)
		return
	}
	remasterText("PAINT", w*.37, h*.66, h*.011, cyan)
	for i, p := range paintPalette {
		columns := 12
		row, col := i/columns, i%columns
		x, y := w*(.43+float64(col)*.036), h*(.645+float64(row)*.039)
		bw, bh := w*.025, h*.027
		ui020.Buttons = append(ui020.Buttons, button020{fmt.Sprintf("paint:%d", i), x, y, bw, bh})
		hudRect(x, y, bw, bh, p[0], p[1], p[2], 1)
		if i == beta.paint {
			hudRect(x-2, y-2, bw+4, 2, 1, .75, .32, 1)
			hudRect(x-2, y+bh, bw+4, 2, 1, .75, .32, 1)
		}
	}
}

var loadingStart036 = time.Now()

func drawLoading036Legacy(progress float64, label string) {
	drawMenuBackdrop036()
	w, h := float64(app.width), float64(app.height)
	t := time.Since(loadingStart036).Seconds()
	progress = engine.Clamp(progress, 0, 1)
	titleWord037("SPEED IN NEED", w*.12, h*.23, h*.073, w*.76)
	remasterText("PREPARING YOUR DRIVE", w*.12, h*.39, h*.018, white)
	headline035(label, w*.12, h*.50, h*.016, w*.75, cyan)
	hudRect(w*.12, h*.58, w*.76, h*.012, .08, .15, .20, 1)
	hudRect(w*.12, h*.58, w*.76*progress, h*.012, .12, .84, .89, 1)
	x := w*.12 + math.Mod(t*.18, 1)*w*.73
	hudRect(x, h*.58, w*.03, h*.012, .75, 1, 1, .65)
	remasterText(fmt.Sprintf("%02d%%", int(progress*100)), w*.12, h*.63, h*.021, white)
	remasterText("MOUSE WHEEL: ZOOM  /  R3 + STICK: ZOOM", w*.12, h*.76, h*.013, cyan)
	remasterText("RT + STICK NORTH: DRIVE  /  LT + SOUTH: REVERSE", w*.12, h*.81, h*.013, white)
}
