package main

import (
	"encoding/json"
	"math"
	"os"
	"speedinneed/engine"
	"strconv"
	"testing"
)

// Opt-in simulated-time endurance check; this is not an interactive Windows soak.
func TestRemasteredLongSimulationSoak(t *testing.T) {
	if os.Getenv("SPEED_LONG_SOAK") != "1" {
		t.Skip("set SPEED_LONG_SOAK=1 for the 30-minute-per-world simulation")
	}
	seconds := 1800
	if n, err := strconv.Atoi(os.Getenv("SPEED_SOAK_SECONDS")); err == nil && n >= 60 && n <= 1800 && n%4 == 0 {
		seconds = n
	}
	type result struct {
		Mode, Steps, CarsChecked                        int
		Seconds, MaxSpeed, MaxRoll, MaxBrakeTemperature float64
	}
	results := []result{}
	for mode := 0; mode < 2; mode++ {
		d := shippedRemasterDefinition(t, mode)
		w := engine.NewRemasteredWorld(d, mode, 0, mode*2, 8, 7)
		r := result{Mode: mode, Seconds: float64(seconds), CarsChecked: 4}
		hint := -1
		for step := 0; step < seconds*120; step++ {
			if step > 0 && step%(seconds/4*120) == 0 {
				class := step / (seconds / 4 * 120)
				w = engine.NewRemasteredWorld(d, mode, class, (class+mode)%3, 8, 7)
				hint = -1
			}
			_, _, routeID := w.RemasteredRoadAt(w.Car.Pos)
			if mode == 1 {
				routeID = 0
			}
			route := w.Routes[engine.ClampInt(routeID, 0, len(w.Routes)-1)]
			distance, next := route.ClosestDistance(w.Car.Pos, hint)
			hint = next
			target, _ := route.AtDistance(distance + 12 + math.Abs(w.Car.Speed)*.45)
			delta := target.Sub(w.Car.Pos)
			angle := math.Atan2(delta.X, delta.Z) - w.Car.Heading
			angle = math.Atan2(math.Sin(angle), math.Cos(angle))
			in := engine.Input{Throttle: .65, Steer: engine.Clamp(angle*2.2, -1, 1)}
			if w.Car.Speed > 24 {
				in.Throttle = 0
				in.Brake = .2
			}
			if step%3600 < 90 && step > 600 {
				in.Handbrake = true
			}
			in.Boost = step%7200 > 5000 && math.Abs(angle) < .08
			if step%7200 > 6900 && math.Abs(w.Car.Speed) < 1 {
				w.RecoverRemasteredCar()
				hint = -1
			}
			w.Step(in, 1./120)
			cars := []engine.Car{w.Car}
			for _, ai := range w.Remaster.Racers {
				cars = append(cars, ai.Car)
			}
			for _, c := range cars {
				for _, value := range []float64{c.Pos.X, c.Pos.Z, c.Speed, c.Lateral, c.Heading, c.Body.Height, c.Rig.Pitch, c.Rig.Roll, c.Nitro} {
					if math.IsNaN(value) || math.IsInf(value, 0) {
						t.Fatalf("mode %d step %d non-finite vehicle", mode, step)
					}
				}
				if math.Abs(c.Speed) > 110 || math.Abs(c.Rig.Roll) > .65 || c.Nitro < 0 || c.Nitro > 100.001 {
					t.Fatalf("unbounded vehicle state mode=%d step=%d", mode, step)
				}
				r.MaxSpeed = math.Max(r.MaxSpeed, math.Abs(c.Speed))
				r.MaxRoll = math.Max(r.MaxRoll, math.Abs(c.Rig.Roll))
				r.MaxBrakeTemperature = math.Max(r.MaxBrakeTemperature, c.BrakeTemperature)
			}
			r.Steps++
		}
		results = append(results, r)
		t.Logf("mode %d simulated %.0f seconds, %d fixed steps", mode, r.Seconds, r.Steps)
	}
	raw, _ := json.MarshalIndent(map[string]any{"status": "PASS", "kind": "Automated simulated-time endurance, not interactive hardware soak", "worlds": results}, "", "  ")
	if err := os.WriteFile("SIMULATION_SOAK_BETA_002.json", raw, 0644); err != nil {
		t.Fatal(err)
	}
}
