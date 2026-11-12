package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"unsafe"
)

var customize051 bool
var decorated051 = map[*engine.WorldDefinition]bool{}
var rootedRocks051 = map[*engine.WorldDefinition]bool{}
var streetPools051 struct {
	Def   *engine.WorldDefinition
	VBO   uint32
	Count int
}

func rootRocks051(d *engine.WorldDefinition) {
	if rootedRocks051[d] {
		return
	}
	rootedRocks051[d] = true
	for i := range d.Objects {
		o := &d.Objects[i]
		if o.Kind != "ground-rock" && o.Kind != "rock" && o.Asset != "b035_fallen_branch" {
			continue
		}
		a := pbrAssets[o.Asset]
		if a == nil {
			continue
		}
		old := o.Pos[1]
		ground := math.Inf(1)
		for _, x := range []float64{a.Bounds[0][0], 0, a.Bounds[1][0]} {
			for _, z := range []float64{a.Bounds[0][2], 0, a.Bounds[1][2]} {
				px, pz := x*o.Scale[0], z*o.Scale[2]
				at := engine.Vec{X: o.Pos[0] + math.Cos(o.Heading)*px + math.Sin(o.Heading)*pz, Z: o.Pos[2] - math.Sin(o.Heading)*px + math.Cos(o.Heading)*pz}
				ground = math.Min(ground, d.Grid.Height(at))
			}
		}
		o.Pos[1] = ground - a.Bounds[0][1]*o.Scale[1] - .16
		for j := range d.Colliders {
			if d.Colliders[j].Object == i {
				d.Colliders[j].MinY += o.Pos[1] - old
				d.Colliders[j].MaxY += o.Pos[1] - old
			}
		}
	}
}

// A cached ground-light pass keeps distant streets illuminated. Nearby lamps
// still use the material lighting/shadow budget; these pools cost one draw call.
func drawStreetPools051() {
	if app.world == nil || app.world.Remaster == nil || workshop051.Active || !(engine.AutomaticLights(sceneHour()) || weather050.Rain) {
		return
	}
	d := app.world.Remaster.Def
	if streetPools051.Def != d {
		if streetPools051.VBO != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&streetPools051.VBO)))
		}
		vertices := make([][7]float32, 0, len(remaster.Lamps)*30)
		for _, id := range remaster.Lamps {
			o := d.Objects[id]
			x, z := o.Pos[0]+math.Cos(o.Heading)*3.9, o.Pos[2]-math.Sin(o.Heading)*3.9
			for i := 0; i < 10; i++ {
				for j := 0; j < 3; j++ {
					px, pz, energy := x, z, .08
					if j > 0 {
						a := float64(i+j-1) * engine.Tau / 10
						px += math.Cos(a) * 9
						pz += math.Sin(a) * 9
						energy = 0
					}
					y := app.world.GroundHeight(engine.Vec{X: px, Z: pz}) + .065
					vertices = append(vertices, [7]float32{float32(px), float32(y), float32(pz), float32(energy), float32(energy * .80), float32(energy * .48), 1})
				}
			}
		}
		streetPools051.Def = d
		streetPools051.Count = len(vertices)
		if len(vertices) > 0 {
			glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&streetPools051.VBO)))
			glExt("glBindBuffer", glArrayBuffer, uintptr(streetPools051.VBO))
			glExt("glBufferData", glArrayBuffer, uintptr(len(vertices)*28), uintptr(unsafe.Pointer(&vertices[0])), glStaticDraw)
			glExt("glBindBuffer", glArrayBuffer, 0)
		}
	}
	if streetPools051.Count == 0 {
		return
	}
	useLighting(false)
	dis(GL_TEXTURE_2D)
	en(GL_BLEND)
	gl("glBlendFunc", 1, 1)
	gl("glDepthMask", 0)
	glExt("glBindVertexArray", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(streetPools051.VBO))
	gl("glEnableClientState", glVertexArray)
	gl("glEnableClientState", glColorArray)
	gl("glVertexPointer", 3, glFloat, 28, 0)
	gl("glColorPointer", 4, glFloat, 28, 12)
	gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(streetPools051.Count))
	gl("glDisableClientState", glColorArray)
	gl("glDisableClientState", glVertexArray)
	glExt("glBindBuffer", glArrayBuffer, 0)
	gl("glDepthMask", 1)
	gl("glBlendFunc", GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA)
	useLighting(true)
}

func decorateMeadow051(d *engine.WorldDefinition) {
	if decorated051[d] {
		return
	}
	decorated051[d] = true
	original := len(d.Objects)
	for i := 0; i < original; i++ {
		o := d.Objects[i]
		if o.Kind != "bush" || i%3 != 0 {
			continue
		}
		o.Asset = "b051_daisy"
		if i%2 == 0 {
			o.Asset = "b051_daffodil"
		}
		if i%5 == 0 {
			o.Asset = "b052_pink_flower"
		}
		o.Solid = false
		o.Collider = -1
		o.Radius = 1.5
		sc := .85 + float64(i%5)*.12
		o.Scale = [3]float64{sc, sc * 1.5, sc}
		o.Pos[0] += math.Sin(float64(i)*2.4) * 2
		o.Pos[2] += math.Cos(float64(i)*2.4) * 2
		o.Pos[1] = d.Grid.Height(engine.Vec{X: o.Pos[0], Z: o.Pos[2]}) - .015
		d.Objects = append(d.Objects, o)
	}
}

func drawClouds051() {
	if workshop051.Active || weather050.Rain || app.quality == 0 {
		return
	}
	if pbrAssets["b051_cloud"] == nil {
		loadPBRAsset("b051_cloud")
	}
	sceneInt("uShadowEnabled", 0)
	for i := 0; i < 20; i++ {
		x := -1500 + float64(i%5)*750 + math.Sin(float64(i)*5)*130 + math.Mod(app.t*1.1, 150)
		z := -1500 + float64(i/5)*1000 + math.Cos(float64(i)*7)*140
		if math.Hypot(x-app.cam.x, z-app.cam.z) > 2100 {
			continue
		}
		push()
		move(x, 290+float64(i%4)*36, z)
		rotate(float64(i)*137, 0, 1, 0)
		sc := 1.4 + float64(i%3)*.6
		resize(sc, sc*.65, sc)
		drawPBRAsset("b051_cloud", 0)
		pop()
	}
	updateSceneLights(app.renderCar)
}

// Player catalogue is deliberately independent from internal traffic/model IDs.
var playerVariants052 = []int{1, 2, 3, 13, 14, 15, 16, 17, 18}
var liveryNames052 = []string{"SOLID", "TWIN STRIPES", "CHECKER", "DIAGONAL", "CENTRE STRIPE", "CLUBSPORT", "HERITAGE", "TWO TONE", "RALLY", "CARBON HOOD", "PINSTRIPE", "HEXAGON", "FACTORY FINISH"}

func selectedCar051() int {
	for i, v := range playerVariants052 {
		if setup050().Model == v {
			return i
		}
	}
	return 0
}
func carName051(index int) string {
	return garageCars051()[playerVariants052[engine.ClampInt(index, 0, len(playerVariants052)-1)]].Name
}
func normalizePlayer052() {
	if qa.enabled && qa.stage != 0 {
		return
	}
	v := setup050().Model
	for _, allowed := range playerVariants052 {
		if allowed == v && beta.carSelect == 3 {
			return
		}
	}
	selectCar051(0)
}
func selectCar051(index int) {
	count := len(playerVariants052)
	index = (index%count + count) % count
	saveSetup050(setup050())
	beta.carSelect = 3
	c := garageSetup050{Model: playerVariants052[index]}
	if saved, ok := garage050.Setups[fmt.Sprintf("%s-model%d", garageKey050(), c.Model)]; ok {
		c = saved
		beta.paint = engine.ClampInt(c.Paint051, 0, len(paintPalette)-1)
	}
	saveSetup050(c)
	applyGarage050()
	preview014.ready = false
	saveRemasterSettings()
}
func cycleCar051(delta int) { selectCar051(selectedCar051() + delta) }
func drawCarRemastered051(w, h float64) {
	c := setup050()
	spec := engine.VehicleVariantSpec050(beta.carSelect, c.Model)
	headline035(carName051(selectedCar051()), w*.055, h*.255, h*.026, w*.28, white)
	if customize051 {
		remasterText("VEHICLE CUSTOMIZATION", w*.055, h*.315, h*.012, cyan)
		labels := []string{"TYRES", "RIMS", "LIVERY"}
		values := []string{[]string{"ROAD", "DRY PERFORMANCE", "WET WEATHER", "ALL TERRAIN"}[c.Tyre], []string{"ORIGINAL", "SPORT 01", "SPORT 02", "SPORT 03", "SPORT 04", "SPORT 05"}[c.Wheel], liveryNames052[c.Livery]}
		for i, label := range labels {
			y := h * (.37 + float64(i)*.105)
			remasterText(label, w*.055, y, h*.010, mutedRefine)
			beta020Button(fmt.Sprintf("modify051:%d:-1", i+1), "<", w*.055, y+h*.025, w*.035, h*.046, false)
			headline035(values[i], w*.10, y+h*.038, h*.011, w*.16, cyan)
			beta020Button(fmt.Sprintf("modify051:%d:1", i+1), ">", w*.28, y+h*.025, w*.035, h*.046, false)
		}
		headline035("FINISH / "+paintNamesRefine[beta.paint%len(paintNamesRefine)], w*.36, h*.68, h*.013, w*.55, white)
		for i, p := range paintPalette {
			x := w * (.36 + float64(i%12)*.045)
			y := h * (.725 + float64(i/12)*.036)
			ui020.Buttons = append(ui020.Buttons, button020{fmt.Sprintf("paint:%d", i), x, y, w * .038, h * .028})
			hudRect(x, y, w*.038, h*.028, p[0], p[1], p[2], 1)
			if i == beta.paint {
				hudRect(x, y+h*.029, w*.038, 2, cyan[0], cyan[1], cyan[2], 1)
			}
		}
		beta020Button("playGarage051", "PLAY GARAGE / TEST CAR", w*.69, h*.86, w*.275, h*.067, false)
		headline035("Paint, tyres and livery are saved to your profile", w*.055, h*.745, h*.009, w*.27, mutedRefine)
	} else {
		values := []string{fmt.Sprintf("%.0f KM/H", spec.TopSpeed*3.6), fmt.Sprintf("%.0f BHP", spec.PowerKW*1.341), fmt.Sprintf("%.0f NM", spec.P.PeakTorque), fmt.Sprintf("%.0f KG", spec.P.Mass)}
		for i, label := range []string{"TOP SPEED", "POWER", "TORQUE", "MASS"} {
			y := h * (.355 + float64(i)*.064)
			remasterText(label, w*.055, y, h*.010, mutedRefine)
			remasterText(values[i], w*.20, y, h*.012, white)
			hudRect(w*.055, y+h*.028, w*.24, 1, .35, .45, .45, .6)
		}
		beta020Button("customize051", "MODIFY / TYRES / LIVERY / COLOR", w*.055, h*.65, w*.27, h*.055, false)
		count := len(playerVariants052)
		for i := 0; i < count; i++ {
			x := w * (.055 + float64(i%5)*.182)
			y := h * (.745 + float64(i/5)*.05)
			beta020Button(fmt.Sprintf("vehicle051:%d", i), carName051(i), x, y, w*.176, h*.041, selectedCar051() == i)
		}
	}
	beta020Button("car-", "<", w*.32, h*.44, w*.034, h*.055, false)
	beta020Button("car+", ">", w*.95, h*.44, w*.034, h*.055, false)
	remasterText("DRAG TO ROTATE CAR / SCROLL TO ZOOM", w*.40, h*.63, h*.009, mutedRefine)
}

// Distant lamp lenses remain lit independently of the local lighting budget.
// A single depth-tested triangle batch provides their stable far representation.
func drawLampLenses051() {
	if app.world == nil || app.world.Remaster == nil || !(engine.AutomaticLights(sceneHour()) || weather050.Rain) {
		return
	}
	useLighting(false)
	dis(GL_TEXTURE_2D)
	night := engine.NightAmount(sceneHour())
	if weather050.Rain {
		night = math.Max(night, .7)
	}
	begin(GL_TRIANGLES)
	for _, id := range remaster.Lamps {
		if app.world.Remaster.FallenPoles[id] != nil {
			continue
		}
		o := app.world.Remaster.Def.Objects[id]
		p := pos3{o.Pos[0] + math.Cos(o.Heading)*3.9, o.Pos[1] + 9.7, o.Pos[2] - math.Sin(o.Heading)*3.9}
		d := math.Hypot(p.x-app.cam.x, p.z-app.cam.z)
		if d > 1700*viewDistance026() {
			continue
		}
		r := engine.Clamp(d*.00055, .14, .65)
		col(2.4*night, 2.1*night, 1.55*night, 1)
		for _, q := range []pos3{{-r, 0, -r}, {r, 0, -r}, {r, 0, r}, {-r, 0, -r}, {r, 0, r}, {-r, 0, r}} {
			vtx(p.x+q.x, p.y, p.z+q.z)
		}
	}
	end()
	useLighting(true)
}
