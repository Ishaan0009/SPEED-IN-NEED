package main

import (
	"math"
	"testing"
)

func TestDriver035ArmLengthsAndGripContinuity(t *testing.T) {
	for id := 0; id < 4; id++ {
		last := seatedDriverPose032(id, -.7)
		lastBasis := driverSkinBasis035(last)
		for i := -69; i <= 70; i++ {
			p := seatedDriverPose032(id, float64(i)*.01)
			basis := driverSkinBasis035(p)
			for _, b := range []int{3, 4, 6, 7} {
				if basis[b].U.Dot(lastBasis[b].U) < .95 {
					t.Fatal("driver skin basis twisted abruptly")
				}
			}
			for _, arm := range [][3]int{{3, 4, 5}, {6, 7, 8}} {
				upper := p[arm[1]].Sub(p[arm[0]]).Length()
				lower := p[arm[2]].Sub(p[arm[1]]).Length()
				if math.Abs(upper-.305) > .0001 || math.Abs(lower-.335) > .0001 {
					t.Fatalf("car%d arm length %.4f %.4f", id, upper, lower)
				}
				if p[arm[1]].Sub(last[arm[1]]).Length() > .025 || p[arm[2]].Sub(last[arm[2]]).Length() > .025 {
					t.Fatal("driver arm snapped between steering samples")
				}
			}
			last = p
			lastBasis = basis
		}
	}
}
