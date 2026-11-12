//go:build windows

package main

import (
	"math"
	"runtime"
	"unsafe"
)

var barrier020LastCount = -1
var barrier020Scratch [][8]float32

// Keep the original placements immutable so a reset can restore the full batch.
func syncBeta020Barriers() {
	state := app.world.Remaster
	if state == nil || barrier020LastCount == len(state.DynamicProps)+len(state.FallenPoles) {
		return
	}
	barrier020LastCount = len(state.DynamicProps) + len(state.FallenPoles)
	for i := range remaster.Instances {
		g := &remaster.Instances[i]
		if len(g.ObjectIDs) == 0 {
			continue
		}
		barrier020Scratch = barrier020Scratch[:0]
		for j, id := range g.ObjectIDs {
			if state.DynamicProps[id] == nil && state.FallenPoles[id] == nil {
				barrier020Scratch = append(barrier020Scratch, g.Data[j])
			}
		}
		if g.Count == len(barrier020Scratch) {
			continue
		}
		g.Count = len(barrier020Scratch)
		if g.Count > 0 {
			glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
			glExt("glBufferData", glArrayBuffer, uintptr(g.Count*32), uintptr(unsafe.Pointer(&barrier020Scratch[0])), 0x88E8)
			runtime.KeepAlive(barrier020Scratch)
		}
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
}

func drawBeta020MovedBarriers(eye, target pos3, shadow bool, radius float64) {
	state := app.world.Remaster
	if state == nil {
		return
	}
	limit := 340 * viewDistance026()
	if shadow {
		limit = math.Min(limit, radius)
	}
	sceneInt("uFlowWater", 0)
	for id, p := range state.DynamicProps {
		if id < 0 || id >= len(state.Def.Objects) {
			continue
		}
		o := state.Def.Objects[id]
		if o.Radius > 80 || (o.Kind != "barrier" && o.Kind != "prop") || !remasterVisible(p.Pos.X, p.Pos.Z, o.Radius, limit, eye, target, shadow) {
			continue
		}
		push()
		move(p.Pos.X, p.Height, p.Pos.Z)
		rotate(p.Heading*180/math.Pi, 0, 1, 0)
		rotate(p.Roll*180/math.Pi, 0, 0, 1)
		resize(o.Scale[0], o.Scale[1], o.Scale[2])
		drawPBRAsset(o.Asset, 0)
		pop()
	}
	for id, p := range state.FallenPoles {
		if id < 0 || id >= len(state.Def.Objects) {
			continue
		}
		o := state.Def.Objects[id]
		if o.Kind == "signal" || o.Radius > 80 || !remasterVisible(o.Pos[0], o.Pos[2], o.Radius, limit, eye, target, shadow) {
			continue
		}
		push()
		move(o.Pos[0], o.Pos[1], o.Pos[2])
		rotate(p.Angle*180/math.Pi, p.FallDirection.Z, 0, -p.FallDirection.X)
		rotate(o.Heading*180/math.Pi, 0, 1, 0)
		resize(o.Scale[0], o.Scale[1], o.Scale[2])
		drawPBRAsset(o.Asset, 0)
		pop()
	}

}
