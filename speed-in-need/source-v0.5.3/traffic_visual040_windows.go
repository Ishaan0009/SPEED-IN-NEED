//go:build windows

package main

import (
	"fmt"
	"math"
	"runtime"
	"speedinneed/engine"
	"strings"
	"unsafe"
)

// Render-only filtering: position/heading remain the interpolated simulation
// values. Small steering and suspension discontinuities are softened without
// moving a vehicle away from its physical collision pose.
type trafficVisualState040 struct {
	Variant            int
	Pos                engine.Vec
	Steer, Pitch, Roll float64
}

var trafficVisual040 struct {
	Definition *engine.WorldDefinition
	States     map[int]trafficVisualState040
	Drivers    int
}

func beginTrafficVisual040(shadow bool) {
	if shadow || app.world == nil || app.world.Remaster == nil {
		return
	}
	if trafficVisual040.Definition != app.world.Remaster.Def || trafficVisual040.States == nil {
		trafficVisual040.Definition = app.world.Remaster.Def
		trafficVisual040.States = make(map[int]trafficVisualState040)
	}
	trafficVisual040.Drivers = 0
}

func smoothTrafficVisual040(index int, t engine.Traffic, dt float64) engine.Traffic {
	if trafficVisual040.States == nil {
		trafficVisual040.States = make(map[int]trafficVisualState040)
	}
	state, exists := trafficVisual040.States[index]
	reset := !exists || state.Variant != t.Variant || state.Pos.Dist(t.Pos) > 8 || t.Impulse.Length() > 3
	if reset {
		state = trafficVisualState040{t.Variant, t.Pos, t.SteerAngle, t.Pitch, t.Roll}
	} else {
		dt = engine.Clamp(dt, .001, .05)
		steer := 1 - math.Exp(-dt/.045)
		body := 1 - math.Exp(-dt/.075)
		state.Steer += (t.SteerAngle - state.Steer) * steer
		state.Pitch += (t.Pitch - state.Pitch) * body
		state.Roll += (t.Roll - state.Roll) * body
		// Maximum permitted visual lag: 3.4 degrees at the wheel, 1.4 degrees
		// at the chassis. This never smooths a collision or teleport position.
		state.Steer = engine.Clamp(state.Steer, t.SteerAngle-.06, t.SteerAngle+.06)
		state.Pitch = engine.Clamp(state.Pitch, t.Pitch-.025, t.Pitch+.025)
		state.Roll = engine.Clamp(state.Roll, t.Roll-.025, t.Roll+.025)
	}
	state.Pos = t.Pos
	trafficVisual040.States[index] = state
	t.SteerAngle, t.Pitch, t.Roll = state.Steer, state.Pitch, state.Roll
	return t
}

func trafficAssetKey040(variant int) string {
	if variant >= 4 {
		return fmt.Sprintf("b011_npc%d", (variant-4)%10)
	}
	return fmt.Sprintf("rm_npc%d", engine.ClampInt(variant, 0, 3))
}

// These imported batches were named Glass but marked glass:false. Correct the
// known civilian assets only, so windows get dielectric response and do not
// write opaque depth over the seated driver. Baked opaque rm_npc windows need
// an asset-level separation; they are intentionally not guessed here.
func prepareTrafficGlass040(key string) {
	if !strings.HasPrefix(key, "b011_npc") {
		return
	}
	a := pbrAssets[key]
	if a == nil {
		return
	}
	for i := range a.Materials {
		m := &a.Materials[i]
		if !strings.HasPrefix(strings.ToLower(m.Name), "glass") {
			continue
		}
		m.Glass, m.Double, m.Alpha = true, true, "BLEND"
		m.Metal, m.Rough, m.Coat = 0, .14, .18
		m.Base = [4]float64{.78, .88, .94, .22}
	}
}

type trafficCabin040 struct {
	GlassTop, GlassFront, HalfWidth float64
	SeatX                           float64
}

// Measured local glass/outer-body envelopes from b011_npc0..9. The supplied
// civilian meshes contain no steering or seat-side metadata. +X follows the
// existing game's left-seat convention; it is declared, not inferred by name.
var trafficCabins040 = [...]trafficCabin040{
	{1.693, .986, 1.093, .31},
	{1.320, .355, .985, .32},
	{1.450, .844, .978, .33},
	{1.589, 1.142, .902, .31},
	{1.700, .315, 1.099, .35},
	{1.397, .825, .898, .29},
	{1.332, .612, 1.002, .33},
	{1.061, .927, 1.012, .32},
	{1.578, .669, 1.005, .32},
	{1.330, .666, 1.001, .33},
}

func trafficDriverPose040(variant int, steer float64) ([engine.PedJointCount]engine.PedVec, bool) {
	if variant < 4 {
		return [engine.PedJointCount]engine.PedVec{}, false
	}
	c := trafficCabins040[(variant-4)%len(trafficCabins040)]
	pose := seatedDriverPose032(1, 0)
	delta := engine.PedVec{X: c.SeatX - .34, Y: c.GlassTop - .36 - 1.04, Z: c.GlassFront - .30 + .10}
	for i := range pose {
		pose[i] = pose[i].Add(delta)
	}
	return fitDriverArms035(pose, steer), true
}

type trafficDriverMesh040 struct {
	Body, Hands characterBatch032
	Pose        [engine.PedJointCount]engine.PedVec
	Bounds      [2]engine.PedVec
	Ready, Safe bool
}

// At most 17 small steering poses per civilian body; all use the provided
// driver's lowest-detail mesh. Each unique pose is skinned/uploaded once.
var trafficDriverMeshes040 = make(map[[2]int]*trafficDriverMesh040)

func trafficSteerBin040(steer float64) (int, float64) {
	i := engine.ClampInt(int(math.Round(engine.Clamp(steer, -.4, .4)/.05))+8, 0, 16)
	return i, float64(i-8) * .05
}

func buildTrafficDriverMesh040(variant int, steer float64) *trafficDriverMesh040 {
	pose, ok := trafficDriverPose040(variant, steer)
	result := &trafficDriverMesh040{Ready: true, Pose: pose}
	a := characters032.Models[3]
	if !ok || a == nil || len(a.LODs) < 3 {
		return result
	}
	mesh, matrices := a.LODs[2], driverSkinBasis035(pose)
	lo := engine.PedVec{X: math.Inf(1), Y: math.Inf(1), Z: math.Inf(1)}
	hi := engine.PedVec{X: math.Inf(-1), Y: math.Inf(-1), Z: math.Inf(-1)}
	for _, v := range mesh.Vertices {
		weight := float64(v.Weight)
		p := skinPoint032(v.A, matrices[v.BoneA]).Mul(weight).Add(skinPoint032(v.B, matrices[v.BoneB]).Mul(1 - weight))
		n := skinNormal032(v.NA, matrices[v.BoneA]).Mul(weight).Add(skinNormal032(v.NB, matrices[v.BoneB]).Mul(1 - weight)).Unit()
		lo = engine.PedVec{X: math.Min(lo.X, p.X), Y: math.Min(lo.Y, p.Y), Z: math.Min(lo.Z, p.Z)}
		hi = engine.PedVec{X: math.Max(hi.X, p.X), Y: math.Max(hi.Y, p.Y), Z: math.Max(hi.Z, p.Z)}
		result.Body.Vertices = append(result.Body.Vertices, [12]float32{float32(p.X), float32(p.Y), float32(p.Z), float32(n.X), float32(n.Y), float32(n.Z), v.UV[0], v.UV[1], 1, 0, 0, 1})
	}
	for at := 0; at < len(mesh.Indices); at += 3 {
		if !driverHandTriangle035(mesh, at) {
			result.Body.Indices = append(result.Body.Indices, mesh.Indices[at:at+3]...)
		}
	}
	// Closed gloves replace the source model's stretched open fingers. Bake the
	// small geometry once; never call the player's mutable hand/motion cache.
	previous := pedestrianMesh.vertices
	pedestrianMesh.vertices = nil
	glove := [4]uint8{255, 255, 255, 255}
	for _, wrist := range []int{5, 8} {
		pedEllipsoid(pose[wrist], engine.PedVec{X: .041, Y: .038, Z: .033}, glove, 6, 4)
		pedLimb(pose[wrist].Add(engine.PedVec{Y: -.012, Z: -.010}), pose[wrist].Add(engine.PedVec{Y: -.028, Z: .018}), .014, glove, 5)
	}
	for _, v := range pedestrianMesh.vertices {
		result.Hands.Vertices = append(result.Hands.Vertices, [12]float32{v.P[0], v.P[1], v.P[2], v.N[0], v.N[1], v.N[2], 0, 0, 1, 0, 0, 1})
		result.Hands.Indices = append(result.Hands.Indices, uint32(len(result.Hands.Indices)))
	}
	pedestrianMesh.vertices = previous
	result.Bounds = [2]engine.PedVec{lo, hi}
	c := trafficCabins040[(variant-4)%len(trafficCabins040)]
	result.Safe = lo.X > -c.HalfWidth+.04 && hi.X < c.HalfWidth-.04 && hi.Y < c.GlassTop+.035 && hi.Z < c.GlassFront+.55
	return result
}

func uploadTrafficDriverBatch040(b *characterBatch032) {
	if b.VBO != 0 || len(b.Vertices) == 0 || len(b.Indices) == 0 {
		return
	}
	glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.VBO)))
	glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.EBO)))
	glExt("glBindVertexArray", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(b.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(b.Vertices))*48, uintptr(unsafe.Pointer(&b.Vertices[0])), glStaticDraw)
	glExt("glBindBuffer", 0x8893, uintptr(b.EBO))
	glExt("glBufferData", 0x8893, uintptr(len(b.Indices))*4, uintptr(unsafe.Pointer(&b.Indices[0])), glStaticDraw)
	runtime.KeepAlive(b.Vertices)
	runtime.KeepAlive(b.Indices)
}

func drawTrafficDriverBatch040(b *characterBatch032, material pbrMaterial, key string) {
	uploadTrafficDriverBatch040(b)
	if b.VBO == 0 {
		return
	}
	sceneInt("uInstanced", 0)
	sceneInt("uGrassWind", 0)
	sceneInt("uConformGround", 0)
	glExt("glBindVertexArray", 0)
	for _, attribute := range []uintptr{4, 5, 6} {
		glExt("glDisableVertexAttribArray", attribute)
	}
	objectMotion038(key, 0)
	defer sceneInt("uObjectMotion038", 0)
	bindPBRMaterial(material)
	pbrPointers(pbrPart{vbo: b.VBO, ebo: b.EBO})
	gl("glDrawElements", GL_TRIANGLES, uintptr(len(b.Indices)), 0x1405, 0)
	gpuWorld.draws++
	unbindWorld()
	resetPBRMaterial()
}

// Call inside the NPC's existing local transform, BEFORE drawBeta026NPC(t),
// while temporal038.Entity still identifies that NPC. No driver is added to a
// baked opaque-window model. The body/glass pass then occludes the seated mesh.
func drawTrafficDriver040(t engine.Traffic, distance float64, shadow bool) bool {
	key := trafficAssetKey040(t.Variant)
	prepareTrafficGlass040(key)
	if shadow || sunShadow.drawing || t.Variant < 4 || characters032.Models[3] == nil || !lightPipeline.active {
		return false
	}
	limit, count := 38., 10
	if app.quality == 0 {
		limit, count = 22, 4
	}
	if distance > limit || trafficVisual040.Drivers >= count {
		return false
	}
	bin, steer := trafficSteerBin040(t.SteerAngle)
	cacheKey := [2]int{(t.Variant - 4) % 10, bin}
	mesh := trafficDriverMeshes040[cacheKey]
	if mesh == nil {
		mesh = buildTrafficDriverMesh040(t.Variant, steer)
		trafficDriverMeshes040[cacheKey] = mesh
	}
	if !mesh.Safe {
		return false
	}
	trafficVisual040.Drivers++
	name := fmt.Sprintf("traffic-driver040/%d/%d", cacheKey[0], bin)
	a := characters032.Models[3]
	drawTrafficDriverBatch040(&mesh.Body, pbrMaterial{Name: "Seated traffic driver", Albedo: a.Albedo, Base: [4]float64{1, 1, 1, 1}, Rough: .90, Double: true, NormalScale: 1}, name)
	drawTrafficDriverBatch040(&mesh.Hands, pbrMaterial{Name: "Traffic driving gloves", Base: [4]float64{.085, .10, .12, 1}, Rough: .93, Double: true, NormalScale: 1}, name+"/hands")
	// Reuse the supplied steering assembly. The civilian meshes themselves do
	// not include a wheel, so this is explicitly a shared cockpit accessory.
	if pbrAssets["b033_cabin_1"] != nil {
		center := mesh.Pose[5].Add(mesh.Pose[8]).Mul(.5)
		push()
		move(center.X, center.Y, center.Z)
		rotate(steeringWheelAngle035(steer)*180/math.Pi, 0, .25, 1)
		move(-.314, -.820, -.368)
		drawPBRAsset("b033_cabin_1", 9)
		pop()
	}
	return true
}

// Invoke before destroying the GL context; independent from player/pedestrian
// buffers. The pose cache remains tiny and is not rebuilt on every map switch.
func releaseTrafficVisual040() {
	for _, mesh := range trafficDriverMeshes040 {
		for _, batch := range []*characterBatch032{&mesh.Body, &mesh.Hands} {
			if batch.VBO != 0 {
				glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&batch.VBO)))
			}
			if batch.EBO != 0 {
				glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&batch.EBO)))
			}
		}
	}
	trafficDriverMeshes040 = make(map[[2]int]*trafficDriverMesh040)
	trafficVisual040.States = nil
}
