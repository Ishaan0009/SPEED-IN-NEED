package main

import (
	"math"
	"unsafe"
)

var motion034 struct {
	Ready          bool
	ViewProjection [16]float32
	Eye            pos3
	Screen         int
}

// Reconstruct a world point from depth and project it through the previous
// camera. This captures camera rotation and translation, unlike radial blur.
func postMotionMatrices034(projection, view [16]float32) {
	vp := matrixMultiply(projection, view)
	inv := [16]float32{view[0], view[4], view[8], 0, view[1], view[5], view[9], 0, view[2], view[6], view[10], 0, 0, 0, 0, 1}
	for i := 0; i < 3; i++ {
		inv[12+i] = -(inv[i]*view[12] + inv[4+i]*view[13] + inv[8+i]*view[14])
	}
	if !motion034.Ready || motion034.Screen != beta.screen || math.Hypot(app.cam.x-motion034.Eye.x, app.cam.z-motion034.Eye.z) > 20 || remasterSceneOverlay() || beta.screen != screenDrive {
		motion034.ViewProjection = vp
	}
	for name, m := range map[string][16]float32{"uPreviousVP": motion034.ViewProjection, "uInverseCamera": inv} {
		if loc := postLocation024(name); loc >= 0 {
			glExt("glUniformMatrix4fv", uintptr(loc), 1, 0, uintptr(unsafe.Pointer(&m[0])))
		}
	}
	if loc := postLocation024("uShutter"); loc >= 0 {
		glExt("glUniform1f", uintptr(loc), b32(math.Min(1, 16.667/math.Max(1, app.frameMS))))
	}
	motion034.Ready = true
	motion034.ViewProjection = vp
	motion034.Eye = app.cam
	motion034.Screen = beta.screen
}
