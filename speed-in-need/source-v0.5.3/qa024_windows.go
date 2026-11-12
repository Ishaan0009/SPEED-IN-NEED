//go:build windows

package main

import (
	"os"
	"strconv"
	"unsafe"
)

func applyQATextureBudget024() {
	if n, e := strconv.Atoi(os.Getenv("SPEED_QA_TEXTURE_BUDGET")); e == nil && n >= 0 && n <= 4 {
		beta.cfg.TextureBudget = n
	}
}

func postValue024(name string) float32 {
	var value float32
	if hdrScene.program != 0 {
		if loc := postLocation024(name); loc >= 0 {
			glExt("glGetUniformfv", uintptr(hdrScene.program), uintptr(loc), uintptr(unsafe.Pointer(&value)))
		}
	}
	return value
}
func grassTextureWidth024() int {
	texture := pbrTextures["b024_sward_8k.jpg"]
	if texture == 0 {
		return 0
	}
	glExt("glActiveTexture", 0x84C0)
	var previous, width int32
	gl("glGetIntegerv", 0x8069, uintptr(unsafe.Pointer(&previous)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
	gl("glGetTexLevelParameteriv", GL_TEXTURE_2D, 0, 0x1000, uintptr(unsafe.Pointer(&width)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(previous))
	return int(width)
}
