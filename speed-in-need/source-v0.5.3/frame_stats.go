package main

import (
	"math"
	"sort"
)

type frameSummary struct {
	Frames           int     `json:"frames"`
	AverageFPS       float64 `json:"average_fps"`
	LowOnePercentFPS float64 `json:"one_percent_low_fps"`
	MedianMS         float64 `json:"median_ms"`
	P95MS            float64 `json:"p95_ms"`
	P99MS            float64 `json:"p99_ms"`
	WorstMS          float64 `json:"worst_ms"`
}

func summarizeFrames(samples []float64) frameSummary {
	values := make([]float64, 0, len(samples))
	total := 0.
	for _, v := range samples {
		if v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0) {
			values = append(values, v)
			total += v
		}
	}
	if len(values) == 0 {
		return frameSummary{}
	}
	sort.Float64s(values)
	n := len(values)
	percentile := func(p float64) float64 {
		i := int(math.Ceil(float64(n)*p)) - 1
		if i < 0 {
			i = 0
		}
		return values[i]
	}
	count := int(math.Ceil(float64(n) * .01))
	slowest := 0.
	for _, v := range values[n-count:] {
		slowest += v
	}
	return frameSummary{Frames: n, AverageFPS: 1000 * float64(n) / total, LowOnePercentFPS: 1000 * float64(count) / slowest, MedianMS: percentile(.5), P95MS: percentile(.95), P99MS: percentile(.99), WorstMS: values[n-1]}
}
