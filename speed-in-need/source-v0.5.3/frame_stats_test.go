package main

import (
	"math"
	"testing"
)

func TestFrameStatisticsIncludeStalls(t *testing.T) {
	samples := make([]float64, 1000)
	for i := range samples {
		samples[i] = 10
	}
	for i := 0; i < 10; i++ {
		samples[i] = 100
	}
	s := summarizeFrames(samples)
	if s.LowOnePercentFPS != 10 || s.P99MS != 10 || s.WorstMS != 100 || math.Abs(s.AverageFPS-1000000./10900) > 1e-8 {
		t.Fatalf("incorrect frame statistics: %+v", s)
	}
}

func TestFrameStatisticsRejectInvalidNumbers(t *testing.T) {
	s := summarizeFrames([]float64{0, -1, math.NaN(), math.Inf(1), 20})
	if s.Frames != 1 || s.AverageFPS != 50 {
		t.Fatalf("invalid frame samples leaked: %+v", s)
	}
}
