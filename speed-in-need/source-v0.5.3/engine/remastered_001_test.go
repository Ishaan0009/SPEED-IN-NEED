package engine

import (
	"math"
	"testing"
)

func remasteredFixture(mode, class int) *World {
	grid := HeightGrid{W: 17, H: 17, MinX: -800, MinZ: -800, Step: 100, Values: make([]float32, 17*17)}
	points := SampleSpline([]Vec{{0, -300}, {250, -300}, {300, 0}, {250, 300}, {0, 300}, {-250, 300}, {-300, 0}, {-250, -300}}, true, 12)
	d := &WorldDefinition{Name: "TEST WORLD", Grid: grid, Roads: []RoadDefinition{{Name: "TEST LOOP", Points: points, Width: 15}}, Spawn: Vec{0, -300}, SpawnHeading: math.Pi / 2, Bounds: [2]Vec{{-800, -800}, {800, 800}}}
	w := NewRemasteredWorld(d, mode, class, 0, 2, 3)
	w.Traffic = nil
	return w
}

func TestRemastered001TunnelRoofAndBoreStaySeparate(t *testing.T) {
	w := remasteredFixture(0, 0)
	for i := range w.Remaster.Def.Grid.Values {
		w.Remaster.Def.Grid.Values[i] = 15
	}
	w.Remaster.Def.Tunnels = []TunnelDefinition{{Center: Vec{}, Length: 100, Width: 12, StartY: 0, EndY: 0}}
	c := w.Car
	c.Body.Initialized = true
	c.Body.Height = 15.74
	if y := w.remasteredSupport(&c, Vec{}); y != 15 {
		t.Fatalf("mountain support = %v", y)
	}
	c.Body.Height = .74
	if y := w.remasteredSupport(&c, Vec{}); y != 0 {
		t.Fatalf("bore support = %v", y)
	}
}
func TestRemastered001NormalDrivingNeverEnablesDrift(t *testing.T) {
	for class := 0; class < 4; class++ {
		for mode := 0; mode < 3; mode++ {
			w := remasteredFixture(0, class)
			w.DriveMode = mode
			for tick := 0; tick < 1200; tick++ {
				w.Step(Input{Throttle: .8, Steer: .6 * math.Sin(float64(tick)*.004)}, 1./120)
				if w.Car.DriftActive || w.Car.DriftScore != 0 {
					t.Fatalf("class %d mode %d drifting without handbrake", class, mode)
				}
				if math.Abs(w.Car.Rig.Roll) > .191 {
					t.Fatal("exaggerated chassis roll")
				}
			}
		}
	}
}
func TestRemastered001HandbrakeLocksRearWheels(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 24
		w.Car.Heading = 0
		w.Step(Input{Handbrake: true, Steer: .8}, 1./120)
		if !w.Car.DriftActive {
			t.Fatal("handbrake did not open drift state")
		}
		for i := 2; i < 4; i++ {
			if w.Car.Rig.Wheels[i].AngularVelocity != 0 {
				t.Fatalf("rear wheel %d not locked", i)
			}
		}
		if w.Car.Rig.Wheels[0].AngularVelocity == 0 {
			t.Fatal("front wheel incorrectly locked")
		}
	}
}
func TestRemastered001NitroUsesBoundedMassDependentForce(t *testing.T) {
	for class := 0; class < 4; class++ {
		normal, boost := remasteredFixture(0, class), remasteredFixture(0, class)
		normal.Car.Speed = 40
		boost.Car.Speed = 40
		// Tire torque builds through wheel inertia; measure a 0.5 s boost window,
		// not the first substep before longitudinal tire force has relaxed.
		for tick := 0; tick < 120; tick++ {
			normal.Step(Input{Throttle: .7}, 1./120)
			boost.Step(Input{Throttle: .7}, 1./120)
		}
		for tick := 0; tick < 60; tick++ {
			normal.Step(Input{Throttle: .7}, 1./120)
			boost.Step(Input{Throttle: .7, Boost: true}, 1./120)
		}
		extra := (boost.Car.Speed - normal.Car.Speed) * 2
		if extra < .5 || extra > 2.3 {
			t.Fatalf("class %d nitro acceleration %.3f", class, extra)
		}
		airborne := remasteredFixture(0, class)
		airborne.Car.Speed = 24
		airborne.Car.Body.Height = 25
		airborne.Car.Body.Grounded = false
		airborne.Car.Body.AirVelocity = HeadingVec(airborne.Car.Heading).Mul(24)
		airborne.Step(Input{Boost: true}, 1./120)
		if airborne.Car.Body.VerticalSpeed >= 0 {
			t.Fatal("level air nitro introduced upward acceleration")
		}
	}
}
func TestRemastered001OnlyVisibleObjectGeometryCollides(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Remaster.Def.Objects = []PlacedObject{{Asset: "house", Solid: true}}
	w.Remaster.Def.Colliders = []ColliderShape{{Center: Vec{5, 0}, Half: Vec{1, 3}, MinY: 0, MaxY: 9, Object: 0}}
	w.Remaster.buildSpatial(w)
	w.Car.Pos = Vec{2, 0}
	w.Car.Heading = 0
	before := w.Car.Pos
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Car.Pos != before {
		t.Fatal("ghost collision outside visible wall footprint")
	}
	w.Car.Pos = Vec{4.1, 0}
	w.Car.Lateral = 8
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Car.Pos.X > 3.061 {
		t.Fatal("wall penetration was not separated")
	}
	w.Car.Gearbox.Gear = 1
	w.Car.Speed = 0
	w.Car.Heading = math.Pi / 2
	for i := 0; i < 30; i++ {
		w.Step(Input{Brake: 1, Steer: .3}, 1./120)
	}
	if w.Car.Speed >= -.01 {
		t.Fatal("reverse recovery delayed after collision")
	}
}
func TestRemastered001ProfilesAreDistinctAndFourthPaintFixed(t *testing.T) {
	a, b, c, d := RemasteredVehicleSpec(0), RemasteredVehicleSpec(1), RemasteredVehicleSpec(2), RemasteredVehicleSpec(3)
	if a.P.Mass == b.P.Mass || a.PowerKW == c.PowerKW || d.TopSpeed >= a.TopSpeed || d.Paintable {
		t.Fatal("vehicle identities not differentiated")
	}
}
func TestRemastered001RaceCountdownAndNoCivilianTraffic(t *testing.T) {
	w := remasteredFixture(1, 0)
	if len(w.Traffic) != 0 || len(w.Remaster.Racers) != 3 {
		t.Fatal("race population incorrect")
	}
	for i := 0; i < 300; i++ {
		w.Step(Input{Throttle: 1, Boost: true}, 1./120)
	}
	if w.Remaster.Started || math.Abs(w.Car.Speed) > .1 {
		t.Fatal("player accelerated before start")
	}
	for i := 0; i < 125; i++ {
		w.Step(Input{}, 1./120)
	}
	if !w.Remaster.Started {
		t.Fatal("countdown did not finish")
	}
}
func TestRemastered001LandingCannotEnterGround(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Car.Body.Height = 25
	w.Car.Body.VerticalSpeed = -18
	w.Car.Body.Grounded = false
	for i := 0; i < 500; i++ {
		w.Step(Input{}, 1./120)
		if w.Car.Body.Height < w.GroundHeight(w.Car.Pos)+.539 {
			t.Fatal("car entered terrain")
		}
		if math.IsNaN(w.Car.Body.Height) || math.IsNaN(w.Car.Speed) {
			t.Fatal("nonfinite chassis")
		}
	}
}

func TestRemastered001PlayerLapsRequireSequentialGates(t *testing.T) {
	w := remasteredFixture(1, 0)
	s := w.Remaster
	s.Started = true
	s.Countdown = 0
	visit := func(gate int) {
		_, f := w.Routes[0].AtDistance(float64(gate) * w.Routes[0].Len / float64(len(s.Gates)))
		w.Car.Pos = s.Gates[gate].Add(f.Mul(.5))
		s.RaceTime += 1
		w.updatePlayerRace(1. / 120)
	}
	visit(3)
	if s.NextGate != 0 || s.Lap != 1 {
		t.Fatal("skipped checkpoint accepted")
	}
	visit(0)
	for lap := 0; lap < s.Laps; lap++ {
		for i := 1; i < len(s.Gates); i++ {
			visit(i)
		}
		visit(0)
	}
	if !s.Finished || !w.Finished || s.BestLap <= 0 || s.FinishTime <= 0 {
		t.Fatal("complete laps did not finish race")
	}
}

func TestRemastered001PoleFallUsesWorldTerrainAndSettles(t *testing.T) {
	w := remasteredFixture(0, 0)
	pole := &Pole{Pos: Vec{10, 10}}
	pole.Hit(Vec{12, 0})
	w.Remaster.FallenPoles[0] = pole
	for i := 0; i < 120*12; i++ {
		w.stepRemasteredPoles(1. / 120)
	}
	if !pole.Settled || pole.Angle < 1.3 || pole.Angle > math.Pi*.5+.01 {
		t.Fatal("pole failed to settle on visible terrain")
	}
}
