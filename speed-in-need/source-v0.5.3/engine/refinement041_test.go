package engine

import (
	"math"
	"testing"
)

func Test041ResponseAndProgression(t *testing.T) {
	for _, speed := range []float64{0, 60, 100, 120} {
		if math.Abs(steeringResponse041(speed/3.6)-1.4) > 1e-9 {
			t.Fatalf("slow rack at %.0f", speed)
		}
	}
	if steeringResponse041(180/3.6) != 1 {
		t.Fatal("motorway rack must retain stability")
	}
	s := NewProfileStore040()
	if s.Find040("ishaan").Level040() != 50 {
		t.Fatal("protected profile level")
	}
	p, e := s.Create040("041 progression", "custom_001")
	if e != nil {
		t.Fatal(e)
	}
	p.XP = XPForLevel040(40)
	for class := 0; class < 3; class++ {
		if !p.CanUseCar040(class) || p.PaintCount040(class, 10) != 10 {
			t.Fatalf("incomplete level40 unlocks car%d", class)
		}
	}
}
func Test041BrakeRemainsAvailableWithoutEngine(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := flat22(class, true)
		w.Car.Speed = 100 / 3.6
		for i := 0; i < 240; i++ {
			w.Step(Input{Brake: 1, Handbrake: true, EngineGated: true, EngineEnabled: false}, PhysicsStep038)
		}
		if math.Abs(w.Car.Speed) > .2 {
			t.Fatalf("engine-off brake car%d %.3fm/s", class, w.Car.Speed)
		}
	}
}
func Test041NearPassRejectsStoppedAndDistantVehicles(t *testing.T) {
	for _, item := range []struct{ speed, side float64 }{{0, 2.7}, {2, 2.7}, {8, 4.5}} {
		w, p := driveEventFixture040(t)
		w.Traffic[0].Speed = item.speed
		w.Traffic[0].Pos.X = item.side
		var observer DriveProgress040
		if len(sweptPass040(&observer, w, p)) != 0 || p.NearPasses != 0 {
			t.Fatal("non-close or stopped vehicle awarded a pass")
		}
	}
}
