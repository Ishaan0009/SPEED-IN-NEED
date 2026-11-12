package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"speedinneed/engine"
	"strings"
	"testing"
)

func TestMap035CameraRightAndHeadingRoundTrip(t *testing.T) {
	center := engine.Vec{X: -725.8, Z: 310.3}
	for _, heading := range []float64{0, .43, math.Pi / 2, math.Pi, 4.8, engine.Tau} {
		f := engine.HeadingVec(heading)
		// This is applyCamera's cross(viewDirection, worldUp), independently
		// calculated in 3D to guard against the original mirrored minimap.
		right3 := cross(pos3{f.X, 0, f.Z}, pos3{0, 1, 0})
		right := engine.Vec{X: right3.x, Z: right3.z}
		for _, scale := range []float64{.07, .37, 4.2} {
			p := navProject035(center.Add(right.Mul(50)), center, heading, scale, true)
			if math.Abs(p.X-50*scale) > 1e-8 || math.Abs(p.Z) > 1e-8 {
				t.Fatalf("camera-right road must appear on map right: heading %.2f, %+v", heading, p)
			}
			p = navProject035(center.Add(f.Mul(100)), center, heading, scale, true)
			if math.Abs(p.X) > 1e-8 || math.Abs(p.Z+100*scale) > 1e-8 {
				t.Fatal("ahead must appear above the player", p)
			}
			for _, headingUp := range []bool{true, false} {
				for _, world := range []engine.Vec{{X: -1200, Z: -1320}, {X: 850, Z: 1120}, center} {
					screen := navProject035(world, center, heading, scale, headingUp)
					got := navUnproject035(screen, center, heading, scale, headingUp)
					if got.Dist(world) > 1e-8 {
						t.Fatal("atlas, route and pointer inverse diverged", world, got)
					}
				}
			}
		}
	}
}

func TestMap035FullMapPointerAndZoomAnchor(t *testing.T) {
	oldNav, oldW, oldH := nav033, app.width, app.height
	defer func() { nav033, app.width, app.height = oldNav, oldW, oldH }()
	app.width, app.height = 1920, 1080
	nav033.Definition = &engine.WorldDefinition{Bounds: [2]engine.Vec{{X: -1500, Z: -1500}, {X: 1500, Z: 1500}}}
	nav033.Center = engine.Vec{X: -300, Z: 250}
	x, y, size := navMapRect033()
	for _, zoom := range []float64{.65, 1, 3, 14} {
		nav033.Zoom = zoom
		for _, pixel := range []engine.Vec{{X: x + 11, Z: y + 29}, {X: x + size*.72, Z: y + size*.31}} {
			world := navMapWorld033(pixel.X, pixel.Z)
			q := navProject035(world, nav033.Center, 0, size/navMapSpan033(), false)
			if q.Add(engine.Vec{X: x + size*.5, Z: y + size*.5}).Dist(pixel) > 1e-7 {
				t.Fatal("clicked waypoint is not at the displayed location")
			}
		}
	}
	nav033.Zoom = 1
	px, py := x+size*.78, y+size*.24
	before := navMapWorld033(px, py)
	nav033.Zoom *= 1.2
	nav033.Center = nav033.Center.Add(before.Sub(navMapWorld033(px, py)))
	if navMapWorld033(px, py).Dist(before) > 1e-8 {
		t.Fatal("zoom moved the world location under the mouse pointer")
	}
}

func TestMap035PlaceAnchorsAndModeIsolation(t *testing.T) {
	raw, err := assets.ReadFile("assets/b037_places.json")
	if err != nil {
		t.Fatal(err)
	}
	var places []place034
	if err := json.Unmarshal(raw, &places); err != nil {
		t.Fatal(err)
	}
	if len(places) < 190 {
		t.Fatal("missing detailed road names", len(places))
	}
	d := worldDefinition034(t, "free")
	freeDistricts, raceDistricts := 0, 0
	for _, place := range places {
		if place.Name != friendlyPlaceName035(place.OriginalName) {
			t.Fatal("generated map name differs from minimap naming", place.Name, place.OriginalName)
		}
		if place.Mode == 1 {
			if place.Name != "Coastal Circuit" && place.Name != "Start / Finish" {
				t.Fatal("free-world district leaked into race map", place.Name)
			}
			if place.Kind == "district" {
				raceDistricts++
			}
		}
		if place.Kind != "district" || place.Mode != 0 {
			continue
		}
		freeDistricts++
		best := math.Inf(1)
		for _, road := range d.Roads {
			if !strings.HasPrefix(strings.ToLower(road.Name), strings.ToLower(place.OriginalName)) {
				continue
			}
			for _, p := range road.Points {
				best = math.Min(best, p.Dist(engine.Vec{X: place.X, Z: place.Z}))
			}
		}
		if best > .01 {
			t.Fatal("district name is outside its actual road region", place.Name, best)
		}
		if place.Name == "Fake City" && place.X < 0 {
			t.Fatal("Fake City label returned to the wrong western district")
		}
	}
	if freeDistricts != 4 || raceDistricts != 1 {
		t.Fatal("missing main area names", freeDistricts, raceDistricts)
	}
}

func TestMap035FriendlyNamesPreserveOriginalRoadIdentity(t *testing.T) {
	cases := map[string]string{
		"WHAT ALANDS":               "What Alands",
		"WHAT ALANDS 1":             "What Alands 1",
		"REAL CITY":                 "Real City",
		"REAL CITY 189":             "Real City 189",
		"FAKE CITY":                 "Fake City",
		"fake city 190":             "Fake City 190",
		"MYSTERIOUS JUNGLE":         "Mysterious Jungle",
		"MYSTERIOUS JUNGLE 5":       "Mysterious Jungle 5",
		"COASTAL CIRCUIT":           "Coastal Circuit",
		"THE TUNNEL SOUTH APPROACH": "The Tunnel South Approach",
		"North Portal":              "North Portal",
		"Start / Finish":            "Start / Finish",
	}
	for original, want := range cases {
		if got := friendlyPlaceName035(original); got != want {
			t.Fatalf("friendly name for %q: got %q want %q", original, got, want)
		}
		if got := friendlyPlaceName035(want); got != want {
			t.Fatal("friendly names are not idempotent", want, got)
		}
	}
	d := worldDefinition034(t, "free")
	for _, road := range d.Roads {
		friendly := friendlyPlaceName035(road.Name)
		matched := false
		for _, district := range []string{"What Alands", "Real City", "Fake City", "Mysterious Jungle", "The Tunnel"} {
			matched = matched || strings.HasPrefix(friendly, district)
		}
		if !matched {
			t.Fatal("road has no friendly minimap area name", road.Name, friendly)
		}
	}
}

func TestMap035LabelPlacementRetainsEdgeAndAvoidsPlayer(t *testing.T) {
	for _, anchor := range []engine.Vec{{X: 103, Z: 103}, {X: 895, Z: 895}, {X: 500, Z: 500}} {
		occupied := [][4]float64{{anchor.X - 14, anchor.Z - 14, 28, 28}}
		box, ok := navLabelBox035(anchor.X, anchor.Z, 160, 28, 100, 100, 800, occupied)
		if !ok {
			t.Fatal("visible edge or player-adjacent label was discarded", anchor)
		}
		if box[0] < 105 || box[1] < 105 || box[0]+box[2] > 895 || box[1]+box[3] > 895 {
			t.Fatal("label outside map", box)
		}
	}
}

func referenceCircleClip035(a, b engine.Vec, r float64) (engine.Vec, engine.Vec, bool) {
	d := b.Sub(a)
	den := d.X*d.X + d.Z*d.Z
	if den < 1e-9 {
		return a, b, false
	}
	bb := 2 * (a.X*d.X + a.Z*d.Z)
	cc := a.X*a.X + a.Z*a.Z - r*r
	disc := bb*bb - 4*den*cc
	if disc < 0 {
		return a, b, false
	}
	root := math.Sqrt(disc)
	lo, hi := math.Max(0, (-bb-root)/(2*den)), math.Min(1, (-bb+root)/(2*den))
	return a.Add(d.Mul(lo)), a.Add(d.Mul(hi)), lo <= hi
}

func TestMap035FastClippingMatchesOriginal(t *testing.T) {
	rng := rand.New(rand.NewSource(350351))
	for i := 0; i < 50000; i++ {
		a, b := engine.Vec{X: rng.Float64()*2000 - 1000, Z: rng.Float64()*2000 - 1000}, engine.Vec{X: rng.Float64()*2000 - 1000, Z: rng.Float64()*2000 - 1000}
		if i%100 == 0 {
			b = a
		}
		r := rng.Float64()*180 + 1
		wantA, wantB, wantOK := referenceCircleClip035(a, b, r)
		gotA, gotB, gotOK := navCircleSegment033(a, b, r)
		if wantOK != gotOK || gotOK && (gotA.Dist(wantA) > 1e-7 || gotB.Dist(wantB) > 1e-7) {
			t.Fatal("circle route clipping changed", a, b, r, wantOK, gotOK)
		}
	}
}

var benchmarkMapPoint035 engine.Vec
var benchmarkMapClip035 bool

func BenchmarkMap035Projection(b *testing.B) {
	for _, cached := range []bool{false, true} {
		name := "BasisPerEndpoint"
		if cached {
			name = "BasisOncePerFrame"
		}
		b.Run(name, func(b *testing.B) {
			center := engine.Vec{X: -800, Z: 600}
			projection := newMapProjection035(center, .714, .29, true)
			for i := 0; i < b.N; i++ {
				p := engine.Vec{X: float64(i%400)*5 - 1000, Z: float64(i%373)*5 - 932.5}
				if cached {
					benchmarkMapPoint035 = projection.project(p)
				} else {
					benchmarkMapPoint035 = navProject035(p, center, .714, .29, true)
				}
			}
		})
	}
}

func BenchmarkMap035CircleClipping(b *testing.B) {
	for _, fast := range []bool{false, true} {
		name := "Original"
		if fast {
			name = "TrivialBounds"
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				a := engine.Vec{X: float64(i%400)*5 - 1000, Z: float64(i%373)*5 - 932.5}
				end := a.Add(engine.Vec{X: 3, Z: 2})
				if fast {
					_, _, benchmarkMapClip035 = navCircleSegment033(a, end, 95)
				} else {
					_, _, benchmarkMapClip035 = referenceCircleClip035(a, end, 95)
				}
			}
		})
	}
}
