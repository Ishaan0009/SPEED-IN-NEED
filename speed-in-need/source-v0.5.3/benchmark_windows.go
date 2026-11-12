//go:build windows

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var profiler struct {
	ring         [1800]float64
	count, index int
	last         time.Time
	summary      frameSummary
	active       bool
	samples      []float64
	gpuSamples   []float64
	seconds      float64
	message      string
}

func toggleBenchmark() {
	if profiler.active {
		finishBenchmark()
		return
	}
	profiler.active = true
	profiler.samples = make([]float64, 0, 7200)
	profiler.gpuSamples = make([]float64, 0, 7200)
	profiler.seconds = 0
	profiler.message = "BENCHMARK RECORDING - F9 TO FINISH"
}

func recordFrame(ms float64) {
	if beta.screen != screenDrive || remaster.FullMap {
		return
	}
	profiler.ring[profiler.index] = ms
	profiler.index = (profiler.index + 1) % len(profiler.ring)
	if profiler.count < len(profiler.ring) {
		profiler.count++
	}
	if time.Since(profiler.last) > time.Second && (app.showPerf || profiler.active) {
		profiler.summary = summarizeFrames(profiler.ring[:profiler.count])
		profiler.last = time.Now()
	}
	if profiler.active {
		profiler.samples = append(profiler.samples, ms)
		if app.gpuTiming && app.gpuMS > 0 {
			profiler.gpuSamples = append(profiler.gpuSamples, app.gpuMS)
		}
		profiler.seconds += ms / 1000
		if profiler.seconds >= 60 {
			finishBenchmark()
		}
	}
}

func finishBenchmark() {
	if !profiler.active {
		return
	}
	profiler.active = false
	report := struct {
		Version, RecordedUTC, OS, Arch             string
		Map, Car, Quality, Width, Height           int
		VSync, GPUTiming, WorldVBO, InstancedTrees bool
		GPUTimeMS, ActiveSeconds                   float64
		FrameStatistics                            frameSummary
		Certification                              string
		GPUMedianMS, GPUP95MS                      float64
		Graphics                                   nativeConfig
		GPUDevice, OpenGL                          string
		Optimization026                            map[string]any
	}{Version: "2.2-beta", RecordedUTC: time.Now().UTC().Format(time.RFC3339), OS: runtime.GOOS, Arch: runtime.GOARCH, Map: app.mapID, Car: beta.carSelect, Quality: app.quality, Width: app.width, Height: app.height, VSync: app.vsync, GPUTiming: app.gpuTiming, WorldVBO: gpuWorld.ready, InstancedTrees: gpuWorld.instancing, GPUTimeMS: app.gpuMS, ActiveSeconds: profiler.seconds, FrameStatistics: summarizeFrames(profiler.samples), Certification: "Measured gameplay sample; not release certification."}
	if remaster.Enabled {
		report.Version = "0.2.6-beta"
		report.Graphics = beta.cfg
		report.Optimization026 = optimizationReport026()
		report.GPUDevice = glString(0x1F01)
		report.OpenGL = glString(0x1F02)
		report.Certification = "Measured gameplay sample; automated and interactive tests are reported separately."
	}
	gpu := summarizeFrames(profiler.gpuSamples)
	report.GPUMedianMS = gpu.MedianMS
	report.GPUP95MS = gpu.P95MS
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		profiler.message = "BENCHMARK LOG FAILED"
		return
	}
	dir := filepath.Join(filepath.Dir(cfgPath()), "diagnostics")
	if os.MkdirAll(dir, 0755) != nil {
		profiler.message = "BENCHMARK LOG FAILED"
		return
	}
	name := filepath.Join(dir, "benchmark-"+time.Now().Format("20060102-150405")+".json")
	if os.WriteFile(name, data, 0644) != nil {
		profiler.message = "BENCHMARK LOG FAILED"
		return
	}
	profiler.message = "BENCHMARK SAVED TO DIAGNOSTICS"
}
