package main

import (
	"math"
	"speedinneed/engine"
)

// Start at the requested 45% closer distance, then fit the complete model into
// the clear part of the screen. This also handles a long car seen broadside.
func beta013PreviewCamera(angle, pitch, aspect float64) (pos3, pos3) {
	bounds := [2][3]float64{{-1.05, 0, -2.4}, {1.05, 1.65, 2.4}}
	key := modelKey050(engine.Car{}, beta.carSelect, true)
	if a := pbrAssets[key]; a != nil {
		bounds = a.Bounds
	}
	tan := math.Tan(44 * math.Pi / 360)
	maxWidth, centerX, maxHeight, centerY := 1.25, .30, 1.05, -.04
	left, top := -.30, .60
	if beta.screen == screenTitle {
		maxWidth, centerX, maxHeight, centerY = 1.26, .32, 1.12, -.08
		left, top = -.30, .55
	}
	distance := 6.5 / 1.45
	var eye, target pos3
	for step := 0; step < 100; step++ {
		eye = pos3{math.Sin(angle) * distance, 1.8 + pitch*2.4, math.Cos(angle) * distance}
		target = pos3{0, .65, 0}
		for align := 0; align < 8; align++ {
			f := norm(sub(target, eye))
			r := norm(cross(f, pos3{0, 1, 0}))
			up := norm(cross(r, f))
			loX, loY, hiX, hiY := 1e6, 1e6, -1e6, -1e6
			depthMean := 0.
			for _, x := range []float64{bounds[0][0], bounds[1][0]} {
				for _, y := range []float64{bounds[0][1], bounds[1][1]} {
					for _, z := range []float64{bounds[0][2], bounds[1][2]} {
						q := sub(pos3{x, y, z}, eye)
						depth := q.x*f.x + q.y*f.y + q.z*f.z
						depth = math.Max(.1, depth)
						depthMean += depth / 8
						px := (q.x*r.x + q.y*r.y + q.z*r.z) / (depth * tan * aspect)
						py := (q.x*up.x + q.y*up.y + q.z*up.z) / (depth * tan)
						loX = math.Min(loX, px)
						hiX = math.Max(hiX, px)
						loY = math.Min(loY, py)
						hiY = math.Max(hiY, py)
					}
				}
			}
			if align == 7 {
				if hiX-loX <= maxWidth && hiY-loY <= maxHeight && loX >= left+.005 && hiX <= .985 && loY >= -.655 && hiY <= top-.005 {
					return eye, target
				}
				break
			}
			shift := add(scale(r, ((loX+hiX)/2-centerX)*depthMean*tan*aspect), scale(up, ((loY+hiY)/2-centerY)*depthMean*tan))
			eye = add(eye, shift)
			target = add(target, shift)
		}
		distance += .10
	}
	return eye, target
}

// A soft contact patch makes the studio's ground plane visible under the tires.
func beta013PreviewContactShadow() {
	useLighting(false)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	begin(GL_TRIANGLES)
	for i := 0; i < 48; i++ {
		a, b := float64(i)*2*math.Pi/48, float64(i+1)*2*math.Pi/48
		col(.005, .008, .012, .65)
		vtx(0, .002, 0)
		col(.005, .008, .012, 0)
		vtx(math.Cos(a)*1.45, .002, math.Sin(a)*2.6)
		vtx(math.Cos(b)*1.45, .002, math.Sin(b)*2.6)
	}
	end()
	gl("glDepthMask", 1)
	useLighting(true)
}
