package main

import (
	"math"
	"os"
	"testing"
)

func Test053EveryDriverFitsItsOwnCabin(t *testing.T) {
	raw, e := assets.ReadFile("assets/characters/driver.chr")
	if e != nil {
		t.Fatal(e)
	}
	a, e := decodeCharacter032(raw)
	if e != nil {
		t.Fatal(e)
	}
	for _, variant := range playerVariants052 {
		cabin := cabinProfile053(3, variant)
		for _, steer := range []float64{-.7, 0, .7} {
			pose := seatedDriverPose053(3, variant, steer)
			basis := driverSkinBasis035(pose)
			high := math.Inf(-1)
			for _, v := range a.LODs[0].Vertices {
				w := float64(v.Weight)
				p := skinPoint032(v.A, basis[v.BoneA]).Mul(w).Add(skinPoint032(v.B, basis[v.BoneB]).Mul(1 - w))
				high = math.Max(high, p.Y)
			}
			if high > cabin.Roof-.015 {
				t.Errorf("model %d head/body crosses roof: top %.3f roof %.3f", variant, high, cabin.Roof)
			}
		}
	}
}

func Test053OfflineAuditDoesNotLoadNativeSDK(t *testing.T) {
	old := os.Args
	os.Args = []string{"game", "--offline-audit"}
	defer func() { os.Args = old }()
	if native038 != nil {
		t.Fatal("native SDK already loaded")
	}
	queryGPU038()
	if native038 != nil || gpu038.DLSS != -1 || gpu038.DX12 {
		t.Fatal("offline audit loaded a vendor runtime")
	}
}
