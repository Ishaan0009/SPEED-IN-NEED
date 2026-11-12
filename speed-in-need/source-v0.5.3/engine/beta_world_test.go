package engine

import (
	"math"
	"testing"
)

func TestBetaBuildingFootprintsClearAllRoads(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		if len(w.Buildings) == 0 {
			t.Fatal("empty building layout")
		}
		for _, b := range w.Buildings {
			for x := -6.5; x <= 6.5; x += 1.3 {
				for z := -7.5; z <= 7.5; z += 1.5 {
					p := b.Add(Vec{x, z})
					if DistRoad(p, w.Routes) < 10.5 || w.IsWater(p, 0) {
						t.Fatalf("map %d building footprint on road/water: %v", mapID, p)
					}
				}
			}
		}
	}
}
func TestBetaRampsFollowRoadAndClearCrossings(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		for _, r := range w.Ramps {
			last := math.Inf(-1)
			for d := 0.; d < r.Length; d += .5 {
				p, dir := r.Sample(d)
				if DistRoad(p, w.Routes) > .1 {
					t.Fatalf("ramp off road: %v", p)
				}
				h, ok := r.Surface(p)
				if !ok || h < last-.02 {
					t.Fatalf("invalid ramp surface: %v %g", p, h)
				}
				last = h
				after, _ := r.Sample(d + .1)
				if after.Sub(p).Unit().X*dir.X+after.Sub(p).Unit().Z*dir.Z < .97 {
					t.Fatal("ramp points against road tangent")
				}
				for k, route := range w.Routes {
					if k != r.RouteID && DistRoad(p, []Route{route}) < 16 {
						t.Fatalf("map %d ramp %s obstructs crossing", mapID, r.Name)
					}
				}
			}
		}
		for _, npc := range w.Traffic {
			if npc.Path != 0 {
				t.Fatal("civilian traffic routed onto launch ramp")
			}
		}
	}
}
func TestBetaPoleFollowsImpactAndStopsAboveTerrain(t *testing.T) {
	for _, velocity := range []Vec{{20, 0}, {-20, 0}, {0, 20}, {0, -20}} {
		p := Pole{Pos: Vec{35, 12}, Heading: .4}
		p.Hit(velocity)
		for i := 0; i < 2400; i++ {
			p.Step(1. / 120)
		}
		if !p.Settled || math.Abs(p.AngularSpeed) > .001 || p.FallDirection.Sub(velocity.Unit()).Length() > .001 {
			t.Fatalf("unstable pole: %+v", p)
		}
		for d := 2.; d <= 10; d += 2 {
			q := p.Pos.Add(p.FallDirection.Mul(d * math.Sin(p.Angle)))
			y := Terrain(p.Pos.X, p.Pos.Z) + .1 + d*math.Cos(p.Angle)
			if y < Terrain(q.X, q.Z)+.10 {
				t.Fatal("fallen pole penetrates terrain")
			}
		}
	}
}
func TestBetaLampArmFacesRoad(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		for _, p := range w.Poles {
			head, y := p.LampHead()
			if DistRoad(head, w.Routes) >= DistRoad(p.Pos, w.Routes) {
				t.Fatal("streetlight points away from road")
			}
			if y < Terrain(p.Pos.X, p.Pos.Z)+9 {
				t.Fatal("lamp height incorrect")
			}
		}
	}
}
func TestBetaNPCNightSchedule(t *testing.T) {
	for _, hour := range []float64{12, 14, 16} {
		if AutomaticLights(hour) {
			t.Fatal("daytime NPC lights")
		}
	}
	for _, hour := range []float64{18.5, 22, 2, 6} {
		if !AutomaticLights(hour) {
			t.Fatal("missing dusk/night lights")
		}
	}
}
