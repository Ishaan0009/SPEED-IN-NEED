//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

func hdrRenderWidth() int {
	if hdrScene.active {
		return hdrScene.width
	}
	return app.width
}
func hdrRenderHeight() int {
	if hdrScene.active {
		return hdrScene.height
	}
	return app.height
}
func postOptions() {
	set := func(name string, value int) {
		loc := postLocation024(name)
		if loc >= 0 {
			glExt("glUniform1i", uintptr(loc), uintptr(value))
		}
	}
	glExt("glActiveTexture", 0x84C1)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.depth))
	set("uDepth", 1)
	set("uAO", boolInt(beta.cfg.SSAO && !reducedEffects026()))
	set("uFXAA", boolInt(beta.cfg.FXAA && !experimentDLAA038()))
	set("uGrade", beta.cfg.ColorGrade)
	set("uReflections", boolInt(app.quality >= 1 && !reducedEffects026() && beta.screen == screenDrive))
	var projection, view [16]float32
	gl("glGetFloatv", 0x0BA7, uintptr(unsafe.Pointer(&projection[0])))
	gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
	postMotionMatrices034(projection, view)
	values := [4]float32{projection[0], projection[5], .12, 4600}
	if at := postLocation024("uProjection"); at >= 0 {
		glExt("glUniform4fv", uintptr(at), 1, uintptr(unsafe.Pointer(&values[0])))
	}
	up := [3]float32{view[4], view[5], view[6]}
	if at := postLocation024("uWaterUp"); at >= 0 {
		glExt("glUniform3fv", uintptr(at), 1, uintptr(unsafe.Pointer(&up[0])))
	}
	loc := postLocation024("uMotion")
	motion := 0.0
	if beta.cfg.MotionBlur && !reducedEffects026() && beta.screen == screenDrive && !remaster.FullMap && !app.paused {
		motion = engine.Clamp((math.Abs(app.renderCar.Speed)-13)/45, 0, 1) * .042 * beta.cfg.BlurStrength
		if app.boolFPP {
			motion *= .60
		}
		if app.renderCar.Braking {
			motion *= .55
		}
	}
	if loc >= 0 {
		glExt("glUniform1f", uintptr(loc), b32(motion))
	}
	setFloat := func(name string, value float64) {
		if slot := postLocation024(name); slot >= 0 {
			glExt("glUniform1f", uintptr(slot), b32(value))
		}
	}
	setFloat("uExposure", beta.cfg.Exposure)
	setFloat("uSaturation", beta.cfg.Saturation)
	setFloat("uBloomStrength", beta.cfg.BloomStrength)
	glExt("glActiveTexture", 0x84C0)
}
func aokawaResizeShadows() {
	want := []int{512, 2048, 3072, 4096}[shadowQuality026()]
	if sunShadow.ready && want != sunShadow.size {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&sunShadow.texture)))
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&sunShadow.farTexture)))
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&sunShadow.fbo)))
		sunShadow.ready = false
		sunShadow.frame = 0
	}
}
