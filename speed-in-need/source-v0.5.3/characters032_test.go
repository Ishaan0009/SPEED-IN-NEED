package main

import (
	"testing"
)

func TestCharacter032SuppliedModelsHaveBoundedLODsAndValidSkin(t *testing.T) {
	for _, name := range []string{"manuel", "nathan", "sophia", "driver"} {
		raw, err := assets.ReadFile("assets/characters/" + name + ".chr")
		if err != nil {
			t.Fatal(err)
		}
		a, err := decodeCharacter032(raw)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if a.Name != name {
			t.Fatalf("wrong model %s", a.Name)
		}
		f,err:=assets.Open("assets/pbr/"+a.Albedo)
 if err != nil {
			t.Fatal(err)
		}
        f.Close()
		if len(a.LODs[0].Indices) > 100000 || len(a.LODs[1].Indices) > 9500 || len(a.LODs[2].Indices) > 3750 {
			t.Fatalf("%s exceeds skinning budget", name)
		}
		if len(a.LODs[0].Indices) <= len(a.LODs[1].Indices) || len(a.LODs[1].Indices) <= len(a.LODs[2].Indices) {
			t.Fatalf("%s has duplicate/nondecreasing LODs", name)
		}
		for _, bad := range [][]byte{raw[:7], raw[:len(raw)-1], append(append([]byte{}, raw...), 0)} {
			if _, err := decodeCharacter032(bad); err == nil {
				t.Fatalf("%s accepted corrupt character", name)
			}
		}
	}
}
