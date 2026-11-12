//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

var hdrScene struct {
	fbo, texture, depth, normals, motion, sun, program uint32
	width, height                                      int
	active, failed                                     bool
}
var rayNormals038 bool
var sunShadow struct {
	fbo, texture, farTexture uint32
	matrix, farMatrix        [16]float32
	size, frame              int
	ready, drawing           bool
}

func buildPostProgram() uint32 {
	vraw, _ := assets.ReadFile("shaders/post.vert")
	fraw, _ := assets.ReadFile("shaders/post.frag")
	v, err := compileGLSL(0x8B31, string(vraw))
	if err != nil {
		return 0
	}
	f, err := compileGLSL(0x8B30, string(fraw))
	if err != nil {
		glExt("glDeleteShader", uintptr(v))
		return 0
	}
	p := uint32(glExt("glCreateProgram"))
	glExt("glAttachShader", uintptr(p), uintptr(v))
	glExt("glAttachShader", uintptr(p), uintptr(f))
	glExt("glLinkProgram", uintptr(p))
	var status int32
	glExt("glGetProgramiv", uintptr(p), 0x8B82, uintptr(unsafe.Pointer(&status)))
	glExt("glDeleteShader", uintptr(v))
	glExt("glDeleteShader", uintptr(f))
	if status == 0 {
		pbrErrors = append(pbrErrors, "Post link: "+shaderLog(p, true))
		glExt("glDeleteProgram", uintptr(p))
		return 0
	}
	return p
}
func destroyHDR() {
	rayNormals038 = false
	releaseTemporalTransport038()
	invalidateTemporal038()
	if hdrScene.sun != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&hdrScene.sun)))
		hdrScene.sun = 0
	}
	if hdrScene.motion != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&hdrScene.motion)))
		hdrScene.motion = 0
	}
	releaseCinematicTargets034()
	if hdrScene.normals != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&hdrScene.normals)))
		hdrScene.normals = 0
	}
	if hdrScene.texture != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&hdrScene.texture)))
	}
	if hdrScene.depth != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&hdrScene.depth)))
	}
	if hdrScene.fbo != 0 {
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&hdrScene.fbo)))
	}
	hdrScene.texture = 0
	hdrScene.depth = 0
	hdrScene.fbo = 0
}
func beginHDRScene() {
	fullFrame037(app.width, app.height)
	if lightPipeline.program == 0 || hdrScene.failed {
		return
	}
	if !glExtAvailable("glGenFramebuffers") || !glExtAvailable("glGenRenderbuffers") {
		hdrScene.failed = true
		return
	}
	if hdrScene.program == 0 {
		hdrScene.program = buildPostProgram()
		if hdrScene.program == 0 {
			hdrScene.failed = true
			return
		}
	}
	scale := renderScale026()
	rw, rh := int(float64(app.width)*scale), int(float64(app.height)*scale)
	rw, rh = temporalInputSize038(rw, rh)
	if hdrScene.width != rw || hdrScene.height != rh || hdrScene.fbo == 0 || (hdrScene.sun != 0) != experimentRayShadows038() || rayNormals038 != experimentRayReflections038() {
		destroyHDR()
		hdrScene.width = rw
		hdrScene.height = rh
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(&hdrScene.texture)))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.texture))
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
		gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881A, uintptr(rw), uintptr(rh), 0, GL_RGBA, glFloat, 0)
		glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&hdrScene.fbo)))
		glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
		glExt("glFramebufferTexture2D", 0x8D40, 0x8CE0, GL_TEXTURE_2D, uintptr(hdrScene.texture), 0)
		// View normals plus an explicit reflective-material mask make screen
		// reflections respond to the geometry surrounding the car.
		if glExtAvailable("glDrawBuffers") {
			gl("glGenTextures", 1, uintptr(unsafe.Pointer(&hdrScene.normals)))
			gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.normals))
			gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2600)
			gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2600)
			gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
			gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
			if experimentRayReflections038() {
				rayNormals038 = true
				gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x8814, uintptr(rw), uintptr(rh), 0, GL_RGBA, glFloat, 0)
			} else {
				gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x8058, uintptr(rw), uintptr(rh), 0, GL_RGBA, GL_UNSIGNED_BYTE, 0)
			}
			glExt("glFramebufferTexture2D", 0x8D40, 0x8CE1, GL_TEXTURE_2D, uintptr(hdrScene.normals), 0)
			if outputFBO038 != 0 {
				motionTexture038(rw, rh)
				if experimentRayShadows038() {
					sunTexture038(rw, rh)
				}
				temporalOutputs038(true)
			} else {
				outputs := [2]uint32{0x8CE0, 0x8CE1}
				glExt("glDrawBuffers", 2, uintptr(unsafe.Pointer(&outputs[0])))
			}
		}
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(&hdrScene.depth)))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.depth))
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2601)
		// This is sampled as a regular depth texture by AO and motion blur.
		gl("glTexParameteri", GL_TEXTURE_2D, 0x884c, 0)
		gl("glTexParameteri", GL_TEXTURE_2D, 0x884d, 0x0203)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2601)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
		gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x81A6, uintptr(rw), uintptr(rh), 0, 0x1902, 0x1405, 0)
		glExt("glFramebufferTexture2D", 0x8D40, 0x8D00, GL_TEXTURE_2D, uintptr(hdrScene.depth), 0)
		if glExt("glCheckFramebufferStatus", 0x8D40) != 0x8CD5 {
			destroyHDR()
			hdrScene.failed = true
			glExt("glBindFramebuffer", 0x8D40, 0)
			return
		}
	}
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	hdrScene.active = true
}
func finishHDRScene() {
	if !hdrScene.active {
		return
	}
	restore038 := reconstructForPost038()
	defer restore038()
	useLighting(false)
	glExt("glBindFramebuffer", 0x8D40, 0)
	hdrScene.active = false
	gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	dis(GL_BLEND)
	gl("glDepthMask", 0)
	fullFrame037(app.width, app.height)
	prepareCinematic034()
	fullFrame037(app.width, app.height)
	glExt("glUseProgram", uintptr(hdrScene.program))
	glExt("glActiveTexture", 0x84C0)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.texture))
	loc := postLocation024("uScene")
	glExt("glUniform1i", uintptr(loc), 0)
	loc = postLocation024("uPixel")
	pixel := [3]float32{1 / float32(hdrScene.width), 1 / float32(hdrScene.height), float32(boolInt(beta.cfg.Bloom && !reducedEffects026()))}
	glExt("glUniform3fv", uintptr(loc), 1, uintptr(unsafe.Pointer(&pixel[0])))
	postOptions()
	bindCinematic034()
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
	glExt("glUseProgram", 0)
	gl("glDepthMask", 1)
	en(GL_BLEND)
	en(GL_DEPTH_TEST)
}
func dot(a, b pos3) float64 { return a.x*b.x + a.y*b.y + a.z*b.z }
func matrixMultiply(a, b [16]float32) (c [16]float32) {
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			for k := 0; k < 4; k++ {
				c[col*4+row] += a[k*4+row] * b[col*4+k]
			}
		}
	}
	return
}
func initSunShadows() {
	if sunShadow.ready || !glExtAvailable("glGenFramebuffers") {
		return
	}
	sunShadow.size = []int{512, 2048, 3072, 4096}[engine.ClampInt(shadowQuality026(), 0, 3)]
	glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&sunShadow.fbo)))
	for i, p := range []*uint32{&sunShadow.texture, &sunShadow.farTexture} {
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(p)))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(*p))
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2601)
		gl("glTexParameteri", GL_TEXTURE_2D, 0x884c, 0x884e)
		gl("glTexParameteri", GL_TEXTURE_2D, 0x884d, 0x0203)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2601)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
		size := sunShadow.size
		if i == 1 {
			size /= 2
		}
		gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x81A6, uintptr(size), uintptr(size), 0, 0x1902, 0x1405, 0)
	}
	glExt("glBindFramebuffer", 0x8D40, uintptr(sunShadow.fbo))
	glExt("glFramebufferTexture2D", 0x8D40, 0x8D00, GL_TEXTURE_2D, uintptr(sunShadow.texture), 0)
	gl("glDrawBuffer", 0)
	gl("glReadBuffer", 0)
	sunShadow.ready = glExt("glCheckFramebufferStatus", 0x8D40) == 0x8CD5
	glExt("glBindFramebuffer", 0x8D40, 0)
}
func renderSunShadows(c engine.Car) {
	// The night lighting shader does not sample these sun maps. Avoid rendering
	// two unused cascades, and refresh both when daylight returns.

	if shadowQuality026() == 0 || lightPipeline.program == 0 {
		return
	}
	aokawaResizeShadows()
	initSunShadows()
	if !sunShadow.ready {
		return
	}
	sunShadow.frame++
	glExt("glBindFramebuffer", 0x8D40, uintptr(sunShadow.fbo))
	sunShadow.drawing = true
	gl("glColorMask", 0, 0, 0, 0)
	dis(GL_BLEND)
	en(GL_DEPTH_TEST)
	en(0x8037)
	gl("glPolygonOffset", b32(2), b32(3))
	for cascade := 0; cascade < 2; cascade++ {
		if sunShadow.frame > 1 && cascade == 1 && sunShadow.frame%4 != 0 {
			continue
		}
		radius := []float64{40, 52, 70, 95}[shadowQuality026()]
		size := sunShadow.size
		texture := sunShadow.texture
		if cascade == 1 {
			radius = []float64{120, 190, 280, 380}[shadowQuality026()]
			size /= 2
			texture = sunShadow.farTexture
		}
		glExt("glFramebufferTexture2D", 0x8D40, 0x8D00, GL_TEXTURE_2D, uintptr(texture), 0)
		gl("glViewport", 0, 0, uintptr(size), uintptr(size))
		gl("glClear", GL_DEPTH_BUFFER_BIT)
		gl("glMatrixMode", GL_PROJECTION)
		proj := [16]float32{float32(1 / radius), 0, 0, 0, 0, float32(1 / radius), 0, 0, 0, 0, -2. / 1000, 0, 0, 0, -1, 1}
		gl("glLoadMatrixf", uintptr(unsafe.Pointer(&proj[0])))
		gl("glMatrixMode", GL_MODELVIEW)
		gl("glLoadIdentity")
		// Quantize the light-space origin to avoid shadow crawling during motion.
		light := beta011CelestialDirection()
		target := pos3{c.Pos.X, app.world.GroundHeight(c.Pos), c.Pos.Z}
		right := norm(cross(scale(light, -1), pos3{0, 1, 0}))
		up := norm(cross(right, scale(light, -1)))
		texel := radius * 2 / float64(size)
		rx, uy := dot(target, right), dot(target, up)
		target = add(target, add(scale(right, math.Round(rx/texel)*texel-rx), scale(up, math.Round(uy/texel)*texel-uy)))
		eye := add(target, scale(light, 450))
		applyCamera(eye, target)
		var view [16]float32
		gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
		m := matrixMultiply(proj, view)
		if cascade == 0 {
			sunShadow.matrix = m
		} else {
			sunShadow.farMatrix = m
		}
		useLighting(true)
		updateSceneLights(c)
		if remaster.Enabled && app.world.Remaster != nil {
			drawRemasterWorld(eye, target, true, radius)
			drawRemasterActors(true, radius)
			continue
		}
		drawBufferedWorld(eye, target)
		drawPBRBuildings()
		renderPoles()
		drawCar(c.Pos, c.Heading, 0, true, false)
		for _, t := range app.world.Traffic {
			if t.Pos.Dist(c.Pos) < radius {
				drawCar(t.Pos, t.Heading, t.Variant, false, false)
			}
		}
	}
	useLighting(false)
	sunShadow.drawing = false
	gl("glColorMask", 1, 1, 1, 1)
	dis(0x8037)
	en(GL_BLEND)
	if hdrScene.active {
		glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	} else {
		glExt("glBindFramebuffer", 0x8D40, 0)
	}
}
func releaseBetaRenderTargets() {
	releaseCinematic034()
	destroyHDR()
	if hdrScene.program != 0 {
		glExt("glDeleteProgram", uintptr(hdrScene.program))
	}
	for _, p := range []*uint32{&sunShadow.texture, &sunShadow.farTexture} {
		if *p != 0 {
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(p)))
		}
	}
	if sunShadow.fbo != 0 {
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&sunShadow.fbo)))
	}
}
