//go:build windows

package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"
)

func limitTexture023(im *image.NRGBA) *image.NRGBA {
	cap := textureCap024()
	for im.Bounds().Dx() > cap || im.Bounds().Dy() > cap {
		w, h := im.Bounds().Dx(), im.Bounds().Dy()
		nw, nh := (w+1)/2, (h+1)/2
		small := image.NewNRGBA(image.Rect(0, 0, nw, nh))
		for y := 0; y < nh; y++ {
			for x := 0; x < nw; x++ {
				for c := 0; c < 4; c++ {
					sum, n := 0, 0
					for yy := y * 2; yy < int(math.Min(float64(h), float64(y*2+2))); yy++ {
						for xx := x * 2; xx < int(math.Min(float64(w), float64(x*2+2))); xx++ {
							sum += int(im.Pix[yy*im.Stride+xx*4+c])
							n++
						}
					}
					small.Pix[y*small.Stride+x*4+c] = uint8(sum / n)
				}
			}
		}
		im = small
	}
	return im
}

func crashReport023() {
	if failure := recover(); failure != nil {
		if qa.enabled {
			os.WriteFile(filepath.Join(qa.dir,"CRASH040.log"),append([]byte(fmt.Sprintln(failure)),debug.Stack()...),0644)
			os.Exit(1)
		}
		folder := filepath.Join(os.Getenv("LOCALAPPDATA"), "SpeedInNeedBeta026", "Diagnostics")
		os.MkdirAll(folder, 0755)
		path := filepath.Join(folder, "crash-"+time.Now().Format("20060102-150405")+".log")
		os.WriteFile(path, append([]byte(fmt.Sprintln(failure)), debug.Stack()...), 0644)
		message023("The game stopped. Diagnostic saved to:\n" + path)
	}
}
func message023(text string) {
	body, _ := syscall.UTF16PtrFromString(text)
	title, _ := syscall.UTF16PtrFromString("SPEED IN NEED v0.5.3")
	win(user32, "MessageBoxW", 0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), 0x10)
}

type memorySample023 struct{ WorkingSetMB, PeakWorkingSetMB, PrivateMB, GoHeapMB float64 }

func memory023() memorySample023 {
	var c struct {
		Size, Faults                                                                          uint32
		PeakWS, WS, PeakPaged, Paged, PeakNonPaged, NonPaged, Pagefile, PeakPagefile, Private uintptr
	}
	c.Size = uint32(unsafe.Sizeof(c))
	win(kernel32, "K32GetProcessMemoryInfo", ^uintptr(0), uintptr(unsafe.Pointer(&c)), uintptr(c.Size))
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return memorySample023{float64(c.WS) / (1 << 20), float64(c.PeakWS) / (1 << 20), float64(c.Private) / (1 << 20), float64(stats.HeapAlloc) / (1 << 20)}
}
