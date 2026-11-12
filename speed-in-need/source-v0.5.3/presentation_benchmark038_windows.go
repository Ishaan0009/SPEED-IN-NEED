package main

import "time"

var presentationBenchmark038 struct {
	Start                          time.Time
	Rendered, Displayed, Generated float64
}
var frozenPresentation038 map[string]any

func number038(m map[string]any, key string) float64 { value, _ := m[key].(float64); return value }
func beginPresentationBenchmark038() {
	frozenPresentation038 = nil
	s, f := streamlineReport038(), fsrReport038()
	presentationBenchmark038.Start = time.Now()
	presentationBenchmark038.Rendered = float64(nativePresented038())
	presentationBenchmark038.Displayed = number038(s, "dxgi_last_present_count")
	presentationBenchmark038.Generated = number038(f, "generated_frames")
}
func presentationBenchmarkReport038() map[string]any {
	if frozenPresentation038 != nil {
		return frozenPresentation038
	}
	q := &presentationBenchmark038
	if q.Start.IsZero() || outputFBO038 == 0 {
		return map[string]any{"available": false}
	}
	elapsed := time.Since(q.Start).Seconds()
	s, f := streamlineReport038(), fsrReport038()
	rendered := float64(nativePresented038()) - q.Rendered
	displayed := number038(s, "dxgi_last_present_count") - q.Displayed
	generated := number038(f, "generated_frames") - q.Generated
	return map[string]any{"available": elapsed > 0, "seconds": elapsed, "rendered_frames": rendered, "rendered_average_fps": rendered / elapsed, "sdk_generated_frames": generated, "dxgi_present_count_delta": displayed, "presented_average_fps": displayed / elapsed, "dxgi_result": s["dxgi_present_count_result"], "source": "DXGI GetLastPresentCount on the underlying swapchain; present counts do not guarantee that every frame reaches a monitor refresh; generated frames do not advance physics", "presentation_pacing_measured": false}
}
func endPresentationBenchmark038() { frozenPresentation038 = presentationBenchmarkReport038() }
