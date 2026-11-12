package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"speedinneed/engine"
	"testing"
)

// Exercise the shipped spawn, chassis and drivetrain in every selection.
// This is a deterministic simulation matrix, not a rendered/interactive test.
func TestBeta021AllCarsModesAndLighting(t *testing.T) {
	type result struct {
		Mode, Car, Lighting int
		MaxSpeed, MaxSlip   float64
		Status              string
	}
	rows := []result{}
	for mode := 0; mode < 2; mode++ {
		d := shippedRemasterDefinition(t, mode)
		for car := 0; car < 4; car++ {
			for lighting := 0; lighting < 3; lighting++ {
				t.Run(fmt.Sprintf("mode%d_car%d_light%d", mode, car, lighting), func(t *testing.T) {
					w := engine.NewRemasteredWorld(d, mode, car, lighting, 1, 3)
					r := result{Mode: mode, Car: car, Lighting: lighting, Status: "PASS"}
					for tick := 0; tick < 720; tick++ {
						in := engine.Input{Throttle: .5}
						if tick >= 360 && tick < 420 {
							in.Handbrake, in.Steer = true, .4
							in.Throttle = 0
						}
						w.Step(in, 1./120)
						c := w.Car
						for _, v := range []float64{c.Pos.X, c.Pos.Z, c.Body.Height, c.Speed, c.Lateral, c.Heading, c.Gearbox.RPM} {
							if math.IsNaN(v) || math.IsInf(v, 0) {
								t.Fatal("invalid vehicle state")
							}
						}
						if tick < 360 && (c.DriftActive || c.DriftScore != 0) {
							t.Fatal("normal driving enabled drift")
						}
						if c.Body.Height < -100 || math.Abs(c.Rig.Roll) > .65 || c.Gearbox.Gear < -1 || c.Gearbox.Gear > 8 {
							t.Fatal("chassis/drivetrain escaped bounds")
						}
						r.MaxSpeed = math.Max(r.MaxSpeed, math.Abs(c.Speed))
						r.MaxSlip = math.Max(r.MaxSlip, math.Abs(c.Lateral))
					}
					if r.MaxSpeed < 1 {
						t.Fatal("vehicle never left its spawn")
					}
					rows = append(rows, r)
				})
			}
		}
	}
	if len(rows) != 24 {
		t.Fatalf("only %d of 24 selections passed", len(rows))
	}
	raw, _ := json.MarshalIndent(map[string]any{"method": "6 simulated seconds per selection; not a rendered playtest", "selections": rows}, "", "  ")
	if err := os.WriteFile("VEHICLE_MATRIX_BETA_021.json", raw, 0644); err != nil {
		t.Fatal(err)
	}
}
