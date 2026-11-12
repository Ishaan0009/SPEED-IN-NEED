//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

func debugCircle(p engine.Vec, radius, y float64) {
	begin(GL_LINE_STRIP)
	for i := 0; i <= 24; i++ {
		angle := float64(i) * 2 * math.Pi / 24
		vtx(p.X+math.Cos(angle)*radius, y, p.Z+math.Sin(angle)*radius)
	}
	end()
}
func drawRemasterDebug(c engine.Car) {
	if !app.debugCollision && !app.debugRoutes && !app.debugWheels {
		return
	}
	dis(GL_TEXTURE_2D)
	if app.debugCollision {
		col(.95, .43, .12, 1)
		for _, i := range app.world.NearbyTreeIndices(c.Pos, 90) {
			t := app.world.Trees[i]
			if c.Pos.Dist(t.Pos) < 85 {
				debugCircle(t.Pos, t.Size*.42+1.18, engine.Terrain(t.Pos.X, t.Pos.Z)+.30)
			}
		}
		col(.98, .18, .35, 1)
		for _, i := range app.world.NearbyRockIndices(c.Pos, 90) {
			r := app.world.Rocks[i]
			if c.Pos.Dist(r.Pos) < 85 {
				debugCircle(r.Pos, r.Radius+1.05, engine.Terrain(r.Pos.X, r.Pos.Z)+.30)
			}
		}
		col(.95, .85, .10, 1)
		for _, p := range app.world.Poles {
			if c.Pos.Dist(p.Pos) < 85 && !p.Falling {
				debugCircle(p.Pos, 1.75, engine.Terrain(p.Pos.X, p.Pos.Z)+.28)
			}
		}
	}
	if app.debugRoutes {
		col(.18, .99, .77, 1)
		for _, route := range app.world.Routes {
			begin(GL_LINE_STRIP)
			for _, p := range route.Points {
				vtx(p.X, engine.Terrain(p.X, p.Z)+.52, p.Z)
			}
			end()
		}
		col(.95, .77, .08, 1)
		for _, npc := range app.world.Traffic {
			if c.Pos.Dist(npc.Pos) < 120 {
				debugCircle(npc.Pos, 1.5, engine.Terrain(npc.Pos.X, npc.Pos.Z)+.60)
			}
		}
	}
	if app.debugWheels {
		params := engine.VehicleSpec(app.world.CarClass)
		f := engine.HeadingVec(c.Heading)
		right := engine.Vec{X: f.Z, Z: -f.X}
		for i, wheel := range c.Rig.Wheels {
			front, side := -1., -1.
			if i < 2 {
				front = 1
			}
			if i%2 == 1 {
				side = 1
			}
			pt := c.Pos.Add(f.Mul(params.Wheelbase * .5 * front)).Add(right.Mul(params.Track * .5 * side))
			if wheel.Touching {
				col(.2, 1, .3, 1)
			} else {
				col(1, .2, .2, 1)
			}
			debugCircle(pt, .24, wheel.GroundHeight+.20)
			begin(GL_LINES)
			vtx(pt.X, wheel.GroundHeight+.15, pt.Z)
			vtx(pt.X, wheel.GroundHeight+.15+wheel.Compression*2.5, pt.Z)
			end()
		}
	}
}
