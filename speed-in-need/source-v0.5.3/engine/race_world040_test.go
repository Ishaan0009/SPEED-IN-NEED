package engine

import (
	"math"
	"os"
	"testing"
)

func shippedRaceFinal040(t *testing.T) *World {
	t.Helper()
	raw, err := os.ReadFile("../assets/b038_world_race.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	return NewRemasteredWorld(d, 1, 0, 0, 1, 0)
}

// This sweeps the full usable driving lanes, rather than testing only the
// centre line. Leaves and opaque scenery must not create invisible walls.
func Test040FinalCourseThreeLanesClear(t *testing.T) {
	w := shippedRaceFinal040(t)
	for class := 0; class < 4; class++ {
		for distance := 0.; distance < w.Routes[0].Len; distance += 3 {
			centre, f := w.Routes[0].AtDistance(distance)
			right := Vec{f.Z, -f.X}
			for _, side := range []float64{-6.4, 0, 6.4} {
				p := centre.Add(right.Mul(side))
				c := Car{Pos: p, Heading: math.Atan2(f.X, f.Z)}
				c.Body = ChassisState{Initialized: true, Grounded: true, Height: w.RemasteredGround(p) + .74}
				w.remasteredStaticContacts(&c, class)
				staticShift:=c.Pos.Dist(p)
				w.remasteredMeshContacts(&c, class)
				if shift := c.Pos.Dist(p); shift > .035 {
					cell:=remasterMeshCell(p);candidates:=w.Remaster.triangleGrid[cell]
					for _,index:=range candidates {probe:=c;probe.Pos=p;w.Remaster.triangleGrid[cell]=[]int{index};w.remasteredMeshContacts(&probe,class);if probe.Pos.Dist(p)>.035 {t.Logf("contact triangle %d %+v body %.3f",index,w.Remaster.Def.CollisionTriangles[index],c.Body.Height)}}
					w.Remaster.triangleGrid[cell]=candidates
					t.Fatalf("car %d at %.0fm lane %.1f obstructed: %.3fm displacement (static %.3f)", class, distance, side, shift,staticShift)
				}
			}
		}
	}
}

func Test040FinalCourseTireSupportAcrossWholeWidth(t *testing.T) {
	w := shippedRaceFinal040(t)
	maxGrade := 0.
	previous := math.NaN()
	for distance := 0.; distance < w.Routes[0].Len; distance += 1.5 {
		centre, f := w.Routes[0].AtDistance(distance)
		right := Vec{f.Z, -f.X}
		for _, side := range []float64{-8, -4, 0, 4, 8} {
			p := centre.Add(right.Mul(side))
			supported := false
			for _, index := range w.Remaster.surfaceGrid[remasterMeshCell(p)] {
				tri := w.Remaster.Def.Surfaces[index]
				if tri.Road && tri.Support {
					if _, ok := triangleHeight(tri, p); ok {
						supported = true
						break
					}
				}
			}
			if !supported {
				t.Fatalf("unsupported tire patch at %.1fm side %.1f", distance, side)
			}
			y := w.RemasteredGround(p)
			if !navFinite033(Vec{y, y}) {
				t.Fatalf("nonfinite deck at %.1fm", distance)
			}
			if side == 0 {
				if !math.IsNaN(previous) {
					maxGrade = math.Max(maxGrade, math.Abs(y-previous)/1.5)
				}
				previous = y
			}
		}
	}
	// A missing deck used to select a distant terrain layer. Such a seam creates
	// an abrupt step much larger than any authored elevation change.
	if maxGrade > 1.1 {
		t.Fatalf("disconnected deck/landing seam: %.2fm/m", maxGrade)
	}
	t.Logf("full-width support sweep passed; maximum authored grade %.3fm/m", maxGrade)
}
