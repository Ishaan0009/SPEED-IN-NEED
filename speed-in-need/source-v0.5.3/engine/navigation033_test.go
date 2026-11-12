package engine

import (
	"math"
	"testing"
)

func TestNavigation033JunctionShortestRouteAndOffRoadDestination(t *testing.T) {
	d := &WorldDefinition{Grid: HeightGrid{W: 2, H: 2, MinX: -128, MinZ: -128, Step: 256, Values: make([]float32, 4)}, Roads: []RoadDefinition{
		{Name: "main", Width: 8, Points: []Vec{{0, 0}, {100, 0}}},
		{Name: "cross", Width: 8, Points: []Vec{{50, -50}, {50, 50}}},
		{Name: "detour", Width: 8, Points: []Vec{{0, 0}, {0, 100}, {100, 100}, {100, 0}}},
	}}
	g := NewRoadNavigator(d)
	p := g.FindPath(Vec{10, 0}, Vec{50, 57}, math.NaN())
	if !p.Reachable || math.Abs(p.Metres-90) > .1 || p.RoadEnd.Dist(Vec{50, 50}) > .01 || math.Abs(p.OffRoadMetres-7) > .01 {
		t.Fatalf("incorrect route or off-road snap: %+v", p)
	}
	sawCrossing := false
	for _, v := range p.Points {
		if v.Dist(Vec{50, 0}) < .01 {
			sawCrossing = true
		}
	}
	if !sawCrossing {
		t.Fatal("route skipped the actual junction")
	}
}
func TestNavigation033DoesNotTurnAcrossAnOverpass(t *testing.T) {
	d := &WorldDefinition{Grid: HeightGrid{W: 2, H: 2, MinX: -128, MinZ: -128, Step: 256, Values: make([]float32, 4)}, Roads: []RoadDefinition{{Width: 8, Points: []Vec{{-100, 0}, {100, 0}}, Elevations: []float64{0, 0}}, {Width: 8, Points: []Vec{{0, -100}, {0, 100}}, Elevations: []float64{12, 12}}}}
	g := NewRoadNavigator(d)
	if g.Components != 2 {
		t.Fatalf("overpass became an at-grade junction: %d", g.Components)
	}
	p := g.FindPath(Vec{-90, 0}, Vec{0, 80}, 0)
	if !p.Reachable || p.RoadEnd.Dist(Vec{0, 0}) > .01 || p.OffRoadMetres < 79 {
		t.Fatalf("route left its reachable road: %+v", p)
	}
	q := g.FindPath(Vec{0, 0}, Vec{0, 80}, 12)
	if q.RoadEnd.Dist(Vec{0, 80}) > .01 {
		t.Fatal("player elevation did not select the upper deck")
	}
}
func TestNavigation033EndpointJoinsPavedTJunction(t *testing.T) {
	d := &WorldDefinition{Grid: HeightGrid{W: 2, H: 2, MinX: -128, MinZ: -128, Step: 256, Values: make([]float32, 4)}, Roads: []RoadDefinition{{Width: 8, Points: []Vec{{0, 0}, {100, 0}}}, {Width: 8, Points: []Vec{{50, 1.5}, {50, 60}}}}}
	g := NewRoadNavigator(d)
	p := g.FindPath(Vec{0, 0}, Vec{50, 60}, math.NaN())
	if g.Components != 1 || !p.Reachable || math.Abs(p.Metres-110) > .05 {
		t.Fatalf("paved T junction disconnected: %+v", p)
	}
	if g.FindPath(Vec{math.NaN(), 0}, Vec{}, 0).Reachable {
		t.Fatal("accepted a non-finite waypoint")
	}
}
