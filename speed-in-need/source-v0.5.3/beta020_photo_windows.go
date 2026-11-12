package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"
	"unsafe"
)

var photo020Pending bool

// Read the rendered back buffer before presentation. Borderless fullscreen and
// relative mouse input do not depend on desktop capture or Print Screen.
func beta020SavePhoto() {
	if !photo020Pending || app.width <= 0 || app.height <= 0 {
		return
	}
	photo020Pending = false
	w, h := app.width, app.height
	pixels := make([]byte, w*h*4)
	gl("glPixelStorei", 0x0d05, 1)
	gl("glReadPixels", 0, 0, uintptr(w), uintptr(h), 0x1908, 0x1401, uintptr(unsafe.Pointer(&pixels[0])))
	frame := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		copy(frame.Pix[y*frame.Stride:y*frame.Stride+w*4], pixels[(h-y-1)*w*4:(h-y)*w*4])
	}
	exe, err := os.Executable()
	if err != nil {
		remaster.Status = "PHOTO SAVE FAILED"
		return
	}
	folder := filepath.Join(filepath.Dir(exe), "Screenshots")
	if err = os.MkdirAll(folder, 0755); err != nil {
		remaster.Status = "PHOTO FOLDER NOT WRITABLE"
		return
	}
	file, err := os.Create(filepath.Join(folder, "SpeedInNeed-"+time.Now().Format("20060102-150405.000")+".png"))
	if err != nil {
		remaster.Status = "PHOTO SAVE FAILED"
		return
	}
	err = png.Encode(file, frame)
	closeErr := file.Close()
	if err == nil && closeErr == nil {
		remaster.Status = "PHOTO SAVED TO SCREENSHOTS"
	} else {
		remaster.Status = "PHOTO SAVE FAILED"
	}
}
