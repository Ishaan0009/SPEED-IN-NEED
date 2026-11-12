package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func TestRemastered051SuppliedBasementSupportsAllVariants(t *testing.T) {
	raw, err := assets.ReadFile("assets/b051_basement.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.CollisionTriangles) < 100 || d.Objects[0].Asset != "b051_basement" {
		t.Fatal("supplied basement geometry missing")
	}
	for variant := 0; variant < len(garageCars051()); variant++ {
		w := engine.NewRemasteredWorld(d, 0, 3, 0, 1, 0)
		w.Car.Variant050 = variant
		w.Car.Body.Initialized = false
		w.Car.TiresReady = false
		for tick := 0; tick < 120; tick++ {
			w.Step(engine.Input{Brake: 1, Handbrake: true}, 1./120)
		}
		if math.IsNaN(w.Car.Body.Height) || math.Abs(w.Car.Body.Height-w.GroundHeight(w.Car.Pos)) > 2 {
			t.Fatalf("variant %d has no stable floor support", variant)
		}
		if w.Car.Pos.Dist(d.Spawn) > 1 {
			t.Fatalf("variant %d spawn is inside a wall", variant)
		}
	}
}
func TestRemastered051CarCatalogContainsRealAssets(t *testing.T) {
	cars := garageCars051()
	if len(playerVariants052) != 9 {
		t.Fatal("player selector must contain nine distinct models")
	}
	seen := map[string]bool{}
	for i, c := range cars {
		if i == 0 {
			continue
		}
		if seen[c.Key] {
			t.Fatal("duplicate vehicle", c.Key)
		}
		seen[c.Key] = true
		f, e := assets.Open("assets/pbr/" + c.Key + ".pbr")
		if e != nil {
			t.Fatal(c.Name, e)
		}
		f.Close()
		spec := engine.VehicleVariantSpec050(3, i)
		if (i < 15 && spec.Name == engine.RemasteredVehicleSpec(3).Name) || spec.TopSpeed <= 0 {
			t.Fatal("new model has old base-car physics", c.Name)
		}
	}
	if !seen["b051_r34"] {
		t.Fatal("R34 asset omitted from internal catalog")
	}
}
func TestRemastered051BasementAndMenuAreSeparateAssets(t *testing.T) {
	raw, _ := assets.ReadFile("assets/b051_workshop.json")
	var old struct{ Objects []struct{ Asset string } }
	if json.Unmarshal(raw, &old) != nil || len(old.Objects) == 0 {
		t.Fatal("menu garage asset missing")
	}
	for _, o := range old.Objects {
		if o.Asset == "b051_basement" {
			t.Fatal("test basement replaced menu garage")
		}
	}
}

func TestRemastered051TruenoPaintCycling(t *testing.T) {
	oldClass, oldPaint, oldGarage := beta.carSelect, beta.paint, garage050
	oldStore := profiles040.Store
	defer func() {
		beta.carSelect = oldClass
		beta.paint = oldPaint
		garage050 = oldGarage
		profiles040.Store = oldStore
	}()
	profiles040.Store = nil
	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	beta.carSelect = 3
	beta.paint = 1
	if remasterPaintCount() < 3 {
		t.Fatal("Trueno customization has only the old fixed finish")
	}
	remasterCycle(3, 1)
	if beta.paint != 2 {
		t.Fatal("paint keys ignored on the original Trueno", beta.paint)
	}
}
