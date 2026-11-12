package main

import (
	"math"
	"testing"
)

func TestAudio032ProcessedLoopsAreBoundedAndNonSilent(t *testing.T) {
	for _, name := range []string{"engine0", "engine1", "engine2", "engine3", "idle0", "idle1", "idle2", "idle3", "city", "nature", "surf"} {
		raw, err := assets.ReadFile("assets/audio032/" + name + ".wav")
		if err != nil {
			t.Fatal(err)
		}
		x, err := decodeWAV030(raw)
		if err != nil {
			t.Fatal(err)
		}
		if len(x) < audioRate030 || len(x) > audioRate030*11 {
			t.Fatalf("%s is not a compact loop", name)
		}
		energy := 0.
		for _, v := range x {
			if math.Abs(float64(v)) > .7 {
				t.Fatalf("%s has an excessive peak", name)
			}
			energy += float64(v * v)
		}
		if energy/float64(len(x)) < .0001 {
			t.Fatalf("%s is silent", name)
		}
	}
}
