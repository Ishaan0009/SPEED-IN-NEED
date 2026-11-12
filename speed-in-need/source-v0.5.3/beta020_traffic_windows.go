package main

import (
	"math"
	"sort"
	"unsafe"
)

var traffic020 struct{ Indices []int }

// The population is a simulation setting. GPU submissions have an independent
// near-camera budget, including a smaller shadow budget.
func beta020TrafficSelection(shadow bool, radius float64) []int {
	traffic020.Indices = traffic020.Indices[:0]
	var view [16]float32
	gl("glGetFloatv", 0x0ba6, uintptr(unsafe.Pointer(&view[0])))
	aspect := float64(app.width) / float64(app.height)
	for i, t := range app.world.Traffic {
		delta := t.Pos.Sub(app.renderCar.Pos)
		if delta.X*delta.X+delta.Z*delta.Z > radius*radius {
			continue
		}
		if !shadow {
			y := app.world.GroundHeight(t.Pos) + 1.0
			ex := float64(view[0])*t.Pos.X + float64(view[4])*y + float64(view[8])*t.Pos.Z + float64(view[12])
			ez := float64(view[2])*t.Pos.X + float64(view[6])*y + float64(view[10])*t.Pos.Z + float64(view[14])
			if ez > 8 || math.Abs(ex) > -ez*.70*aspect+8 {
				continue
			}
		}
		traffic020.Indices = append(traffic020.Indices, i)
	}
	sort.Slice(traffic020.Indices, func(i, j int) bool {
		a := app.world.Traffic[traffic020.Indices[i]].Pos.Sub(app.renderCar.Pos)
		b := app.world.Traffic[traffic020.Indices[j]].Pos.Sub(app.renderCar.Pos)
		return a.X*a.X+a.Z*a.Z < b.X*b.X+b.Z*b.Z
	})
	limit := []int{20, 32, 48, 64}[app.quality]
	limit = max(8, int(float64(limit)*foliageDistance026()))
	if shadow {
		limit = limit / 2
	}
	if !shadow {
		near := 0
		for _, i := range traffic020.Indices {
			d := app.world.Traffic[i].Pos.Sub(app.renderCar.Pos)
			if d.X*d.X+d.Z*d.Z < 45*45 {
				near++
			}
		}
		if near > limit {
			limit = near
		}
	}
	if len(traffic020.Indices) > limit {
		traffic020.Indices = traffic020.Indices[:limit]
	}
	return traffic020.Indices
}
