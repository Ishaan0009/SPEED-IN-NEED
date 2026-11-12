//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

// Soft volume impostors preserve the supplied cloud mesh lobes while avoiding
// its flat scan triangles. Fourteen layers have no per-frame buffer uploads.
func drawCloudMeshes035(camera pos3) {
	if app.quality == 0 {
		return
	}
	useLighting(false)
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	dis(0x0B44)
	en(GL_TEXTURE_2D)
	en(GL_BLEND)
	gl("glDepthMask", 0)
	glExt("glActiveTexture", 0x84C0)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture("b035_cloud_volume.png")))
	night := engine.NightAmount(sceneHour())
	sunset := engine.Clamp((sceneHour()-15.5)/3., 0, 1) * (1 - night)
	color := pos3{.94 - .76*night, .96 - .76*night - .14*sunset, .99 - .74*night - .26*sunset}
	if hdrScene.active {
		color = pos3{math.Pow(color.x, 2.2), math.Pow(color.y, 2.2), math.Pow(color.z, 2.2)}
	}
	windTime := 0.0
	if beta.cfg.Wind {
		windTime = app.t
	}
	for i := 0; i < 7; i++ {
		angle := float64(i)*2.399963 + windTime*.0005
		radius := 950. + float64(i%3)*160.
		delta := pos3{math.Sin(angle) * radius, 290. + float64(i%4)*48., math.Cos(angle) * radius}
		direction := norm(delta)
		right := norm(cross(direction, pos3{0, 1, 0}))
		up := norm(cross(right, direction))
		for layer := 0; layer < 2; layer++ {
			center := add(camera, delta)
			width := 240. + float64(i%3)*40.
			height := width * .39
			alpha := .68
			if layer == 0 {
				width *= 1.15
				height *= .72
				center = add(center, scale(up, -height*.10))
				alpha = .24
			}
			col(color.x, color.y, color.z, alpha)
			begin(GL_QUADS)
			for _, q := range [][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}} {
				p := add(center, add(scale(right, (q[0]-.5)*width), scale(up, (.5-q[1])*height)))
				tex(q[0], q[1])
				vtx(p.x, p.y, p.z)
			}
			end()
			gpuWorld.draws++
		}
	}
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
}
