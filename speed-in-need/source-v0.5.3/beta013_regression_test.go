package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta013BridgesSupportTheFullCarriageway(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, b := range d.BridgeSpans {
		if b.Width < 18 {
			t.Fatal("undersized carriageway")
		}
		f := b.End.Sub(b.Start).Unit()
		r := engine.Vec{X: f.Z, Z: -f.X}
		for _, u := range []float64{.05, .5, .95} {
			for _, side := range []float64{-8.5, 0, 8.5} {
				p := b.Start.Add(b.End.Sub(b.Start).Mul(u)).Add(r.Mul(side))
				if math.Abs(w.GroundHeight(p)-b.Y) > .01 {
					t.Fatalf("bridge edge has no support at %v: ground %.3f deck %.3f", p, w.GroundHeight(p), b.Y)
				}
			}
		}
	}
}

func TestBeta013MouseSettingsCannotChangeKeyboardSteering(t *testing.T) {
	for _, oldMouse := range []bool{false, true} {
		for _, invert := range []bool{false, true} {
			app.boolMouse = oldMouse
			app.invertMouseX = invert
			if aokawaKeyboardSteer(true, false) != 1 || aokawaKeyboardSteer(false, true) != -1 || aokawaKeyboardSteer(true, true) != 0 {
				t.Fatal("keyboard handedness changed")
			}
		}
	}
	app.boolMouse = false
	app.invertMouseX = false
	if beta013MiniSpan() < 300 || beta013MiniSpan() > 700 {
		t.Fatal("minimap is not a local driving crop")
	}
}

func TestBeta013PreviewFitsWholeCarThroughoutOrbit(t *testing.T) {
	old := beta.screen
	defer func() { beta.screen = old }()
	for _, screen := range []int{screenTitle, screenMaps} {
		beta.screen = screen
		for _, aspect := range []float64{16. / 9, 4. / 3} {
			for angle := 0.; angle < engine.Tau; angle += .31 {
				for _, pitch := range []float64{-.25, 0, .5} {
					eye, target := beta013PreviewCamera(angle, pitch, aspect)
					f := norm(sub(target, eye))
					r := norm(cross(f, pos3{0, 1, 0}))
					up := norm(cross(r, f))
					tan := math.Tan(44 * math.Pi / 360)
					for _, x := range []float64{-1.05, 1.05} {
						for _, y := range []float64{0, 1.65} {
							for _, z := range []float64{-2.4, 2.4} {
								q := sub(pos3{x, y, z}, eye)
								depth := dot(q, f)
								px := dot(q, r) / (depth * tan * aspect)
								py := dot(q, up) / (depth * tan)
								left, top := -.86, .55
								if screen == screenTitle {
									left, top = -.85, .50
								}
								if depth <= 0 || px < left || px > .99 || py < -.66 || py > top {
									t.Fatalf("preview clips: screen=%d aspect=%.2f angle=%.2f x=%.2f y=%.2f", screen, aspect, angle, px, py)
								}
							}
						}
					}
				}
			}
		}
	}
}
