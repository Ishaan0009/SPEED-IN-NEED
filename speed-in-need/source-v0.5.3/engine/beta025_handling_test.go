package engine

import (
	"fmt"
	"math"
	"testing"
)

func TestBeta025EveryCarTurnsWhileAcceleratingAtDifferentSpeeds(t *testing.T) {
	for car := 0; car < 4; car++ {
		for _, speed := range []float64{8, 20, 38, 55} {
			for _, sign := range []float64{-1, 1} {
				t.Run(fmt.Sprintf("car%d-speed%.0f-sign%.0f", car, speed, sign), func(t *testing.T) {
					w := flat22(car, true)
					w.Car.Speed = speed
					before := w.Car.Heading
					for tick := 0; tick < 96; tick++ {
						w.Step(Input{Throttle: 1, Steer: sign * .65}, 1./120)
					}
					turn := math.Remainder(w.Car.Heading-before, Tau) * sign
					if turn < .025 {
						t.Fatalf("full-throttle steering failed: heading change %.5f, rack %.5f, speed %.2f", turn, w.Car.VisualSteer, w.Car.Speed)
					}
					if !isFinite025(w.Car.Speed) || !isFinite025(w.Car.YawRate) || math.Abs(w.Car.YawRate) > 2.4 {
						t.Fatal("unstable acceleration turn")
					}
					if w.Car.DriftActive {
						t.Fatal("ordinary acceleration turn enabled handbrake drift")
					}
				})
			}
		}
	}
}
func TestBeta025FullThrottleRetainsUsefulTurnAuthority(t *testing.T) {
	for car := 0; car < 4; car++ {
		coast, drive := flat22(car, true), flat22(car, true)
		for _, w := range []*World{coast, drive} {
			w.Car.Speed = 20
		}
		startCoast, startDrive := coast.Car.Heading, drive.Car.Heading
		for tick := 0; tick < 96; tick++ {
			coast.Step(Input{Steer: .7}, 1./120)
			drive.Step(Input{Throttle: 1, Steer: .7}, 1./120)
		}
		a, b := math.Abs(math.Remainder(coast.Car.Heading-startCoast, Tau)), math.Abs(math.Remainder(drive.Car.Heading-startDrive, Tau))
		if b < a*.40 {
			t.Fatalf("car %d loses steering under throttle: coast %.4f, throttle %.4f", car, a, b)
		}
	}
}
func isFinite025(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
