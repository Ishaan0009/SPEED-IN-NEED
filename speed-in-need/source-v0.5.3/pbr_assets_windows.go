//go:build windows

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"math"
	"runtime"
	"speedinneed/engine"
	"strings"
	"unsafe"
)

type pbrMaterial struct {
	Surface                                 string
	Emissive                                string
	Name, Albedo, MR, Normal, AO, Alpha     string
	Base                                    [4]float64
	Metal, Rough, Coat, Cutoff, NormalScale float64
	Emission                                [3]float64
	Paint, Glass, Double                    bool
}

// Set per vehicle so an opponent's lamps never follow the player's H key.
var remasterVehicleLamps struct{ Active, Front, Tail, Brake, Reverse bool }
var remasterPlayerCockpit bool

type pbrPart struct {
	Slot, Material, Vertices, Indices int
	vbo, ebo, vao                     uint32
}
type pbrAsset struct {
	Key        string
	Materials  []pbrMaterial
	Batches    []pbrPart
	Anchors    map[string][3]float64
	WheelRadii map[string]float64
	Bounds     [2][3]float64
}

var pbrAssets = map[string]*pbrAsset{}
var pbrTextures = map[string]uint32{}
var pbrErrors []string

func pbrTexture(name string) uint32 {
	if name == "" {
		return app.whiteTexture
	}
	if t, ok := pbrTextures[name]; ok {
		return t
	}
	if texture, ok := pbrLeafTexture040(name); ok {
		pbrTextures[name] = texture
		return texture
	}
	raw, err := assets.ReadFile("assets/pbr/" + textureSource024(name))
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		pbrTextures[name] = app.whiteTexture
		return app.whiteTexture
	}
	im, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		pbrTextures[name] = app.whiteTexture
		return app.whiteTexture
	}
	b := im.Bounds()
	rgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), im, b.Min, draw.Src)
	rgba = limitTexture023(rgba)
	b = rgba.Bounds()
	var t uint32
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&t)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(t))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2703)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, GL_RGBA, uintptr(b.Dx()), uintptr(b.Dy()), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&rgba.Pix[0])))
	if (strings.HasPrefix(name, "b052_grass") || strings.HasPrefix(name, "b053_grass")) || strings.Contains(name, "_far052") {
		coverageMips052(rgba.Pix, b.Dx(), b.Dy())
	} else if glExtAvailable("glGenerateMipmap") {
		glExt("glGenerateMipmap", GL_TEXTURE_2D)
	} else {
		uploadTextureMipLevels(rgba.Pix, b.Dx(), b.Dy())
	}
	// Anisotropy is advertised by both supported desktop driver families.
	if supportsAnisotropy() {
		gl("glTexParameterf", GL_TEXTURE_2D, 0x84FE, b32(anisotropy026()))
	}
	gl("glTexParameterf", GL_TEXTURE_2D, 0x8501, b32(float64(beta.cfg.TextureBias)))
	pbrTextures[name] = t
	runtime.KeepAlive(rgba)
	return t
}

var anisotropyCaps026 struct {
	Checked, Supported bool
	Maximum            float32
}

func supportsAnisotropy() bool {
	if !anisotropyCaps026.Checked {
		anisotropyCaps026.Checked = true
		anisotropyCaps026.Supported = strings.Contains(glString(0x1F03), "GL_EXT_texture_filter_anisotropic")
		if anisotropyCaps026.Supported {
			gl("glGetFloatv", 0x84FF, uintptr(unsafe.Pointer(&anisotropyCaps026.Maximum)))
		}
	}
	return anisotropyCaps026.Supported
}
func anisotropy026() float64 {
	return math.Max(1, math.Min(float64(beta.cfg.Anisotropy), float64(anisotropyCaps026.Maximum)))
}

func loadPBRAsset(key string) {
	if pbrAssets[key] != nil {
		return
	}
	raw, err := assets.ReadFile("assets/pbr/" + key + ".pbr")
	if err != nil {
		return
	}
	if len(raw) < 8 || string(raw[:4]) != "PBR1" {
		pbrErrors = append(pbrErrors, key+": invalid header")
		return
	}
	n := int(binary.LittleEndian.Uint32(raw[4:8]))
	if n < 2 || n > len(raw)-8 {
		return
	}
	a := new(pbrAsset)
	if json.Unmarshal(raw[8:8+n], a) != nil {
		return
	}
	for i := range a.Materials {
		fullLeafMaterial040(&a.Materials[i])
	}
	offset := 8 + n
	for i := range a.Batches {
		p := &a.Batches[i]
		vb, ib := p.Vertices*48, p.Indices*4
		if p.Vertices < 3 || p.Indices < 3 || p.Material < 0 || p.Material >= len(a.Materials) || vb < 0 || ib < 0 || offset > len(raw)-vb-ib {
			pbrErrors = append(pbrErrors, key+": malformed mesh")
			return
		}
		reliefVertices051(key, raw[offset:offset+vb])
		if glExtAvailable("glGenVertexArrays") {
			glExt("glGenVertexArrays", 1, uintptr(unsafe.Pointer(&p.vao)))
			glExt("glBindVertexArray", uintptr(p.vao))
		}
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&p.vbo)))
		glExt("glBindBuffer", glArrayBuffer, uintptr(p.vbo))
		glExt("glBufferData", glArrayBuffer, uintptr(vb), uintptr(unsafe.Pointer(&raw[offset])), glStaticDraw)
		offset += vb
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&p.ebo)))
		glExt("glBindBuffer", 0x8893, uintptr(p.ebo))
		glExt("glBufferData", 0x8893, uintptr(ib), uintptr(unsafe.Pointer(&raw[offset])), glStaticDraw)
		offset += ib
		initialPointers := *p
		initialPointers.vao = 0
		pbrPointers(initialPointers)
		glExt("glBindVertexArray", 0)
	}
	if offset != len(raw) {
		pbrErrors = append(pbrErrors, key+": trailing mesh bytes")
		return
	}
	for _, m := range a.Materials {
		for _, name := range []string{m.Albedo, m.MR, m.Normal, m.AO, m.Emissive} {
			if name != "" {
				pbrTexture(name)
			}
		}
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
	glExt("glBindBuffer", 0x8893, 0)
	pbrAssets[key] = a
	runtime.KeepAlive(raw)
}
func initPBRAssets() {
	if lightPipeline.program == 0 || !glExtAvailable("glGenBuffers") {
		return
	}
	for _, key := range []string{"ferrari", "porsche", "streetlight", "tree", "building0", "building1", "building2", "building3", "road", "ground"} {
		loadPBRAsset(key)
	}
}
func installBetaWorldTextures() {
	for key, name := range map[string]string{"asphalt.jpg": "road_surface_color.jpg", "grass.jpg": "ground_6_color.jpg"} {
		if old := app.tex[key]; old != 0 {
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&old)))
		}
		app.tex[key] = pbrTexture(name)
	}
}
func bindWorldPBR(texture uint32) {
	resetPBRMaterial()
	var m pbrMaterial
	found := false
	if texture == app.tex["asphalt.jpg"] && pbrAssets["road"] != nil {
		m = pbrAssets["road"].Materials[0]
		m.Albedo = "road_surface_color.jpg"
		m.MR = "road_surface_mr.jpg"
		m.Normal = "road_surface_normal.jpg"
		m.Rough = .93
		m.NormalScale = .20
		found = true
	}
	if texture == app.tex["grass.jpg"] && pbrAssets["ground"] != nil {
		m = pbrAssets["ground"].Materials[2]
		m.Rough = 1
		found = true
	}
	if found {
		bindPBRMaterial(m)
		sceneInt("uMapped", 2)
	}
}
func bindPBRMaterial(m pbrMaterial) {
	if m.Albedo == "b031_asphalt_lanes.png" {
		m.Albedo = "b034_asphalt_lanes.jpg"
		m.Normal = "b034_asphalt_normal.jpg"
		m.NormalScale = .24
		if weather050.Rain {
			m.Normal = "b052_rain_asphalt_normal.png"
			m.NormalScale = .18
		}
	}
	if m.Name == "Primary" && remasterVehicleLamps.Active {
		m.Paint = true
	}
	m, response := beta021Surface(m)
	if remasterPlayerCockpit && m.Glass {
		m.Base[3] = .055
		m.Metal = 0
		m.Rough = .16
		m.Coat = 0
		response = pos3{.035, .08, .16}
	}
	sceneVector("uSurfaceResponse", response)
	sceneInt("uFoliage", boolInt(m.Surface == "foliage"))
	if m.Surface == "foliage" {
		m.Double = true
	}
	if strings.Contains(strings.ToLower(m.Name), "glow") && !remasterVehicleLamps.Active {
		energy := engine.Clamp((sceneHour()-15.0)/3.0, 0, 1)
		if engine.NightAmount(sceneHour()) > .5 {
			energy = 1
		}
		if weather050.Rain {
			energy = math.Max(energy, .7)
		}
		m.Emission = [3]float64{m.Emission[0] * energy, m.Emission[1] * energy, m.Emission[2] * energy}
	}

	sceneInt("uDoubleSided", boolInt(m.Double))
	sceneInt("uTerrain", boolInt(m.Name == "remaster_ground"))
	sceneInt("uRockSurface", boolInt(m.Surface == "rock"))
	sceneInt("uRoadBacking", boolInt(m.Name == "b022r_road_backing"))
	sceneInt("uPavement", boolInt(m.Name == "coast_paving"))
	sceneInt("uRoadSurface", boolInt(strings.Contains(strings.ToLower(m.Name), "asphalt")))
	if m.Name == "sugar_maple_leaf" || m.Name == "sugar_maple_bark" {
		m.Metal = 0
		m.Rough = .95
		m.MR = ""
		m.AO = ""
	}
	front := beta.cfg.Headlights
	if remasterVehicleLamps.Active {
		front = remasterVehicleLamps.Front
	}
	if m.Name == "HGLOW3" || m.Name == "GLOW" || m.Name == "beta_rear_lamp" || strings.HasPrefix(m.Name, "tail_light_light") || m.Name == "back_extra_3_lights" {
		// Only rear-lamp materials need brake-key state. Polling it for each
		// terrain, grass and building material cost a native OS call per bind,
		// and vehicle-local state already supplies the authoritative lamps.
		tail, brake := engine.AutomaticLights(sceneHour()), false
		if remasterVehicleLamps.Active {
			tail, brake = remasterVehicleLamps.Tail, remasterVehicleLamps.Brake
		} else {
			brake = beta.screen == screenDrive && pressed(1)
		}
		energy := .35
		if !tail {
			energy = 0
		}
		if brake {
			energy = 2.5
		}
		m.Base = [4]float64{.9, .025, .01, 1}
		m.Emission = [3]float64{energy, .008, .002}
		m.Emissive = ""
	}
	if m.Name == "beta_front_lamp" {
		m.Emissive = m.Albedo
	}
	if m.Paint && sceneIntValues["uLivery050"] == 12 {
		m.Paint = false
	}
	sceneInt("uPaintSurface051", boolInt(m.Paint))
	if m.Paint {
		m.NormalScale = math.Min(m.NormalScale, .045)
		m.Metal = .18
		m.Rough = .28
		m.Coat = .80
	}
	if m.Name == "HGLOW" || m.Name == "HGLOW2" || m.Name == "LED_WHITE" || m.Name == "beta_front_lamp" || m.Name == "headlight_light" {
		if front {
			m.Emission = [3]float64{2.8, 2.8, 2.65}
		} else {
			m.Emission = [3]float64{}
		}
	}

	if m.Surface == "headlamp053" || m.Surface == "taillamp053" || m.Surface == "brakelamp053" || m.Surface == "reverselamp053" {
		m.Paint = false
		m.Metal = 0
		m.Rough = .18
		m.Glass = false
		m.Emissive = ""
		on := false
		switch m.Surface {
		case "headlamp053":
			on = front
			if on {
				m.Emission = [3]float64{4.2, 4.1, 3.9}
			}
		case "taillamp053":
			on = remasterVehicleLamps.Tail || remasterVehicleLamps.Brake
			if on {
				e := 1.1
				if remasterVehicleLamps.Brake {
					e = 3.8
				}
				m.Emission = [3]float64{e, .018, .005}
			}
		case "brakelamp053":
			on = remasterVehicleLamps.Brake
			if on {
				m.Emission = [3]float64{4, .015, .004}
			}
		case "reverselamp053":
			on = remasterVehicleLamps.Reverse
			if on {
				m.Emission = [3]float64{2.1, 2.0, 1.8}
			}
		}
		if !on {
			m.Emission = [3]float64{}
		}
	}
	if m.Name == "Street fixture 053" {
		e := engine.NightAmount(sceneHour())
		if weather050.Rain {
			e = math.Max(e, .7)
		}
		m.Emission = [3]float64{3.8 * e, 3.25 * e, 2.4 * e}
	}

	setPBRMaterial(true, m.Metal, m.Rough)
	sceneInt("uMapped", 1)
	sceneInt("uPaintTintRefine", boolInt(m.Paint && m.Albedo != ""))
	if m.Paint && m.Name == "car_chassis" {
		sceneInt("uPaintTintRefine", 2)
	}
	sceneInt("uTreeWindRefine", 0)
	sceneFloat("uClearcoat", m.Coat)
	sceneFloat("uNormalScale", m.NormalScale)
	sceneInt("uHasMR", boolInt(m.MR != ""))
	sceneInt("uHasNormal", boolInt(m.Normal != "" && (m.Surface != "ground" || beta.cfg.GroundDetail > 0)))
	sceneInt("uGroundDetail", beta.cfg.GroundDetail)
	sceneInt("uWindEnabled", boolInt(beta.cfg.Wind))
	sceneInt("uHasAO", boolInt(m.AO != ""))
	sceneFloat("uAlphaCutoff", 0)
	if m.Alpha == "MASK" {
		sceneFloat("uAlphaCutoff", m.Cutoff)
	}
	sceneVector("uEmission", pos3{m.Emission[0], m.Emission[1], m.Emission[2]})
	sceneInt("uHasEmissive", boolInt(m.Emissive != ""))
	bindPBRTexture035(7, m.Emissive)
	for i, name := range []string{m.Albedo, m.MR, m.Normal, m.AO} {
		bindPBRTexture035(i, name)
	}
	glExt("glActiveTexture", 0x84C0)
	c := m.Base
	if m.Paint {
		p := paintPalette[beta.paint]
		c = [4]float64{p[0], p[1], p[2], 1}
	}
	col(c[0], c[1], c[2], c[3])
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func resetPBRMaterial() {
	sceneInt("uPaintSurface051", 0)
	sceneInt("uPaintTintRefine", 0)
	sceneInt("uTreeWindRefine", 0)
	sceneInt("uRockSurface", 0)
	sceneVector("uSurfaceResponse", pos3{.015, .15, .76})
	sceneInt("uFoliage", 0)
	sceneInt("uDoubleSided", 0)
	sceneInt("uTerrain", 0)
	sceneInt("uRoadBacking", 0)
	sceneInt("uPavement", 0)
	sceneInt("uRoadSurface", 0)
	sceneInt("uMapped", 0)
	sceneFloat("uClearcoat", 0)
	sceneFloat("uAlphaCutoff", 0)
	sceneVector("uEmission", pos3{})
	glExt("glActiveTexture", 0x84C0)
	col(1, 1, 1, 1)
	setPBRMaterial(false, 0, .8)
}
func pbrPointers(p pbrPart) {
	if p.vao != 0 {
		glExt("glBindVertexArray", uintptr(p.vao))
		return
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(p.vbo))
	glExt("glBindBuffer", 0x8893, uintptr(p.ebo))
	for _, s := range []uintptr{glVertexArray, glNormalArray, 0x8078} {
		gl("glEnableClientState", s)
	}
	gl("glDisableClientState", glColorArray)
	gl("glVertexPointer", 3, glFloat, 48, 0)
	gl("glNormalPointer", glFloat, 48, 12)
	gl("glTexCoordPointer", 2, glFloat, 48, 24)
	glExt("glEnableVertexAttribArray", 7)
	glExt("glVertexAttribPointer", 7, 4, glFloat, 0, 48, 32)
}
func drawPBRAsset(key string, slot int) bool {
	objectMotion038(key, slot)
	defer sceneInt("uObjectMotion038", 0)
	a := pbrAssets[key]
	if a == nil {
		return false
	}
	for _, p := range a.Batches {
		if p.Slot != slot {
			continue
		}
		m := a.Materials[p.Material]
		if remasterPlayerCockpit && slot == 0 && m.Name == "SteeringWheel" {
			continue // The supplied animated steering prop replaces the static interior wheel.
		}
		bindPBRMaterial(m)
		pbrPointers(p)
		if m.Glass {
			gl("glDepthMask", 0)
		}
		gl("glDrawElements", GL_TRIANGLES, uintptr(p.Indices), 0x1405, 0)
		gpuWorld.draws++
		if m.Glass {
			gl("glDepthMask", 1)
		}
	}
	unbindWorld()
	resetPBRMaterial()
	return true
}
func drawPBRCar(id int, spin, steer float64) bool {
	key := []string{"ferrari", "porsche"}[id]
	a := pbrAssets[key]
	if a == nil || !lightPipeline.active {
		return false
	}
	drawPBRAsset(key, 0)
	for i := 0; i < 4; i++ {
		anchor, ok := a.Anchors[fmt.Sprint(i+1)]
		if !ok {
			continue
		}
		angle, yaw, travel := spin, 0., 0.
		if beta.screen == screenDrive {
			w := app.renderCar.Rig.Wheels[i]
			angle = w.RollAngle
			yaw = w.SteerAngle
			p := engine.VehicleSpec(id)
			travel = engine.Clamp(w.Compression-p.Mass*9.81/(4*p.SpringRate), -.18, .22)
		} else if i < 2 {
			yaw = steer
		}
		push()
		move(anchor[0], anchor[1]+travel, anchor[2])
		rotate(yaw*180/math.Pi, 0, 1, 0)
		rotate(angle*180/math.Pi, 1, 0, 0)
		drawPBRAsset(key, i+1)
		pop()
	}
	return true
}
func drawPBRBuildings() {
	if !lightPipeline.active {
		return
	}
	for i, p := range app.world.Buildings {
		if p.Dist(app.world.Car.Pos) > 580 && beta.screen == screenDrive {
			continue
		}
		push()
		move(p.X, engine.Terrain(p.X, p.Z)-.04, p.Z)
		// Asset footprints are capped at 12m; the pavement/collider is 13 by 15m.
		drawPBRAsset(fmt.Sprintf("building%d", i%6), 0)
		pop()
	}
}
func drawPBRTreeInstances(instances uint32, count int) bool {
	a := pbrAssets["tree"]
	if a == nil || !lightPipeline.active {
		return false
	}
	setInstancedWorld(true)
	for _, p := range a.Batches {
		bindPBRMaterial(a.Materials[p.Material])
		pbrPointers(p)
		glExt("glBindBuffer", glArrayBuffer, uintptr(instances))
		loc := uintptr(gpuWorld.instanceLoc)
		glExt("glEnableVertexAttribArray", loc)
		glExt("glVertexAttribPointer", loc, 4, glFloat, 0, 16, 0)
		glExt("glVertexAttribDivisor", loc, 1)
		gl("glDepthMask", 1)
		glExt("glDrawElementsInstanced", GL_TRIANGLES, uintptr(p.Indices), 0x1405, 0, uintptr(count))
		gpuWorld.draws++
		glExt("glVertexAttribDivisor", loc, 0)
		glExt("glDisableVertexAttribArray", loc)
	}

	unbindWorld()
	resetPBRMaterial()
	setInstancedWorld(false)
	return true
}
func releasePBRAssets() {
	for _, a := range pbrAssets {
		for _, p := range a.Batches {
			if p.vao != 0 {
				glExt("glDeleteVertexArrays", 1, uintptr(unsafe.Pointer(&p.vao)))
			}
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&p.vbo)))
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&p.ebo)))
		}
	}
	for _, t := range pbrTextures {
		shared := false
		for _, owned := range app.tex {
			if owned == t {
				shared = true
				break
			}
		}
		if shared {
			continue
		}
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&t)))
	}
	pbrAssets = map[string]*pbrAsset{}
	pbrTextures = map[string]uint32{}
}
