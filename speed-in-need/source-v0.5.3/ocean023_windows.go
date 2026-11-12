//go:build windows

package main

func drawOcean023() {
	if app.mapID != 0 || app.world.Remaster == nil {
		return
	}
	d := app.world.Remaster.Def
	g := d.Grid
	loX, loZ := g.MinX, g.MinZ
	hiX := loX + float64(g.W-1)*g.Step
	hiZ := loZ + float64(g.H-1)*g.Step
	extent := 50000.0
	y := d.WaterHeight
	resetPBRMaterial()
	setPBRMaterial(true, 0, .26)
	sceneInt("uFlowWater", 1)
	sceneFloat("uFlowTime", app.world.Elapsed)
	dis(GL_TEXTURE_2D)
	col(1, 1, 1, 1)
	nrm(0, 1, 0)
	for _, r := range [][4]float64{{loX - extent, loZ - extent, loX, hiZ + extent}, {hiX, loZ - extent, hiX + extent, hiZ + extent}, {loX, loZ - extent, hiX, loZ}, {loX, hiZ, hiX, hiZ + extent}} {
		begin(GL_QUADS)
		for _, p := range [][2]float64{{r[0], r[1]}, {r[0], r[3]}, {r[2], r[3]}, {r[2], r[1]}} {
			tex(0, 0)
			vtx(p[0], y, p[1])
		}
		end()
	}
	sceneInt("uFlowWater", 0)
	resetPBRMaterial()
}
