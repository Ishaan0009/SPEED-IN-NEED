//go:build windows

package main

import (
	"fmt"
	"speedinneed/engine"
	"unsafe"
)

var postLocations024 = map[uint32]map[string]int32{}

func postLocation024(name string) int32 {
	p := hdrScene.program
	if postLocations024[p] == nil {
		postLocations024[p] = map[string]int32{}
	}
	if v, ok := postLocations024[p][name]; ok {
		return v
	}
	s := append([]byte(name), 0)
	v := int32(glExt("glGetUniformLocation", uintptr(p), uintptr(unsafe.Pointer(&s[0]))))
	postLocations024[p][name] = v
	return v
}

var maxTexture024 int32

func textureCap024() int {
	if maxTexture024 == 0 {
		gl("glGetIntegerv", 0x0D33, uintptr(unsafe.Pointer(&maxTexture024)))
		if maxTexture024 < 512 {
			maxTexture024 = 512
		}
	}
	cap := []int{512, 1024, 2048, 4096, 8192}[engine.ClampInt(beta.cfg.TextureBudget, 0, 4)]
	if beta.cfg.GPUTextureAuto {
		cap = min(cap, gpuTextureCap038())
	}
	if beta.cfg.Lightweight {
		cap = min(cap, 512)
	}
	if adaptiveEnabled026() && hardware026.Budget.TextureCap > 0 {
		cap = min(cap, hardware026.Budget.TextureCap)
	}
	for cap > int(maxTexture024) {
		cap /= 2
	}
	return cap
}

func textureSource024(name string) string {
	cap := textureCap024()
	if name == "b024_sward_8k.jpg" && cap < 8192 {
		return fmt.Sprintf("b024_sward_%d.jpg", cap)
	}
	if name == "b024_sward_normal.png" && cap < 4096 {
		return fmt.Sprintf("b024_sward_normal_%d.png", cap)
	}
	return name
}
