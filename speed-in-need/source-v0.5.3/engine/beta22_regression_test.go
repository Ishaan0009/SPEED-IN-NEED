package engine

import (
	"math"
	"testing"
)

func flat22(class int, road bool) *World {
	w := remasteredFixture(0, class)
	d := w.Remaster.Def
	width := 10.0
	if road {
		width = 1200
	}
	d.Roads = []RoadDefinition{{Points: []Vec{{-600, 0}, {600, 0}}, Width: width}}
	d.Spawn = Vec{}
	d.SpawnHeading = 0
	if !road {
		d.Spawn = Vec{0, 200}
	}
	w = NewRemasteredWorld(d, 0, class, 0, 1, 0)
	w.Traffic = nil
	return w
}

func TestBeta22BothSteeringDirectionsMakeCompactCircles(t *testing.T) {
	for class := 0; class < 4; class++ {
		var radii [2]float64
		for side, sign := range []float64{-1, 1} {
			w := flat22(class, true)
			w.Car.Speed = 9
			total := 0.0
			for tick := 0; tick < 1800; tick++ {
				throttle := Clamp(.11+(9-w.Car.Speed)*.14, 0, .65)
				w.Step(Input{Throttle: throttle, Steer: sign}, 1./120)
				if w.Car.DriftActive {
					t.Fatal("normal steering enabled drift")
				}
				if tick > 240 {
					total += math.Abs(w.Car.YawRate) / 120
				}
			}
			if total < math.Pi*2 {
				t.Fatalf("car %d direction %.0f failed full circle: heading %.2f", class, sign, total)
			}
			radii[side] = 9 / (total / float64(1559./120))
			if radii[side] > 12 {
				t.Fatalf("car %d excessively wide circle %.2fm", class, radii[side])
			}
		}
		if math.Abs(radii[0]-radii[1]) > .6 {
			t.Fatalf("car %d asymmetric turning radii %v", class, radii)
		}
	}
}

func TestBeta22GrassSlowsEveryCarMoreThanAsphalt(t *testing.T) {
	for class := 0; class < 4; class++ {
		road, turf := flat22(class, true), flat22(class, false)
		road.Car.Speed, turf.Car.Speed = 25, 25
		for i := 0; i < 360; i++ {
			road.Step(Input{}, 1./120)
			turf.Step(Input{}, 1./120)
		}
		if !road.Car.OnRoad || turf.Car.OnRoad {
			t.Fatal("surface fixture did not select two distinct surfaces")
		}
		if turf.Car.Speed > road.Car.Speed-.7 || turf.Car.Speed < road.Car.Speed-3 {
			t.Fatalf("class %d insufficient soil resistance: road %.2f grass %.2f", class, road.Car.Speed, turf.Car.Speed)
		}
		if turf.Car.Speed < 0 {
			t.Fatal("rolling resistance reversed the vehicle")
		}
	}
}

func TestBeta22BrokenPoleDoesNotStopCarOnFollowingTicks(t *testing.T) {
	w := flat22(0, true)
	d := w.Remaster.Def
	d.Objects = []PlacedObject{{Kind: "lamp", Pos: [3]float64{0, 0, 4}, Scale: [3]float64{1, 1, 1}}}
	d.Colliders = []ColliderShape{{Object: 0, Kind: "lamp", Center: Vec{0, 4}, Radius: .16, MinY: 0, MaxY: 9}}
	w.Remaster.staticGrid = map[gridKey][]int{gridOf(Vec{0, 4}): {0}}
	w.Car.Pos = Vec{0, 1.8}
	w.Car.Heading = 0
	w.Car.Speed = 12
	w.Car.Body.Height = .74
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Remaster.FallenPoles[0] == nil {
		t.Fatal("lamp did not break away")
	}
	after := w.Car.Speed
	for i := 0; i < 120; i++ {
		w.remasteredStaticContacts(&w.Car, 0)
		w.stepRemasteredPoles(1. / 120)
	}
	if w.Car.Speed < after-.01 || after < 9 {
		t.Fatalf("falling pole repeatedly stopped car: %.2f -> %.2f", after, w.Car.Speed)
	}
}

func TestBeta22BrakeMapZoomHasWiderSpan(t *testing.T) {
	if BrakeMapSpan(true) <= BrakeMapSpan(false) {
		t.Fatal("braking did not zoom out")
	}
}

func TestBeta22TireSupportEnvelopeIncludesRolledAndSteeredTread(t *testing.T) {
	for _, pitch := range []float64{-.6, 0, .6} {
		for _, roll := range []float64{-.19, 0, .19} {
			for _, steer := range []float64{-.6, 0, .6} {
				extent := WheelVerticalExtent(.34, .17, pitch, roll, steer)
				for j := 0; j < 90; j++ {
					for _, side := range []float64{-.17, .17} {
						a := float64(j) * Tau / 90
						x, y, z := side, .34*math.Cos(a), .34*math.Sin(a)
						x, z = math.Cos(steer)*x+math.Sin(steer)*z, -math.Sin(steer)*x+math.Cos(steer)*z
						y = math.Sin(roll)*x + math.Cos(roll)*y
						y = math.Cos(pitch)*y - math.Sin(pitch)*z
						if math.Abs(y) > extent+1e-9 {
							t.Fatalf("tread exceeds contact envelope: %.4f > %.4f", math.Abs(y), extent)
						}
					}
				}
			}
		}
	}
}
