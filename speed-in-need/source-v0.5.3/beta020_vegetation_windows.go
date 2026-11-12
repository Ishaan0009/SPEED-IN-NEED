package main

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"speedinneed/engine"
	"strings"
	"unsafe"
)

// Compact cell streams upload once. Grass has no per-instance CPU physics or
// JSON objects. Trunks and substantial rocks retain independent visible hulls.
func loadBeta020Vegetation(d *engine.WorldDefinition) {
	if d.VegetationFile == "" {
		return
	}
	raw, err := assets.ReadFile("assets/" + d.VegetationFile)
	if err != nil || len(raw) < 12 || string(raw[:4]) != "VEG2" {
		pbrErrors = append(pbrErrors, "Invalid vegetation stream")
		return
	}
	groups := int(binary.LittleEndian.Uint32(raw[4:8]))
	total := int(binary.LittleEndian.Uint32(raw[8:12]))
	off, loaded := 12, 0
	if groups > 20000 || total > 5000000 {
		pbrErrors = append(pbrErrors, "Vegetation stream exceeds budget")
		return
	}
	for group := 0; group < groups; group++ {
		if off+2 > len(raw) {
			break
		}
		n := int(binary.LittleEndian.Uint16(raw[off:]))
		off += 2
		if n > 64 || off+n+12 > len(raw) {
			break
		}
		key := string(raw[off : off+n])
		off += n
		x := int(binary.LittleEndian.Uint32(raw[off:]))
		z := int(binary.LittleEndian.Uint32(raw[off+4:]))
		count := int(binary.LittleEndian.Uint32(raw[off+8:]))
		off += 12
		if count < 1 || count > 100000 || off+count*32 > len(raw) {
			break
		}
		if pbrAssets[key] == nil {
			loadPBRAsset(key)
		}
		if pbrAssets[key+"_lod"] == nil {
			loadPBRAsset(key + "_lod")
		}
		// The stream is already the validated little-endian float32 layout
		// used by the Windows GL instance buffer; no second array is needed.
		g := remasterInstance{Key: key, Kind: "grass", X: float64(x*48 - 1500 + 24), Z: float64(z*48 - 1500 + 24), Radius: 35, Count: count, Fade: 1}
		if strings.HasPrefix(key, "b026_shrub") {
			g.Kind = "bush"
		}
		if g.Kind == "grass" {
			remaster.Instances = append(remaster.Instances, splitGrass050(raw[off:off+count*32], count, g)...)
		} else {
			glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&g.VBO)))
			glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
			glExt("glBufferData", glArrayBuffer, uintptr(count*32), uintptr(unsafe.Pointer(&raw[off])), glStaticDraw)
			remaster.Instances = append(remaster.Instances, g)
		}
		off += count * 32
		loaded += count
	}
	if loaded != total || off != len(raw) {
		pbrErrors = append(pbrErrors, fmt.Sprintf("Vegetation stream incomplete: %d/%d", loaded, total))
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
	runtime.KeepAlive(raw)
}
