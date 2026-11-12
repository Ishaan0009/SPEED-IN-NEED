package main

import (
	"math"
	"speedinneed/engine"
	"time"
)

var beta011MapOpened time.Time
var beta011MapProgress float64

func drawBeta011TitleLogo(x, y, w, h float64) {
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex["title_beta011.png"]))
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	tex(0, 0)
	vtx(x, y, 0)
	tex(1, 0)
	vtx(x+w, y, 0)
	tex(1, 1)
	vtx(x+w, y+h, 0)
	tex(0, 1)
	vtx(x, y+h, 0)
	end()
	dis(GL_TEXTURE_2D)
}
func beta011MiniRect() (float64, float64, float64) {
	w, h := float64(app.width), float64(app.height)
	size := math.Min(246, h*.28) * math.Min(beta.cfg.UIScale, 1.25)
	return w * .035, h - size - h*.055, size
}

func beta011AtlasDomain() (float64, float64, float64, float64) {
	d := app.world.Remaster.Def
	if d.Beta020 {
		return d.Bounds[0].X, d.Bounds[0].Z, d.Bounds[1].X - d.Bounds[0].X, d.Bounds[1].Z - d.Bounds[0].Z
	}
	a := d.Atlas
	if a.Size[0] > 0 && a.Size[1] > 0 && a.MetresPerPixel > 0 {
		return -a.OriginPixel[0] * a.MetresPerPixel, -a.OriginPixel[1] * a.MetresPerPixel, float64(a.Size[0]) * a.MetresPerPixel, float64(a.Size[1]) * a.MetresPerPixel
	}
	return d.Bounds[0].X, d.Bounds[0].Z, d.Bounds[1].X - d.Bounds[0].X, d.Bounds[1].Z - d.Bounds[0].Z
}

// North-up atlas. The driving crop follows the car without rotating the map.
func beta011MapPanel(c engine.Car, x, y, size, span float64, center engine.Vec, rotation float64) {
	h := float64(app.height)
	cx, cy := x+size/2, y+size/2

	en(GL_SCISSOR_TEST)
	gl("glScissor", uintptr(math.Max(0, x)), uintptr(math.Max(0, h-y-size)), uintptr(size), uintptr(size))
	push()
	move(cx, cy, 0)
	rotate(rotation, 0, 0, 1)
	resize(size/span, size/span, 1)
	move(-center.X, -center.Z, 0)
	key := "aokawa_reference.png"
	if app.world.Remaster.Def.Beta020 && app.tex["beta020_atlas.png"] != 0 {
		key = "beta020_atlas.png"
	}
	loX, loZ, extentX, extentZ := beta011AtlasDomain()
	if app.world.Remaster.Mode == 1 {
		key = "aokawa_circuit.png"
	}
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex[key]))
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	tex(0, 0)
	vtx(loX, loZ, 0)
	tex(1, 0)
	vtx(loX+extentX, loZ, 0)
	tex(1, 1)
	vtx(loX+extentX, loZ+extentZ, 0)
	tex(0, 1)
	vtx(loX, loZ+extentZ, 0)
	end()
	dis(GL_TEXTURE_2D)
	pop()
	delta := c.Pos.Sub(center)
	a := rotation * math.Pi / 180
	px := cx + (math.Cos(a)*delta.X-math.Sin(a)*delta.Z)*size/span
	py := cy + (math.Sin(a)*delta.X+math.Cos(a)*delta.Z)*size/span
	push()
	move(px, py, 0)
	rotate(180-c.Heading*180/math.Pi+rotation, 0, 0, 1)
	col(0, 0, 0, .9)
	begin(GL_TRIANGLES)
	vtx(0, -11, 0)
	vtx(-7, 8, 0)
	vtx(7, 8, 0)
	end()
	col(1, .55, .12, 1)
	begin(GL_TRIANGLES)
	vtx(0, -8, 0)
	vtx(-4.5, 5, 0)
	vtx(4.5, 5, 0)
	end()
	pop()
	dis(GL_SCISSOR_TEST)
}
func drawRemasterMinimap(c engine.Car) {
	drawRoadMinimap033(c)
}
func drawAokawaFullMap() { drawFullMap033() }
