package main

import (
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

// Light assignment is indexed by world location, never the player's location.
// Four lamps per cell fit the local influence radius. Tunnel fixtures use their
// complete analytic row in the shader. Both paths shade diffuse AND specular.
var lightfield052 struct {
	Def     *engine.WorldDefinition
	Texture uint32
	W, H    int
	Fallen  int
	X, Z    float64
}

func bindLightfield052() {
	if app.world == nil || app.world.Remaster == nil {
		sceneFloat("uLampEnable052", 0)
		return
	}
	d := app.world.Remaster.Def
	if lightfield052.Def != d || lightfield052.Fallen != len(app.world.Remaster.FallenPoles) {
		if lightfield052.Texture != 0 {
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&lightfield052.Texture)))
		}
		g := d.Grid
		w, h := int(math.Ceil(float64(g.W-1)*g.Step/32))+2, int(math.Ceil(float64(g.H-1)*g.Step/32))+2
		w = engine.ClampInt(w, 2, 256)
		h = engine.ClampInt(h, 2, 256)
		data := make([]float32, w*h*4*4)
		type lamp struct{ x, y, z float64 }
		cells := map[[2]int][]lamp{}
		for id, o := range d.Objects {
			if o.Kind != "lamp" || app.world.Remaster.FallenPoles[id] != nil {
				continue
			}
			q := lamp{o.Pos[0] + math.Cos(o.Heading)*3.9, o.Pos[1] + 9.7, o.Pos[2] - math.Sin(o.Heading)*3.9}
			cx, cz := int(math.Floor((q.x-g.MinX)/32)), int(math.Floor((q.z-g.MinZ)/32))
			for x := cx - 2; x <= cx+2; x++ {
				for z := cz - 2; z <= cz+2; z++ {
					cells[[2]int{x, z}] = append(cells[[2]int{x, z}], q)
				}
			}
		}
		for z := 0; z < h; z++ {
			for x := 0; x < w; x++ {
				px, pz := g.MinX+(float64(x)+.5)*32, g.MinZ+(float64(z)+.5)*32
				var near [4]lamp
				ds := [4]float64{1e20, 1e20, 1e20, 1e20}
				for _, q := range cells[[2]int{x, z}] {
					dist := (q.x-px)*(q.x-px) + (q.z-pz)*(q.z-pz)
					for k := 0; k < 4; k++ {
						if dist < ds[k] {
							for j := 3; j > k; j-- {
								near[j] = near[j-1]
								ds[j] = ds[j-1]
							}
							near[k] = q
							ds[k] = dist
							break
						}
					}
				}
				for k, q := range near {
					at := ((z*w+x)*4 + k) * 4
					if ds[k] < 1e19 {
						copy(data[at:], []float32{float32(q.x), float32(q.y), float32(q.z), 1})
					}
				}
			}
		}
		lightfield052.Fallen = len(app.world.Remaster.FallenPoles)
		lightfield052.Def = d
		lightfield052.W = w
		lightfield052.H = h
		lightfield052.X = g.MinX
		lightfield052.Z = g.MinZ
		glExt("glActiveTexture", 0x84CB)
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(&lightfield052.Texture)))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(lightfield052.Texture))
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2600)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2600)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
		gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x8814, uintptr(w*4), uintptr(h), 0, GL_RGBA, glFloat, uintptr(unsafe.Pointer(&data[0])))
		runtime.KeepAlive(data)
	}
	glExt("glActiveTexture", 0x84CB)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(lightfield052.Texture))
	glExt("glActiveTexture", 0x84C0)
	sceneInt("uLampGrid052", 11)
	sceneVec4("uLampDomain052", lightfield052.X, lightfield052.Z, float64(lightfield052.W), float64(lightfield052.H))
	energy := engine.NightAmount(sceneHour())
	if weather050.Rain {
		energy = math.Max(energy, .70)
	}
	if garage033 || workshop051.Active {
		energy = 0
	}
	sceneFloat("uLampEnable052", energy)
}
