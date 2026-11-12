package main

import (
	"encoding/json"
	"math"
	"os"
	"speedinneed/engine"
	"testing"
)

func shippedRemasterDefinition(t *testing.T, mode int) *engine.WorldDefinition {
	t.Helper()
	raw, e := os.ReadFile([]string{"assets/remaster_world_free.json", "assets/remaster_world_race.json"}[mode])
	if e != nil {
		t.Fatal(e)
	}
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func TestShippedRemasteredMapsAndVisibleCollision(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		d := shippedRemasterDefinition(t, mode)
		w := engine.NewRemasteredWorld(d, mode, 0, 0, 2, 4)
		if mode == 0 && d.AreaKM2 < 5 {
			t.Fatal("free roam area too small")
		}
		if mode == 1 && len(w.Traffic) > 0 {
			t.Fatal("civilian traffic in circuit")
		}
		for _, o := range d.Objects {
			if o.Asset == "building5" || o.Asset == "building4" {
				t.Fatal("car or prop catalogue used as building")
			}
			if o.Solid && (o.Collider < 0 || o.Collider >= len(d.Colliders)) {
				t.Fatal("unpaired solid")
			}
		}
		for _, r := range w.Routes {
			for s := 0.; s < r.Len; s += 8 {
				p, _ := r.AtDistance(s)
				y := w.GroundHeight(p)
				if math.IsNaN(y) || math.IsInf(y, 0) {
					t.Fatal("invalid support")
				}
			}
		}
		for i := 0; i < 360; i++ {
			w.Step(engine.Input{}, 1./120)
		}
		if w.Car.Body.Height < w.GroundHeight(w.Car.Pos)+.53 {
			t.Fatal("spawn below ground")
		}
		t.Logf("map %d roads=%d objects=%d colliders=%d triangles=%d", mode, len(d.Roads), len(d.Objects), len(d.Colliders), len(d.CollisionTriangles))
	}
}
func TestShippedCircuitRacersMakeProgress(t *testing.T) {
	if testing.Short() {
		t.Skip("physical simulation")
	}
	d := shippedRemasterDefinition(t, 1)
	w := engine.NewRemasteredWorld(d, 1, 0, 0, 1, 4)
	for i := 0; i < 120*360; i++ {
		w.Step(engine.Input{Brake: 1, Handbrake: true}, 1./120)
	}
	result := map[string]any{"race_seconds": w.Remaster.RaceTime, "racers": w.Remaster.Racers}
	raw, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile("RACE_SIMULATION_BETA_001.json", raw, 0644)
	for _, r := range w.Remaster.Racers {
		t.Logf("%s class=%d progress=%.1fm lap=%d speed=%.1f damage=%.1f finished=%t", r.Name, r.Class, r.Progress, r.Lap, r.Car.Speed, r.Car.Damage, r.Finished)
		if !r.Finished {
			t.Errorf("%s did not finish circuit at %.1f m", r.Name, r.Progress)
		}
		if math.IsNaN(r.Car.Speed) || math.IsNaN(r.Car.Body.Height) {
			t.Fatal("invalid racer state")
		}
	}
}
func TestShippedTunnelAndBridgeSupport(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, tunnel := range d.Tunnels {
		previous := 0.
		for along := -tunnel.Length*.5 + 3; along < tunnel.Length*.5-3; along += 2 {
			p := tunnel.Center.Add(engine.HeadingVec(tunnel.Heading).Mul(along))
			h := w.GroundHeight(p)
			if along > -tunnel.Length*.5+4 && math.Abs(h-previous) > .5 {
				t.Fatalf("tunnel support step %.2f", h-previous)
			}
			previous = h
		}
	}
	found := false
	for _, tr := range d.Surfaces {
		if !tr.Road {
			continue
		}
		p := engine.Vec{X: (tr.A[0] + tr.B[0] + tr.C[0]) / 3, Z: (tr.A[2] + tr.B[2] + tr.C[2]) / 3}
		level, water := w.RemasteredWaterLevel(p)
		if water {
			found = true
			if w.GroundHeight(p)-level < 2 {
				t.Fatal("bridge submerged")
			}
		}
	}
	if !found {
		t.Fatal("no bridge surface crosses river")
	}
}
func TestShippedStuntRampLaunchAndLanding(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	w.Traffic = nil
	var ramp engine.PlacedObject
	found := false
	for _, o := range d.Objects {
		if o.Asset == "rm_stunt_ramp0" {
			ramp = o
			found = true
			break
		}
	}
	if !found {
		if d.Name == "SPEED IN NEED" {
			t.Skip("Historical stunt layout; REMASTERED follows the supplied roads. Stunt assets remain in source.")
		}
		t.Fatal("stunt ramp lost")
	}
	w.Car.Pos = engine.Vec{X: ramp.Pos[0], Z: ramp.Pos[2] - 35}
	w.Car.Heading = 0
	w.Car.Body = engine.ChassisState{}
	w.InitializeChassis()
	w.Car.Speed = 25
	air, landed := false, false
	for i := 0; i < 120*10; i++ {
		w.Step(engine.Input{Throttle: .8, Boost: i > 240 && i < 420}, 1./120)
		if !w.Car.Body.Grounded && w.Car.Body.Height > w.GroundHeight(w.Car.Pos)+1.8 {
			air = true
		}
		if air && w.Car.Body.Grounded {
			landed = true
		}
		if w.Car.Body.Height < w.GroundHeight(w.Car.Pos)+.689 {
			t.Fatal("landing penetrated terrain")
		}
	}
	if !air || !landed {
		t.Fatalf("ramp launch=%t landing=%t pos=%+v", air, landed, w.Car.Pos)
	}
}
func TestShippedTunnelCanBeDrivenBothWays(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	tunnel := d.Tunnels[0]
	f := engine.HeadingVec(tunnel.Heading)
	for _, direction := range []float64{-1, 1} {
		w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
		w.Traffic = nil
		w.Car.Pos = tunnel.Center.Sub(f.Mul(direction * (tunnel.Length/2 + 15)))
		w.Car.Heading = tunnel.Heading
		if direction < 0 {
			w.Car.Heading += math.Pi
		}
		w.Car.Body = engine.ChassisState{}
		w.Car.Rig = engine.SuspensionRig{}
		w.InitializeChassis()
		w.Car.Speed = 18
		reached := false
		for i := 0; i < 120*30; i++ {
			w.Step(engine.Input{Throttle: .65}, 1./120)
			delta := w.Car.Pos.Sub(tunnel.Center)
			if (delta.X*f.X+delta.Z*f.Z)*direction > tunnel.Length/2+8 {
				reached = true
				break
			}
		}
		if !reached {
			t.Fatalf("cave blocked direction %.0f pos=%+v speed=%.2f", direction, w.Car.Pos, w.Car.Speed)
		}
	}
}
