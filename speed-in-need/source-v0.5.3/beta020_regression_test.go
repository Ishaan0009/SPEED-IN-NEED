package main

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func TestOneIslandUsesBoundedDenseVegetation(t *testing.T) {
	raw, err := assets.ReadFile("assets/remaster_world_free.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Beta020 || d.AreaKM2 != 9 || d.Bounds[1].X-d.Bounds[0].X != 3000 {
		t.Fatal("v0.2 world data missing or map size changed")
	}
	stream, err := assets.ReadFile("assets/" + d.VegetationFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(stream) < 12 || string(stream[:4]) != "VEG2" {
		t.Fatal("ground-cover stream missing")
	}
	total := int(binary.LittleEndian.Uint32(stream[8:12]))
	for _, o := range d.Objects {
		if o.Kind == "tree" || o.Kind == "road-tree" || o.Kind == "bush" {
			total++
		}
	}
	if total < 400000 || total > 1000000 {
		t.Fatalf("vegetation count = %d", total)
	}
}

func TestBeta020FleetAddsOneThousandToNormalPreset(t *testing.T) {
	raw, _ := assets.ReadFile("assets/remaster_world_free.json")
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	if len(w.Traffic) != 1032 {
		t.Fatalf("normal traffic = %d; expected 32 + 1000", len(w.Traffic))
	}
	for _, actor := range w.Traffic {
		if actor.Lane <= 0 || math.IsNaN(actor.Heading) {
			t.Fatal("traffic is not in a valid right-hand lane")
		}
	}
	w.SetTrafficDensity(0)
	if len(w.Traffic) != 0 {
		t.Fatal("traffic-off setting failed")
	}
}

func TestBeta020HDRSkiesRetainLinearRadiance(t *testing.T) {
	for _, label := range []string{"day", "afternoon", "night"} {
		raw, err := assets.ReadFile("assets/b020_sky_" + label + ".f16")
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) < 12 || string(raw[:4]) != "HDR2" {
			t.Fatal("invalid linear HDR header")
		}
		w, h := int(binary.LittleEndian.Uint32(raw[4:])), int(binary.LittleEndian.Uint32(raw[8:]))
		if w != 2048 || h != 1024 || len(raw) != 12+w*h*6 {
			t.Fatal("invalid HDR payload")
		}
	}
	var fonts map[string]remasterGlyph
	raw, _ := assets.ReadFile("assets/remaster_alphabet.json")
	if err := json.Unmarshal(raw, &fonts); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"SPEED IN NEED", "100%", "< >", "F12 PHOTO", "BETA 0.2"} {
		for _, c := range label {
			if _, ok := fonts[string(c)]; !ok {
				t.Fatalf("missing interface glyph %c", c)
			}
		}
	}
}
