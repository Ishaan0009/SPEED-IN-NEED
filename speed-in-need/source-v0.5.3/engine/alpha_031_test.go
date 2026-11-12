package engine

import (
	"math"
	"testing"
)

func setupImpact(variant int) *World {
	w := NewWorld(0)
	w.Trees = nil
	w.VisibleTrees = 0
	w.Buildings = nil
	w.Poles = nil
	w.Pickups = nil
	t := w.Traffic[0]
	t.Variant = variant
	t.Direction = 1
	t.Lane = 3.1
	p, d := w.Routes[t.Path].AtDistance(t.T)
	t.Heading = math.Atan2(d.X, d.Z)
	t.Pos = p.Add(Vec{-d.Z, d.X}.Mul(t.Lane))
	t.CruiseSpeed = 9
	t.Speed = 6
	w.Traffic = []Traffic{t}
	w.Car = Car{Pos: t.Pos.Sub(d.Mul(2.9)), Heading: t.Heading, Speed: 28, Nitro: 100}
	return w
}

func TestAlpha031CarAndTruckImpact(t *testing.T) {
	car := setupImpact(0)
	truck := setupImpact(3)
	before := car.Car.Speed
	car.Step(Input{}, 1.0/120)
	truck.Step(Input{}, 1.0/120)
	carKick := car.Traffic[0].Impulse.Length()
	truckKick := truck.Traffic[0].Impulse.Length()
	if !(carKick > 0.05 && truckKick > 0.005 && truckKick < carKick*.4) {
		t.Fatalf("impact impulse missing or truck not heavier: car %g truck %g", carKick, truckKick)
	}
	if car.Car.Speed >= before {
		t.Fatalf("player did not lose impact momentum: %g", car.Car.Speed)
	}
}

func TestAlpha031TrafficReturnsToLane(t *testing.T) {
	w := setupImpact(0)
	w.Step(Input{}, 1./120)
	w.Car.Pos = Vec{X: 1000, Z: 1000}
	kick := w.Traffic[0].Impulse.Length()
	for i := 0; i < 600; i++ {
		w.Step(Input{}, 1./120)
	}
	tcar := w.Traffic[0]
	if tcar.Impulse.Length() >= kick*.2 {
		t.Fatalf("traffic did not settle after impact: %g vs %g", tcar.Impulse.Length(), kick)
	}
	if tcar.Offset.Length() > 1 {
		t.Fatalf("traffic failed to recover its lane: %+v", tcar.Offset)
	}
}

func TestAlpha031TireForceSaturates(t *testing.T) {
	if TireLateralForce(0, 30, 4) != 0 {
		t.Fatal("zero slip should have zero lateral force")
	}
	if TireLateralForce(.7, 30, 4) > 0 {
		t.Fatal("positive slip must create opposite lateral force")
	}
	if math.Abs(TireLateralForce(3, 50, 4)) > 4.00001 {
		t.Fatal("tire friction limit exceeded")
	}
}

func TestAlpha031ReverseIsUsefulAndBounded(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Buildings = nil
	w.Poles = nil
	for i := 0; i < 360; i++ {
		w.Step(Input{Brake: 1}, 1./120)
	}
	if w.Car.Speed > -4 || w.Car.Speed < -13.01 {
		t.Fatalf("unusable reverse speed %g", w.Car.Speed)
	}
}
