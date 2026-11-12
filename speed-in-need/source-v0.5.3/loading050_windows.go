package main

import (
	"fmt"
	"speedinneed/engine"
)

// The value is supplied by real loading stages, not an elapsed-time animation.
func drawLoading036(progress float64, label string) {
	drawMenuBackdrop036()
	w, h := float64(app.width), float64(app.height)
	p := engine.Clamp(progress, 0, 1)
	remasterText("SPEED IN NEED / 0.5.3", w*.065, h*.09, h*.016, white)
	remasterText("PREPARING YOUR DRIVE", w*.16, h*.40, h*.030, white)
	headline035(label, w*.16, h*.47, h*.013, w*.68, cyan)
	hudRect(w*.16, h*.54, w*.68, 4, .15, .20, .23, 1)
	hudRect(w*.16, h*.54, w*.68*p, 4, cyan[0], cyan[1], cyan[2], 1)
	remasterText(fmt.Sprintf("%.0f%%", p*100), w*.16, h*.58, h*.018, white)
	headline035("READING ASSETS / PREPARING WORLD / UPLOADING GRAPHICS", w*.16, h*.66, h*.010, w*.68, mutedRefine)
}
