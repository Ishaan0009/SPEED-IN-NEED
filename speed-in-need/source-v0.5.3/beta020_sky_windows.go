package main

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"speedinneed/engine"
	"unsafe"
)

var sky020Locations = map[string]uintptr{}

var sky020 struct {
	Textures   [3]uint32
	Program    uint32
	FOV        float64
	Directions [3]pos3
	Sizes      [3][2]int
}

func initBeta020Sky() {
	pbrTexture("b035_cloud_volume.png")
	raw, _ := assets.ReadFile("assets/beta020_sky.json")
	var metadata []struct{ Direction [3]float64 }
	json.Unmarshal(raw, &metadata)
	for i, label := range []string{"day", "afternoon", "night"} {
		if i < len(metadata) {
			v := metadata[i].Direction
			sky020.Directions[i] = pos3{v[0], v[1], v[2]}
		}
		if bootstrapSky040(i) {
			continue
		}
		suffix := ""
		cap := textureCap024()
		if cap <= 512 {
			suffix = "_512"
		} else if cap <= 1024 {
			suffix = "_1024"
		}
		var maxTexture int32
		gl("glGetIntegerv", 0x0d33, uintptr(unsafe.Pointer(&maxTexture)))
		data, err := assets.ReadFile("assets/b034_sky_" + label + suffix + ".f16")
		if err != nil {
			continue
		}
		if len(data) < 12 || string(data[:4]) != "HDR2" {
			pbrErrors = append(pbrErrors, "Invalid HDR sky "+label)
			continue
		}
		w, h := int(binary.LittleEndian.Uint32(data[4:])), int(binary.LittleEndian.Uint32(data[8:]))
		if w < 1 || h < 1 || w > 8192 || h > 4096 || len(data) != 12+w*h*6 {
			pbrErrors = append(pbrErrors, "HDR sky dimensions invalid")
			continue
		}
		sky020.Sizes[i] = [2]int{w, h}
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(&sky020.Textures[i])))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(sky020.Textures[i]))
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2703)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2601)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x2901)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812f)
		gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881b, uintptr(w), uintptr(h), 0, 0x1907, 0x140b, uintptr(unsafe.Pointer(&data[12])))
		glExt("glGenerateMipmap", GL_TEXTURE_2D)
		if i < len(metadata) {
			v := metadata[i].Direction
			sky020.Directions[i] = pos3{v[0], v[1], v[2]}
		}
	}
	vr, ev := assets.ReadFile("shaders/sky020.vert")
	fr, ef := assets.ReadFile("shaders/sky020.frag")
	if ev != nil || ef != nil {
		return
	}
	v, e := compileGLSL(0x8b31, string(vr))
	if e != nil {
		pbrErrors = append(pbrErrors, e.Error())
		return
	}
	f, e := compileGLSL(0x8b30, string(fr))
	if e != nil {
		glExt("glDeleteShader", uintptr(v))
		pbrErrors = append(pbrErrors, e.Error())
		return
	}
	p := uint32(glExt("glCreateProgram"))
	glExt("glAttachShader", uintptr(p), uintptr(v))
	glExt("glAttachShader", uintptr(p), uintptr(f))
	glExt("glLinkProgram", uintptr(p))
	glExt("glDeleteShader", uintptr(v))
	glExt("glDeleteShader", uintptr(f))
	var valid int32
	glExt("glGetProgramiv", uintptr(p), 0x8b82, uintptr(unsafe.Pointer(&valid)))
	if valid == 0 {
		pbrErrors = append(pbrErrors, shaderLog(p, true))
		glExt("glDeleteProgram", uintptr(p))
		return
	}
	sky020.Program = p
}

func beta020Environment() uint32 {
	if garage033 {
		return sky020.Textures[0]
	}
	i, _, _ := skyBlend034()
	return sky020.Textures[i]
}

var skyContextLoaded051 = -1

func ensureSkyContext051() {
	context := 0
	if weather050.Rain {
		context = 1
	}
	if garage033 {
		context = 2
	}
	if workshop051.Active {
		context = 3
	}
	if context != skyContextLoaded051 {
		skyContextLoaded051 = context
		reloadEnvironment051()
	}
}
func drawBeta020Sky() bool {
	ensureSkyContext051()

	pumpSky040()
	texture := beta020Environment()
	if texture == 0 || sky020.Program == 0 {
		return false
	}
	var view [16]float32
	gl("glGetFloatv", 0x0ba6, uintptr(unsafe.Pointer(&view[0])))
	inv := [16]float32{view[0], view[4], view[8], 0, view[1], view[5], view[9], 0, view[2], view[6], view[10], 0, 0, 0, 0, 1}
	useLighting(false)
	dis(GL_DEPTH_TEST)
	dis(GL_BLEND)
	gl("glDepthMask", 0)
	glExt("glUseProgram", uintptr(sky020.Program))
	glExt("glActiveTexture", 0x84c0)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
	loc := func(name string) uintptr {
		if value, ok := sky020Locations[name]; ok {
			return value
		}
		b := append([]byte(name), 0)
		value := glExt("glGetUniformLocation", uintptr(sky020.Program), uintptr(unsafe.Pointer(&b[0])))
		sky020Locations[name] = value
		return value
	}
	glExt("glUniform1i", loc("uSky"), 0)
	_, next, blend := skyBlend034()
	glExt("glActiveTexture", 0x84C1)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(sky020.Textures[next]))
	glExt("glUniform1i", loc("uSkyNext"), 1)
	glExt("glUniform1f", loc("uSkyBlend"), b32(blend))
	gains040 := skyDisplayGains040()
	glExt("glUniform2fv", loc("uSkyGain040"), 1, uintptr(unsafe.Pointer(&gains040[0])))
	align040 := skyAlignment040()
	glExt("glUniform4fv", loc("uPanoramaAlign040"), 1, uintptr(unsafe.Pointer(&align040[0][0])))
	glExt("glUniform4fv", loc("uPanoramaAlignNext040"), 1, uintptr(unsafe.Pointer(&align040[1][0])))
	glExt("glUniform1f", loc("uProceduralCelestial040"), b32(float64(proceduralCelestialWeight040())))
	dir := beta011CelestialDirection()
	light := [3]float32{float32(dir.x), float32(dir.y), float32(dir.z)}
	glExt("glUniform3fv", loc("uCelestial"), 1, uintptr(unsafe.Pointer(&light[0])))
	glExt("glUniform1f", loc("uNight"), b32(engine.NightAmount(sceneHour())))
	glExt("glUniform1f", loc("uSkyTime"), b32(app.t))
	glExt("glUniform1f", loc("uRainSky051"), b32(float64(boolInt(weather050.Rain && !workshop051.Active))))
	glExt("glUniform1f", loc("uIndoor051"), b32(float64(boolInt(workshop051.Active))))
	glExt("glActiveTexture", 0x84C0)
	glExt("glUniform1i", loc("uHDR"), uintptr(boolInt(hdrScene.active)))
	glExt("glUniform1f", loc("uAspect"), b32(float64(app.width)/float64(app.height)))
	glExt("glUniform1f", loc("uTanHalfFov"), b32(math.Tan(sky020.FOV*math.Pi/360)))
	jitter038 := [2]float32{}
	if hdrScene.active && experimentDLAA038() {
		jitter038 = [2]float32{2 * temporal038.JitterX / float32(hdrScene.width), 2 * temporal038.JitterY / float32(hdrScene.height)}
	}
	glExt("glUniform2fv", loc("uSkyJitter038"), 1, uintptr(unsafe.Pointer(&jitter038[0])))
	glExt("glUniformMatrix4fv", loc("uCameraRotation"), 1, 0, uintptr(unsafe.Pointer(&inv[0])))
	begin(GL_QUADS)
	vtx(-1, -1, 0)
	vtx(1, -1, 0)
	vtx(1, 1, 0)
	vtx(-1, 1, 0)
	end()
	glExt("glUseProgram", 0)
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
	en(GL_BLEND)
	gpuWorld.draws++
	return true
}

func releaseBeta020Sky() {
	releaseSky040()
	for i := range sky020.Textures {
		if sky020.Textures[i] != 0 {
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&sky020.Textures[i])))
			sky020.Textures[i] = 0
		}
	}
	if sky020.Program != 0 {
		glExt("glDeleteProgram", uintptr(sky020.Program))
		sky020.Program = 0
	}
}
