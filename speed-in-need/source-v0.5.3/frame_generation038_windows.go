package main

import (
	"encoding/json"
	"os"
	"strings"
	"unsafe"
)

var hudless038 interopTarget038
var hudlessDevice038 uintptr

func experimentFSR038() bool {
	return (qa.enabled && os.Getenv("SPEED_QA_FSR_FG") == "1" || !qa.enabled && beta.cfg.FrameGeneration038 == 1) && outputFBO038 != 0 && beta.screen == screenDrive
}
func experimentNVIDIAFG038() bool {
	return qa.enabled && os.Getenv("SPEED_QA_NVIDIA_FG") == "1" && gpu038.NVIDIAFG == 0 && outputFBO038 != 0 && beta.screen == screenDrive
}
func releaseHUDless038() {
	if hudless038.Handle != 0 {
		glExt("wglDXUnregisterObjectNV", hudlessDevice038, hudless038.Handle)
	}
	if hudless038.Buffer != 0 {
		glExt("glDeleteRenderbuffers", 1, uintptr(unsafe.Pointer(&hudless038.Buffer)))
	}
	if hudless038.FBO != 0 {
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&hudless038.FBO)))
	}
	hudless038 = interopTarget038{}
}
func snapshotHUDless038() {
	if !experimentFSR038() && !experimentNVIDIAFG038() {
		return
	}
	target := nativeCall038("SIN038_HUDSurface")
	if target == 0 {
		return
	}
	d, ok := backend038.(*hybridDX12038)
	if !ok {
		return
	}
	if hudless038.Handle == 0 {
		hudlessDevice038 = d.device
		glExt("glGenRenderbuffers", 1, uintptr(unsafe.Pointer(&hudless038.Buffer)))
		hudless038.Handle = glExt("wglDXRegisterObjectNV", d.device, target, uintptr(hudless038.Buffer), 0x8D41, 0x0001)
		if hudless038.Handle == 0 {
			temporalTransport038.Error = "HUD-less interop registration failed"
			return
		}
		glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&hudless038.FBO)))
	}
	if glExt("wglDXLockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&hudless038.Handle))) == 0 {
		temporalTransport038.Error = "HUD-less interop lock failed"
		return
	}
	glExt("glBindFramebuffer", 0x8CA9, uintptr(hudless038.FBO))
	glExt("glFramebufferRenderbuffer", 0x8CA9, 0x8CE0, 0x8D41, uintptr(hudless038.Buffer))
	gl("glDrawBuffer", 0x8CE0)
	glExt("glBindFramebuffer", 0x8CA8, uintptr(outputFBO038))
	gl("glReadBuffer", 0x8CE0)
	if glExt("glCheckFramebufferStatus", 0x8CA9) != 0x8CD5 {
		temporalTransport038.Error = "HUD-less framebuffer incomplete"
	} else {
		glExt("glBlitFramebuffer", 0, 0, uintptr(app.width), uintptr(app.height), 0, 0, uintptr(app.width), uintptr(app.height), GL_COLOR_BUFFER_BIT, 0x2600)
	}
	if glExt("wglDXUnlockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&hudless038.Handle))) == 0 {
		temporalTransport038.Error = "HUD-less interop unlock failed"
		return
	}
	nativeCall038("SIN038_HUDReady")
	glExt("glBindFramebuffer", 0x8D40, uintptr(outputFBO038))
	gl("glReadBuffer", 0x8CE0)
	gl("glDrawBuffer", 0x8CE0)
}
func fsrReport038() map[string]any {
	result := map[string]any{}
	if outputFBO038 == 0 {
		return result
	}
	data := make([]byte, 8192)
	nativeCall038("SIN038_FSRStats", uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	json.Unmarshal([]byte(strings.TrimRight(string(data), "\x00")), &result)
	return result
}
func streamlineReport038() map[string]any {
	result := map[string]any{}
	if outputFBO038 == 0 {
		return result
	}
	data := make([]byte, 8192)
	nativeCall038("SIN038_StreamlineStats", uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	json.Unmarshal([]byte(strings.TrimRight(string(data), "\x00")), &result)
	return result
}
