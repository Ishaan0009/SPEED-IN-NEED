package main

import "unsafe"

func reflectionMaskPixels035() int {
	if hdrScene.normals == 0 || hdrScene.width < 1 || hdrScene.height < 1 {
		return 0
	}
	count := hdrScene.width * hdrScene.height
	pixels := make([]byte, count*4)
	depth := make([]float32, count)
	var previous int32
	gl("glGetIntegerv", 0x8CA6, uintptr(unsafe.Pointer(&previous)))
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	gl("glReadBuffer", 0x8CE1)
	gl("glReadPixels", 0, 0, uintptr(hdrScene.width), uintptr(hdrScene.height), GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&pixels[0])))
	gl("glReadPixels", 0, 0, uintptr(hdrScene.width), uintptr(hdrScene.height), 0x1902, glFloat, uintptr(unsafe.Pointer(&depth[0])))
	gl("glReadBuffer", 0x8CE0)
	glExt("glBindFramebuffer", 0x8D40, uintptr(previous))
	n := 0
	for i, d := range depth {
		b := pixels[i*4+2]
		if d < .99998 && b >= 31 && b <= 140 {
			n++
		}
	}
	return n
}
func checks035() {
	checks036()
	checks037()
	checksProfiles040()
	checksWorldFinal040()
    checksQA041()
	qa.checks["035-scalable-font-ready"] = font035.Program != 0 && app.tex["b037_ui_font_sdf.png"] != 0
	if beta.screen == screenDrive && app.quality > 0 && !reducedEffects026() {
		if experimentRayReflections038() {
			qa.checks["038-ray-reflection-target-ready"] = hdrScene.normals != 0 && rayNormals038 && postValue024("uSurfaceReflections") == 0
			if qa.stage == 5 {
				r := rayReport038()
				native, ok := r["native"].(map[string]any)
				qa.checks["038-reflected-hardware-hits-present"] = ok && number038(native, "reflection_pixels") > 100
			}
			return
		}
		qa.checks["035-local-reflection-buffer-ready"] = hdrScene.normals != 0 && postValue024("uSurfaceReflections") == 1
		if qa.stage == 5 {
			qa.checks["035-reflective-surfaces-present"] = reflectionMaskPixels035() > 100
		}
	}
}
