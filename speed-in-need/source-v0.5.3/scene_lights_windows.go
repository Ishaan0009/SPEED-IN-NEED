//go:build windows

package main

import (
	"fmt"
	"math"
	"sort"
	"speedinneed/engine"
	"unsafe"
)

var sceneUniforms = map[string]int32{}
var sceneIntValues = map[string]int{}
var sceneFloatValues = map[string]float64{}
var sceneVectorValues = map[string]pos3{}

func sceneLocation(name string) int32 {
	if loc, ok := sceneUniforms[name]; ok {
		return loc
	}
	n := append([]byte(name), 0)
	loc := int32(glExt("glGetUniformLocation", uintptr(lightPipeline.program), uintptr(unsafe.Pointer(&n[0]))))
	sceneUniforms[name] = loc
	return loc
}
func sceneInt(name string, value int) {
	if !lightPipeline.active {
		return
	}
	if previous, ok := sceneIntValues[name]; ok && previous == value {
		return
	}
	sceneIntValues[name] = value
	if loc := sceneLocation(name); loc >= 0 {
		glExt("glUniform1i", uintptr(loc), uintptr(value))
	}
}
func sceneFloat(name string, value float64) {
	if !lightPipeline.active {
		return
	}
	if previous, ok := sceneFloatValues[name]; ok && previous == value {
		return
	}
	sceneFloatValues[name] = value
	if loc := sceneLocation(name); loc >= 0 {
		glExt("glUniform1f", uintptr(loc), b32(value))
	}
}
func sceneVector(name string, value pos3) {
	if !lightPipeline.active {
		return
	}
	if previous, ok := sceneVectorValues[name]; ok && previous == value {
		return
	}
	sceneVectorValues[name] = value
	if loc := sceneLocation(name); loc >= 0 {
		v := [3]float32{float32(value.x), float32(value.y), float32(value.z)}
		glExt("glUniform3fv", uintptr(loc), 1, uintptr(unsafe.Pointer(&v[0])))
	}
}
func setWorldFade(value float64) { sceneFloat("uFade", value) }
func setInstancedWorld(value bool) {
	v := 0
	if value {
		v = 1
	}
	sceneInt("uInstanced", v)
}

func sceneMatrix(name string, m [16]float32) {
	if !lightPipeline.active {
		return
	}
	if loc := sceneLocation(name); loc >= 0 {
		glExt("glUniformMatrix4fv", uintptr(loc), 1, 0, uintptr(unsafe.Pointer(&m[0])))
	}
}
func sceneHour() float64 {
	if multiplayer050.Session != nil && !multiplayer050.Session.IsHost() && multiplayer050.Session.Connected() {
		return multiplayer050.View.World.Hour
	}
	if remaster.Enabled {
		return cycleHour034()
	}
	switch beta.cfg.TimeOfDay {
	case 1:
		return 13
	case 2:
		return 18.5
	case 3:
		return 22
	}
	return engine.MapHour(app.mapID)
}
func updateSceneLights(c engine.Car) {
	if !lightPipeline.active {
		return
	}
	var view [16]float32
	gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
	transform := func(p pos3, w float64) pos3 {
		return pos3{float64(view[0])*p.x + float64(view[4])*p.y + float64(view[8])*p.z + float64(view[12])*w, float64(view[1])*p.x + float64(view[5])*p.y + float64(view[9])*p.z + float64(view[13])*w, float64(view[2])*p.x + float64(view[6])*p.y + float64(view[10])*p.z + float64(view[14])*w}
	}
	inv := [16]float32{view[0], view[4], view[8], 0, view[1], view[5], view[9], 0, view[2], view[6], view[10], 0, 0, 0, 0, 1}
	for i := 0; i < 3; i++ {
		inv[12+i] = -(inv[i]*view[12] + inv[4+i]*view[13] + inv[8+i]*view[14])
	}
	sceneMatrix("uInverseView", inv)
	sceneVector("uSunEye", norm(transform(beta011CelestialDirection(), 0)))
	sceneVector("uUpEye", transform(pos3{0, 1, 0}, 0))
	sceneFloat("uWetness050", weather050.Amount)
	sceneFloat("uRainSky051", float64(boolInt(weather050.Rain && !workshop051.Active && !garage033)))
	night := engine.NightAmount(sceneHour())
	sceneFloat("uNight", night)
	sceneFloat("uSunset", sunset034())
	sceneInt("uFogEnabled", boolInt(beta.screen == screenDrive || remasterSceneOverlay()))
	sceneInt("uHDR", boolInt(hdrScene.active))
	sceneInt("uShadowPass", boolInt(sunShadow.drawing))
	sceneInt("uShadowEnabled", boolInt(sunShadow.ready && !sunShadow.drawing && shadowQuality026() > 0 && (beta.screen == screenDrive || remasterSceneOverlay())))
	sceneMatrix("uShadowMatrix", sunShadow.matrix)
	sceneMatrix("uShadowMatrix2", sunShadow.farMatrix)
	sceneFloat("uShadowTexel", 1./float64(sunShadow.size))
	sceneFloat("uShadowSplit", []float64{28, 38, 52, 72}[shadowQuality026()])
	sky := "sky_day.jpg"
	if night > .8 {
		sky = "sky_night.jpg"
	} else if sceneHour() > 15.5 {
		sky = "sky_sunset.jpg"
	}
	glExt("glActiveTexture", 0x84C4)
	envTexture := app.tex[sky]
	if beta020Environment() != 0 {
		envTexture = beta020Environment()
	}
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(envTexture))
	sceneInt("uEnvironmentHDR", boolInt(beta020Environment() != 0))
	_, next, blend := skyBlend034()
	glExt("glActiveTexture", 0x84CC)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(sky020.Textures[next]))
	sceneInt("uEnvironmentNext", 12)
	sceneFloat("uEnvironmentBlend", blend)
	gains040 := environmentSkyGains040()
	if location040 := sceneLocation("uEnvironmentGain040"); location040 >= 0 {
		glExt("glUniform2fv", uintptr(location040), 1, uintptr(unsafe.Pointer(&gains040[0])))
	}
	setEnvironmentAlignment040()
	glExt("glActiveTexture", 0x84C5)
	if !sunShadow.drawing {
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(sunShadow.texture))
	} else {
		gl("glBindTexture", GL_TEXTURE_2D, 0)
	}
	glExt("glActiveTexture", 0x84C6)
	if !sunShadow.drawing {
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(sunShadow.farTexture))
	} else {
		gl("glBindTexture", GL_TEXTURE_2D, 0)
	}
	glExt("glActiveTexture", 0x84C0)
	if remaster.Enabled {
		updateRemasterLights(c, view)
		setWorldFade(1)
		setInstancedWorld(false)
		return
	}
	count := 0
	addLight := func(p, d, color pos3, distance, outer, inner float64) {
		if count >= 12 {
			return
		}
		sceneVector(fmt.Sprintf("uLightPos[%d]", count), transform(p, 1))
		sceneVector(fmt.Sprintf("uLightDir[%d]", count), norm(transform(d, 0)))
		sceneVector(fmt.Sprintf("uLightColor[%d]", count), color)
		sceneVector(fmt.Sprintf("uLightParams[%d]", count), pos3{distance, outer, inner})
		count++
	}
	if beta.screen == screenDrive && !sunShadow.drawing {
		f := engine.HeadingVec(c.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		if beta.cfg.Headlights {
			for _, side := range []float64{-1, 1} {
				p := c.Pos.Add(f.Mul(2.2)).Add(r.Mul(.77 * side))
				addLight(pos3{p.X, c.Body.Height - .19, p.Z}, pos3{f.X * math.Cos(c.Rig.Pitch), -math.Sin(c.Rig.Pitch) - .045, f.Z * math.Cos(c.Rig.Pitch)}, pos3{.85, .93, 1}, 105, .88, .987)
			}
		}
		if c.Speed < -.1 || c.Gearbox.Gear < 0 {
			p := c.Pos.Sub(f.Mul(2.24))
			addLight(pos3{p.X, c.Body.Height - .19, p.Z}, pos3{-f.X, -.18, -f.Z}, pos3{1, .98, .91}, 22, .45, .9)
		}
		if engine.AutomaticLights(sceneHour()) {
			traffic := append([]engine.Traffic(nil), app.world.Traffic...)
			sort.Slice(traffic, func(i, j int) bool { return traffic[i].Pos.Dist(c.Pos) < traffic[j].Pos.Dist(c.Pos) })
			for _, t := range traffic {
				if count > 10 || t.Pos.Dist(c.Pos) > 100 {
					break
				}
				f := engine.HeadingVec(t.Heading)
				r := engine.Vec{X: f.Z, Z: -f.X}
				z, y := 2.1, .62
				if t.Variant%4 == 3 {
					z = 2.8
					y = 1.1
				}
				for _, side := range []float64{-1, 1} {
					p := t.Pos.Add(f.Mul(z)).Add(r.Mul(side * .78))
					addLight(pos3{p.X, engine.Terrain(p.X, p.Z) + y + .22, p.Z}, pos3{f.X, -.055, f.Z}, pos3{1, .90, .74}, 75, .86, .982)
				}
			}
		}
	}
	sceneInt("uVehicleLightCount", count)
	type lamp struct {
		p        pos3
		distance float64
	}
	nearest := make([]lamp, 0, 8)
	if engine.AutomaticLights(sceneHour()) && !sunShadow.drawing {
		for _, pole := range app.world.Poles {
			if pole.Falling {
				continue
			}
			d := pole.Pos.Dist(c.Pos)
			if d > 140 {
				continue
			}
			p, y := pole.LampHead()
			nearest = append(nearest, lamp{pos3{p.X, y, p.Z}, d})
		}
		sort.Slice(nearest, func(i, j int) bool { return nearest[i].distance < nearest[j].distance })
		if len(nearest) > 8 {
			nearest = nearest[:8]
		}
	}
	for i, l := range nearest {
		sceneVector(fmt.Sprintf("uStreetPos[%d]", i), transform(l.p, 1))
	}
	sceneInt("uStreetCount", len(nearest))
	sceneVector("uStreetDir", transform(pos3{0, -1, 0}, 0))
	setWorldFade(1)
	setInstancedWorld(false)
}

func sceneVec4(name string, x, y, z, w float64) {
	if !lightPipeline.active {
		return
	}
	if loc := sceneLocation(name); loc >= 0 {
		v := [4]float32{float32(x), float32(y), float32(z), float32(w)}
		glExt("glUniform4fv", uintptr(loc), 1, uintptr(unsafe.Pointer(&v[0])))
	}
}
