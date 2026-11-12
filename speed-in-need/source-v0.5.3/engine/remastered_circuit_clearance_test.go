package engine

import (
	"math"
	"os"
	"testing"
)

func TestShippedCircuitCentrelineClearance(t *testing.T) {
	raw, e := os.ReadFile("../assets/remaster_world_race.json")
	if e != nil {
		t.Fatal(e)
	}
	d, e := DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	w := NewRemasteredWorld(d, 1, 0, 0, 1, 0)
	hits := 0
	for distance := 0.; distance < w.Routes[0].Len; distance += 3 {
		p, f := w.Routes[0].AtDistance(distance)
		c := Car{Pos: p, Heading: math.Atan2(f.X, f.Z)}
		c.Body.Height = w.GroundHeight(p) + .74
		c.Body.Initialized = true
		c.Body.Grounded = true
		w.remasteredMeshContacts(&c, 0)
		if shift := c.Pos.Dist(p); shift > .03 {
			if hits < 15 {
				t.Logf("blocked at %.0fm pos=%+v shift=%.2f height=%.2f", distance, p, shift, c.Body.Height)
			}
			hits++
		}
	}
	if hits > 0 {
		t.Fatalf("%d centreline samples intersected collision walls", hits)
	}
}
