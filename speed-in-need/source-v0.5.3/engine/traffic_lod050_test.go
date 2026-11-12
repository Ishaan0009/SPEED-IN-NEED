package engine

import (
	"math"
	"testing"
)

func TestTrafficLOD050AccountsForTimeAndAllPlayers(t *testing.T) {
	w := &World{SimulationLOD050: true, Remaster: &RemasteredState{}}
	car := Traffic{Pos: Vec{X: 1000}}
	total := 0.
	updates := 0
	for i := 1; i <= 120; i++ {
		w.Remaster.TrafficTick020 = uint64(i)
		step := w.trafficStep050(&car, 0, 1./120, w.observerDistanceSquared050(car.Pos))
		total += step
		if step > 0 {
			updates++
		}
	}
	if math.Abs(total-1) > 1e-9 || updates != 10 {
		t.Fatalf("far traffic lost simulation time: %f / %d", total, updates)
	}
	w.NetworkObservers050 = []Vec{{X: 1000}}
	for i := 0; i < 120; i++ {
		w.Remaster.TrafficTick020++
		if w.trafficStep050(&car, 0, 1./120, w.observerDistanceSquared050(car.Pos)) != 1./120 {
			t.Fatal("traffic near a guest was throttled")
		}
	}
	w.NetworkObservers050 = nil
	car.ImpactCooldown = 1
	if w.trafficStep050(&car, 0, 1./120, 1000000) != 1./120 {
		t.Fatal("impact recovery was throttled")
	}
}
func TestTrafficLOD050FarVehicleContinuesAlongRoad(t *testing.T) {
	route := newRoute("Straight", []Vec{{X: 900}, {X: 2000}})
	w := &World{SimulationLOD050: true, Routes: []Route{route}, Remaster: &RemasteredState{Def: &WorldDefinition{Beta020: true, Roads: []RoadDefinition{{Width: 14}}}}, Traffic: []Traffic{{T: 100, Speed: 10, CruiseSpeed: 10, Direction: 1, Lane: 2, Pos: Vec{X: 1000, Z: 2}}}}
	for tick := 0; tick < 240; tick++ {
		w.Elapsed += 1. / 120
		w.stepRemasteredTraffic(1. / 120)
	}
	if w.Traffic[0].T < 115 || w.Traffic[0].Pos.X < 1015 {
		t.Fatalf("far vehicle froze: %+v", w.Traffic[0])
	}
}
