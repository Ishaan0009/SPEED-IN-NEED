package engine

import (
	"math"
	"testing"
)

func Test040CombinedBrakesStopEveryCar(t *testing.T) {
	for class := 0; class < 4; class++ {
		previous := 0
		for _, hand := range []bool{false, true} {
			w := flat22(class, true)
			w.Car.Speed = 120 / 3.6
			ticks := 0
			for ; ticks < 600 && math.Abs(w.Car.Speed) > .2; ticks++ {
				w.Step(Input{Brake: 1, Handbrake: hand}, 1./120)
			}
			if ticks >= 300 {
				t.Fatalf("car %d handbrake %v needed %.2fs", class, hand, float64(ticks)/120)
			}
			t.Logf("car %d handbrake %v: %.2fs", class, hand, float64(ticks)/120)
			if hand && ticks > previous+3 { t.Fatalf("combined braking slower than pedal: %d > %d", ticks, previous) }; previous=ticks
		}
	}
}

func Test040ParkAndReverseOnGrade(t *testing.T) {
	for class:=0;class<4;class++ {
		base:=flat22(class,true);d:=base.Remaster.Def
		for z:=0;z<d.Grid.H;z++ {for x:=0;x<d.Grid.W;x++ {d.Grid.Values[z*d.Grid.W+x]=float32((d.Grid.MinZ+float64(z)*d.Grid.Step)*.15)}}
		w:=NewRemasteredWorld(d,0,class,0,1,0);w.Traffic=nil;start:=w.Car.Pos
		for tick:=0;tick<1200;tick++ {w.Step(Input{},1./120)}
		if shift:=w.Car.Pos.Dist(start);shift>.35 {t.Fatalf("parked car %d crept %.3fm",class,shift)}
		for tick:=0;tick<480;tick++ {w.Step(Input{Brake:1},1./120)}
		if w.Car.Speed>-.5 || w.Car.Pos.Z>start.Z-1 {t.Fatalf("car %d cannot reverse on grade: speed %.2f displacement %.2f",class,w.Car.Speed,w.Car.Pos.Z-start.Z)}
	}
}

func Test040RaceBoundaryLeavesFreeRoamAlone(t *testing.T) {
	for _, mode := range []int{0, 1} {
		w := remasteredFixture(mode, 0)
		p, f := w.Routes[0].AtDistance(30)
		right := Vec{f.Z, -f.X}
		w.Car.Pos = p.Add(right.Mul(18))
		before := w.Car.Pos
		w.constrainRaceCorridor040(&w.Car, 0)
		if mode == 0 && w.Car.Pos.Dist(before) > .001 {
			t.Fatal("free roam constrained")
		}
		if mode == 1 && w.Car.Pos.Dist(p) > 7 {
			t.Fatalf("race escaped: %.2fm", w.Car.Pos.Dist(p))
		}
	}
}

func Test040OpenBuildingCourtAndSolidWall(t *testing.T) {
	w := remasteredFixture(0, 0)
	d := w.Remaster.Def
	mask := CollisionFootprint040{MinX: -10, MinZ: -10, Step: 1, W: 20, H: 20, Cells: make([]uint8, 400)}
	for z := 0; z < 20; z++ {
		mask.Cells[z*20] = 1
		mask.Cells[z*20+19] = 1
	}
	d.Objects = []PlacedObject{{Asset: "court", Kind: "building", Scale: [3]float64{1, 1, 1}}}
	d.CollisionFootprints040 = map[string]CollisionFootprint040{"court": mask}
	shape := ColliderShape{Kind: "building", Object: 0}
	c := Car{Pos: Vec{}}
	if _, _, hit, detailed := w.buildingContact040(&c, RemasteredVehicleSpec(0), shape); hit || !detailed {
		t.Fatal("open court blocked")
	}
	c.Pos.X = 9
	if _, _, hit, _ := w.buildingContact040(&c, RemasteredVehicleSpec(0), shape); !hit {
		t.Fatal("wall pass-through")
	}
}

func Test040DifficultyPaceOrdered(t *testing.T) {
	a, at := RacePace040(0)
	b, bt := RacePace040(1)
	c, ct := RacePace040(2)
	if !(a < b && b < c && at < bt && bt < ct) {
		t.Fatal("difficulty pace is not ordered")
	}
}

func Test040SpawnRejectsWaterAndPreservesRaceGrid(t *testing.T) {
	w := remasteredFixture(0, 0)
	d := w.Remaster.Def
	d.Colliders = []ColliderShape{{Center: Vec{40, 40}, Radius: 8, MinY: -1, MaxY: 10}}
	p, _, ok := SetupSpawnPoint(d, Vec{40, 40}, 0)
	if !ok || p.Dist(Vec{40, 40}) < 10 {
		t.Fatalf("unsafe requested spawn: %+v %v", p, ok)
	}
	race := remasteredFixture(1, 0)
	before := race.Car.Pos
	if race.ApplySetupSpawn(Vec{100, 100}) || race.Car.Pos != before {
		t.Fatal("setup changed race grid")
	}
}
