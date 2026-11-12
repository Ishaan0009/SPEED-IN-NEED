package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestAokawaControlsTurnTowardCameraSide(t *testing.T) {
	for _, left := range []bool{true, false} {
		d := shippedRemasterDefinition(t, 1)
		w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
		w.Car.Speed = 10
		initial := w.Car.Heading
		f := engine.HeadingVec(initial)
		screenRight := norm(cross(pos3{f.X, 0, f.Z}, pos3{0, 1, 0}))
		for i := 0; i < 90; i++ {
			w.Step(engine.Input{Throttle: .25, Steer: aokawaKeyboardSteer(left, !left)}, 1./120)
		}
		now := engine.HeadingVec(w.Car.Heading)
		projection := now.X*screenRight.x + now.Z*screenRight.z
		if (left && projection >= -.01) || (!left && projection <= .01) {
			t.Fatalf("left=%v turns away from selected screen side: %.3f", left, projection)
		}
	}
}

func TestAokawaVegetationRootsFollowFinalTerrain(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	count := map[string]int{}
	for _, o := range d.Objects {
		switch o.Kind {
		case "tree", "bush", "grass":
			h := d.Grid.Height(engine.Vec{X: o.Pos[0], Z: o.Pos[2]})
			offset := 0.
			if o.Kind == "tree" {
				offset = .14
			}
			if o.Kind == "bush" {
				offset = .1
			}
			if o.Asset == "b022_fern" {
				offset = .018
			}
			if d.Name == "SPEED IN NEED" {
				// New prefabs retain their source pivot; validate the actual root.
				meta := readTestPBR030(t, o.Asset)
				offset = meta.Bounds[0][1]*o.Scale[1] + .06
			}
			if math.Abs(o.Pos[1]+offset-h) > .002 {
				t.Fatalf("floating %s at %v: ground %.2f", o.Kind, o.Pos, h)
			}
			count[o.Kind]++
		}
	}
	if count["tree"] < 3000 || count["bush"] < 3000 {
		t.Fatal("forest understory lost")
	}
}

func TestAokawaCoastHasWaterAndNoPhantomMountain(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, p := range []engine.Vec{{X: -1480, Z: 0}, {X: -1490, Z: 1100}, {X: 1490, Z: -1450}} {
		if !w.IsWater(p, 0) {
			t.Fatalf("ocean mask missing at %v", p)
		}
		h := w.GroundHeight(p)
		if h > 14 {
			t.Fatalf("phantom terrain above ocean: %.2f", h)
		}
		level, ok := w.RemasteredWaterLevel(p)
		if !ok || level != 14 {
			t.Fatal("water render/recovery mismatch")
		}
	}
}
func TestAokawaAllBuildingFootprintsStayOffRoads(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, c := range d.Colliders {
		if c.Kind != "building" {
			continue
		}
		f := engine.HeadingVec(c.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		for _, s := range [][2]float64{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} {
			p := c.Center.Add(r.Mul(c.Half.X * s[0])).Add(f.Mul(c.Half.Z * s[1]))
			dist, width, _ := w.RemasteredRoadAt(p)
			if dist < width/2-.1 {
				t.Fatalf("building intersects road: object %d at %+v", c.Object, p)
			}
		}
	}
}
func TestAokawaRoadAtlasAndRegionIdentity(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	if d.Name != "SPEED IN NEED" || len(d.Roads) < 150 || len(d.Signals) < 6 || len(d.Tunnels) != 1 {
		t.Fatal("map content missing")
	}
	raw, e := assets.ReadFile("assets/beta020_atlas.png")
	if e != nil || len(raw) < 100000 {
		t.Fatal("full map reference missing")
	}
	for id := 0; id < 4; id++ {
		spec := engine.RemasteredVehicleSpec(id)
		if spec.P.Mass < 1000 {
			t.Fatal("unrealistically light car")
		}
	}
}

func TestAokawaPortalCarCannotHideUnderVisibleTerrain(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	tunnel := d.Tunnels[0]
	f := engine.HeadingVec(tunnel.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	for _, sign := range []float64{-1, 1} {
		for along := 2.; along <= 36; along += 2 {
			for _, side := range []float64{-4, 0, 4} {
				p := tunnel.Center.Add(f.Mul(sign * (tunnel.Length/2 + along))).Add(r.Mul(side))
				support := w.GroundHeight(p)
				visible := d.Grid.Height(p)
				if support < visible-.05 {
					t.Fatalf("car under portal terrain: along %.1f side %.1f support %.2f visible %.2f", along, side, support, visible)
				}
			}
		}
	}
}
