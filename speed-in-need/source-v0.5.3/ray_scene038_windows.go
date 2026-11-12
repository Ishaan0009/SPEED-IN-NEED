package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"speedinneed/engine"
	"strconv"
	"strings"
	"unsafe"
)

type rayInstance038 struct {
	Transform [12]float32
	Mesh      uint32
}

var rayScene038 struct {
	Definition               *engine.WorldDefinition
	Meshes                   map[string][]uint32
	Next                     uint32
	Static, Actors           []rayInstance038
	CellX, CellZ             int
	StaticReady, Initialized bool
	Error                    string
}

func experimentRays038() bool {
	return (qa.enabled && (os.Getenv("SPEED_QA_RAY_SCENE") == "1" || os.Getenv("SPEED_QA_RT_MASK") == "1" || os.Getenv("SPEED_QA_RT_SHADOWS") == "1" || os.Getenv("SPEED_QA_RT_REFLECTIONS") == "1") || !qa.enabled && beta.cfg.RayShadows038 >= 1 && beta.screen == screenDrive) && gpu038.RayTier >= 11 && outputFBO038 != 0
}
func rayQuality038() int {
	if qa.enabled {
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_RT_QUALITY")); e == nil && n >= 0 && n <= 3 {
			return n
		}
	}
	return beta.cfg.RayQuality038
}
func rayRadius038() int { return [4]int{130, 175, 230, 290}[rayQuality038()] }
func experimentRayMask038() bool {
	return experimentRays038() && (os.Getenv("SPEED_QA_RT_MASK") == "1" || experimentRayShadows038())
}
func experimentRayShadows038() bool {
	return experimentRays038() && (os.Getenv("SPEED_QA_RT_SHADOWS") == "1" || os.Getenv("SPEED_QA_RT_REFLECTIONS") == "1" || !qa.enabled && beta.cfg.RayShadows038 >= 1)
}
func experimentRayReflections038() bool {
	return experimentRays038() && (os.Getenv("SPEED_QA_RT_REFLECTIONS") == "1" || !qa.enabled && beta.cfg.RayShadows038 == 2)
}
func resetRayScene038() {
	if native038 != nil {
		nativeCall038("SIN038_RTDestroy")
	}
	rayScene038.Definition = nil
	rayScene038.Meshes = map[string][]uint32{}
	rayScene038.Next = 1
	rayScene038.Static = nil
	rayScene038.Actors = nil
	rayScene038.StaticReady = false
	rayScene038.Initialized = false
}
func beginRayFrame038() {
	if !experimentRays038() || app.world == nil || app.world.Remaster == nil {
		return
	}
	if rayScene038.Definition != app.world.Remaster.Def {
		resetRayScene038()
		rayScene038.Definition = app.world.Remaster.Def
	}
	rayScene038.Actors = rayScene038.Actors[:0]
}
func ensureRayUpload038() error {
	if rayScene038.Error != "" {
		return fmt.Errorf("%s", rayScene038.Error)
	}
	if int32(nativeCall038("SIN038_RTBegin")) != 0 {
		return bridgeError038()
	}
	rayScene038.Initialized = true
	return nil
}
func rayMeshes038(key string, slot int) ([]uint32, error) {
	cache := fmt.Sprintf("%s/%d", key, slot)
	if ids, ok := rayScene038.Meshes[cache]; ok {
		return ids, nil
	}
	raw, e := assets.ReadFile("assets/pbr/" + key + ".pbr")
	if e != nil {
		return nil, e
	}
	if len(raw) < 8 || string(raw[:4]) != "PBR1" {
		return nil, fmt.Errorf("invalid ray source %s", key)
	}
	size := int(binary.LittleEndian.Uint32(raw[4:8]))
	if size < 2 || size > len(raw)-8 {
		return nil, fmt.Errorf("invalid ray source header %s", key)
	}
	var meta pbrAsset
	if e = json.Unmarshal(raw[8:8+size], &meta); e != nil {
		return nil, e
	}
	off := 8 + size
	ids := []uint32{}
	for _, batch := range meta.Batches {
		vb, ib := batch.Vertices*48, batch.Indices*4
		if batch.Material < 0 || batch.Material >= len(meta.Materials) || batch.Vertices < 3 || batch.Indices < 3 || off > len(raw)-vb-ib {
			return nil, fmt.Errorf("invalid ray mesh data %s", key)
		}
		material := meta.Materials[batch.Material]
		opaque := !material.Glass && material.Alpha != "BLEND" && material.Base[3] >= .99
		// Alpha-tested foliage stays in the raster shadow path until its alpha
		// intersection resources are implemented. JPEG bark is entirely opaque.
		if material.Alpha == "MASK" && !(strings.HasSuffix(strings.ToLower(material.Albedo), ".jpg") || strings.HasSuffix(strings.ToLower(material.Albedo), ".jpeg")) {
			opaque = false
		}
		if batch.Slot != slot || !opaque {
			off += vb + ib
			continue
		}
		reliefVertices051(key, raw[off:off+vb])
		positions := make([]float32, batch.Vertices*3)
		indices := make([]uint32, batch.Indices)
		for i := 0; i < batch.Vertices; i++ {
			for k := 0; k < 3; k++ {
				positions[i*3+k] = math.Float32frombits(binary.LittleEndian.Uint32(raw[off+i*48+k*4:]))
			}
		}
		for i := range indices {
			indices[i] = binary.LittleEndian.Uint32(raw[off+vb+i*4:])
		}
		id := rayScene038.Next
		rayScene038.Next++
		if e = ensureRayUpload038(); e != nil {
			return nil, e
		}
		if int32(nativeCall038("SIN038_RTMesh", uintptr(id), uintptr(unsafe.Pointer(&positions[0])), uintptr(batch.Vertices), uintptr(unsafe.Pointer(&indices[0])), uintptr(len(indices)))) != 0 {
			return nil, bridgeError038()
		}
		ids = append(ids, id)
		off += vb + ib
	}
	rayScene038.Meshes[cache] = ids
	return ids, nil
}
func rayInstanceMatrix038(matrix [16]float32, id uint32) rayInstance038 {
	return rayInstance038{[12]float32{matrix[0], matrix[4], matrix[8], matrix[12], matrix[1], matrix[5], matrix[9], matrix[13], matrix[2], matrix[6], matrix[10], matrix[14]}, id}
}
func placementMatrix038(o engine.PlacedObject) [16]float32 {
	c, s := float32(math.Cos(o.Heading)), float32(math.Sin(o.Heading))
	x, y, z := float32(o.Scale[0]), float32(o.Scale[1]), float32(o.Scale[2])
	return [16]float32{c * x, 0, -s * x, 0, 0, y, 0, 0, s * z, 0, c * z, 0, float32(o.Pos[0]), float32(o.Pos[1]), float32(o.Pos[2]), 1}
}
func rayObjectNear038(key string, matrix [16]float32, eye engine.Vec, radius float64) bool {
	meta := pbrAssets[key]
	if meta == nil {
		return false
	}
	center := [3]float32{}
	bound := 0.
	for k := 0; k < 3; k++ {
		v := float32((meta.Bounds[0][k] + meta.Bounds[1][k]) * .5)
		for j := 0; j < 3; j++ {
			center[j] += matrix[k*4+j] * v
		}
		extent := (meta.Bounds[1][k] - meta.Bounds[0][k]) * .5
		scale := math.Sqrt(float64(matrix[k*4]*matrix[k*4] + matrix[k*4+2]*matrix[k*4+2]))
		bound += extent * scale
	}
	return math.Hypot(float64(center[0]+matrix[12])-eye.X, float64(center[2]+matrix[14])-eye.Z) < radius+bound
}
func collectStaticRays038() error {
	d := app.world.Remaster.Def
	eye := app.renderCar.Pos
	cx, cz := int(math.Floor(eye.X/64)), int(math.Floor(eye.Z/64))
	if rayScene038.StaticReady && cx == rayScene038.CellX && cz == rayScene038.CellZ {
		return nil
	}
	result := []rayInstance038{}
	appendObject := func(key string, matrix [16]float32) error {
		if !rayObjectNear038(key, matrix, eye, float64(rayRadius038())) {
			return nil
		}
		ids, e := rayMeshes038(key, 0)
		if e != nil {
			return e
		}
		for _, id := range ids {
			result = append(result, rayInstanceMatrix038(matrix, id))
		}
		return nil
	}
	identity := [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	for _, key := range d.TerrainTiles {
		if e := appendObject(key, identity); e != nil {
			return e
		}
	}
	for _, o := range d.Objects {
		switch o.Kind {
		case "grass", "bush", "water", "pickup", "backdrop", "forest", "beach":
			continue
		}
		if e := appendObject(o.Asset, placementMatrix038(o)); e != nil {
			return e
		}
	}
	if len(result) > 7800 {
		return fmt.Errorf("static ray scene exceeds the 7800-instance budget: %d", len(result))
	}
	rayScene038.Static = result
	rayScene038.CellX, rayScene038.CellZ = cx, cz
	rayScene038.StaticReady = true
	return nil
}
func recordRayActor038(key string, slot int, view [16]float32) {
	if !experimentRays038() || rayScene038.Error != "" || pbrAssets[key] == nil {
		return
	}
	ids, e := rayMeshes038(key, slot)
	if e != nil {
		rayScene038.Error = e.Error()
		return
	}
	model := matrixMultiply(inverseRigid038(temporal038.View), view)
	for _, id := range ids {
		rayScene038.Actors = append(rayScene038.Actors, rayInstanceMatrix038(model, id))
	}
}
func updateRayScene038() {
	if !experimentRays038() || rayScene038.Error != "" {
		return
	}
	if e := collectStaticRays038(); e != nil {
		rayScene038.Error = e.Error()
		return
	}
	all := append(append([]rayInstance038{}, rayScene038.Static...), rayScene038.Actors...)
	if len(all) == 0 {
		return
	}
	if len(all) > 8192 {
		rayScene038.Error = "ray scene instance budget exceeded"
		return
	}
	if int32(nativeCall038("SIN038_RTEnd")) != 0 {
		rayScene038.Error = bridgeError038().Error()
		return
	}
	if int32(nativeCall038("SIN038_RTInstances", uintptr(unsafe.Pointer(&all[0])), uintptr(len(all)))) != 0 {
		rayScene038.Error = bridgeError038().Error()
	}
	light := beta011CelestialDirection()
	values := [4]float32{float32(light.x), float32(light.y), float32(light.z), [4]float32{80, 100, 130, 155}[rayQuality038()]}
	nativeCall038("SIN038_RTQuality", uintptr([4]int{1, 2, 4, 8}[rayQuality038()]), uintptr(rayRadius038()))
	mode := boolInt(experimentRayMask038())
	if experimentRayShadows038() {
		mode = 2
	}
	if experimentRayReflections038() {
		mode = 3
	}
	if int32(nativeCall038("SIN038_RTLight", uintptr(unsafe.Pointer(&values[0])), uintptr(mode))) != 0 {
		rayScene038.Error = "Invalid ray-tracing light direction"
	}
}
func rayReport038() map[string]any {
	report := map[string]any{"enabled": experimentRays038(), "error": rayScene038.Error, "shadow_rendering_complete": false}
	if rayScene038.Initialized {
		buf := make([]byte, 4096)
		nativeCall038("SIN038_RTStats", uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		var native map[string]any
		if json.Unmarshal([]byte(strings.TrimRight(string(buf), "\x00")), &native) == nil {
			report["native"] = native
		}
	}
	return report
}
func captureRayScene038(name string) {
	if !experimentRays038() {
		return
	}
	raw, _ := json.MarshalIndent(rayReport038(), "", "  ")
	os.WriteFile(filepath.Join(qa.dir, name+"_rayscene038.json"), raw, 0644)
}
