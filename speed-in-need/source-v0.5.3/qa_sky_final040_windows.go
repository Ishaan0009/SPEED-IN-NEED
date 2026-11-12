//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const qaSkyWaitLimit040 = 90 * time.Second

var qaSkyFinal040 struct {
	Raw       string
	Applied   bool
	Detail    int
	Problem   string
	WaitStart time.Time
	TimedOut  bool
}

type qaSkySourceProof040 struct {
	Label             string  `json:"label"`
	Weight            float64 `json:"blend_weight"`
	Required          bool    `json:"required_at_selected_detail"`
	TrueHDR           bool    `json:"genuine_hdr"`
	StateSize         [2]int  `json:"completed_state_size"`
	TexturePresent    bool    `json:"texture_present"`
	AtTarget          bool    `json:"completed_at_effective_target"`
	True16K           bool    `json:"completed_true_16k"`
	GPUSize           [2]int  `json:"queried_gpu_size"`
	GPUInternalFormat uint32  `json:"queried_gpu_internal_format"`
	GPUChecked        bool    `json:"gpu_checked"`
	GPUStateMatches   bool    `json:"gpu_matches_completed_state_rgb16f"`
}

func parseQASkyDetail040(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	detail, err := strconv.Atoi(raw)
	return detail, err == nil && detail >= 0 && detail <= 4
}

// Call last, after QA presets, RTX profiles and scenario settings. This never
// changes normal gameplay settings or stores an override in a driver profile.
func applyQASkyDetail040() {
	if !qa.enabled {
		return
	}
	qaSkyFinal040.Raw = os.Getenv("SPEED_QA_SKY_DETAIL040")
	qaSkyFinal040.Applied = false
	qaSkyFinal040.Problem = ""
	qaSkyFinal040.WaitStart = time.Time{}
	qaSkyFinal040.TimedOut = false
	detail, valid := parseQASkyDetail040(qaSkyFinal040.Raw)
	if !valid {
		if strings.TrimSpace(qaSkyFinal040.Raw) != "" {
			qaSkyFinal040.Problem = "SPEED_QA_SKY_DETAIL040 must be an integer from 0 to 4"
		}
		return
	}
	qaSkyFinal040.Applied = true
	qaSkyFinal040.Detail = detail
	beta.cfg.SkyDetail040 = detail
}

func requestedSkyWidthQA040(quality, detail int) int {
	if detail >= 1 && detail <= 4 {
		return []int{2048, 4096, 8192, 16384}[detail-1]
	}
	if quality >= 3 {
		return 16384
	}
	if quality == 2 {
		return 8192
	}
	return 4096
}

func requiredSkySourcesQA040(current, next int, blend float64, target int, previous [3]bool, sizes [3][2]int, trueHDR [3]bool, textures [3]uint32) ([3]qaSkySourceProof040, bool) {
	wanted, weights := skyResidencyWants040(current, next, blend, previous)
	sources := [3]qaSkySourceProof040{}
	complete := true
	for label, name := range []string{"day", "afternoon", "night"} {
		s := qaSkySourceProof040{Label: name, Weight: weights[label], Required: wanted[label], TrueHDR: trueHDR[label], StateSize: sizes[label], TexturePresent: textures[label] != 0}
		s.AtTarget = s.TrueHDR && s.TexturePresent && sizes[label] == [2]int{target, target / 2}
		s.True16K = s.TrueHDR && s.TexturePresent && sizes[label] == [2]int{16384, 8192}
		if s.Required && !s.AtTarget {
			complete = false
		}
		sources[label] = s
	}
	return sources, complete
}

func currentQASkySources040() ([3]qaSkySourceProof040, int, bool) {
	current, next, blend := skyBlend034()
	maximum := sky040.MaxTexture
	if maximum == 0 {
		maximum = 2048
	}
	target := skyViewportTarget050(app.quality, beta.cfg.SkyDetail040, gpu038.Dedicated, maximum, app.width)
	sources, complete := requiredSkySourcesQA040(current, next, blend, target, sky040.Wanted, sky020.Sizes, sky040.TrueHDR, sky020.Textures)
	return sources, target, complete
}

// Place after the ordinary QA frame/time limit check, before final capture.
// Returning true advances ordinary render frames, rather than blocking/sleeping.
// Missing assets, failed upgrades and timeout produce explicit failed proof.
func waitQASkyDetail040() bool {
	if !qa.enabled || !qaSkyFinal040.Applied {
		return false
	}
	sources, target, complete := currentQASkySources040()
	if complete {
		return false
	}
	for label, source := range sources {
		if !source.Required {
			continue
		}
		if !source.TrueHDR || !source.TexturePresent {
			qaSkyFinal040.Problem = fmt.Sprintf("required %s HDR bootstrap is unavailable", source.Label)
			return false
		}
		if sky040.Failed[[2]int{label, target}] {
			qaSkyFinal040.Problem = fmt.Sprintf("required %s %d-wide upload failed: %s", source.Label, target, sky040.LastError)
			return false
		}
	}
	now := time.Now()
	if qaSkyFinal040.WaitStart.IsZero() {
		qaSkyFinal040.WaitStart = now
	}
	if now.Sub(qaSkyFinal040.WaitStart) >= qaSkyWaitLimit040 {
		qaSkyFinal040.TimedOut = true
		qaSkyFinal040.Problem = "HDR sky did not reach the effective requested detail within 90 seconds of capture wait"
		return false
	}
	return true
}

// Called on the existing render thread after glFinish, beside the screenshot.
// Query real allocated dimensions/format in addition to the streaming state.
func queryQASkyTextures040(sources *[3]qaSkySourceProof040) {
	var active, previous int32
	gl("glGetIntegerv", 0x84e0, uintptr(unsafe.Pointer(&active)))
	glExt("glActiveTexture", 0x84c0)
	gl("glGetIntegerv", 0x8069, uintptr(unsafe.Pointer(&previous)))
	defer func() {
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(previous))
		glExt("glActiveTexture", uintptr(active))
	}()
	for label, texture := range sky020.Textures {
		s := &sources[label]
		s.GPUChecked = true
		if texture == 0 || gl("glIsTexture", uintptr(texture)) == 0 {
			continue
		}
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
		var width, height, format int32
		gl("glGetTexLevelParameteriv", GL_TEXTURE_2D, 0, 0x1000, uintptr(unsafe.Pointer(&width)))
		gl("glGetTexLevelParameteriv", GL_TEXTURE_2D, 0, 0x1001, uintptr(unsafe.Pointer(&height)))
		gl("glGetTexLevelParameteriv", GL_TEXTURE_2D, 0, 0x1003, uintptr(unsafe.Pointer(&format)))
		s.GPUSize = [2]int{int(width), int(height)}
		s.GPUInternalFormat = uint32(format)
		s.GPUStateMatches = s.GPUSize == s.StateSize && format == 0x881b
	}
}

func captureQASkyProof040(name string) {
	if !qa.enabled || qa.dir == "" {
		return
	}
	sources, target, stateComplete := currentQASkySources040()
	queryQASkyTextures040(&sources)
	gpuComplete, actual16K := stateComplete, true
	for _, source := range sources {
		if !source.Required {
			continue
		}
		gpuComplete = gpuComplete && source.GPUStateMatches
		actual16K = actual16K && source.True16K && source.GPUStateMatches
	}
	status := "complete"
	if !gpuComplete {
		status = "incomplete"
	}
	if qaSkyFinal040.Problem != "" {
		status = "failed"
	}
	if qaSkyFinal040.TimedOut {
		status = "timeout"
	}
	uncapped := requestedSkyWidthQA040(app.quality, beta.cfg.SkyDetail040)
	proof := map[string]any{
		"stage": qa.stage, "name": name, "utc": time.Now().UTC().Format(time.RFC3339),
		"qa_environment_override": qaSkyFinal040.Applied, "environment_value": qaSkyFinal040.Raw, "selected_detail": beta.cfg.SkyDetail040,
		"requested_width_before_hardware_cap": uncapped, "effective_target_width": target, "hardware_capped": target < uncapped,
		"status": status, "streaming_state_complete": stateComplete, "queried_gpu_target_complete": gpuComplete, "actual_required_sources_16k": actual16K,
		"capture_wait_limit_seconds": qaSkyWaitLimit040.Seconds(), "capture_timed_out": qaSkyFinal040.TimedOut, "problem": qaSkyFinal040.Problem,
		"scene_hour": sceneHour(), "sources": sources, "sky_runtime": sky040Report(),
		"validation_boundary": "Actual OpenGL allocations and completed uploads at this captured stage; not an image-quality or FPS acceptance claim.",
	}
	if !qaSkyFinal040.WaitStart.IsZero() {
		proof["capture_wait_seconds"] = time.Since(qaSkyFinal040.WaitStart).Seconds()
	}
	if qaSkyFinal040.Applied && qa.checks != nil {
		qa.checks["040-sky-"+name+"-effective-target-complete"] = gpuComplete && qaSkyFinal040.Problem == ""
		if uncapped == 16384 && target == 16384 {
			qa.checks["040-sky-"+name+"-actual-16k"] = actual16K && qaSkyFinal040.Problem == ""
		}
	}
	filename := "SKY_PROOF040_" + strings.NewReplacer("/", "-", "\\", "-", ":", "-").Replace(name) + ".json"
	if raw, err := json.MarshalIndent(proof, "", "  "); err == nil {
		if err = os.WriteFile(filepath.Join(qa.dir, filename), raw, 0644); err != nil && qa.checks != nil {
			qa.checks["040-sky-proof-written"] = false
		}
	} else if qa.checks != nil {
		qa.checks["040-sky-proof-written"] = false
	}
}
