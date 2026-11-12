package main

import (
	"math"
	"speedinneed/engine"
	"time"
	"unsafe"
)

var input020 struct {
	Raw, Captured   bool
	ScreenshotUntil time.Time
}

func beta020DrivingPointer() bool {
	return !multiplayer050.Open && !garage050.Open && !photo041.Active && !qa.enabled && app.started && beta.screen == screenDrive && !remaster.FullMap && !app.minimized && gameForeground040() && time.Now().After(input020.ScreenshotUntil)
}

// Raw relative movement does not warp the desktop pointer. Windows screenshot
// overlays and Alt-Tab can therefore release the pointer without a warp loop.
func beta020RegisterMouse() {
	type rawDevice struct {
		Page, Usage uint16
		Flags       uint32
		Target      uintptr
	}
	device := rawDevice{Page: 1, Usage: 2, Target: app.hwnd}
	input020.Raw = win(user32, "RegisterRawInputDevices", uintptr(unsafe.Pointer(&device)), 1, unsafe.Sizeof(device)) != 0
}

func beta020Look(dx, dy float64) {
	if beta.cfg.InvertLookYRefine {
		dy = -dy
	}
	if math.Abs(dx) > 4000 || math.Abs(dy) > 4000 {
		return
	}
	sign := -1.0
	if app.invertMouseX {
		sign = 1
	}
	app.camYaw = math.Remainder(app.camYaw+dx*.003*beta.cfg.CameraSensitivity*sign, engine.Tau)
	app.camPitch = engine.Clamp(app.camPitch+dy*.002*beta.cfg.CameraSensitivity, -.45, .65)
	beta.idleMouse = 0
}

func beta020SyncPointer() {
	if !qa.enabled && (key(0x2c) || (key(0x5b) || key(0x5c)) && key(0x10) && key('S')) {
		input020.ScreenshotUntil = time.Now().Add(1500 * time.Millisecond)
	}
	capture := beta020DrivingPointer()
	if capture {
		var rect windowRect
		win(user32, "GetClientRect", app.hwnd, uintptr(unsafe.Pointer(&rect)))
		origin := point{}
		win(user32, "ClientToScreen", app.hwnd, uintptr(unsafe.Pointer(&origin)))
		rect.left += origin.x
		rect.right += origin.x
		rect.top += origin.y
		rect.bottom += origin.y
		win(user32, "ClipCursor", uintptr(unsafe.Pointer(&rect)))
		win(user32, "SetCursor", 0)
	} else if input020.Captured {
		win(user32, "ClipCursor", 0)
		win(user32, "ReleaseCapture")
		win(user32, "SetCursor", win(user32, "LoadCursorW", 0, 32512))
		app.mouseInit = false
	}
	input020.Captured = capture
}

func beta020PointerMessage(hwnd uintptr, message uint32, w, l uintptr) bool {
	if message == 0x20 && l&65535 == 1 {
		cursor := uintptr(0)
		if !beta020DrivingPointer() {
			cursor = win(user32, "LoadCursorW", 0, 32512)
		}
		win(user32, "SetCursor", cursor)
		return true
	}
	if (message == 0x1c && w == 0) || message == 0x08 || message == WM_CLOSE || message == WM_DESTROY {
		if input020.Captured {
			win(user32, "ClipCursor", 0)
			win(user32, "ReleaseCapture")
			input020.Captured = false
		}
	}
	if message == 0xff && input020.Raw {
		if beta020DrivingPointer() {
			type rawHeader struct {
				Type, Size    uint32
				Device, Param uintptr
			}
			type rawMouse struct {
				Flags, Padding      uint16
				Buttons, RawButtons uint32
				X, Y                int32
				Extra               uint32
			}
			var buffer [64]byte
			n := uint32(len(buffer))
			head := unsafe.Sizeof(rawHeader{})
			got := win(user32, "GetRawInputData", l, 0x10000003, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&n)), head)
			if got != uintptr(^uint32(0)) && n >= uint32(head+unsafe.Sizeof(rawMouse{})) {
				h := (*rawHeader)(unsafe.Pointer(&buffer[0]))
				m := (*rawMouse)(unsafe.Pointer(&buffer[head]))
				if h.Type == 0 && m.Flags&1 == 0 {
					beta020Look(float64(m.X), float64(m.Y))
				}
			}
		}
		// DefWindowProc owns cleanup of foreground RAWINPUT handles.
		win(user32, "DefWindowProcW", hwnd, uintptr(message), w, l)
		return true
	}
	if message == WM_MOUSEMOVE && input020.Raw && beta.screen == screenDrive && !remaster.FullMap {
		return true
	}
	return false
}
