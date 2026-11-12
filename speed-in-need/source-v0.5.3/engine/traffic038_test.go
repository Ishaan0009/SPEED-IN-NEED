package engine

import "testing"

func TestTrafficGrid038MovesMembershipWithoutStaleCells(t *testing.T) {
	s := &RemasteredState{}
	for tick := 0; tick < 100; tick++ {
		cars := []Traffic{{Pos: Vec{X: float64(tick * 80)}}, {Pos: Vec{X: float64(tick*80 + 2)}}, {Pos: Vec{X: float64(tick*80 + 100)}}}
		s.beta020TrafficGrid(cars)
		seen := map[int]int{}
		for cell, ids := range s.TrafficGrid020 {
			for _, i := range ids {
				if trafficCell020(cars[i].Pos) != cell {
					t.Fatal("stale membership", tick)
				}
				seen[i]++
			}
		}
		if len(s.TrafficCells038) != 2 || len(seen) != 3 {
			t.Fatal("missing active cells", tick)
		}
		for _, count := range seen {
			if count != 1 {
				t.Fatal("duplicate membership")
			}
		}
	}
}
