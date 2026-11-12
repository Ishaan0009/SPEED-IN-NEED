package engine

import (
	"math"
	"testing"
)

func TestFixedClock038PreservesTimeAcrossRenderRatesAndHitches(t *testing.T) {
	for _, fps := range []int{15, 30, 60, 144, 240} {
		var clock FixedClock038
		position, velocity := 0., 0.
		for frame := 0; frame < fps*10; frame++ {
			clock.Advance(1/float64(fps), func(dt float64) { velocity += 2 * dt; position += velocity * dt })
		}
		if clock.Ticks != 1200 || math.Abs(position-100.08333333333) > 1e-7 || clock.DroppedSeconds != 0 {
			t.Fatalf("%d FPS: ticks=%d position=%f dropped=%f", fps, clock.Ticks, position, clock.DroppedSeconds)
		}
	}
	var clock FixedClock038
	n := clock.Advance(.1, func(float64) {})
	if n != 12 || clock.Ticks != 12 || clock.DroppedSeconds != 0 {
		t.Fatal("100ms hitch lost simulation time", clock)
	}
}

func TestFixedClock038PauseAndSuspendDoNotSpiral(t *testing.T) {
	var clock FixedClock038
	clock.Advance(.004, func(float64) { t.Fatal("partial step") })
	clock.Pause()
	if clock.Alpha() != 0 {
		t.Fatal("paused remainder retained")
	}
	if n := clock.Advance(30, func(float64) {}); n != 30 || math.Abs(clock.DroppedSeconds-29.75) > 1e-9 {
		t.Fatal("unbounded suspend catch-up", n, clock)
	}
	for _, invalid := range []float64{-1, math.NaN(), math.Inf(1)} {
		if clock.Advance(invalid, func(float64) { t.Fatal("invalid elapsed") }) != 0 {
			t.Fatal(invalid)
		}
	}
}
