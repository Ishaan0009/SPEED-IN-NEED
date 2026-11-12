package main

import (
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"time"
	"unsafe"
)

var hardware026 struct {
	TotalRAM, AvailableRAM uint64
	Budget                 memoryBudget026
	GoLimitMB              int64
}
var adaptive026 qualityGovernor026

// Only the explicit adaptive QA mode may alter quality during verification.
var qaAdaptive026 bool
var pacer026 framePacer026

func initMemoryBudget026() {
	var s struct {
		Length, Load                                                                     uint32
		TotalPhys, AvailPhys, TotalPage, AvailPage, TotalVirtual, AvailVirtual, Extended uint64
	}
	s.Length = uint32(unsafe.Sizeof(s))
	if win(kernel32, "GlobalMemoryStatusEx", uintptr(unsafe.Pointer(&s))) != 0 {
		hardware026.TotalRAM, hardware026.AvailableRAM = s.TotalPhys, s.AvailPhys
	}
	hardware026.Budget = memoryPolicy026(hardware026.TotalRAM, hardware026.AvailableRAM)
	if _, explicit := os.LookupEnv("GOMEMLIMIT"); !explicit {
		debug.SetMemoryLimit(int64(hardware026.Budget.HeapMB) << 20)
	}
	hardware026.GoLimitMB = debug.SetMemoryLimit(-1) >> 20
	adaptive026.Reset()
}

func adaptiveEnabled026() bool { return beta.cfg.AdaptiveQuality && (!qa.enabled || qaAdaptive026) }
func adaptiveFactors026() (float64, float64) {
	if !adaptiveEnabled026() || adaptive026.Scale == 0 {
		return 1, 1
	}
	return adaptive026.Scale, adaptive026.Foliage
}
func renderScale026() float64 {
	s, _ := adaptiveFactors026()
	if beta.cfg.Lightweight {
		return math.Max(.4, math.Min(.6, beta.cfg.RenderScale)*s)
	}
	return math.Max(math.Min(.5, beta.cfg.RenderScale), beta.cfg.RenderScale*s)
}
func foliageFactor026() float64 {
	_, f := adaptiveFactors026()
	if beta.cfg.Lightweight {
		f *= .6
	}
	return f
}
func foliageDistance026() float64 { return .75 + .25*foliageFactor026() }

func sceneFactor026() float64 {
	if adaptiveEnabled026() && adaptive026.Scene > 0 {
		return adaptive026.Scene
	}
	return 1
}
func viewDistance026() float64 {
	ceiling := beta.cfg.ViewDistance
	if beta.cfg.Lightweight {
		ceiling = math.Min(.5, ceiling)
	}
	return ceiling * sceneFactor026()
}
func reducedEffects026() bool {
	s, _ := adaptiveFactors026()
	return beta.cfg.Lightweight || sceneFactor026() < .85 || s <= .75
}
func shadowQuality026() int {
	if reducedEffects026() {
		return 0
	}
	return beta.cfg.Shadows
}
func lightingDetail026() int {
	if reducedEffects026() {
		return 0
	}
	return beta.cfg.LightingDetail
}

func updateAdaptive026(dt, workMS float64) {
	if !adaptiveEnabled026() {
		adaptive026.Reset()
		return
	}
	if beta.screen != screenDrive || remaster.FullMap || app.paused || remaster.Loading {
		adaptive026.warmup, adaptive026.window, adaptive026.gpu, adaptive026.cpu, adaptive026.recovery = 0, 0, 0, 0, 0
		adaptive026.frames, adaptive026.overloaded = 0, 0
		return
	}
	target := beta.cfg.FrameCap
	if target <= 0 {
		target = 60
	}
	if app.vsync {
		target = min(target, 60)
	}
	target = max(30, min(120, target))
	// GPU execution overlaps CPU submission. Summing them would double-count.
	gpu, cpu := qualityCosts026(app.gpuTiming, app.gpuMS, app.renderMS+app.simMS, workMS)
	adaptive026.Sample(dt, gpu, cpu, 1000/float64(target))
}

func paceFrame026(start time.Time) {
	fps := beta.cfg.FrameCap
	if beta.screen != screenDrive {
		fps = 60
	}
	deadline := pacer026.Next(start, fps)
	if remaining := time.Until(deadline); remaining > 0 {
		time.Sleep(remaining)
	}
}

func optimizationReport026() map[string]any {
	return map[string]any{
		"sky040":             sky040Report(),
		"native_gl_cache035": nativeGLStats035(),
		"graphics038":        graphicsReport038(),
		"logical_cpus":       runtime.NumCPU(), "total_ram_bytes": hardware026.TotalRAM,
		"available_ram_at_start_bytes": hardware026.AvailableRAM,
		"go_soft_memory_budget_mb":     hardware026.GoLimitMB,
		"adaptive_enabled":             adaptiveEnabled026(), "effective_render_scale": renderScale026(),
		"configured_frame_cap": beta.cfg.FrameCap,
		"lightweight_mode":     beta.cfg.Lightweight, "scene_factor": sceneFactor026(),
		"effective_view_distance": viewDistance026(), "effective_shadow_quality": shadowQuality026(),
		"foliage_factor": foliageFactor026(), "building_cache_hits": buildingStats026.Hits,
		"building_uploads": buildingStats026.Uploads, "building_upload_bytes": buildingStats026.Bytes,
		"internal_width": hdrScene.width, "internal_height": hdrScene.height,
	}
}
