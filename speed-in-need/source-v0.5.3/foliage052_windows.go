package main

import (
	"math"
	"unsafe"
)

// Preserve the atlas' cutout coverage while downsampling. Ordinary alpha
// averaging made distant grass disappear and reappear as the camera advanced.
func coverageMips052(pixels []byte, w, h int) {
	source := append([]byte(nil), pixels...)
	covered := 0
	for i := 3; i < len(source); i += 4 {
		if source[i] > 64 {
			covered++
		}
	}
	target := float64(covered) / float64(w*h)
	for level := 1; w > 1 || h > 1; level++ {
		nw, nh := max(1, w/2), max(1, h/2)
		out := make([]byte, nw*nh*4)
		for y := 0; y < nh; y++ {
			for x := 0; x < nw; x++ {
				for k := 0; k < 4; k++ {
					value := 0
					for j := 0; j < 2; j++ {
						for i := 0; i < 2; i++ {
							value += int(source[(min(h-1, y*2+j)*w+min(w-1, x*2+i))*4+k])
						}
					}
					out[(y*nw+x)*4+k] = byte(value / 4)
				}
			}
		}
		low, high := .1, 8.
		for j := 0; j < 12; j++ {
			scale := (low + high) * .5
			count := 0
			for i := 3; i < len(out); i += 4 {
				if float64(out[i])*scale > 64 {
					count++
				}
			}
			if float64(count)/float64(nw*nh) < target {
				low = scale
			} else {
				high = scale
			}
		}
		for i := 3; i < len(out); i += 4 {
			out[i] = byte(math.Min(255, float64(out[i])*(low+high)*.5))
		}
		gl("glTexImage2D", GL_TEXTURE_2D, uintptr(level), GL_RGBA, uintptr(nw), uintptr(nh), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&out[0])))
		source = out
		w, h = nw, nh
	}
}
