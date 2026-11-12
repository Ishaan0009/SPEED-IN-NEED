package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
)

type garageSetup050 struct{ Model, Tyre, Wheel, Livery, Paint051 int }

var garage050 struct {
	Open, Loaded bool
	Selected     int
	Setups       map[string]garageSetup050
	Message      string
}

func garagePath050() string {
	exe, _ := os.Executable()
	return filepath.Join(filepath.Dir(exe), "UserData", "garage050.json")
}
func garageKey050() string {
	name := "driver"
	if p := activeProfile040(); p != nil {
		name = p.ID
	}
	return fmt.Sprintf("%s-car%d", name, beta.carSelect)
}
func setup050() garageSetup050 {
	if !garage050.Loaded {
		garage050.Loaded = true
		garage050.Setups = map[string]garageSetup050{}
		if b, e := os.ReadFile(garagePath050()); e == nil {
			if json.Unmarshal(b, &garage050.Setups) != nil {
				garage050.Message = "Garage save could not be read; original file preserved"
				garage050.Setups = nil
			}
		}
	}
	c := garage050.Setups[garageKey050()]
	if saved, ok := garage050.Setups[fmt.Sprintf("%s-model%d", garageKey050(), c.Model)]; ok {
		c = saved
	}
	c.Model = engine.ClampInt(c.Model, 0, len(garageCars051())-1)
	c.Tyre = engine.ClampInt(c.Tyre, 0, 3)
	c.Wheel = engine.ClampInt(c.Wheel, 0, 5)
	c.Livery = engine.ClampInt(c.Livery, 0, len(liveryNames052)-1)
	return c
}
func saveSetup050(c garageSetup050) {
	if garage050.Setups == nil {
		return
	}
	c.Paint051 = beta.paint
	garage050.Setups[garageKey050()] = c
	garage050.Setups[fmt.Sprintf("%s-model%d", garageKey050(), c.Model)] = c
	if qa.enabled {
		return
	} // Renderer fixtures keep garage changes in memory.
	b, e := json.MarshalIndent(garage050.Setups, "", "  ")
	if e != nil {
		return
	}
	path := garagePath050()
	os.MkdirAll(filepath.Dir(path), 0755)
	if e = os.WriteFile(path+".tmp", b, 0644); e == nil {
		e = os.Rename(path+".tmp", path)
	}
	if e != nil {
		garage050.Message = e.Error()
	} else {
		garage050.Message = "SAVED TO YOUR PLAYER PROFILE"
	}
}
func modelKey050(c engine.Car, id int, preview bool) string {
	variant := c.Variant050
	if preview {
		variant = setup050().Model
	}
	if variant > 0 {
		return garageCars051()[engine.ClampInt(variant, 0, len(garageCars051())-1)].Key
	}
	return []string{"ferrari", "porsche", "rm_car2", "rm_car3"}[engine.ClampInt(id, 0, 3)]
}
func applyGarage050() {
	if app.world == nil {
		return
	}
	c := setup050()
	app.world.Car.Variant050 = c.Model
	app.world.Car.Tyre050 = c.Tyre
	app.world.Car.Rim050 = c.Wheel
	app.world.Car.Livery050 = c.Livery
	app.world.Car.Wetness050 = weather050.Amount
}
func garageWnd050(m uint32, w, l uintptr) bool {
	if m == WM_KEYDOWN && w == 'G' && beta.screen == screenMaps && ui020.Stage == 2 {
		customize051 = !customize051
		focus033.Ready = false
		return true
	}
	if m == WM_KEYDOWN && w == 'G' && (l>>30)&1 == 0 && !multiplayer050.Open {
		garage050.Open = !garage050.Open
		beta020SyncPointer()
		return true
	}
	if !garage050.Open {
		return false
	}
	if m == WM_KEYDOWN {
		switch w {
		case VK_ESCAPE:
			garage050.Open = false
		case 0x26:
			garage050.Selected = (garage050.Selected + 4) % 5
		case 0x28:
			garage050.Selected = (garage050.Selected + 1) % 5
		case 0x25:
			adjustGarage050(-1)
		case 0x27, VK_RETURN:
			adjustGarage050(1)
		}
		return true
	}
	if m == WM_LBUTTONDOWN {
		x := float64(int16(l&65535)) / float64(app.width)
		y := float64(int16(l>>16)) / float64(app.height)
		row := int((y - .30) / .09)
		if row >= 0 && row < 5 && x > .66 && x < .99 {
			garage050.Selected = row
			delta := 1
			if x < .82 {
				delta = -1
			}
			adjustGarage050(delta)
		}
		return true
	}
	return m == WM_MOUSEMOVE || m == 0x202 || m == WM_RBUTTONDOWN || m == WM_RBUTTONUP
}
func adjustGarage050(delta int) {
	c := setup050()
	switch garage050.Selected {
	case 0:
		selectCar051(selectedCar051() + delta)
		return
	case 1:
		c.Tyre = (c.Tyre + delta + 4) % 4
	case 2:
		c.Wheel = (c.Wheel + delta + 6) % 6
	case 3:
		c.Livery = (c.Livery + delta + len(liveryNames052)) % len(liveryNames052)
	case 4:
		beta.paint = (beta.paint + delta + len(paintPalette)) % len(paintPalette)
		saveRemasterSettings()
	}
	saveSetup050(c)
	key := modelKey050(engine.Car{Variant050: c.Model}, beta.carSelect, true)
	if pbrAssets[key] == nil {
		loadPBRAsset(key)
	}
	applyGarage050()
	if garage050.Selected == 0 && app.world != nil {
		app.world.Car.Body.Initialized = false
		app.world.Car.TiresReady = false
	}
}
func drawGarage050() {
	if !garage050.Open || qa.enabled && qa.stage != 204 && qa.stage != 208 {
		return
	}
	remasterUIBegin()
	defer en(GL_DEPTH_TEST)
	w, h := float64(app.width), float64(app.height)
	hudRect(w*.66, h*.10, w*.325, h*.80, .015, .025, .032, .96)
	remasterText("CUSTOMIZE", w*.68, h*.145, h*.027, white)
	headline035("ARROWS / CLICK: CHANGE    G: TEST DRIVE", w*.68, h*.22, h*.009, w*.29, mutedRefine)
	c := setup050()
	labels := []string{"VEHICLE MODEL", "TYRES", "RIM STYLE", "LIVERY", "PAINT"}
	values := []string{garageCars051()[c.Model].Name, []string{"ROAD", "DRY PERFORMANCE", "WET WEATHER", "ALL TERRAIN"}[c.Tyre], []string{"ORIGINAL", "SPORT 01", "SPORT 02", "SPORT 03", "SPORT 04", "SPORT 05"}[c.Wheel], liveryNames052[c.Livery], paintNamesRefine[beta.paint%len(paintNamesRefine)]}
	for i, label := range labels {
		y := h * (.30 + float64(i)*.09)
		if garage050.Selected == i {
			hudRect(w*.675, y-7, w*.30, h*.075, .12, .18, .19, 1)
		}
		remasterText(label, w*.69, y, h*.010, mutedRefine)
		headline035(values[i], w*.69, y+h*.023, h*.016, w*.27, cyan)
	}
	headline035(garage050.Message, w*.68, h*.80, h*.009, w*.28, mutedRefine)
}

// Rims use real rotating geometry. Radius follows the car's fitted wheel size.
func drawRim050(radius float64, side float64, style int) {
	if style < 1 {
		return
	}
	key := fmt.Sprintf("b051_wheel%d", engine.ClampInt(style, 1, 5))
	if pbrAssets[key] == nil {
		loadPBRAsset(key)
	}
	if pbrAssets[key] != nil {
		push()
		resize(radius, radius, radius)
		if side < 0 {
			rotate(180, 0, 1, 0)
		}
		drawPBRAsset(key, 0)
		pop()
		return
	}
	dis(GL_TEXTURE_2D)
	sceneInt("uPBR", 1)
	sceneFloat("uMetallic", .75)
	sceneFloat("uRoughness", .25)
	col(.28, .31, .34, 1)
	x := math.Copysign(.155, side)
	spokes := []int{0, 5, 6, 7, 10, 12}[engine.ClampInt(style, 0, 5)]
	begin(GL_QUADS)
	for i := 0; i < spokes; i++ {
		a := float64(i) * engine.Tau / float64(spokes)
		b := a + .11
		for _, p := range [][2]float64{{a, .12}, {b, .12}, {b, radius * .77}, {a, radius * .77}} {
			nrm(math.Copysign(1, side), 0, 0)
			vtx(x, math.Cos(p[0])*p[1], math.Sin(p[0])*p[1])
		}
	}
	end()
}
