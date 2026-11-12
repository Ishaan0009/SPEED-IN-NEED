package main

import "math"

// RGBE has only 256 exponents; do the exact conversion once, rather than once
// per pixel of a panorama containing up to 134 million pixels.
var rgbeScale050 = func() (table [256]float32) {
	for e := 1; e < 256; e++ {
		table[e] = float32(math.Ldexp(1, e-136))
	}
	return
}()

func skyViewportTarget050(quality, setting int, vram uint64, maxTexture, viewportWidth int) int {
	if workshop051.Active || weather050.Rain || garage033 {
		return 2048
	}
	target := skyTargetSize040(quality, setting, vram, maxTexture)
	if setting != 0 {
		return target
	} // Keep explicit 2K/4K/8K/16K choices.
	if quality == 0 {
		return min(target, 2048)
	}
	// Desktop Auto uses 4K, over 11 panorama pixels/degree, to reduce residency and upload stalls.
	// Reserve the 16K automatic tier for a 4K output; it remains manually selectable.
	if viewportWidth > 0 && viewportWidth <= 2560 {
		return min(target, 4096)
	}
	return target
}
