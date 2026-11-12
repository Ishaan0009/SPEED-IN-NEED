//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

func drawAokawaReverseLenses(c engine.Car, id int) {
	if c.Gearbox.Gear >= 0 && c.Speed >= -.1 {
		return
	}
	anchors := [][3]float64{{.67, .787, -2.020}, {.54, .50, -2.161}, {.64, .52, -2.190}, {.48, .62, -2.065}}
	p := anchors[engine.ClampInt(id, 0, 3)]
	push()
	move(c.Pos.X, c.Body.Height-.74, c.Pos.Z)
	rotate(c.Heading*180/math.Pi, 0, 1, 0)
	rotate(c.Rig.Pitch*180/math.Pi, 1, 0, 0)
	rotate(c.Rig.Roll*180/math.Pi, 0, 0, 1)
	dis(GL_TEXTURE_2D)
	col(1, .97, .88, 1)
	sceneVector("uEmission", pos3{2, 1.9, 1.6})
	nrm(0, 0, -1)
	for _, sign := range []float64{-1, 1} {
		x := p[0] * sign
		begin(GL_QUADS)
		vtx(x-.022, p[1]-.012, p[2])
		vtx(x+.022, p[1]-.012, p[2])
		vtx(x+.022, p[1]+.012, p[2])
		vtx(x-.022, p[1]+.012, p[2])
		end()
	}
	sceneVector("uEmission", pos3{})
	pop()
}
