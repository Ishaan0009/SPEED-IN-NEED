package engine

import (
	"math"
	"testing"
)

func Test040FinalRackStaysResponsiveUntilMotorwaySpeeds(t *testing.T) {
	for _, speed := range []float64{0, 60. / 3.6, 100. / 3.6, 150. / 3.6} {
		if motorwaySteerBlend040(speed) != 0 {
			t.Fatalf("early motorway steering attenuation at %.1fkm/h", speed*3.6)
		}
	}
	if motorwaySteerBlend040(180./3.6) <= 0 || motorwaySteerBlend040(300./3.6) != 1 {
		t.Fatal("motorway attenuation missing")
	}
	for _, speed := range []float64{60. / 3.6, 100. / 3.6, 150. / 3.6} {
		angle := 0.
		for tick := 0; tick < 20; tick++ {
			angle = ProgressiveSteer(angle, 1, .15, speed, 1./120)
		}
		if angle < .149 {
			t.Fatalf("slow mid-speed steering rack %.3f at %.0fkm/h", angle, speed*3.6)
		}
		for tick := 0; tick < 32; tick++ {
			angle = ProgressiveSteer(angle, -1, .15, speed, 1./120)
		}
		if angle > -.149 {
			t.Fatalf("slow countersteer %.3f at %.0fkm/h", angle, speed*3.6)
		}
	}
}

func Test040FinalNormalTurningDoesNotCreateAutomaticDrift(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, kmh := range []float64{60, 100, 150, 220} {
			w := flat22(class, true)
			w.Car.Speed = kmh / 3.6
			for tick := 0; tick < 360; tick++ {
				steer := .6 * math.Sin(float64(tick)*.02)
				w.Step(Input{Throttle: .45, Steer: steer}, 1./120)
				if w.Car.DriftActive || w.Car.RearLocked || math.IsNaN(w.Car.Pos.X) || math.Abs(w.Car.YawRate) > 2.4 {
					t.Fatalf("unstable ordinary corner car%d at%.0fkm/h", class, kmh)
				}
			}
		}
	}
}

func Test040FinalRivalsUseControlledDriftAndPower(t *testing.T) {
	for level := 0; level < 3; level++ {
		fixture := shippedRaceFinal040(t)
		w := NewRemasteredWorld(fixture.Remaster.Def, 1, 0, 0, 1, 7)
		w.Remaster.RaceDifficulty040 = level
		peak := 0.
		for tick := 0; tick < 120*240; tick++ {
			w.Step(Input{Brake: 1, Handbrake: true}, 1./120)
			if tick%120 == 0 {
				for _, r := range w.Remaster.Racers {
					peak = math.Max(peak, math.Abs(r.Car.Speed))
				}
			}
		}
		finished, drifts := 0, 0
		seconds, slip, nitro, boost := 0., 0., 0., 0.
		best := math.Inf(1)
		for _, r := range w.Remaster.Racers {
			if r.Finished {
				finished++
				best = math.Min(best, r.FinishTime)
			}
			drifts += r.IntentionalDrifts040
			seconds += r.IntentionalDriftSeconds040
			slip = math.Max(slip, r.DriftPeakAngle040)
			nitro += r.NitroSpent
			boost += r.BoostSeconds040
		}
		t.Logf("difficulty%d finish%d/7, peak%.1fkm/h, best%.2fs, intentional drifts%d/%.2fs, peak slip%.2fdeg, boost%.2fs, nitro%.1f", level, finished, peak*3.6, best, drifts, seconds, slip, boost, nitro)
		if finished != 7 {
			t.Fatalf("difficulty%d failed to finish all rivals", level)
		}
		if level == 0 && drifts != 0 {
			t.Fatal("Easy rivals unexpectedly used drift")
		}
		if level > 0 && (drifts == 0 || seconds < .1 || slip < 1) {
			t.Fatalf("difficulty%d did not demonstrate a controlled physical drift", level)
		}
		if nitro < 1 || boost < .1 {
			t.Fatalf("difficulty%d did not use drivetrain nitro", level)
		}
	}
}
