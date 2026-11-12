package engine

import (
	"math"
	"os"
	"testing"
)

func Test040ShippedRaceAllDifficulties(t *testing.T) {
	raw, err := os.ReadFile("../assets/b038_world_race.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	for level := 0; level < 3; level++ {
		w := NewRemasteredWorld(d, 1, 0, 0, 1, 7)
		w.SetRaceDifficulty051(level)
		maxSpeed := 0.
		finished := 0
		best := math.Inf(1)
		for tick := 0; tick < 120*240; tick++ {
			w.Step(Input{Brake: 1, Handbrake: true}, 1./120)
			if tick%120 == 0 {
				for _, r := range w.Remaster.Racers {
					if math.IsNaN(r.Car.Pos.X) || math.IsNaN(r.Car.Speed) {
						t.Fatalf("difficulty %d unstable", level)
					}
					at, _ := routeProjection(w.Routes[0], r.Car.Pos, r.Segment)
					p, _ := w.Routes[0].AtDistance(at)
					if r.Car.Pos.Dist(p) > d.Roads[0].Width*.5+1 {
						t.Fatalf("difficulty %d rival left course %.2fm", level, r.Car.Pos.Dist(p))
					}
					maxSpeed = math.Max(maxSpeed, math.Abs(r.Car.Speed))
				}
			}
		}
		for _, r := range w.Remaster.Racers {
			if r.Finished {
				finished++
				best = math.Min(best, r.FinishTime)
			} else {
				t.Logf("difficulty %d %s progress %.0f lap %d gate %d", level, r.Name, r.Progress, r.Lap, r.NextGate)
			}
		}
		t.Logf("difficulty %d: %d/7 finished, peak %.1f km/h, fastest lap %.2fs, lap average %.1f km/h", level, finished, maxSpeed*3.6, best, w.Routes[0].Len/best*3.6)
		if finished < 5 {
			t.Fatalf("difficulty %d: too many rivals stuck, only %d finished", level, finished)
		}
	}
}
