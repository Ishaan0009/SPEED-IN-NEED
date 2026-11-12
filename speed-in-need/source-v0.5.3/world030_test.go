package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"sync"
	"testing"
)

var pbrTest030 struct {
	sync.Mutex
	Cache map[string]pbrAsset
}

func readTestPBR030(t *testing.T, key string) pbrAsset {
	t.Helper()
	pbrTest030.Lock()
	defer pbrTest030.Unlock()
	if pbrTest030.Cache == nil {
		pbrTest030.Cache = map[string]pbrAsset{}
	}
	if m, ok := pbrTest030.Cache[key]; ok {
		return m
	}
	f, err := assets.Open("assets/pbr/" + key + ".pbr")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var header [8]byte
	if _, err = io.ReadFull(f, header[:]); err != nil || string(header[:4]) != "PBR1" {
		t.Fatal("invalid PBR header", key)
	}
	n := binary.LittleEndian.Uint32(header[4:])
	if n < 2 || n > 4<<20 {
		t.Fatal("invalid PBR metadata", key)
	}
	raw := make([]byte, n)
	if _, err = io.ReadFull(f, raw); err != nil {
		t.Fatal(err)
	}
	var result pbrAsset
	if err = json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	pbrTest030.Cache[key] = result
	return result
}
func TestOneIsland030ReferenceAndLongCentralTunnel(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	if d.Name != "SPEED IN NEED" || len(d.Roads) < 150 || len(d.Tunnels) != 1 {
		t.Fatal("REMASTERED layout incomplete")
	}
	tunnel := d.Tunnels[0]
	if tunnel.Length < 490 || tunnel.Length > 520 || tunnel.Width < 12 {
		t.Fatal("central long tunnel dimensions incorrect")
	}
	if tunnel.Center.Z < -750 || tunnel.Center.Z > -600 || tunnel.Center.X < 40 || tunnel.Center.X > 110 {
		t.Fatal("tunnel was left in the old northern location")
	}
	for _, o := range d.Objects {
		if o.Solid && (o.Collider < 0 || o.Collider >= len(d.Colliders)) {
			t.Fatal("unpaired visible collider", o.Asset)
		}
	}
	if len(d.BridgeSpans) < 5 || len(d.Signals) < 6 {
		t.Fatal("bridges/signals incomplete")
	}
}
