package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var unityHost040 struct {
	Parent       uintptr
	FirstFrame   bool
	InputFocused bool
}

// The Unity project hosts this exact renderer/simulation in a native child
// window. No second physics implementation or copied scene is substituted.
func configureUnityHost040() {
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--unity-parent=") {
			value, err := strconv.ParseUint(strings.TrimPrefix(arg, "--unity-parent="), 10, 64)
			if err == nil && value != 0 && win(user32, "IsWindow", uintptr(value)) != 0 {
				unityHost040.Parent = uintptr(value)
				beta.cfg.Fullscreen = false
			}
		}
	}
}

func unityStage040(stage string) {
	if unityHost040.Parent == 0 {
		return
	}
	exe, e := os.Executable()
	if e != nil {
		return
	}
	folder := filepath.Join(filepath.Dir(exe), "Logs")
	os.MkdirAll(folder, 0755)
	f, e := os.OpenFile(filepath.Join(folder, "UNITY_NATIVE040.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e == nil {
		f.WriteString(time.Now().UTC().Format(time.RFC3339Nano) + " " + stage + "\n")
		f.Close()
	}
}

func gameForeground040() bool {
	foreground := win(user32, "GetForegroundWindow")
	if foreground == app.hwnd {
		return true
	}
	return unityHost040.Parent != 0 && foreground == win(user32, "GetAncestor", app.hwnd, 2) && (unityHost040.InputFocused || win(user32, "GetFocus") == app.hwnd)
}

func unityFirstFrame040() {
	if unityHost040.Parent != 0 && !unityHost040.FirstFrame {
		unityHost040.FirstFrame = true
		unityStage040(fmt.Sprintf("first frame %dx%d visible=%d", app.width, app.height, win(user32, "IsWindowVisible", app.hwnd)))
	}
}

// Unity posts requests, leaving both GUI threads free throughout startup.
func unityHostMessage040(hwnd uintptr, message uint32, w, l uintptr) bool {
	if unityHost040.Parent == 0 {
		return false
	}
	switch message {
	case 0x802a:
		unityHost040.InputFocused = w != 0
		if w == 0 && beta.screen == screenDrive && !qa.enabled {
			selectScreen(screenPause)
		}
		return true
	case 0x802b:
		if unityHost040.InputFocused && beta.screen == screenDrive && !remaster.FullMap {
			beta020Look(float64(int32(w)), float64(int32(l)))
		}
		return true
	case 0x8028:
		x, y := int32(int16(w&65535)), int32(int16((w>>16)&65535))
		width, height := uint32(l&65535), uint32((l>>16)&65535)
		if width >= 320 && height >= 200 {
			win(user32, "SetWindowPos", hwnd, 0, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0x14)
		}
		return true
	case 0x8029:
		show := uintptr(0)
		if w != 0 {
			show = 4
		}
		win(user32, "ShowWindow", hwnd, show)
		return true
	}
	return false
}
