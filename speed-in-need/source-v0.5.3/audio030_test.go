package main

import (
	"encoding/binary"
	"math"
	"speedinneed/engine"
	"testing"
)

func TestAudio030PCMRejectsTruncationAndWrongRate(t *testing.T) {
	raw := make([]byte, 52)
	copy(raw, "RIFF")
	binary.LittleEndian.PutUint32(raw[4:], 44)
	copy(raw[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(raw[16:], 16)
	binary.LittleEndian.PutUint16(raw[20:], 1)
	binary.LittleEndian.PutUint16(raw[22:], 1)
	binary.LittleEndian.PutUint32(raw[24:], audioRate030)
	binary.LittleEndian.PutUint16(raw[34:], 16)
	copy(raw[36:], "data")
	binary.LittleEndian.PutUint32(raw[40:], 8)
	binary.LittleEndian.PutUint16(raw[44:], 16384)
	samples, err := decodeWAV030(raw)
	if err != nil || len(samples) != 4 || samples[0] != .5 {
		t.Fatalf("valid PCM: %v %v", samples, err)
	}
	if _, err = decodeWAV030(raw[:51]); err == nil {
		t.Fatal("accepted truncated PCM")
	}
	binary.LittleEndian.PutUint32(raw[24:], 48000)
	if _, err = decodeWAV030(raw); err == nil {
		t.Fatal("accepted unsupported sample rate")
	}
}

func TestRemasteredAmbienceDoesNotPlaySurfInland(t *testing.T) {
	clip := make([]float32, 2205)
	for i := range clip {
		clip[i] = float32(math.Sin(float64(i)*.1) * .25)
	}
	m := audioMixer030{Clips: map[string][]float32{"surf": clip}, Gain: 1}
	out := make([]int16, 1024)
	f := audioFrame030{Master: 1, Ambience: 1, Driving: true, Nature: 1}
	m.render(out, f)
	for _, v := range out {
		if v != 0 {
			t.Fatal("surf leaked into inland ambience")
		}
	}
	f.Coast = 1
	m.render(out, f)
	audible := false
	for _, v := range out {
		if v != 0 {
			audible = true
			break
		}
	}
	if !audible {
		t.Fatal("coastal recording is silent")
	}
	d := &engine.WorldDefinition{WaterGrid: engine.HeightGrid{W: 2, H: 2, MinX: -1500, MinZ: -1500, Step: 3000, Values: []float32{0, 0, 0, 0}}, Roads: []engine.RoadDefinition{{Kind: "town", Points: []engine.Vec{{X: 0, Z: 0}, {X: 0, Z: 100}}}}}
	coast, city, nature := ambienceGainsRemastered(engine.Vec{}, d)
	if coast != 0 || city < .9 || nature > .1 {
		t.Fatalf("city mix: %.2f %.2f %.2f", coast, city, nature)
	}
}

func TestRemasteredImportedRecordingsAreNonemptyPCM(t *testing.T) {
	for _, name := range []string{"engine0", "engine1", "engine2", "engine3", "surf", "city", "nature"} {
		raw, err := assets.ReadFile("assets/audio030/" + name + ".wav")
		if err != nil {
			t.Fatal(err)
		}
		clip, err := decodeWAV030(raw)
		if err != nil || len(clip) < audioRate030*10 {
			t.Fatalf("%s: %v", name, err)
		}
		energy := 0.
		for _, v := range clip {
			energy += float64(v * v)
		}
		if energy/float64(len(clip)) < .00001 {
			t.Fatalf("silent imported recording %s", name)
		}
	}
}

func TestRemasteredProcessedEngineTracksRPMWithoutLegacyLayers(t *testing.T) {
	clip := make([]float32, 2205)
	for i := range clip {
		clip[i] = float32(math.Sin(float64(i)*.1) * .25)
	}
	a := audioMixer030{Clips: map[string][]float32{"engine0": clip}, Gain: 1}
	b := audioMixer030{Clips: map[string][]float32{"engine0": clip}, Gain: 1}
	for _, name := range []string{"engine0_load", "wind", "gravel", "forest"} {
		b.Clips[name] = clip
	}
	frame := audioFrame030{Master: 1, Effects: 1, Ambience: 1, RPM: 8000, Speed: 55, Gravel: 1, Forest: 1, Throttle: 1, Driving: true}
	x, y := make([]int16, 1024), make([]int16, 1024)
	a.render(x, frame)
	b.render(y, frame)
	for i := range x {
		if x[i] != y[i] {
			t.Fatal("legacy engine or ambience layer entered the mixer")
		}
	}
	for i := 0; i < 24; i++ {
		a.render(x, frame)
	}
	if a.EnginePitch < 1.2 || a.EnginePitch > 1.361 {
		t.Fatalf("high RPM pitch outside bounded range: %v", a.EnginePitch)
	}
	frame.RPM = 900
	for i := 0; i < 24; i++ {
		a.render(x, frame)
	}
	if a.EnginePitch < .77 || a.EnginePitch > .80 {
		t.Fatalf("engine did not settle to idle pitch: %v", a.EnginePitch)
	}
}
func TestAudio030StationsProduceSoundAndMuteSettles(t *testing.T) {
	clip := make([]float32, 2205)
	for i := range clip {
		clip[i] = float32(math.Sin(float64(i)*.1) * .3)
	}
	clips := map[string][]float32{}
	for _, name := range []string{"engine0", "radio0", "radio1", "radio2", "radio3", "wind"} {
		clips[name] = clip
	}
	mixer := audioMixer030{Clips: clips}
	out := make([]int16, 1024)
	for station := 1; station <= 4; station++ {
		mixer.render(out, audioFrame030{Master: 1, Effects: 1, Music: 1, RPM: 3000, Speed: 30, Driving: true, Station: station})
		nonzero := false
		for _, x := range out {
			if x != 0 {
				nonzero = true
				break
			}
		}
		if !nonzero {
			t.Fatalf("silent station %d", station)
		}
	}
	for i := 0; i < 40; i++ {
		mixer.render(out, audioFrame030{Master: 1, Music: 1, Driving: true, Station: 1, Muted: true})
	}
	for _, x := range out {
		if x != 0 {
			t.Fatal("mute did not converge to silence")
		}
	}
}
