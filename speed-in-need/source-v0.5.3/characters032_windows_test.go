package main

import (
	"speedinneed/engine"
	"testing"
)

func checkCharacterEnvelope032(t *testing.T, name string, pose [engine.PedJointCount]engine.PedVec, lo, hi engine.PedVec) {
	t.Helper()
	raw, err := assets.ReadFile("assets/characters/" + name + ".chr")
	if err != nil {
		t.Fatal(err)
	}
	a, err := decodeCharacter032(raw)
	if err != nil {
		t.Fatal(err)
	}
	basis := characterPoseBasis032(pose)
	if name == "driver" {
		basis = driverSkinBasis035(pose)
	}
	for _, lod := range a.LODs {
		for _, v := range lod.Vertices {
			w := float64(v.Weight)
			p := skinPoint032(v.A, basis[v.BoneA]).Mul(w).Add(skinPoint032(v.B, basis[v.BoneB]).Mul(1 - w))
			if p.X < lo.X || p.X > hi.X || p.Y < lo.Y || p.Y > hi.Y || p.Z < lo.Z || p.Z > hi.Z {
				t.Fatalf("%s skin escaped its physical envelope: %+v", name, p)
			}
		}
	}
}
func TestCharacter032BindCoordinatesFitHumanAndSeatedDriver(t *testing.T) {
	// Independent physical envelope: a normal 1.8m walking person, and the
	// bounded cabin/footwell of each car. Catches mixed centimetre/metre and
	// Z-up bind transforms which can otherwise pass finite-buffer validation.
	pose := [engine.PedJointCount]engine.PedVec{{Y: .91}, {Y: 1.43}, {Y: 1.66}, {X: -.23, Y: 1.48}, {X: -.29, Y: 1.18, Z: .06}, {X: -.28, Y: .90, Z: .10}, {X: .23, Y: 1.48}, {X: .29, Y: 1.18, Z: .06}, {X: .28, Y: .90, Z: .10}, {X: -.12, Y: .87}, {X: -.13, Y: .49, Z: .045}, {X: -.13, Y: .09, Z: .055}, {X: .12, Y: .87}, {X: .13, Y: .49, Z: .045}, {X: .13, Y: .09, Z: .055}}
	for _, name := range []string{"manuel", "nathan", "sophia"} {
		checkCharacterEnvelope032(t, name, pose, engine.PedVec{X: -.8, Y: -.2, Z: -.7}, engine.PedVec{X: .8, Y: 2.10, Z: .7})
	}
	for id := 0; id < 4; id++ {
		// These limits follow the measured roof underside in the supplied car
		// meshes. A generic 1.6m envelope missed visible roof penetration.
		roof := []float64{1.105, 1.285, 1.14, 1.245}[id]
		for _, steer := range []float64{-.25, 0, .25} {
			p := seatedDriverPose032(id, steer)
			x := p[0].X
			checkCharacterEnvelope032(t, "driver", p, engine.PedVec{X: x - .75, Y: -.2, Z: -1.3}, engine.PedVec{X: x + .75, Y: roof, Z: 1.05})
		}
	}
}
