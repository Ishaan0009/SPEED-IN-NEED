package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"unsafe"
)

type motionDraw038 struct {
	Matrix [16]float32
	Frame  uint64
}

var temporal038 struct {
	Frame                            uint64
	Projection, View, VP, PreviousVP [16]float32
	Flow, PreviousFlow               float64
	JitterX, JitterY                 float32
	Valid, Active                    bool
	Width, Height, Map, Camera       int
	Entity                           int
	Draws                            map[string]motionDraw038
	ObjectSamples, Resets            uint64
}

func inverseRigid038(view [16]float32) [16]float32 {
	inv := [16]float32{view[0], view[4], view[8], 0, view[1], view[5], view[9], 0, view[2], view[6], view[10], 0, 0, 0, 0, 1}
	for i := 0; i < 3; i++ {
		inv[12+i] = -(inv[i]*view[12] + inv[4+i]*view[13] + inv[8+i]*view[14])
	}
	return inv
}
func invalidateTemporal038() {
	temporal038.Valid = false
	temporal038.Active = false
	temporal038.Resets++
}

func beginTemporal038() {
	if hdrScene.motion == 0 || sunShadow.drawing {
		return
	}
	var projection, view [16]float32
	gl("glGetFloatv", 0x0BA7, uintptr(unsafe.Pointer(&projection[0])))
	gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
	camera := beta.cfg.CameraMode
	if app.boolFPP {
		camera += 100
	}
	reset := !temporal038.Valid || temporal038.Width != hdrScene.width || temporal038.Height != hdrScene.height || temporal038.Map != app.mapID || temporal038.Camera != camera || app.world.Elapsed < temporal038.Flow
	if temporal038.Valid {
		// The inverse camera translation is available without an extra GL query.
		a, b := inverseRigid038(view), inverseRigid038(temporal038.View)
		distance := math.Sqrt(float64((a[12]-b[12])*(a[12]-b[12]) + (a[13]-b[13])*(a[13]-b[13]) + (a[14]-b[14])*(a[14]-b[14])))
		reset = reset || distance > 18 || math.Abs(float64(projection[0]-temporal038.Projection[0])) > .02
	}
	temporal038.Frame++
	temporal038.Projection = projection
	temporal038.View = view
	vp := matrixMultiply(projection, view)
	if reset {
		temporal038.PreviousVP = vp
		temporal038.PreviousFlow = app.world.Elapsed
		temporal038.Draws = map[string]motionDraw038{}
		skinHistory038 = map[string]*skinHistoryFrame038{}
		temporal038.Resets++
	} else {
		temporal038.PreviousVP = temporal038.VP
		temporal038.PreviousFlow = temporal038.Flow
	}
	temporal038.VP = vp
	temporal038.Flow = app.world.Elapsed
	temporal038.Valid = true
	temporal038.Active = true
	temporal038.Entity = 0
	beginRayFrame038()
	temporal038.Width = hdrScene.width
	temporal038.Height = hdrScene.height
	temporal038.Map = app.mapID
	temporal038.Camera = camera
	if temporal038.Frame%120 == 0 {
		for key, history := range skinHistory038 {
			if history.Frame+120 < temporal038.Frame {
				delete(skinHistory038, key)
			}
		}
	}
	temporal038.JitterX, temporal038.JitterY = 0, 0
	if experimentDLAA038() {
		halton := func(index, base uint64) float32 {
			fraction, value := float32(1), float32(0)
			for index > 0 {
				fraction /= float32(base)
				value += fraction * float32(index%base)
				index /= base
			}
			return value - .5
		}
		temporal038.JitterX = halton(temporal038.Frame%16+1, 2)
		temporal038.JitterY = halton(temporal038.Frame%16+1, 3)
		jittered := projection
		jittered[8] -= 2 * temporal038.JitterX / float32(hdrScene.width)
		jittered[9] -= 2 * temporal038.JitterY / float32(hdrScene.height)
		gl("glMatrixMode", 0x1701)
		gl("glLoadMatrixf", uintptr(unsafe.Pointer(&jittered[0])))
		gl("glMatrixMode", 0x1700)
	}
}
func setupTemporal038() {
	sceneInt("uRayReflections038", boolInt(experimentRayReflections038()))
	sceneInt("uObjectMotion038", 0)
	sceneInt("uPreviousVertex038", 0)
	sceneFloat("uPreviousFlow038", temporal038.PreviousFlow)
	sceneMatrix("uPreviousVP038", temporal038.PreviousVP)
	sceneMatrix("uUnjitteredProjection038", temporal038.Projection)
}
func objectMotion038(key string, slot int) {
	if !temporal038.Active || !hdrScene.active || sunShadow.drawing || temporal038.Entity == 0 {
		return
	}
	var view [16]float32
	gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
	recordRayActor038(key, slot, view)
	current := matrixMultiply(temporal038.Projection, view)
	name := fmt.Sprintf("%d/%s/%d", temporal038.Entity, key, slot)
	old, ok := temporal038.Draws[name]
	previous := current
	if ok && old.Frame+1 == temporal038.Frame {
		previous = old.Matrix
	}
	temporal038.Draws[name] = motionDraw038{current, temporal038.Frame}
	temporal038.ObjectSamples++
	sceneMatrix("uPreviousObject038", previous)
	sceneInt("uObjectMotion038", 1)
}

func motionTexture038(width, height int) {
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&hdrScene.motion)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.motion))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881A, uintptr(width), uintptr(height), 0, GL_RGBA, glFloat, 0)
	glExt("glFramebufferTexture2D", 0x8D40, 0x8CE2, GL_TEXTURE_2D, uintptr(hdrScene.motion), 0)
}
func clearMotion038() {
	if hdrScene.active && hdrScene.motion != 0 && glExtAvailable("glClearBufferfv") {
		zero := [4]float32{}
		glExt("glClearBufferfv", 0x1800, 2, uintptr(unsafe.Pointer(&zero[0])))
		if hdrScene.sun != 0 {
			glExt("glClearBufferfv", 0x1800, 3, uintptr(unsafe.Pointer(&zero[0])))
		}
	}
}
func captureMotion038(name string) {
	if hdrScene.motion == 0 {
		return
	}
	var oldFBO, oldRead int32
	gl("glGetIntegerv", 0x8CA6, uintptr(unsafe.Pointer(&oldFBO)))
	gl("glGetIntegerv", 0x0C02, uintptr(unsafe.Pointer(&oldRead)))
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	gl("glReadBuffer", 0x8CE2)
	w, h := hdrScene.width, hdrScene.height
	data := make([]float32, w*h*4)
	gl("glReadPixels", 0, 0, uintptr(w), uintptr(h), GL_RGBA, glFloat, uintptr(unsafe.Pointer(&data[0])))
	glExt("glBindFramebuffer", 0x8D40, uintptr(oldFBO))
	gl("glReadBuffer", uintptr(oldRead))
	valid, moving, invalid := 0, 0, 0
	maximum, sum := 0., 0.
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(data); i += 4 {
		x, y := float64(data[i]), float64(data[i+1])
		if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || math.Abs(x) > 1 || math.Abs(y) > 1 {
			invalid++
			continue
		}
		if data[i+3] < .5 {
			continue
		}
		valid++
		length := math.Hypot(x*float64(w), y*float64(h))
		sum += length
		maximum = math.Max(maximum, length)
		if length > .05 {
			moving++
		}
		pixel := i / 4
		at := ((h-1-pixel/w)*w + pixel%w) * 4
		im.Pix[at] = uint8(math.Max(0, math.Min(255, 128+x*float64(w)*4)))
		im.Pix[at+1] = uint8(math.Max(0, math.Min(255, 128+y*float64(h)*4)))
		im.Pix[at+2] = uint8(math.Min(255, length*10))
		im.Pix[at+3] = 255
	}
	report := map[string]any{"status": "PASS_INPUT_FINITE", "scene": name, "width": w, "height": h, "valid_pixels": valid, "moving_pixels": moving, "invalid_pixels": invalid, "maximum_motion_pixels": maximum, "mean_motion_pixels": sum / math.Max(1, float64(valid)), "object_matrix_samples": temporal038.ObjectSamples, "history_resets": temporal038.Resets, "skin_particle_history_complete": false, "dlss_evaluated": false}
	if invalid != 0 || valid == 0 {
		report["status"] = "FAIL"
	}
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(filepath.Join(qa.dir, name+"_motion038.json"), raw, 0644)
	if f, e := os.Create(filepath.Join(qa.dir, name+"_motion038.png")); e == nil {
		png.Encode(f, im)
		f.Close()
	}
}
func temporalOutputs038(enabled bool) {
	if hdrScene.motion == 0 {
		return
	}
	outputs := [4]uint32{0x8CE0, 0x8CE1, 0x8CE2, 0x8CE3}
	count := uintptr(1)
	if enabled {
		count = 3
		if hdrScene.sun != 0 {
			count = 4
		}
	}
	glExt("glDrawBuffers", count, uintptr(unsafe.Pointer(&outputs[0])))
}
