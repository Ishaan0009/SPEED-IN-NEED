package engine

import (
	"math"
	"testing"
)

func TestLANGrid050DoesNotOverlapCars(t *testing.T) {
	for ai := 0; ai <= 7; ai++ {
		type point050 struct{ x, z float64 }
		var points []point050
		for i := 0; i < ai; i++ {
			lane := -1.9
			if i%2 == 1 {
				lane = 1.9
			}
			points = append(points, point050{lane, -5 - float64(i/2)*7})
		}
		for slot := 0; slot < 4; slot++ {
			back, lane := LANGrid050(ai, slot)
			p := point050{lane, -back}
			for _, other := range points {
				if math.Abs(p.x-other.x) < 2.2 && math.Abs(p.z-other.z) < 5 {
					t.Fatalf("overlapping grid: AI=%d slot=%d", ai, slot)
				}
			}
			points = append(points, p)
		}
	}
}
