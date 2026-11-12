//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

var photo041 struct {
	Active, Drag, PreviousFPP         bool
	Yaw, Pitch, Distance, Height, FOV float64
	X, Y                              int
}

func togglePhoto041() {
	if beta.screen != screenDrive || remaster.FullMap {
		return
	}
	photo041.Active = !photo041.Active
	photo041.Drag = false
	if photo041.Active {
		photo041.PreviousFPP = app.boolFPP
		app.boolFPP = false
		photo041.Yaw, photo041.Pitch = app.world.Car.Heading+math.Pi*.75, .18
		photo041.Distance, photo041.Height, photo041.FOV = 7.5, 1, 55
	} else {
		app.boolFPP = photo041.PreviousFPP
	}
	remaster.CameraReset = true
	invalidateTemporal038()
	beta020SyncPointer()
}
func photoKey041(k uintptr) bool {
	if !photo041.Active {
		return false
	}
	switch k {
	case VK_ESCAPE, 'C':
		togglePhoto041()
	case VK_RETURN, 0x7b:
		photo020Pending = true
	case 0x25:
		photo041.Yaw -= .08
	case 0x27:
		photo041.Yaw += .08
	case 0x26:
		photo041.Height = engine.Clamp(photo041.Height+.15, -.25, 5)
	case 0x28:
		photo041.Height = engine.Clamp(photo041.Height-.15, -.25, 5)
	case 'Q':
		photo041.FOV = engine.Clamp(photo041.FOV-2, 25, 85)
	case 'E':
		photo041.FOV = engine.Clamp(photo041.FOV+2, 25, 85)
	}
	return true
}
func photoPointer041(m uint32, w, l uintptr) bool {
	if !photo041.Active {
		return false
	}
	x, y := int(int16(l&65535)), int(int16((l>>16)&65535))
	switch m {
	case WM_LBUTTONDOWN:
		if y > int(float64(app.height)*.86) {
			if x > int(float64(app.width)*.72) {
				togglePhoto041()
			} else if x > int(float64(app.width)*.52) {
				photo020Pending = true
			}
		} else {
			photo041.Drag = true
			photo041.X, photo041.Y = x, y
		}
	case 0x202:
		photo041.Drag = false
	case WM_MOUSEMOVE:
		if photo041.Drag {
			photo041.Yaw += float64(x-photo041.X) * .006
			photo041.Pitch = engine.Clamp(photo041.Pitch+float64(y-photo041.Y)*.004, -.25, 1.15)
		}
		photo041.X, photo041.Y = x, y
	case 0x20a:
		photo041.Distance = engine.Clamp(photo041.Distance-float64(int16(w>>16))*.004, 2.5, 25)
	default:
		return false
	}
	return true
}
func photoPad041(s padSample033, down uint32, dt float64) {
	if down&0x1000 != 0 {
		photo020Pending = true
	}
	if down&(0x2000|0x10) != 0 {
		togglePhoto041()
		return
	}
	photo041.Yaw += padDeadzone033(s.Axes[2], .18) * dt * 2
	photo041.Pitch = engine.Clamp(photo041.Pitch-padDeadzone033(s.Axes[3], .18)*dt, -.25, 1.15)
	photo041.Height = engine.Clamp(photo041.Height+padDeadzone033(s.Axes[1], .18)*dt*2, -.25, 5)
	photo041.Distance = engine.Clamp(photo041.Distance+padDeadzone033(s.Axes[0], .18)*dt*5, 2.5, 25)
	if s.Buttons&0x100 != 0 {
		photo041.FOV = engine.Clamp(photo041.FOV-dt*20, 25, 85)
	}
	if s.Buttons&0x200 != 0 {
		photo041.FOV = engine.Clamp(photo041.FOV+dt*20, 25, 85)
	}
}
func photoView041(c engine.Car, target *pos3) {
	if !photo041.Active {
		return
	}
	floor := c.Body.Height - .74
	*target = pos3{c.Pos.X, floor + photo041.Height, c.Pos.Z}
	r := photo041.Distance * math.Cos(photo041.Pitch)
	app.cam = pos3{c.Pos.X - math.Sin(photo041.Yaw)*r, target.y + math.Sin(photo041.Pitch)*photo041.Distance, c.Pos.Z - math.Cos(photo041.Yaw)*r}
	app.cam.y = math.Max(app.cam.y, app.world.GroundHeight(engine.Vec{X: app.cam.x, Z: app.cam.z})+.25)
}
func photoHUD041() {
	if photo020Pending {
		return
	} // Saved image contains the scene without controls.
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	roundRect032(w*.035, h*.84, w*.93, h*.12, 4, [4]float64{.025, .04, .05, .86})
	remasterText("PHOTO MODE", w*.055, h*.858, h*.018, white)
	remasterText("DRAG: ORBIT / WHEEL: ZOOM / ARROWS: HEIGHT / Q,E: LENS", w*.055, h*.895, h*.009, mutedRefine)
	remasterText("PAD: STICKS FRAME SHOT / L1,R1: LENS", w*.055, h*.922, h*.008, mutedRefine)
	remasterText("[ ENTER / A: CAPTURE ]", w*.54, h*.892, h*.012, cyan)
	remasterText("[ ESC / B: BACK ]", w*.77, h*.892, h*.012, white)
	en(GL_DEPTH_TEST)
}
