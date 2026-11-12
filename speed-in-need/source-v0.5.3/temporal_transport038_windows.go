package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"unsafe"
)

type interopTarget038 struct {
	Handle      uintptr
	Buffer, FBO uint32
}

var temporalTransport038 struct {
	Surfaces                  [5]interopTarget038
	Width, Height             int
	OutputWidth, OutputHeight int
	Device                    uintptr
	Program, SunProgram       uint32
	Frames                    uint64
	LastReset                 uint64
	Error                     string
	Locked                    bool
}
var reconstructed038 cinematicTarget034
var beforeParticles038 cinematicTarget034
var beforeParticlesReady038 bool
var reconstructedReady038 bool
var optimalSize038 struct{ Mode, Width, Height, InputWidth, InputHeight int }

func temporalMode038() int {
	if beta.screen != screenDrive || temporalTransport038.Error != "" || outputFBO038 == 0 || gpu038.DLSS != 0 || gpu038.SLInit != 0 {
		return 0
	}
	switch os.Getenv("SPEED_QA_DLSS_MODE") {
	case "quality":
		return 3
	case "balanced":
		return 2
	case "performance":
		return 1
	case "ultraperformance":
		if app.width >= 3840 {
			return 4
		}
	case "dlaa":
		return 6
	}
	if os.Getenv("SPEED_QA_DLAA") == "1" {
		return 6
	}
	if !qa.enabled {
		mode := [6]int{0, 3, 2, 1, 6, 4}[beta.cfg.Upscaling038]
		if mode == 4 && app.width < 3840 {
			return 3
		}
		return mode
	}
	return 0
}
func postSceneWidth038() int {
	if reconstructedReady038 {
		return reconstructed038.Width
	}
	return hdrScene.width
}
func postSceneHeight038() int {
	if reconstructedReady038 {
		return reconstructed038.Height
	}
	return hdrScene.height
}
func experimentDLAA038() bool { return temporalMode038() != 0 }
func temporalInputSize038(w, h int) (int, int) {
	mode := temporalMode038()
	if mode == 0 || outputFBO038 == 0 {
		return w, h
	}
	if optimalSize038.Mode == mode && optimalSize038.Width == app.width && optimalSize038.Height == app.height {
		return optimalSize038.InputWidth, optimalSize038.InputHeight
	}
	size := [2]uint32{}
	if int32(nativeCall038("SIN038_Optimal", uintptr(mode), uintptr(app.width), uintptr(app.height), uintptr(unsafe.Pointer(&size[0])))) != 0 || size[0] < 64 || size[1] < 64 {
		temporalTransport038.Error = "Cannot query supported DLSS input size: " + bridgeError038().Error()
		return w, h
	}
	optimalSize038.Mode, optimalSize038.Width, optimalSize038.Height = mode, app.width, app.height
	optimalSize038.InputWidth, optimalSize038.InputHeight = int(size[0]), int(size[1])
	return int(size[0]), int(size[1])
}
func reconstructForPost038() func() {
	if !reconstructedReady038 {
		return func() {}
	}
	texture, w, h := hdrScene.texture, hdrScene.width, hdrScene.height
	hdrScene.texture, hdrScene.width, hdrScene.height = reconstructed038.Texture, reconstructed038.Width, reconstructed038.Height
	gl("glMatrixMode", 0x1701)
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&temporal038.Projection[0])))
	gl("glMatrixMode", 0x1700)
	return func() { hdrScene.texture, hdrScene.width, hdrScene.height = texture, w, h }
}

type temporalNativeFrame038 struct {
	ABI, Bytes, Index, Reset     uint32
	Projection, View, PreviousVP [16]float32
	JitterX, JitterY, DeltaMS    float32
	Mode                         uint32
}

func releaseTemporalTransport038() {
	beforeParticles038.release()
	beforeParticlesReady038 = false
	reconstructedReady038 = false
	reconstructed038.release()
	t := &temporalTransport038
	for i := range t.Surfaces {
		s := &t.Surfaces[i]
		if s.Handle != 0 {
			if t.Locked {
				glExt("wglDXUnlockObjectsNV", t.Device, 1, uintptr(unsafe.Pointer(&s.Handle)))
			}
			glExt("wglDXUnregisterObjectNV", t.Device, s.Handle)
		}
		if s.FBO != 0 {
			glExt("glDeleteFramebuffers", 1, uintptr(unsafe.Pointer(&s.FBO)))
		}
		if s.Buffer != 0 {
			glExt("glDeleteRenderbuffers", 1, uintptr(unsafe.Pointer(&s.Buffer)))
		}
		*s = interopTarget038{}
	}
	t.Locked = false
	t.Width, t.Height = 0, 0
	if native038 != nil {
		nativeCall038("SIN038_TemporalDestroy")
	}
}
func snapshotBeforeParticles038() {
	beforeParticlesReady038 = false
	if (!experimentDLAA038() && !experimentRayShadows038()) || !hdrScene.active || outputFBO038 == 0 {
		return
	}
	if !beforeParticles038.ensure(hdrScene.width, hdrScene.height, true) {
		return
	}
	glExt("glBindFramebuffer", 0x8CA8, uintptr(hdrScene.fbo))
	gl("glReadBuffer", 0x8CE0)
	glExt("glBindFramebuffer", 0x8CA9, uintptr(beforeParticles038.FBO))
	glExt("glBlitFramebuffer", 0, 0, uintptr(hdrScene.width), uintptr(hdrScene.height), 0, 0, uintptr(hdrScene.width), uintptr(hdrScene.height), GL_COLOR_BUFFER_BIT, 0x2600)
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	temporalOutputs038(true)
	beforeParticlesReady038 = true
}
func ensureTemporalTransport038() error {
	t := &temporalTransport038
	ow, oh := hdrScene.width, hdrScene.height
	if experimentDLAA038() {
		ow, oh = app.width, app.height
	}
	if t.Width == hdrScene.width && t.Height == hdrScene.height && t.OutputWidth == ow && t.OutputHeight == oh {
		return nil
	}
	releaseTemporalTransport038()
	d, ok := backend038.(*hybridDX12038)
	if !ok {
		return fmt.Errorf("temporal transport requires the DirectX backend")
	}
	t.Device = d.device
	if int32(nativeCall038("SIN038_TemporalCreate", uintptr(hdrScene.width), uintptr(hdrScene.height), uintptr(ow), uintptr(oh))) != 0 {
		return bridgeError038()
	}
	for i := range t.Surfaces {
		s := &t.Surfaces[i]
		glExt("glGenRenderbuffers", 1, uintptr(unsafe.Pointer(&s.Buffer)))
		s.Handle = glExt("wglDXRegisterObjectNV", t.Device, nativeCall038("SIN038_TemporalSurface", uintptr(i)), uintptr(s.Buffer), 0x8D41, 0x0001)
		if s.Handle == 0 {
			return fmt.Errorf("temporal surface %d registration failed: %d", i, interopError038)
		}
		if glExt("wglDXLockObjectsNV", t.Device, 1, uintptr(unsafe.Pointer(&s.Handle))) == 0 {
			return fmt.Errorf("temporal surface %d lock failed", i)
		}
		t.Locked = true
		glExt("glGenFramebuffers", 1, uintptr(unsafe.Pointer(&s.FBO)))
		glExt("glBindFramebuffer", 0x8D40, uintptr(s.FBO))
		glExt("glFramebufferRenderbuffer", 0x8D40, 0x8CE0, 0x8D41, uintptr(s.Buffer))
		gl("glDrawBuffer", 0x8CE0)
		gl("glReadBuffer", 0x8CE0)
		if glExt("glCheckFramebufferStatus", 0x8D40) != 0x8CD5 {
			return fmt.Errorf("temporal surface %d incomplete", i)
		}
	}
	if t.Program == 0 {
		t.Program = cinematicProgram034("temporal_pack038.frag")
		if t.Program == 0 {
			return fmt.Errorf("temporal packing shader failed")
		}
	}
	if t.SunProgram == 0 {
		t.SunProgram = cinematicProgram034("sun_pack038.frag")
		if t.SunProgram == 0 {
			return fmt.Errorf("sun packing shader failed")
		}
	}
	t.Width, t.Height = hdrScene.width, hdrScene.height
	t.OutputWidth, t.OutputHeight = ow, oh
	if experimentDLAA038() && !reconstructed038.ensure(ow, oh, false) {
		return fmt.Errorf("reconstructed scene framebuffer unavailable")
	}
	return nil
}
func readHDR038() []float32 {
	data := make([]float32, hdrScene.width*hdrScene.height*4)
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	gl("glReadBuffer", 0x8CE0)
	gl("glReadPixels", 0, 0, uintptr(hdrScene.width), uintptr(hdrScene.height), GL_RGBA, glFloat, uintptr(unsafe.Pointer(&data[0])))
	return data
}
func runTemporalTransport038() {
	reconstructedReady038 = false
	if outputFBO038 == 0 || hdrScene.motion == 0 || !hdrScene.active || (os.Getenv("SPEED_QA_TEMPORAL_TRANSPORT") != "1" && !experimentDLAA038() && !experimentRayMask038() && !experimentFSR038() && !experimentNVIDIAFG038()) {
		return
	}
	t := &temporalTransport038
	if t.Error != "" {
		return
	}
	if e := ensureTemporalTransport038(); e != nil {
		t.Error = e.Error()
		releaseTemporalTransport038()
		glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
		return
	}
	capture := qa.frame+1 == qa.limit
	var before []float32
	if capture {
		before = readHDR038()
	}
	glExt("glBindFramebuffer", 0x8CA8, uintptr(hdrScene.fbo))
	gl("glReadBuffer", 0x8CE0)
	glExt("glBindFramebuffer", 0x8CA9, uintptr(t.Surfaces[0].FBO))
	glExt("glBlitFramebuffer", 0, 0, uintptr(t.Width), uintptr(t.Height), 0, 0, uintptr(t.Width), uintptr(t.Height), GL_COLOR_BUFFER_BIT, 0x2600)
	glExt("glBindFramebuffer", 0x8D40, uintptr(t.Surfaces[1].FBO))
	fullFrame037(t.Width, t.Height)
	dis(GL_BLEND)
	dis(GL_DEPTH_TEST)
	gl("glDepthMask", 0)
	glExt("glUseProgram", uintptr(t.Program))
	glExt("glActiveTexture", 0x84C0)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.motion))
	glExt("glActiveTexture", 0x84C1)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.depth))
	glExt("glActiveTexture", 0x84C2)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.texture))
	glExt("glActiveTexture", 0x84C3)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(beforeParticles038.Texture))
	cinematicInt034(t.Program, "uFullScene038", 2)
	cinematicInt034(t.Program, "uBeforeParticles038", 3)
	cinematicInt034(t.Program, "uMaskReady038", boolInt(beforeParticlesReady038))
	cinematicInt034(t.Program, "uMotion038", 0)
	cinematicInt034(t.Program, "uDepth038", 1)
	glExt("glActiveTexture", 0x84C0)
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
	if hdrScene.sun != 0 {
		glExt("glBindFramebuffer", 0x8D40, uintptr(t.Surfaces[3].FBO))
		glExt("glUseProgram", uintptr(t.SunProgram))
		glExt("glActiveTexture", 0x84C0)
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.sun))
		glExt("glActiveTexture", 0x84C1)
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.motion))
		cinematicInt034(t.SunProgram, "uSun038", 0)
		cinematicInt034(t.SunProgram, "uMotion038", 1)
		glExt("glActiveTexture", 0x84C0)
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
	}
	if experimentRayReflections038() {
		glExt("glBindFramebuffer", 0x8CA8, uintptr(hdrScene.fbo))
		gl("glReadBuffer", 0x8CE1)
		glExt("glBindFramebuffer", 0x8CA9, uintptr(t.Surfaces[4].FBO))
		gl("glDrawBuffer", 0x8CE0)
		glExt("glBlitFramebuffer", 0, 0, uintptr(t.Width), uintptr(t.Height), 0, 0, uintptr(t.Width), uintptr(t.Height), GL_COLOR_BUFFER_BIT, 0x2600)
	}
	handles := [5]uintptr{t.Surfaces[0].Handle, t.Surfaces[1].Handle, t.Surfaces[2].Handle, t.Surfaces[3].Handle, t.Surfaces[4].Handle}
	if glExt("wglDXUnlockObjectsNV", t.Device, 5, uintptr(unsafe.Pointer(&handles[0]))) == 0 {
		t.Error = "temporal surface unlock failed"
		return
	}
	t.Locked = false
	result := int32(0)
	if experimentDLAA038() || experimentRayMask038() || experimentFSR038() || experimentNVIDIAFG038() {
		frame := temporalNativeFrame038{ABI: 38, Index: applicationFrame038, Reset: uint32(boolInt(t.LastReset != temporal038.Resets)), Projection: temporal038.Projection, View: temporal038.View, PreviousVP: temporal038.PreviousVP, JitterX: temporal038.JitterX, JitterY: temporal038.JitterY, DeltaMS: float32(app.frameMS), Mode: uint32(temporalMode038())}
		frame.Bytes = uint32(unsafe.Sizeof(frame))
		result = int32(nativeCall038("SIN038_TemporalEvaluate", uintptr(unsafe.Pointer(&frame))))
		if result == 0 {
			t.LastReset = temporal038.Resets
		}
	} else {
		result = int32(nativeCall038("SIN038_TemporalDispatch"))
	}
	if glExt("wglDXLockObjectsNV", t.Device, 5, uintptr(unsafe.Pointer(&handles[0]))) == 0 {
		t.Error = "temporal surface relock failed"
		return
	}
	t.Locked = true
	if result != 0 {
		t.Error = bridgeError038().Error()
	} else {
		glExt("glBindFramebuffer", 0x8CA8, uintptr(t.Surfaces[2].FBO))
		gl("glReadBuffer", 0x8CE0)
		target := hdrScene.fbo
		if experimentDLAA038() {
			target = reconstructed038.FBO
		}
		glExt("glBindFramebuffer", 0x8CA9, uintptr(target))
		gl("glDrawBuffer", 0x8CE0)
		glExt("glBlitFramebuffer", 0, 0, uintptr(t.OutputWidth), uintptr(t.OutputHeight), 0, 0, uintptr(t.OutputWidth), uintptr(t.OutputHeight), GL_COLOR_BUFFER_BIT, 0x2600)
		t.Frames++
		reconstructedReady038 = experimentDLAA038()
	}
	glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	temporalOutputs038(true)
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
	en(GL_BLEND)
	if capture && before != nil {
		captureTemporalProof038(remasterQANames[qa.stage], before)
	}
}

func captureTemporalProof038(name string, before []float32) {
	t := &temporalTransport038
	var oldFBO, oldRead int32
	gl("glGetIntegerv", 0x8CA6, uintptr(unsafe.Pointer(&oldFBO)))
	gl("glGetIntegerv", 0x0C02, uintptr(unsafe.Pointer(&oldRead)))
	defer func() { glExt("glBindFramebuffer", 0x8D40, uintptr(oldFBO)); gl("glReadBuffer", uintptr(oldRead)) }()
	if before == nil {
		before = readHDR038()
	}

	after := readHDR038()
	if reconstructedReady038 {
		after = make([]float32, t.OutputWidth*t.OutputHeight*4)
		glExt("glBindFramebuffer", 0x8D40, uintptr(reconstructed038.FBO))
		gl("glReadBuffer", 0x8CE0)
		gl("glReadPixels", 0, 0, uintptr(t.OutputWidth), uintptr(t.OutputHeight), GL_RGBA, glFloat, uintptr(unsafe.Pointer(&after[0])))
		glExt("glBindFramebuffer", 0x8D40, uintptr(hdrScene.fbo))
	}
	maximum, total, invalid := 0., 0., 0
	sourceInvalid, outputInvalid := 0, 0
	locations := []string{}
	for _, value := range before {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			sourceInvalid++
		}
	}
	for i, value := range after {
		reference := i
		if reconstructedReady038 {
			pixel := i / 4
			x, y := pixel%t.OutputWidth, pixel/t.OutputWidth
			reference = ((y*t.Height/t.OutputHeight)*t.Width+x*t.Width/t.OutputWidth)*4 + i%4
		}
		d := math.Abs(float64(value - before[reference]))
		if math.IsNaN(d) || math.IsInf(d, 0) {
			invalid++
			if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
				outputInvalid++
			}
			if len(locations) < 8 {
				locations = append(locations, fmt.Sprintf("output channel %d value %v / source %v", i, value, before[reference]))
			}
			continue
		}
		maximum = math.Max(maximum, d)
		total += d
	}
	status := "PASS"
	sourceEnergy, outputEnergy := 0., 0.
	for i, v := range before {
		if i%4 != 3 && !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) {
			sourceEnergy += float64(v)
		}
	}
	for i, v := range after {
		if i%4 != 3 && !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) {
			outputEnergy += float64(v)
		}
	}
	sourceEnergy /= float64(len(before) * 3 / 4)
	outputEnergy /= float64(len(after) * 3 / 4)
	if experimentRayReflections038() && sourceEnergy > .0001 && outputEnergy < sourceEnergy*.25 {
		status = "FAIL"
	}
	if (!experimentDLAA038() && !experimentRayShadows038() && maximum > .001) || invalid != 0 || t.Error != "" {
		status = "FAIL"
	}
	report := map[string]any{"status": status, "scene": name, "gpu_round_trips": t.Frames, "maximum_hdr_channel_error": maximum, "mean_hdr_channel_error": total / float64(len(after)), "invalid_values": invalid, "error": t.Error, "pipeline": "GL RGBA16F and packed depth/motion -> D3D11 shared GPU resources -> D3D12 compute -> D3D11 -> GL RGBA16F; QA-only CPU comparison"}
	report["dlss_evaluations"] = uint64(nativeCall038("SIN038_TemporalEvaluations"))
	report["dlaa_experiment"] = experimentDLAA038()
	report["sdk_mode"] = temporalMode038()
	report["source_average_hdr_rgb"] = sourceEnergy
	report["output_average_hdr_rgb"] = outputEnergy
	report["ray_shadow_composition"] = experimentRayShadows038()
	report["input_size"] = []int{t.Width, t.Height}
	report["output_size"] = []int{t.OutputWidth, t.OutputHeight}
	report["feature_quality_validation_complete"] = false
	report["source_nonfinite_channels"] = sourceInvalid
	report["output_nonfinite_channels"] = outputInvalid
	report["nonfinite_locations"] = locations
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(filepath.Join(qa.dir, name+"_transport038.json"), raw, 0644)
}
