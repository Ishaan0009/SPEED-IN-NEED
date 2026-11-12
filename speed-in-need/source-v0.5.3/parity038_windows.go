package main

import (
	"encoding/json"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"unsafe"
)

var parity038 struct {
	Captured  int
	Reference []byte
	Name      string
}

func prepareParity038() bool {
	if outputFBO038 == 0 || !qa.enabled || os.Getenv("SPEED_QA_DX12_PARITY") != "1" || len(qa.results) <= parity038.Captured {
		return false
	}
	parity038.Captured = len(qa.results)
	parity038.Name = qa.results[len(qa.results)-1].Name
	parity038.Reference = make([]byte, app.width*app.height*4)
	gl("glReadPixels", 0, 0, uintptr(app.width), uintptr(app.height), GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&parity038.Reference[0])))
	nativeCall038("SIN038_CaptureNext")
	return true
}
func finishParity038() {
	if experimentFSR038() {
		pixels := make([]byte, app.width*app.height*4)
		if nativeCall038("SIN038_CopyGenerated", uintptr(unsafe.Pointer(&pixels[0])), uintptr(len(pixels))) == uintptr(len(pixels)) {
			img := image.NewRGBA(image.Rect(0, 0, app.width, app.height))
			copy(img.Pix, pixels)
			if f, e := os.Create(filepath.Join(qa.dir, parity038.Name+"_fsr_generated038.png")); e == nil {
				png.Encode(f, img)
				f.Close()
			}
		}
	}
	pixels := make([]byte, len(parity038.Reference))
	n := nativeCall038("SIN038_CopyCapture", uintptr(unsafe.Pointer(&pixels[0])), uintptr(len(pixels)))
	if n != uintptr(len(pixels)) {
		return
	}
	ref := image.NewRGBA(image.Rect(0, 0, app.width, app.height))
	native := image.NewRGBA(ref.Bounds())
	copy(native.Pix, pixels)
	for y := 0; y < app.height; y++ {
		copy(ref.Pix[y*ref.Stride:(y+1)*ref.Stride], parity038.Reference[(app.height-1-y)*ref.Stride:(app.height-y)*ref.Stride])
	}
	sum, maximum, changed := 0., 0., 0
	for i, a := range ref.Pix {
		if i%4 == 3 {
			continue
		}
		d := math.Abs(float64(a) - float64(native.Pix[i]))
		sum += d
		maximum = math.Max(maximum, d)
		if d > 2 {
			changed++
		}
	}
	mean := sum / float64(app.width*app.height*3)
	status := "PASS"
	if mean > .2 || changed > app.width*app.height/1000 {
		status = "FAIL"
	}
	for suffix, img := range map[string]*image.RGBA{"gl_reference": ref, "dx12_output": native} {
		if f, err := os.Create(filepath.Join(qa.dir, parity038.Name+"_"+suffix+".png")); err == nil {
			png.Encode(f, img)
			f.Close()
		}
	}
	report := map[string]any{"status": status, "scene": parity038.Name, "width": app.width, "height": app.height, "mean_absolute_error_rgb": mean, "maximum_channel_error": maximum, "channels_difference_gt_2": changed, "comparison": "Same-frame GL framebuffer versus actual D3D12 backbuffer before Present; QA-only GPU readback."}
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(filepath.Join(qa.dir, parity038.Name+"_parity038.json"), raw, 0644)
}
