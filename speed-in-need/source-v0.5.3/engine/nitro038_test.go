package engine

import (
	"math"
	"testing"
)

func TestNitro038RespectsEngineGearAndPowerBudget(t *testing.T) {
	for id := 0; id < 4; id++ {
		s := RemasteredVehicleSpec(id)
		for _, speed := range []float64{4, 15, 40, 75} {
			force := NitroDriveForce038(s, s.P.RedlineRPM*.65, s.P.Ratios[2], speed, 0)
			if force <= 0 || force*math.Max(7, speed) > s.PowerKW*.55*1000*.88+.001 {
				t.Fatalf("car%d nitro exceeds its additional power budget", id)
			}
			if NitroDriveForce038(s, 0, s.P.Ratios[2], speed, 0) != 0 {
				t.Fatal("stopped engine produced nitro drive")
			}
			shifted := NitroDriveForce038(s, s.P.RedlineRPM*.65, s.P.Ratios[2], speed, .1)
			if math.Abs(shifted/force-.22) > .00001 {
				t.Fatal("nitro bypassed gear-shift torque interruption")
			}
		}
	}
}
