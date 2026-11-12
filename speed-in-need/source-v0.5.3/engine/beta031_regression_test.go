package engine

import (
	"math"
	"testing"
)

func TestBeta031BrakingRespondsAndStopsAllCars(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := flat22(class, true)
		w.Car.Speed = 100 / 3.6
		w.Car.Gearbox.Gear = 3
		before := w.Car.Speed
		w.Step(Input{Brake: 1}, 1./120)
		if a := (before - w.Car.Speed) * 120; a < 5 {
			t.Fatalf("car %d pedal response %.2f m/s2", class, a)
		}
		if w.Car.RearLocked || w.Car.DriftActive || w.Car.Rig.Wheels[0].AngularVelocity == 0 || w.Car.Rig.Wheels[2].AngularVelocity == 0 {
			t.Fatalf("car %d service pedal caused an unrequested axle lock", class)
		}
		distance := before / 120
		tick := 1
		for ; tick < 600 && w.Car.Speed > .2; tick++ {
			old := w.Car.Pos
			w.Step(Input{Brake: 1}, 1./120)
			distance += old.Dist(w.Car.Pos)
		}
		t.Logf("car %d stops from 100 km/h in %.2fs / %.2fm", class, float64(tick)/120, distance)
		if distance > 65 || tick > 540 {
			t.Fatalf("car %d excessive stopping distance/time", class)
		}
		if w.Car.RearLocked || math.Abs(w.Car.Lateral) > 1 {
			t.Fatal("lock persisted at rest or straight braking slid sideways")
		}
	}
}
func TestBeta038ServiceBrakeDoesNotInitiateHandbrakeDrift(t *testing.T) {
	leftFoot := flat22(0, true)
	leftFoot.Car.Speed = 25
	leftFoot.Step(Input{Throttle: 1, Brake: 1}, 1./120)
	if leftFoot.Car.RearLocked || !leftFoot.Car.Braking {
		t.Fatal("combined pedals must service-brake without locking the rear axle")
	}
	w := flat22(0, true)
	w.Car.Speed = 32
	w.Step(Input{Brake: .6}, 1./120)
	if w.Car.RearLocked {
		t.Fatal("normal pedal locked rear")
	}
	w.Step(Input{}, 1./120)
	w.Step(Input{Brake: 1, Steer: .7}, 1./120)
	if w.Car.RearLocked || w.Car.DriftActive {
		t.Fatal("a full service pedal while cornering must not initiate handbrake drift")
	}
	for i := 0; i < 45; i++ {
		w.Step(Input{Brake: 1}, 1./120)
	}
	if w.Car.RearLocked {
		t.Fatal("sustained pedal re-armed rear lock")
	}
	w.Step(Input{}, 1./120)
	w.Step(Input{Handbrake: true}, 1./120)
	if !w.Car.RearLocked {
		t.Fatal("manual handbrake no longer locks rear")
	}
	w.Step(Input{}, 1./120)
	if w.Car.RearLocked {
		t.Fatal("released handbrake remains locked")
	}
}
func TestBeta031HighSpeedSteeringHasGripAndResponse(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, sign := range []float64{-1, 1} {
			w := flat22(class, true)
			w.Car.Speed = 40
			w.Car.Gearbox.Gear = 4
			for i := 0; i < 144; i++ {
				w.Step(Input{Throttle: .45, Steer: sign}, 1./120)
			}
			t.Logf("car %d steer %.0f yaw %.3f speed %.2f slip %.2f", class, sign, w.Car.Heading, w.Car.Speed, w.Car.DriftAngle)
			if math.Abs(w.Car.Heading) < .25 || math.Abs(w.Car.DriftAngle) > 18 || w.Car.Speed < 28 || w.Car.DriftActive {
				t.Fatal("high-speed steering response/grip failed")
			}
		}
	}
}
func TestBeta031CameraShakeIsBoundedSpeedProportional(t *testing.T) {
	for i := 0; i < 300; i++ {
		at := float64(i) * .013
		a, b, c := SpeedCameraShake(0, at, false), SpeedCameraShake(20, at, false), SpeedCameraShake(40, at, false)
		for j := 0; j < 3; j++ {
			if a[j] != 0 || math.Abs(c[j]-2*b[j]) > 1e-12 || math.Abs(c[j]) > .06 {
				t.Fatal("shake is not bounded and proportional")
			}
		}
		if SpeedCameraShake(40, at, false) != c {
			t.Fatal("shake depends on frame history")
		}
	}
}
func TestBeta031NPCImpactPreservesMomentumAndBackwardPush(t *testing.T) {
	for _, side := range []bool{false, true} {
		w := flat22(0, true)
		w.Car.Pos = Vec{}
		w.Car.Speed = 18
		w.Car.Heading = 0
		npc := Traffic{Pos: Vec{0, 3.6}, Heading: math.Pi, Speed: 6, Variant: 1}
		if side {
			w.Car.Pos = Vec{-1.5, 0}
			w.Car.Heading = math.Pi / 2
			npc = Traffic{Pos: Vec{}, Heading: 0, Variant: 1}
		}
		w.Traffic = []Traffic{npc}
		m1, m2 := RemasteredVehicleSpec(0).P.Mass, TrafficBodyMass(1)
		v1, v2 := HeadingVec(w.Car.Heading).Mul(w.Car.Speed), TrafficVelocity(npc)
		momentum := v1.Mul(m1).Add(v2.Mul(m2))
		energy := m1*v1.Length()*v1.Length() + m2*v2.Length()*v2.Length()
		w.resolveActorContacts(1. / 120)
		f := HeadingVec(w.Car.Heading)
		r := Vec{f.Z, -f.X}
		after1 := f.Mul(w.Car.Speed).Add(r.Mul(w.Car.Lateral))
		after2 := TrafficVelocity(w.Traffic[0])
		after := after1.Mul(m1).Add(after2.Mul(m2))
		if after.Sub(momentum).Length() > 1e-6 || m1*after1.Length()*after1.Length()+m2*after2.Length()*after2.Length() > energy*1.001 {
			t.Fatal("impact creates linear momentum/energy")
		}
		if w.Traffic[0].ImpactCooldown < 1 || after2.Length() < 1 {
			t.Fatal("NPC did not receive collision velocity")
		}
		if !side && after2.Z <= 0 {
			t.Fatal("head-on collision did not push NPC backwards")
		}
		for i := 0; i < 20; i++ {
			RecoverImpactedTraffic(&w.Traffic[0], 1./120)
		}
		if w.Traffic[0].Offset.Length() < .2 {
			t.Fatal("NPC snapped onto its route after impact")
		}
	}
}
func TestBeta031NPCAnticipatesCrossingWithoutStoppingForSidewalk(t *testing.T) {
	w := flat22(0, true)
	npc := Traffic{Pos: Vec{}, Heading: 0, Speed: 10, Variant: 1}
	w.Car.Pos = Vec{8, 12}
	w.Car.Heading = -math.Pi / 2
	w.Car.Speed = 8
	if w.trafficHazardSpeed(&npc, 10) >= 8 {
		t.Fatal("NPC did not anticipate player crossing")
	}
	w.Car.Pos = Vec{400, 400}
	w.Pedestrians = []Pedestrian{{Pos: Vec{5, 12}, Heading: 0, Speed: 1.2, Moving: true}}
	w.rebuildPedestrianGrid()
	if w.trafficHazardSpeed(&npc, 10) != 10 {
		t.Fatal("NPC stopped for a parallel sidewalk walker")
	}
	w.Pedestrians[0].Pos = Vec{0, 12}
	w.Pedestrians[0].Moving = false
	w.rebuildPedestrianGrid()
	if w.trafficHazardSpeed(&npc, 10) >= 8 {
		t.Fatal("NPC did not yield to pedestrian in lane")
	}
}
func sidewalkFixture031() *World {
	d := remasteredFixture(0, 0).Remaster.Def
	d.Roads = []RoadDefinition{{Points: []Vec{{-100, 0}, {100, 0}}, Width: 8, Kind: "town"}}
	d.Spawn = Vec{}
	d.SpawnHeading = math.Pi / 2
	for _, side := range []float64{-1, 1} {
		a, b := side*4.1, side*8.2
		d.Surfaces = append(d.Surfaces, SurfaceTriangle{A: [3]float64{-100, .18, a}, B: [3]float64{100, .18, a}, C: [3]float64{100, .18, b}, Support: true}, SurfaceTriangle{A: [3]float64{-100, .18, a}, B: [3]float64{100, .18, b}, C: [3]float64{-100, .18, b}, Support: true})
	}
	w := NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	w.Traffic = nil
	return w
}
func TestBeta031PedestriansWalkOnlyOnAuthoredSidewalks(t *testing.T) {
	w := sidewalkFixture031()
	if len(w.Pedestrians) < 4 {
		t.Fatalf("no sidewalk population: %d", len(w.Pedestrians))
	}
	before := w.Pedestrians[0].Pos
	for tick := 0; tick < 1200; tick++ {
		w.Elapsed += 1. / 120
		w.StepPedestrians(1. / 120)
	}
	if before.Dist(w.Pedestrians[0].Pos) < 1 {
		t.Fatal("walkers did not animate/move")
	}
	for _, p := range w.Pedestrians {
		y, paved := w.pedestrianPavement(p.Pos)
		if !paved || math.Abs(p.Height-y) > .01 || p.Ragdoll {
			t.Fatal("walker left sidewalk or lost floor")
		}
	}
}
func TestBeta031ArticulatedRagdollHasGroundAndJointConstraints(t *testing.T) {
	w := flat22(0, true)
	w.Car.Pos = Vec{-400, -400}
	p := Pedestrian{Pos: Vec{}, Height: 0, Moving: true}
	p.Pose = p.walkingPose()
	p.PreviousPose = p.Pose
	p.BeginRagdoll(PedVec{X: 9, Y: 3, Z: 2})
	for tick := 0; tick < 1200; tick++ {
		w.stepPedRagdoll(&p, 1./120)
		for _, a := range p.Particles {
			if math.IsNaN(a.Pos.X) || math.IsInf(a.Pos.Y, 0) || a.Pos.Y < a.Radius-1e-6 {
				t.Fatal("ragdoll invalid or below ground")
			}
		}
		if tick > 180 {
			for _, b := range PedBones {
				length := p.Particles[b.A].Pos.Sub(p.Particles[b.B].Pos).Length()
				rest := pedRest[b.A].Sub(pedRest[b.B]).Length()
				if math.Abs(length-rest) > .055 {
					t.Fatalf("joint stretched: %v %.3f != %.3f", b, length, rest)
				}
			}
		}
	}
	if p.Pos.Length() < 1 {
		t.Fatal("ragdoll did not receive impulse")
	}
	t.Logf("ragdoll sleeping=%v position=%v", p.Sleeping, p.Pos)
}
func TestBeta031PedestrianVehicleCollisionTransfersImpulse(t *testing.T) {
	w := flat22(0, true)
	w.Car.Pos = Vec{}
	w.Car.Speed = 15
	p := Pedestrian{Pos: Vec{0, 2.15}, Height: 0, Moving: true, Speed: 0}
	p.Pose = p.walkingPose()
	p.PreviousPose = p.Pose
	w.pedVehicleHit(&p, 1./120)
	if !p.Ragdoll || w.Car.Speed >= 15 || w.Car.Speed < 10 || p.Velocity().Z <= 1 {
		t.Fatal("vehicle/pedestrian impact did not exchange impulse")
	}
	a := PedParticle{Pos: PedVec{X: .8, Y: .8, Z: 0}, Previous: PedVec{X: .8, Y: .8, Z: 0}, InverseMass: 1. / 5, Radius: .1}
	received := Vec{}
	pedParticleVehicle(&a, 1./120, Vec{}, 0, Vec{1, 2}, .1, 1.3, Vec{X: 5}, 1300, func(v Vec) { received = received.Add(v) })
	if a.Pos.X < 1.09 || received.X >= 0 {
		t.Fatal("fallen limb passes through moving vehicle")
	}
}
