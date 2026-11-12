//go:build windows

package main

import (
	"fmt"
	"speedinneed/engine"
	"unsafe"
)

// All intermediate targets use linear RGBA16F. HUD and menus are composed
// afterwards and never enter bloom, occlusion, motion blur or sharpening.
type cinematicTarget034 struct {
	FBO, Texture  uint32
	Width, Height int
}

var cinematic034 struct {
	BloomProgram, EffectsProgram       uint32
	Bloom                              [4][2]cinematicTarget034
	Effects                            cinematicTarget034
	BloomActive, EffectsActive, Failed bool
	Locations                          map[uint32]map[string]int32
}

func cinematicLocation034(program uint32, name string) int32 {
	if cinematic034.Locations == nil {
		cinematic034.Locations = map[uint32]map[string]int32{}
	}
	if cinematic034.Locations[program] == nil {
		cinematic034.Locations[program] = map[string]int32{}
	}
	if v, ok := cinematic034.Locations[program][name]; ok {
		return v
	}
	n := append([]byte(name), 0)
	v := int32(glExt("glGetUniformLocation", uintptr(program), uintptr(unsafe.Pointer(&n[0]))))
	cinematic034.Locations[program][name] = v
	return v
}
func cinematicInt034(p uint32, name string, v int) {
	if loc := cinematicLocation034(p, name); loc >= 0 {
		glExt("glUniform1i", uintptr(loc), uintptr(v))
	}
}
func cinematicFloat034(p uint32, name string, v float64) {
	if loc := cinematicLocation034(p, name); loc >= 0 {
		glExt("glUniform1f", uintptr(loc), b32(v))
	}
}
func cinematicVec2034(p uint32, name string, x, y float32) {
	v := [2]float32{x, y}
	if loc := cinematicLocation034(p, name); loc >= 0 {
		glExt("glUniform2fv", uintptr(loc), 1, uintptr(unsafe.Pointer(&v[0])))
	}
}
func cinematicVec3034(p uint32, name string, v [3]float32) {
	if loc := cinematicLocation034(p, name); loc >= 0 {
		glExt("glUniform3fv", uintptr(loc), 1, uintptr(unsafe.Pointer(&v[0])))
	}
}
func cinematicMatrix034(p uint32, name string, v [16]float32) {
	if loc := cinematicLocation034(p, name); loc >= 0 {
		glExt("glUniformMatrix4fv", uintptr(loc), 1, 0, uintptr(unsafe.Pointer(&v[0])))
	}
}
func cinematicProgram034(fragment string) uint32 {
	vraw, err := assets.ReadFile("shaders/post.vert")
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		return 0
	}
	fraw, err := assets.ReadFile("shaders/" + fragment)
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		return 0
	}
	v, err := compileGLSL(0x8B31, string(vraw))
	if err != nil {
		pbrErrors = append(pbrErrors, fragment+": "+err.Error())
		return 0
	}
	f, err := compileGLSL(0x8B30, string(fraw))
	if err != nil {
		glExt("glDeleteShader", uintptr(v))
		pbrErrors = append(pbrErrors, fragment+": "+err.Error())
		return 0
	}
	p := uint32(glExt("glCreateProgram"))
	glExt("glAttachShader", uintptr(p), uintptr(v))
	glExt("glAttachShader", uintptr(p), uintptr(f))
	glExt("glLinkProgram", uintptr(p))
	var ok int32
	glExt("glGetProgramiv", uintptr(p), 0x8B82, uintptr(unsafe.Pointer(&ok)))
	glExt("glDeleteShader", uintptr(v))
	glExt("glDeleteShader", uintptr(f))
	if ok == 0 {
		pbrErrors = append(pbrErrors, fragment+": "+shaderLog(p, true))
		glExt("glDeleteProgram", uintptr(p))
		return 0
	}
	return p
}
func (t *cinematicTarget034) release() {
	if t.Texture != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&t.Texture)))
	}
	if t.FBO != 0 {
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&t.FBO)))
	}
	*t = cinematicTarget034{}
}
func (t *cinematicTarget034) ensure(w, h int, nearest bool) bool {
	if t.FBO != 0 && t.Width == w && t.Height == h {
		return true
	}
	t.release()
	t.Width, t.Height = w, h
	glExt("glActiveTexture", 0x84C0)
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&t.Texture)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(t.Texture))
	filter := uintptr(GL_LINEAR)
	if nearest {
		filter = 0x2600
	}
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, filter)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, filter)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881A, uintptr(w), uintptr(h), 0, GL_RGBA, glFloat, 0)
	glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&t.FBO)))
	glExt("glBindFramebuffer", 0x8D40, uintptr(t.FBO))
	glExt("glFramebufferTexture2D", 0x8D40, 0x8CE0, GL_TEXTURE_2D, uintptr(t.Texture), 0)
	if glExt("glCheckFramebufferStatus", 0x8D40) != 0x8CD5 {
		pbrErrors = append(pbrErrors, fmt.Sprintf("Cinematic target incomplete: %dx%d", w, h))
		t.release()
		return false
	}
	return true
}
func cinematicQuad034(t cinematicTarget034) {
	fullFrame037(t.Width, t.Height)
	glExt("glBindFramebuffer", 0x8D40, uintptr(t.FBO))
	gl("glViewport", 0, 0, uintptr(t.Width), uintptr(t.Height))
	begin(GL_QUADS)
	tex(0, 0)
	vtx(-1, -1, 0)
	tex(1, 0)
	vtx(1, -1, 0)
	tex(1, 1)
	vtx(1, 1, 0)
	tex(0, 1)
	vtx(-1, 1, 0)
	end()
}
func releaseCinematicTargets034() {
	for i := range cinematic034.Bloom {
		for j := range cinematic034.Bloom[i] {
			cinematic034.Bloom[i][j].release()
		}
	}
	cinematic034.Effects.release()
	cinematic034.BloomActive, cinematic034.EffectsActive = false, false
}
func releaseCinematic034() {
	releaseCinematicTargets034()
	for _, p := range []uint32{cinematic034.BloomProgram, cinematic034.EffectsProgram} {
		if p != 0 {
			glExt("glDeleteProgram", uintptr(p))
		}
	}
	cinematic034.BloomProgram, cinematic034.EffectsProgram = 0, 0
	cinematic034.Locations = nil
}
func prepareCinematic034() {
	cinematic034.BloomActive, cinematic034.EffectsActive = false, false
	defer func() {
		glExt("glBindFramebuffer", 0x8D40, 0)
		gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
		glExt("glActiveTexture", 0x84C0)
	}()
	if cinematic034.Failed || reducedEffects026() || app.quality == 0 {
		return
	}
	if beta.cfg.Bloom && beta.cfg.BloomStrength > .001 {
		if cinematic034.BloomProgram == 0 {
			cinematic034.BloomProgram = cinematicProgram034("bloom034.frag")
		}
		p := cinematic034.BloomProgram
		if p == 0 {
			cinematic034.Failed = true
			return
		}
		glExt("glUseProgram", uintptr(p))
		cinematicInt034(p, "uInput", 0)
		source, sw, sh := hdrScene.texture, hdrScene.width, hdrScene.height
		for level := 0; level < 4; level++ {
			w, h := max(1, sw/2), max(1, sh/2)
			pair := &cinematic034.Bloom[level]
			if !pair[0].ensure(w, h, false) || !pair[1].ensure(w, h, false) {
				cinematic034.Failed = true
				return
			}
			gl("glBindTexture", GL_TEXTURE_2D, uintptr(source))
			cinematicVec2034(p, "uTexel", 1/float32(sw), 1/float32(sh))
			cinematicInt034(p, "uMode", 0)
			cinematicInt034(p, "uThreshold", boolInt(level == 0))
			cinematicQuad034(pair[0])
			cinematicInt034(p, "uMode", 1)
			cinematicVec2034(p, "uTexel", 1/float32(w), 1/float32(h))
			cinematicVec2034(p, "uDirection", 1, 0)
			gl("glBindTexture", GL_TEXTURE_2D, uintptr(pair[0].Texture))
			cinematicQuad034(pair[1])
			cinematicVec2034(p, "uDirection", 0, 1)
			gl("glBindTexture", GL_TEXTURE_2D, uintptr(pair[1].Texture))
			cinematicQuad034(pair[0])
			source, sw, sh = pair[0].Texture, w, h
		}
		cinematic034.BloomActive = true
	}
	drive := beta.screen == screenDrive || remasterSceneOverlay()
	contacts := drive && sunShadow.ready && shadowQuality026() > 0
	shafts := contacts && beta.cfg.LightingDetail >= 1
	if beta.cfg.SSAO || contacts || shafts {
		if cinematic034.EffectsProgram == 0 {
			cinematic034.EffectsProgram = cinematicProgram034("effects034.frag")
		}
		p := cinematic034.EffectsProgram
		if p == 0 || !cinematic034.Effects.ensure(max(1, hdrScene.width/2), max(1, hdrScene.height/2), true) {
			cinematic034.Failed = true
			return
		}
		glExt("glUseProgram", uintptr(p))
		cinematicInt034(p, "uDepth", 0)
		cinematicInt034(p, "uFarShadow", 1)
		cinematicInt034(p, "uAO", boolInt(beta.cfg.SSAO))
		cinematicInt034(p, "uContacts", boolInt(contacts))
		cinematicInt034(p, "uShafts", boolInt(shafts))
		cinematicFloat034(p, "uNight", engine.NightAmount(sceneHour()))
		cinematicVec2034(p, "uTexel", 1/float32(hdrScene.width), 1/float32(hdrScene.height))
		var projection, view [16]float32
		gl("glGetFloatv", 0x0BA7, uintptr(unsafe.Pointer(&projection[0])))
		gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
		pv := [4]float32{projection[0], projection[5], .12, 4600}
		if at := cinematicLocation034(p, "uProjection"); at >= 0 {
			glExt("glUniform4fv", uintptr(at), 1, uintptr(unsafe.Pointer(&pv[0])))
		}
		inv := [16]float32{view[0], view[4], view[8], 0, view[1], view[5], view[9], 0, view[2], view[6], view[10], 0, 0, 0, 0, 1}
		for i := 0; i < 3; i++ {
			inv[12+i] = -(inv[i]*view[12] + inv[4+i]*view[13] + inv[8+i]*view[14])
		}
		cinematicMatrix034(p, "uInverseCamera", inv)
		cinematicMatrix034(p, "uFarMatrix", sunShadow.farMatrix)
		sun := beta011CelestialDirection()
		cinematicVec3034(p, "uSunEye", [3]float32{view[0]*float32(sun.x) + view[4]*float32(sun.y) + view[8]*float32(sun.z), view[1]*float32(sun.x) + view[5]*float32(sun.y) + view[9]*float32(sun.z), view[2]*float32(sun.x) + view[6]*float32(sun.y) + view[10]*float32(sun.z)})
		glExt("glActiveTexture", 0x84C1)
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(sunShadow.farTexture))
		glExt("glActiveTexture", 0x84C0)
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.depth))
		cinematicQuad034(cinematic034.Effects)
		cinematic034.EffectsActive = true
	}
}
func bindCinematic034() {
	p := hdrScene.program
	gel := boolInt(hdrScene.normals != 0 && app.quality >= 1 && !reducedEffects026() && beta.screen == screenDrive)
	if experimentRayReflections038() {
		gel = 0
	}
	cinematicInt034(p, "uSurfaceReflections", gel)
	cinematicInt034(p, "uNormals", 7)
	glExt("glActiveTexture", 0x84C7)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.normals))
	cinematicInt034(p, "uBloomEnabled", boolInt(cinematic034.BloomActive))
	cinematicInt034(p, "uEffectsEnabled", boolInt(cinematic034.EffectsActive))
	for i, name := range []string{"uBloom0", "uBloom1", "uBloom2", "uBloom3"} {
		glExt("glActiveTexture", uintptr(0x84C2+i))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(cinematic034.Bloom[i][0].Texture))
		cinematicInt034(p, name, 2+i)
	}
	glExt("glActiveTexture", 0x84C6)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(cinematic034.Effects.Texture))
	cinematicInt034(p, "uEffects", 6)
	if cinematic034.EffectsActive {
		cinematicVec2034(p, "uEffectsPixel", 1/float32(cinematic034.Effects.Width), 1/float32(cinematic034.Effects.Height))
	}
	cinematicFloat034(p, "uNight", engine.NightAmount(sceneHour()))
	cinematicFloat034(p, "uSunset", sunset034())
	glExt("glActiveTexture", 0x84C0)
}

func cinematicReport034() map[string]any {
	return map[string]any{"revision": "native-038.1", "local_reflection_buffer": hdrScene.normals != 0, "font_sdf": font035.Program != 0, "driver_skin_builds": driverCache035.Builds, "driver_skin_reuses": driverCache035.Reuses, "bloom_levels": 4, "bloom_active": cinematic034.BloomActive, "effects_active": cinematic034.EffectsActive, "failed": cinematic034.Failed, "effects_width": cinematic034.Effects.Width, "effects_height": cinematic034.Effects.Height, "ao_enabled": beta.cfg.SSAO, "contacts_enabled": beta.screen == screenDrive && sunShadow.ready && shadowQuality026() > 0, "precision": "linear RGBA16F"}
}
