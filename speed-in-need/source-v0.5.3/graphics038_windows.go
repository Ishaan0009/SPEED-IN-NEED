package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// Rendering/presentation owns no simulation ticks. Generated frames must never
// call World.Step. The OpenGL implementation remains the default safe path.
type graphicsBackend038 interface {
	Name() string
	BeginFrame() error
	Present() error
	Close()
}
type openGL038 struct{}

func (openGL038) Name() string      { return "OPENGL" }
func (openGL038) BeginFrame() error { return nil }
func (openGL038) Present() error {
	if win(gdi32, "SwapBuffers", app.dc) == 0 {
		return fmt.Errorf("OpenGL swap failed")
	}
	return nil
}
func (openGL038) Close() {}

type capability038 struct {
	ABI          int    `json:"abi"`
	Adapter      string `json:"adapter"`
	Vendor       uint32 `json:"vendor_id"`
	Dedicated    uint64 `json:"dedicated_vram_bytes"`
	Budget       uint64 `json:"local_budget_bytes"`
	Usage        uint64 `json:"local_usage_bytes"`
	DX12         bool   `json:"dx12_device"`
	RayTier      int    `json:"raytracing_tier"`
	SLInit       int    `json:"streamline_init"`
	DLSS         int    `json:"dlss_runtime_result"`
	NVIDIAFG     int    `json:"nvidia_fg_runtime_result"`
	FSRSupported bool   `json:"fsr_runtime_supported"`
	FSRProvider  string `json:"fsr_provider"`
	FSRError     string `json:"fsr_runtime_error"`
	Error        string `json:"error"`
}

var gpu038 capability038
var backend038 graphicsBackend038 = openGL038{}
var native038 *syscall.DLL
var nativeProcs038 = map[string]*syscall.Proc{}
var rendererError038 string
var outputFBO038 uint32
var rendererRequested038 int
var interopError038 syscall.Errno

// Native calls consume Go-backed buffers synchronously. Keep pointer arguments
// alive and off the movable stack through export lookup and the Windows call.
//
//go:uintptrescapes
func nativeCall038(name string, args ...uintptr) uintptr {
	if native038 == nil {
		return ^uintptr(0)
	}
	p := nativeProcs038[name]
	if p == nil {
		var err error
		p, err = native038.FindProc(name)
		if err != nil {
			rendererError038 = err.Error()
			return ^uintptr(0)
		}
		nativeProcs038[name] = p
	}
	r, _, _ := p.Call(args...)
	return r
}
func queryGPU038() {
	gpu038 = capability038{SLInit: -1, DLSS: -1, NVIDIAFG: -1}
	if nativeSDKDisabled053() {
		gpu038.Error = "Native SDK capability probe disabled by explicit OpenGL/offline mode"
		return
	}
	exe, _ := os.Executable()
	p := filepath.Join(filepath.Dir(exe), "sin_native052.dll")
	var err error
	native038, err = syscall.LoadDLL(p)
	if err != nil {
		gpu038.Error = "Native graphics bridge not installed; OpenGL fallback"
		return
	}
	buf := make([]byte, 16384)
	r := nativeCall038("SIN038_Probe", 38, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if err = json.Unmarshal([]byte(strings.TrimRight(string(buf), "\x00")), &gpu038); err != nil {
		gpu038.Error = err.Error()
	}
	if int32(r) != 0 && gpu038.Error == "" {
		gpu038.Error = "DirectX capability query failed"
	}
}
func bridgeError038() error {
	b := make([]byte, 4096)
	nativeCall038("SIN038_Error", uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	return fmt.Errorf("%s", strings.TrimRight(string(b), "\x00"))
}

type hybridDX12038 struct {
	texture, fbo, depth uint32
	device, object      uintptr
	width, height       int
	locked              bool
}

func (*hybridDX12038) Name() string { return "DX12 / OPENGL HYBRID" }
func (d *hybridDX12038) attach() error {
	if d.device == 0 {
		d.device = glExt("wglDXOpenDeviceNV", nativeCall038("SIN038_InteropDevice"))
		if d.device == 0 {
			return fmt.Errorf("WGL/DX interop device failed")
		}
	}
	glExt("glGenRenderbuffers", 1, uintptr(unsafe.Pointer(&d.texture)))
	d.object = glExt("wglDXRegisterObjectNV", d.device, nativeCall038("SIN038_InteropTexture"), uintptr(d.texture), 0x8D41, 0x0001)
	if d.object == 0 {
		return fmt.Errorf("WGL/DX shared surface registration failed: Win32 %d", interopError038)
	}
	if glExt("wglDXLockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&d.object))) == 0 {
		return fmt.Errorf("WGL/DX shared surface lock failed")
	}
	d.locked = true
	glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&d.fbo)))
	glExt("glBindFramebuffer", 0x8D40, uintptr(d.fbo))
	glExt("glFramebufferRenderbuffer", 0x8D40, 0x8CE0, 0x8D41, uintptr(d.texture))
	glExt("glGenRenderbuffers", 1, uintptr(unsafe.Pointer(&d.depth)))
	glExt("glBindRenderbuffer", 0x8D41, uintptr(d.depth))
	glExt("glRenderbufferStorage", 0x8D41, 0x81A6, uintptr(d.width), uintptr(d.height))
	glExt("glFramebufferRenderbuffer", 0x8D40, 0x8D00, 0x8D41, uintptr(d.depth))
	gl("glDrawBuffer", 0x8CE0)
	gl("glReadBuffer", 0x8CE0)
	if glExt("glCheckFramebufferStatus", 0x8D40) != 0x8CD5 {
		return fmt.Errorf("Shared output framebuffer incomplete")
	}
	outputFBO038 = d.fbo
	return nil
}
func (d *hybridDX12038) detach() {
	releaseHUDless038()
	outputFBO038 = 0
	glExt("glBindFramebuffer", 0x8D40, 0)
	if d.locked {
		glExt("wglDXUnlockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&d.object)))
		d.locked = false
	}
	if d.object != 0 {
		glExt("wglDXUnregisterObjectNV", d.device, d.object)
		d.object = 0
	}
	if d.fbo != 0 {
		glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&d.fbo)))
		d.fbo = 0
	}
	if d.depth != 0 {
		glExt("glDeleteRenderbuffers", 1, uintptr(unsafe.Pointer(&d.depth)))
		d.depth = 0
	}
	if d.texture != 0 {
		glExt("glDeleteRenderbuffers", 1, uintptr(unsafe.Pointer(&d.texture)))
		d.texture = 0
	}
}
func (d *hybridDX12038) BeginFrame() error {
	if app.width != d.width || app.height != d.height {
		d.detach()
		d.width, d.height = app.width, app.height
		if int32(nativeCall038("SIN038_Resize", uintptr(d.width), uintptr(d.height))) != 0 {
			return bridgeError038()
		}
		return d.attach()
	}
	glExt("glBindFramebuffer", 0x8D40, uintptr(d.fbo))
	return nil
}
func (d *hybridDX12038) Present() error {
	// Ownership transfers on the rendering thread. Both APIs use the same GPU
	// resource: no glReadPixels, staging upload or CPU framebuffer copy.
	if glExt("wglDXUnlockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&d.object))) == 0 {
		return fmt.Errorf("WGL/DX unlock failed")
	}
	d.locked = false
	if int32(nativeCall038("SIN038_Present", uintptr(boolInt(app.vsync)))) != 0 {
		return bridgeError038()
	}
	if glExt("wglDXLockObjectsNV", d.device, 1, uintptr(unsafe.Pointer(&d.object))) == 0 {
		return fmt.Errorf("WGL/DX relock failed")
	}
	d.locked = true
	return nil
}
func (d *hybridDX12038) Close() {
	releaseTemporalTransport038()
	d.detach()
	if d.device != 0 {
		glExt("wglDXCloseDeviceNV", d.device)
		d.device = 0
	}
	nativeCall038("SIN038_Destroy")
}
func initBackend038() {
	if nativeSDKDisabled053() {
		return
	}
	if rendererRequested038 == 0 {
		return
	}
	if !gpu038.DX12 || native038 == nil {
		rendererError038 = "DirectX 12 device unavailable"
		return
	}
	for _, name := range []string{"wglDXOpenDeviceNV", "wglDXRegisterObjectNV", "wglDXLockObjectsNV", "wglDXUnlockObjectsNV", "wglDXUnregisterObjectNV", "wglDXCloseDeviceNV"} {
		if !glExtAvailable(name) {
			rendererError038 = "Driver lacks WGL/DX interoperability"
			return
		}
	}
	features := uintptr(0)
	if gpu038.FSRSupported && (qa.enabled && os.Getenv("SPEED_QA_FSR_FG") == "1" || !qa.enabled && beta.cfg.FrameGeneration038 == 1) {
		features |= 1
	}
	if qa.enabled && os.Getenv("SPEED_QA_NVIDIA_FG") == "1" && gpu038.NVIDIAFG == 0 {
		features = 2
	}
	if int32(nativeCall038("SIN038_Create", app.hwnd, uintptr(app.width), uintptr(app.height), features)) != 0 {
		rendererError038 = bridgeError038().Error()
		return
	}
	d := &hybridDX12038{width: app.width, height: app.height}
	if err := d.attach(); err != nil {
		rendererError038 = err.Error()
		d.Close()
		return
	}
	backend038 = d
}
func beginFrame038() {
	if err := backend038.BeginFrame(); err != nil {
		fallback038(err)
	}
}
func fallback038(err error) {
	rendererError038 = err.Error()
	backend038.Close()
	backend038 = openGL038{}
	destroyHDR()
	gl("glDrawBuffer", 0x405)
	gl("glReadBuffer", 0x405)
}
func present038() {
	capture := prepareParity038()
	if err := backend038.Present(); err != nil {
		fallback038(err)
		return
	}
	if capture {
		finishParity038()
	}
}

func graphicsReport038() map[string]any {
	return map[string]any{"backend": backend038.Name(), "requested_backend": rendererRequested038, "capabilities": gpu038, "fallback_reason": rendererError038, "native_presented": nativePresented038(), "temporal_transport_error": temporalTransport038.Error, "temporal_transport_frames": temporalTransport038.Frames, "dlss_implemented": true, "dlss_fg_implemented": false, "fsr_fg_implemented": true, "fsr_runtime": fsrReport038(), "streamline_runtime": streamlineReport038(), "presentation_measurement": presentationBenchmarkReport038(), "nvidia_fg_compiled_path": true, "rt_shadows_compiled_path": true, "hardware_rt_implemented": true, "rt_scope": "Directional shadows with raster foliage; selective screen-visible clearcoat and wet-asphalt receivers; no ray-traced GI or local-light shadows", "desktop_validated_adapter": gpu038.Adapter, "offline_audit": offlineAudit053(), "native_sdk_disabled": nativeSDKDisabled053(), "rtx4050_hardware_tested": false}
}
func nativePresented038() uint64 {
	if outputFBO038 == 0 {
		return 0
	}
	return uint64(nativeCall038("SIN038_Presented"))
}
