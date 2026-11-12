//go:build windows

package main

import (
	"embed"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"speedinneed/engine"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:embed shaders/*
var shaderAssets embed.FS
var assets gameAssets023

const (
	WM_DESTROY              = 2
	WM_SIZE                 = 5
	WM_KEYDOWN              = 0x100
	WM_MOUSEMOVE            = 0x200
	WM_LBUTTONDOWN          = 0x201
	WM_RBUTTONDOWN          = 0x204
	WM_RBUTTONUP            = 0x205
	WM_CLOSE                = 0x10
	CS_OWNDC                = 0x20
	WS_OVERLAPPEDWINDOW     = 0x00CF0000
	WS_VISIBLE              = 0x10000000
	WS_POPUP                = 0x80000000
	CW_USEDEFAULT           = 0x80000000
	PM_REMOVE               = 1
	VK_SHIFT                = 0x10
	VK_SPACE                = 0x20
	VK_ESCAPE               = 0x1b
	VK_RETURN               = 0x0d
	PFD_DRAW_TO_WINDOW      = 0x4
	PFD_SUPPORT_OPENGL      = 0x20
	PFD_DOUBLEBUFFER        = 0x1
	PFD_TYPE_RGBA           = 0
	GL_COLOR_BUFFER_BIT     = 0x4000
	GL_DEPTH_BUFFER_BIT     = 0x100
	GL_DEPTH_TEST           = 0xb71
	GL_SCISSOR_TEST         = 0xc11
	GL_LEQUAL               = 0x203
	GL_TEXTURE_2D           = 0xde1
	GL_BLEND                = 0xbe2
	GL_SRC_ALPHA            = 0x302
	GL_ONE_MINUS_SRC_ALPHA  = 0x303
	GL_PROJECTION           = 0x1701
	GL_MODELVIEW            = 0x1700
	GL_QUADS                = 7
	GL_TRIANGLES            = 4
	GL_LINES                = 1
	GL_LINE_STRIP           = 3
	GL_COMPILE              = 0x1300
	GL_RGBA                 = 0x1908
	GL_UNSIGNED_BYTE        = 0x1401
	GL_LINEAR               = 0x2601
	GL_LINEAR_MIPMAP_LINEAR = 0x2703
	GL_TEXTURE_MIN_FILTER   = 0x2801
	GL_TEXTURE_MAG_FILTER   = 0x2800
	GL_TEXTURE_WRAP_S       = 0x2802
	GL_TEXTURE_WRAP_T       = 0x2803
	GL_REPEAT               = 0x2901
	GL_CLAMP                = 0x2900
	GL_FOG                  = 0x0b60
	GL_FOG_MODE             = 0xb65
	GL_FOG_START            = 0xb63
	GL_FOG_END              = 0xb64
	GL_FOG_COLOR            = 0xb66
	GL_LINEAR_FOG           = 0x2601
	GL_SMOOTH               = 0x1d01
)

type wndClass struct {
	size, style                        uint32
	wndproc                            uintptr
	clsExtra, wndExtra                 int32
	instance, icon, cursor, background uintptr
	menuName, className                *uint16
	iconSm                             uintptr
}
type pixelFormat struct {
	size, version                                                                                                                                                                                                                     uint16
	flags                                                                                                                                                                                                                             uint32
	pixelType, colorBits, redBits, redShift, greenBits, greenShift, blueBits, blueShift, alphaBits, alphaShift, accumBits, accumRedBits, accumGreenBits, accumBlueBits, accumAlphaBits, depthBits, stencilBits, auxBuffers, layerType uint8
	reserved                                                                                                                                                                                                                          uint8
	layerMask, visibleMask, damageMask                                                                                                                                                                                                uint32
}
type point struct{ x, y int32 }
type windowRect struct{ left, top, right, bottom int32 }
type msg struct {
	wnd            uintptr
	message        uint32
	_              uint32
	wParam, lParam uintptr
	time           uint32
	pt             point
	lPrivate       uint32
	_              uint32
}
type pos3 struct{ x, y, z float64 }

func add(a, b pos3) pos3           { return pos3{a.x + b.x, a.y + b.y, a.z + b.z} }
func sub(a, b pos3) pos3           { return pos3{a.x - b.x, a.y - b.y, a.z - b.z} }
func scale(a pos3, k float64) pos3 { return pos3{a.x * k, a.y * k, a.z * k} }
func norm(a pos3) pos3 {
	l := math.Sqrt(a.x*a.x + a.y*a.y + a.z*a.z)
	if l == 0 {
		return pos3{0, 0, 1}
	}
	return scale(a, 1/l)
}
func cross(a, b pos3) pos3 { return pos3{a.y*b.z - a.z*b.y, a.z*b.x - a.x*b.z, a.x*b.y - a.y*b.x} }

var user32 = syscall.NewLazyDLL("user32.dll")
var gdi32 = syscall.NewLazyDLL("gdi32.dll")
var ogl = syscall.NewLazyDLL("opengl32.dll")
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var funcs = map[string]*syscall.LazyProc{}

// Native arguments may contain Go pointers, including output buffer addresses.
//
//go:uintptrescapes
func gl(name string, args ...uintptr) uintptr {
	if outputFBO038 != 0 && (name == "glDrawBuffer" || name == "glReadBuffer") && len(args) > 0 && (args[0] == 0x405 || args[0] == 0x404) {
		args[0] = 0x8CE0
	}
	if capture != nil {
		return 0
	}
	if skipNativeGL035(name, args) {
		return 0
	}
	p := funcs[name]
	if p == nil {
		p = ogl.NewProc(name)
		funcs[name] = p
	}
	r, _, _ := p.Call(args...)
	return r
}

//go:uintptrescapes
func win(d *syscall.LazyDLL, n string, a ...uintptr) uintptr {
	r, _, _ := nativeProcedures035.lookup(d, n).Call(a...)
	nativeContextChange035(d, n, a, r)
	return r
}
func b32(x float64) uintptr { return uintptr(math.Float32bits(float32(x))) }
func b64(x float64) uintptr { return uintptr(math.Float64bits(x)) }
func en(v uintptr)          { gl("glEnable", v) }
func dis(v uintptr) {
	if capture != nil {
		if v == GL_TEXTURE_2D {
			capture.texture = app.whiteTexture
		}
		return
	}
	gl("glDisable", v)
	// GLSL samples texture regardless of fixed-function enable state. A 1x1
	// white texture preserves untextured meshes inside GL display lists.
	if v == GL_TEXTURE_2D && app.whiteTexture != 0 {
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.whiteTexture))
	}
}
func col(r, g, b, a float64) {
	if capture != nil {
		capture.color = [4]uint8{uint8(engine.Clamp(r, 0, 1) * 255), uint8(engine.Clamp(g, 0, 1) * 255), uint8(engine.Clamp(b, 0, 1) * 255), uint8(engine.Clamp(a, 0, 1) * 255)}
		return
	}
	c := [4]float32{float32(r), float32(g), float32(b), float32(a)}
	gl("glColor4fv", uintptr(unsafe.Pointer(&c[0])))
}
func vtx(x, y, z float64) {
	if capture != nil {
		capture.vertex(x, y, z)
		return
	}
	v := [3]float32{float32(x), float32(y), float32(z)}
	gl("glVertex3fv", uintptr(unsafe.Pointer(&v[0])))
}
func tex(u, v float64) {
	if capture != nil {
		capture.uv = [2]float32{float32(u), float32(v)}
		return
	}
	p := [2]float32{float32(u), float32(v)}
	gl("glTexCoord2fv", uintptr(unsafe.Pointer(&p[0])))
}
func begin(mode uintptr) {
	if capture != nil {
		capture.mode = mode
		capture.pending = capture.pending[:0]
		return
	}
	gl("glBegin", mode)
}
func end() {
	if capture != nil {
		capture.finish()
		return
	}
	gl("glEnd")
}
func push() {
	if capture != nil {
		capture.stack = append(capture.stack, capture.matrix)
		return
	}
	gl("glPushMatrix")
}
func pop() {
	if capture != nil {
		n := len(capture.stack)
		if n > 0 {
			capture.matrix = capture.stack[n-1]
			capture.stack = capture.stack[:n-1]
		}
		return
	}
	gl("glPopMatrix")
}
func move(x, y, z float64) {
	mat := [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, float32(x), float32(y), float32(z), 1}
	if capture != nil {
		capture.multiply(mat)
		return
	}
	gl("glMultMatrixf", uintptr(unsafe.Pointer(&mat[0])))
}
func rotate(a, x, y, z float64) {
	rad := a * math.Pi / 180
	c, s := float32(math.Cos(rad)), float32(math.Sin(rad))
	axis := norm(pos3{x, y, z})
	x, y, z = axis.x, axis.y, axis.z
	f := float64(1 - c)
	m := [16]float32{float32(float64(c) + x*x*f), float32(y*x*f + z*float64(s)), float32(z*x*f - y*float64(s)), 0, float32(x*y*f - z*float64(s)), float32(float64(c) + y*y*f), float32(z*y*f + x*float64(s)), 0, float32(x*z*f + y*float64(s)), float32(y*z*f - x*float64(s)), float32(float64(c) + z*z*f), 0, 0, 0, 0, 1}
	if capture != nil {
		capture.multiply(m)
		return
	}
	gl("glMultMatrixf", uintptr(unsafe.Pointer(&m[0])))
}
func quad(a, b, c, d pos3) {
	meshFaceNormal([3]float64{a.x, a.y, a.z}, [3]float64{b.x, b.y, b.z}, [3]float64{c.x, c.y, c.z})
	vtx(a.x, a.y, a.z)
	vtx(b.x, b.y, b.z)
	vtx(c.x, c.y, c.z)
	vtx(d.x, d.y, d.z)
}
func tq(a, b, c, d pos3, u0, v0, u1, v1 float64) {
	// World ground and road quads have clockwise winding; their geometric
	// normal can point downward. Face horizontal terrain toward the sky.
	ax, ay, az := b.x-a.x, b.y-a.y, b.z-a.z
	bx, by, bz := c.x-a.x, c.y-a.y, c.z-a.z
	nx, ny, nz := ay*bz-az*by, az*bx-ax*bz, ax*by-ay*bx
	if ny < 0 && -ny > 1.5*math.Hypot(nx, nz) {
		meshFaceNormal([3]float64{a.x, a.y, a.z}, [3]float64{c.x, c.y, c.z}, [3]float64{b.x, b.y, b.z})
	} else {
		meshFaceNormal([3]float64{a.x, a.y, a.z}, [3]float64{b.x, b.y, b.z}, [3]float64{c.x, c.y, c.z})
	}
	tex(u0, v0)
	vtx(a.x, a.y, a.z)
	tex(u1, v0)
	vtx(b.x, b.y, b.z)
	tex(u1, v1)
	vtx(c.x, c.y, c.z)
	tex(u0, v1)
	vtx(d.x, d.y, d.z)
}
func key(k int) bool { return (int16(win(user32, "GetAsyncKeyState", uintptr(k))) < 0) }

var app struct {
	hwnd, dc, rc                                                                          uintptr
	width, height                                                                         int
	run, paused, started, boolMouse, boolFPP, camHeld, showHelp, invertMouseX, fullscreen bool
	quality, mapID                                                                        int
	world                                                                                 *engine.World
	cam                                                                                   pos3
	camYaw, camPitch, camDist, mouseSteer, sensitivity, floatFPS, nitroFOV                float64
	lastX, lastY                                                                          int32
	mouseInit                                                                             bool
	windowStyle                                                                           uintptr
	windowRect                                                                            windowRect
	tex                                                                                   map[string]uint32
	staticList, whiteTexture                                                              uint32
	lightingOn                                                                            bool
	prevCar, renderCar                                                                    engine.Car
	renderAlpha                                                                           float64
	debugCollision, debugRoutes, debugWheels                                              bool
	showPerf, vsync                                                                       bool
	minimized                                                                             bool
	frameMS, simMS, renderMS, gpuMS                                                       float64
	gpuTiming                                                                             bool
	t                                                                                     float64
}

func wndproc(hwnd uintptr, m uint32, w, l uintptr) uintptr {
	if garageWnd050(m, w, l) {
		return 0
	}
	if lanWndProc050(hwnd, m, w, l) {
		return 0
	}
	if unityHost040.Parent != 0 && m == WM_LBUTTONDOWN {
		win(user32, "SetFocus", hwnd)
	}
	if unityHostMessage040(hwnd, m, w, l) {
		return 0
	}
	if navMapPointer033(m, w, l) {
		return 0
	}
	if beta020PointerMessage(hwnd, m, w, l) {
		return 0
	}
	if remasterWndProc(hwnd, m, w, l) {
		return 0
	}
	switch m {
	case WM_CLOSE:
		app.run = false
		win(user32, "DestroyWindow", hwnd)
		return 0
	case WM_DESTROY:
		app.run = false
		win(user32, "PostQuitMessage", 0)
		return 0
	case WM_SIZE:
		app.minimized = w == 1
		nw := int(uint16(l & 0xffff))
		nh := int(uint16((l >> 16) & 0xffff))
		if nw > 0 && nh > 0 {
			app.width = nw
			app.height = nh
		}
		return 0
	case 0x001C: // WM_ACTIVATEAPP
		if w == 0 && beta.screen == screenDrive && !qa.enabled {
			selectScreen(screenPause)
			app.camHeld = false
			app.mouseSteer = 0
		}
		return 0
	case 0x0020: // WM_SETCURSOR: free-look during driving, normal pointer in menus.
		if app.started && (beta.screen == screenDrive && !remaster.FullMap && !photo041.Active || multiplayer050.Session != nil) {
			win(user32, "SetCursor", 0)
			return 1
		}
	case WM_RBUTTONDOWN:
		win(user32, "SetCapture", hwnd)
		app.camHeld = true
		return 0
	case WM_RBUTTONUP:
		win(user32, "ReleaseCapture")
		app.camHeld = false
		return 0
	case WM_MOUSEMOVE:
		x := int32(int16(l & 0xffff))
		y := int32(int16((l >> 16) & 0xffff))
		if app.mouseInit && app.started && beta.screen == screenDrive && !remaster.FullMap {
			dx := float64(x - app.lastX)
			dy := float64(y - app.lastY)
			if math.Abs(dx) < 100 && math.Abs(dy) < 100 {
				beta020Look(dx, dy)
			}
		}
		app.lastX = x
		app.lastY = y
		app.mouseInit = true
		return 0
	case WM_LBUTTONDOWN:
		if unityHost040.Parent != 0 {
			win(user32, "SetFocus", hwnd)
		}
		x := int(uint16(l & 0xffff))
		y := int(uint16((l >> 16) & 0xffff))
		betaClick(x, y)
		return 0
	case WM_KEYDOWN:
		if profileKey040(w) {
			return 0
		}
		if (l>>30)&1 != 0 {
			return 0
		}
		betaInput(w)
		return 0
	case 0x0102: // WM_CHAR: profile names use normal text input.
		profileChar040(rune(w))
		return 0
	}
	return win(user32, "DefWindowProcW", hwnd, uintptr(m), w, l)
}

// Borderless fullscreen preserving the previous window size/style.
// F11 is supported on every game screen, including pause and garage.
func toggleFullscreen() {
	if unityHost040.Parent != 0 {
		return
	}
	if outputFBO038 != 0 {
		nativeCall038("SIN038_SuspendFrameGeneration")
	}
	if app.hwnd == 0 {
		return
	}
	styleIndex := ^uintptr(15) // GWL_STYLE (-16)
	if !app.fullscreen {
		win(user32, "GetWindowRect", app.hwnd, uintptr(unsafe.Pointer(&app.windowRect)))
		app.windowStyle = win(user32, "GetWindowLongPtrW", app.hwnd, styleIndex)
		win(user32, "SetWindowLongPtrW", app.hwnd, styleIndex, WS_POPUP|WS_VISIBLE)
		w := win(user32, "GetSystemMetrics", 0)
		h := win(user32, "GetSystemMetrics", 1)
		win(user32, "SetWindowPos", app.hwnd, 0, 0, 0, w, h, 0x24) // SWP_FRAMECHANGED|SWP_NOZORDER
		app.fullscreen = true
	} else {
		style := app.windowStyle
		if style == 0 {
			style = WS_OVERLAPPEDWINDOW | WS_VISIBLE
		}
		win(user32, "SetWindowLongPtrW", app.hwnd, styleIndex, style)
		box := app.windowRect
		win(user32, "SetWindowPos", app.hwnd, 0, uintptr(box.left), uintptr(box.top), uintptr(box.right-box.left), uintptr(box.bottom-box.top), 0x24)
		app.fullscreen = false
	}
}

func initWindow() {
	runtime.LockOSThread()
	win(user32, "SetProcessDPIAware")
	title, _ := syscall.UTF16PtrFromString("SPEED IN NEED  |  BETA v0.5.3")
	cls, _ := syscall.UTF16PtrFromString("SpeedInNeedNativeGame")
	instance := win(kernel32, "GetModuleHandleW", 0)
	gameIcon := win(user32, "LoadIconW", instance, 1) // RT_GROUP_ICON 1 embedded in PE
	wc := wndClass{size: uint32(unsafe.Sizeof(wndClass{})), style: CS_OWNDC, wndproc: syscall.NewCallback(wndproc), instance: instance, className: cls, icon: gameIcon, iconSm: gameIcon, cursor: win(user32, "LoadCursorW", 0, 32512)}
	if win(user32, "RegisterClassExW", uintptr(unsafe.Pointer(&wc))) == 0 {
		panic("window class registration failed")
	}
	app.width = 1280
	app.height = 720
	resolution := [4][2]int{{1280, 720}, {1600, 900}, {1920, 1080}, {2560, 1440}}[engine.ClampInt(beta.cfg.Resolution, 0, 3)]
	app.width, app.height = resolution[0], resolution[1]
	style := uintptr(WS_OVERLAPPEDWINDOW | WS_VISIBLE)
	if qa.enabled {
		app.width = 1920
		app.height = 1080
		style = WS_POPUP
	}
	if unityHost040.Parent != 0 {
		style = 0x40000000 | 0x04000000 | 0x02000000
	}
	windowBox := [4]int32{0, 0, int32(app.width), int32(app.height)}
	win(user32, "AdjustWindowRect", uintptr(unsafe.Pointer(&windowBox[0])), style, 0)
	x040, y040 := uintptr(100), uintptr(55)
	if unityHost040.Parent != 0 {
		x040, y040 = 0, 0
	}
	exStyle040 := uintptr(0)
	if unityHost040.Parent != 0 {
		exStyle040 = 4
	}
	app.hwnd = win(user32, "CreateWindowExW", exStyle040, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)), style, x040, y040, uintptr(windowBox[2]-windowBox[0]), uintptr(windowBox[3]-windowBox[1]), unityHost040.Parent, 0, instance, 0)
	if app.hwnd == 0 {
		panic("native window creation failed")
	}
	var client [4]int32
	win(user32, "GetClientRect", app.hwnd, uintptr(unsafe.Pointer(&client[0])))
	app.width = int(client[2])
	app.height = int(client[3])
	if gameIcon != 0 && unityHost040.Parent == 0 {
		win(user32, "SendMessageW", app.hwnd, 0x80, 1, gameIcon) // ICON_BIG
		win(user32, "SendMessageW", app.hwnd, 0x80, 0, gameIcon) // ICON_SMALL
	}
	app.dc = win(user32, "GetDC", app.hwnd)
	pf := pixelFormat{size: uint16(unsafe.Sizeof(pixelFormat{})), version: 1, flags: PFD_DRAW_TO_WINDOW | PFD_SUPPORT_OPENGL | PFD_DOUBLEBUFFER, pixelType: PFD_TYPE_RGBA, colorBits: 32, depthBits: 24, stencilBits: 8, layerType: 0}
	format := win(gdi32, "ChoosePixelFormat", app.dc, uintptr(unsafe.Pointer(&pf)))
	if format == 0 {
		panic("ChoosePixelFormat failed")
	}
	if win(gdi32, "SetPixelFormat", app.dc, format, uintptr(unsafe.Pointer(&pf))) == 0 {
		panic("SetPixelFormat failed")
	}
	app.rc = win(ogl, "wglCreateContext", app.dc)
	if app.rc == 0 {
		panic("OpenGL context failed")
	}
	if win(ogl, "wglMakeCurrent", app.dc, app.rc) == 0 {
		panic("activate OpenGL failed")
	}
	if !qa.enabled && unityHost040.Parent == 0 {
		win(user32, "ShowWindow", app.hwnd, 5)
	}
	win(user32, "UpdateWindow", app.hwnd)
	beta020RegisterMouse()
}

// loadGameTexture allows players to replace artwork without installing Go,
// altering the EXE, or using a game editor. Font atlas is intentionally fixed:
// its glyph positions must match the built-in HUD renderer.
func loadGameTexture(name string) (image.Image, error) {
	if name != "b037_font_fallback.png" {
		if exe, err := os.Executable(); err == nil {
			base := strings.TrimSuffix(name, filepath.Ext(name))
			for _, ext := range []string{".png", ".jpg", ".jpeg"} {
				path := filepath.Join(filepath.Dir(exe), "Mods", "textures", base+ext)
				if f, err := os.Open(path); err == nil {
					im, _, decodeErr := image.Decode(f)
					f.Close()
					if decodeErr == nil {
						return im, nil
					}
				}
			}
		}
	}
	f, err := assets.Open("assets/" + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	im, _, err := image.Decode(f)
	return im, err
}

func setupGL() {
	en(GL_DEPTH_TEST)
	gl("glDepthFunc", GL_LEQUAL)
	en(GL_TEXTURE_2D)
	en(GL_BLEND)
	gl("glBlendFunc", GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA)
	gl("glShadeModel", GL_SMOOTH)
	app.tex = map[string]uint32{}
	for _, name := range []string{"asphalt.jpg", "grass.jpg", "rock.jpg", "sand.jpg", "ocean.jpg", "building.jpg", "foliage.jpg", "bark.jpg", "sky_sunset.jpg", "sky_night.jpg", "sky_day.jpg", "b037_font_fallback.png", "b037_ui_font_sdf.png", "b033_map_free.png", "b033_map_race.png", "b038_map_free.png", "b038_map_race.png", "b038_title_logo.png", "remaster_smoke.png", "b035_smoke.png", "title_beta011.png", "b011_landcover.png", "aokawa_reference.png", "aokawa_circuit.png", "aokawa_roads.png", "b041_road_clearance.png", "b020_title.png", "beta020_atlas.png", "b020_sky_day.jpg", "b020_sky_afternoon.jpg", "b020_sky_night.jpg"} {
		im, err := loadGameTexture(name)
		if err != nil {
			continue
		}
		b := im.Bounds()
		rgba := texturePixels041(im)
		var t uint32
		gl("glGenTextures", 1, uintptr(unsafe.Pointer(&t)))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(t))
		uiImage := name == "font.png" || name == "b032_ui_font.png" || name == "b034_title_font.png" || name == "b037_ui_font_sdf.png"
		minFilter := uintptr(GL_LINEAR_MIPMAP_LINEAR)
		if uiImage {
			minFilter = GL_LINEAR
		}
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, minFilter)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
		wrap := uintptr(GL_REPEAT)
		if strings.HasPrefix(name, "b033_map_") || strings.HasPrefix(name, "b038_map_") {
			wrap = 0x812f
		}
		if uiImage {
			wrap = 0x812f
		}
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, wrap)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, wrap)
		gl("glTexImage2D", GL_TEXTURE_2D, 0, GL_RGBA, uintptr(b.Dx()), uintptr(b.Dy()), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&rgba[0])))
		if !uiImage {
			uploadTextureMipLevels(rgba, b.Dx(), b.Dy())
		}
		app.tex[name] = t
	}
}

// OpenGL 1.1-compatible software mipmap generator. Distant tarmac/grass
// textures were sparkling because the old renderer always sampled full-size
// textures with no mip levels. Everything is generated once at game startup.
func uploadTextureMipLevels(base []byte, width, height int) {
	level := 1
	prev := base
	w, h := width, height
	for w > 1 || h > 1 {
		nw, nh := w/2, h/2
		if nw < 1 {
			nw = 1
		}
		if nh < 1 {
			nh = 1
		}
		next := make([]byte, nw*nh*4)
		for y := 0; y < nh; y++ {
			for x := 0; x < nw; x++ {
				for c := 0; c < 4; c++ {
					sum, n := 0, 0
					for dy := 0; dy < 2; dy++ {
						for dx := 0; dx < 2; dx++ {
							sx, sy := x*2+dx, y*2+dy
							if sx < w && sy < h {
								sum += int(prev[(sy*w+sx)*4+c])
								n++
							}
						}
					}
					next[(y*nw+x)*4+c] = byte(sum / n)
				}
			}
		}
		gl("glTexImage2D", GL_TEXTURE_2D, uintptr(level), GL_RGBA, uintptr(nw), uintptr(nh), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&next[0])))
		prev, w, h = next, nw, nh
		level++
	}
}

func createWhiteTexture() {
	var id uint32
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&id)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(id))
	white := [4]byte{255, 255, 255, 255}
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, GL_RGBA, 1, 1, 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&white[0])))
	app.whiteTexture = id
}
func bind(name string) {
	if capture != nil {
		capture.texture = app.tex[name]
		return
	}
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex[name]))
}
func applyCamera(eye, target pos3) {
	f := norm(sub(target, eye))
	s := norm(cross(f, pos3{0, 1, 0}))
	u := cross(s, f)
	mat := [16]float32{float32(s.x), float32(u.x), float32(-f.x), 0, float32(s.y), float32(u.y), float32(-f.y), 0, float32(s.z), float32(u.z), float32(-f.z), 0, 0, 0, 0, 1}
	gl("glMultMatrixf", uintptr(unsafe.Pointer(&mat[0])))
	move(-eye.x, -eye.y, -eye.z)
}
func perspective(fov float64) {
	sky020.FOV = fov
	aspect := float64(app.width) / float64(app.height)
	z := .12
	x := z * math.Tan(fov*math.Pi/360)
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	r := x * aspect
	n := z
	f := 4600.0
	matrix := [16]float32{float32(n / r), 0, 0, 0, 0, float32(n / x), 0, 0, 0, 0, float32(-(f + n) / (f - n)), -1, 0, 0, float32(-2 * f * n / (f - n)), 0}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&matrix[0])))
	gl("glMatrixMode", GL_MODELVIEW)
}
func legacyDrawSky(camera pos3) {
	// One seamless 360-degree panoramic image on an inward-facing dome.
	// Repeating a portrait image 18 times made the sky look striped.
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	en(GL_TEXTURE_2D)
	sky := "sky_day.jpg"
	if engine.NightAmount(sceneHour()) > .8 {
		sky = "sky_night.jpg"
	} else if sceneHour() > 17 {
		sky = "sky_sunset.jpg"
	}
	bind(sky)
	col(1, 1, 1, 1)
	gl("glDepthMask", 0)
	const sections = 96
	const rows = 6
	radius := 980.0
	yBottom, yTop := -110.0, 650.0
	begin(GL_QUADS)
	for band := 0; band < rows; band++ {
		vLow, vHigh := float64(band)/rows, float64(band+1)/rows
		yLow, yHigh := yBottom+(yTop-yBottom)*vLow, yBottom+(yTop-yBottom)*vHigh
		for i := 0; i < sections; i++ {
			a, b := float64(i)*2*math.Pi/sections, float64(i+1)*2*math.Pi/sections
			xa, za := camera.x+math.Cos(a)*radius, camera.z+math.Sin(a)*radius
			xb, zb := camera.x+math.Cos(b)*radius, camera.z+math.Sin(b)*radius
			u1, u2 := float64(i)/sections, float64(i+1)/sections
			tq(pos3{xa, yLow, za}, pos3{xb, yLow, zb}, pos3{xb, yHigh, zb}, pos3{xa, yHigh, za}, u1, 1-vLow, u2, 1-vHigh)
		}
	}
	end()
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
}
func box(x0, y0, z0, x1, y1, z1 float64) {
	begin(GL_QUADS)
	quad(pos3{x0, y0, z0}, pos3{x1, y0, z0}, pos3{x1, y1, z0}, pos3{x0, y1, z0})
	quad(pos3{x1, y0, z1}, pos3{x0, y0, z1}, pos3{x0, y1, z1}, pos3{x1, y1, z1})
	quad(pos3{x0, y0, z1}, pos3{x0, y0, z0}, pos3{x0, y1, z0}, pos3{x0, y1, z1})
	quad(pos3{x1, y0, z0}, pos3{x1, y0, z1}, pos3{x1, y1, z1}, pos3{x1, y1, z0})
	quad(pos3{x0, y1, z0}, pos3{x1, y1, z0}, pos3{x1, y1, z1}, pos3{x0, y1, z1})
	quad(pos3{x0, y0, z1}, pos3{x1, y0, z1}, pos3{x1, y0, z0}, pos3{x0, y0, z0})
	end()
}
func texturedBox(w, d, h, tu, tv float64) {
	begin(GL_QUADS)
	tq(pos3{-w / 2, 0, -d / 2}, pos3{w / 2, 0, -d / 2}, pos3{w / 2, h, -d / 2}, pos3{-w / 2, h, -d / 2}, 0, 0, tu, tv)
	tq(pos3{w / 2, 0, d / 2}, pos3{-w / 2, 0, d / 2}, pos3{-w / 2, h, d / 2}, pos3{w / 2, h, d / 2}, 0, 0, tu, tv)
	tq(pos3{-w / 2, 0, d / 2}, pos3{-w / 2, 0, -d / 2}, pos3{-w / 2, h, -d / 2}, pos3{-w / 2, h, d / 2}, 0, 0, tu, tv)
	tq(pos3{w / 2, 0, -d / 2}, pos3{w / 2, 0, d / 2}, pos3{w / 2, h, d / 2}, pos3{w / 2, h, -d / 2}, 0, 0, tu, tv)
	tq(pos3{-w / 2, h, -d / 2}, pos3{w / 2, h, -d / 2}, pos3{w / 2, h, d / 2}, pos3{-w / 2, h, d / 2}, 0, 0, tu, tu)
	end()
}

// Lightweight clouds in every map, drawn in layers and independent of sky JPG.
func renderCloudBanks(camera pos3) {
	dis(GL_TEXTURE_2D)
	dis(GL_DEPTH_TEST)
	en(GL_BLEND)
	gl("glDepthMask", 0)
	for i := 0; i < 33; i++ {
		a := float64(i) * 2 * math.Pi / 33
		r := 470. + float64((i*79)%250)
		x, z := camera.x+math.Cos(a)*r, camera.z+math.Sin(a)*r
		y := 185. + float64((i*31)%160)
		for j := 0; j < 5; j++ {
			dx := float64(j-2) * 26
			shade := .66 + .12*math.Sin(a+float64(j))
			if app.mapID == 1 {
				col(.40*shade, .49*shade, .78*shade, .13)
			} else if app.mapID == 2 {
				col(.96*shade, .96*shade, 1, .18)
			} else {
				col(1, .77*shade, .64*shade, .16)
			}
			radius := 25. + float64((i*7+j*13)%19)
			begin(GL_QUADS)
			vtx(x+dx-radius, y, z-radius)
			vtx(x+dx+radius, y, z-radius)
			vtx(x+dx+radius, y+radius*.46, z+radius)
			vtx(x+dx-radius, y+radius*.46, z+radius)
			end()
		}
	}
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
}

// Particle-looking rear-tire vapor generated during a stationary handbrake burnout.
func renderRearSmoke(c engine.Car) {
	if !c.Burnout && !(math.Abs(c.DriftAngle) > 13 && math.Abs(c.Speed) > 10) {
		return
	}
	dis(GL_TEXTURE_2D)
	en(GL_BLEND)
	gl("glDepthMask", 0)
	f := engine.HeadingVec(c.Heading)
	right := engine.Vec{X: f.Z, Z: -f.X}
	for side := -1; side <= 1; side += 2 {
		for i := 0; i < 8; i++ {
			age := math.Mod(app.t*2.3+float64(i)*.15, 1)
			p := c.Pos.Sub(f.Mul(1.35 + age*2.9)).Add(right.Mul(float64(side)*.88 + age*float64(side)*.2))
			h := engine.Terrain(p.X, p.Z) + .42 + age*2.2
			size := .30 + age*1.15
			col(.81, .86, .90, (1-age)*.23)
			begin(GL_QUADS)
			vtx(p.X-size, h-size*.3, p.Z)
			vtx(p.X+size, h-size*.3, p.Z)
			vtx(p.X+size, h+size, p.Z)
			vtx(p.X-size, h+size, p.Z)
			end()
		}
	}
	gl("glDepthMask", 1)
}

func renderTerrain() {
	en(GL_TEXTURE_2D)
	terrainTex := []string{"grass.jpg", "grass.jpg", "grass.jpg"}[app.mapID]
	bind(terrainTex)
	for z := -990.; z < 1000; z += 18 {
		begin(GL_QUADS)
		for x := -990.; x < 1000; x += 18 {
			sh := math.Sin(x*.03 + z*.07)
			v := .70 + .20*sh
			if app.mapID == 1 {
				col(.70*v, .83*v, .82*v, 1)
			} else if app.mapID == 2 {
				col(.88*v, .87*v, .75*v, 1)
			} else {
				col(.88*v, .96*v, .86*v, 1)
			}
			tq(pos3{x, app.world.TerrainHeight(engine.Vec{X: x, Z: z}) - .065, z}, pos3{x + 18, app.world.TerrainHeight(engine.Vec{X: x + 18, Z: z}) - .065, z}, pos3{x + 18, app.world.TerrainHeight(engine.Vec{X: x + 18, Z: z + 18}) - .065, z + 18}, pos3{x, app.world.TerrainHeight(engine.Vec{X: x, Z: z + 18}) - .065, z + 18}, x/28, z/28, (x+18)/28, (z+18)/28)
		}
		end()
	}
	// Rugged layered rock outcrops. Beta 3.0 used giant untextured pyramid
	// triangles that obscured the driving view. Keep the ridges away from roads.
	en(GL_TEXTURE_2D)
	bind("rock.jpg")
	for j := 0; j < 35; j++ {
		x := -445. + float64((j*139+29)%900)
		z := -415. + float64((j*197+31)%850)
		if engine.DistRoad(engine.Vec{X: x, Z: z}, app.world.Routes) < 92 || app.world.IsWater(engine.Vec{X: x, Z: z}, 30) || app.world.RampClearance(engine.Vec{X: x, Z: z}, 40) {
			continue
		}
		y := engine.Terrain(x, z) - 2
		height := 19. + float64((j*19)%42)
		radius := 24. + float64((j*11)%29)
		// Four concentric belts with irregular silhouette avoid triangular cones.
		for ring := 0; ring < 3; ring++ {
			for side := 0; side < 14; side++ {
				a := float64(side) * 2 * math.Pi / 14
				b := float64(side+1) * 2 * math.Pi / 14
				r1 := radius * (1 - float64(ring)*.28)
				r2 := radius * (1 - float64(ring+1)*.28)
				y1 := y + height*float64(ring)*.31
				y2 := y + height*float64(ring+1)*.31
				noise1 := 1 + .12*math.Sin(float64(j)*.72+float64(side)*2.3)
				noise2 := 1 + .17*math.Sin(float64(j)*1.1+float64(side+1)*2.1)
				p1 := pos3{x + math.Cos(a)*r1*noise1, y1, z + math.Sin(a)*r1*noise1}
				p2 := pos3{x + math.Cos(b)*r1*noise2, y1, z + math.Sin(b)*r1*noise2}
				p3 := pos3{x + math.Cos(b)*r2*noise2, y2, z + math.Sin(b)*r2*noise2}
				p4 := pos3{x + math.Cos(a)*r2*noise1, y2, z + math.Sin(a)*r2*noise1}
				shade := .71 + .17*math.Sin(a+float64(j))
				if app.mapID == 2 {
					col(.89*shade, .77*shade, .64*shade, 1)
				} else {
					col(.72*shade, .78*shade, .75*shade, 1)
				}
				begin(GL_QUADS)
				tq(p1, p2, p3, p4, float64(side)/4, float64(ring)/2, float64(side+1)/4, float64(ring+1)/2)
				end()
			}
		}
	}

}

func legacyRenderHills() {
	// Rugged layered rock outcrops. Beta 3.0 used giant untextured pyramid
	// triangles that obscured the driving view. Keep the ridges away from roads.
	en(GL_TEXTURE_2D)
	bind("rock.jpg")
	for j := 0; j < 35; j++ {
		x := -445. + float64((j*139+29)%900)
		z := -415. + float64((j*197+31)%850)
		if engine.DistRoad(engine.Vec{X: x, Z: z}, app.world.Routes) < 92 || app.world.IsWater(engine.Vec{X: x, Z: z}, 30) || app.world.RampClearance(engine.Vec{X: x, Z: z}, 40) {
			continue
		}
		y := engine.Terrain(x, z) - 2
		height := 19. + float64((j*19)%42)
		radius := 24. + float64((j*11)%29)
		// Four concentric belts with irregular silhouette avoid triangular cones.
		for ring := 0; ring < 3; ring++ {
			for side := 0; side < 14; side++ {
				a := float64(side) * 2 * math.Pi / 14
				b := float64(side+1) * 2 * math.Pi / 14
				r1 := radius * (1 - float64(ring)*.28)
				r2 := radius * (1 - float64(ring+1)*.28)
				y1 := y + height*float64(ring)*.31
				y2 := y + height*float64(ring+1)*.31
				noise1 := 1 + .12*math.Sin(float64(j)*.72+float64(side)*2.3)
				noise2 := 1 + .17*math.Sin(float64(j)*1.1+float64(side+1)*2.1)
				p1 := pos3{x + math.Cos(a)*r1*noise1, y1, z + math.Sin(a)*r1*noise1}
				p2 := pos3{x + math.Cos(b)*r1*noise2, y1, z + math.Sin(b)*r1*noise2}
				p3 := pos3{x + math.Cos(b)*r2*noise2, y2, z + math.Sin(b)*r2*noise2}
				p4 := pos3{x + math.Cos(a)*r2*noise1, y2, z + math.Sin(a)*r2*noise1}
				shade := .71 + .17*math.Sin(a+float64(j))
				if app.mapID == 2 {
					col(.89*shade, .77*shade, .64*shade, 1)
				} else {
					col(.72*shade, .78*shade, .75*shade, 1)
				}
				begin(GL_QUADS)
				tq(p1, p2, p3, p4, float64(side)/4, float64(ring)/2, float64(side+1)/4, float64(ring+1)/2)
				end()
			}
		}
	}

}

// A small original coastal-water surface visually separates the Coast map
// from the city's buildings and the canyon's rock formations.
func renderCoastalWater() {
	en(GL_TEXTURE_2D)
	bind("ocean.jpg")
	for z := engine.CoastWaterMinZ; z < engine.CoastWaterMaxZ; z += 14 {
		begin(GL_QUADS)
		for x := engine.CoastWaterMinX; x < engine.CoastWaterMaxX; x += 14 {
			shade := .82 + .16*math.Sin(z*.024+x*.011)
			col(.68*shade, .88*shade, 1.0*shade, 1)
			x1, z1 := math.Min(x+14, engine.CoastWaterMaxX), math.Min(z+14, engine.CoastWaterMaxZ)
			y := engine.CoastWaterLevel
			tq(pos3{x, y, z}, pos3{x1, y, z}, pos3{x1, y, z1}, pos3{x, y, z1}, x/15, z/15, x1/15, z1/15)
		}
		end()
	}
	// A soft sandy shoreline near the ocean; road geometry is rendered later.
	bind("sand.jpg")
	col(.83, .75, .60, 1)
	begin(GL_QUADS)
	for z := -520.; z < 520; z += 14 {
		for x := -205.; x < -169; x += 12 {
			tq(pos3{x, engine.Terrain(x, z) + .055, z}, pos3{x + 12, engine.Terrain(x+12, z) + .055, z}, pos3{x + 12, engine.Terrain(x+12, z+14) + .055, z + 14}, pos3{x, engine.Terrain(x, z+14) + .055, z + 14}, 0, z/11, 1, (z+14)/11)
		}
	}
	end()
}

func renderRoads() {
	// Shared per-knot tangents and metre-based UVs fix hard seams and stretched
	// asphalt around spline curves. Each crossing road has a tiny depth bias
	// to avoid competing coplanar road faces (z-fighting).
	for routeID, r := range app.world.Routes {
		count := len(r.Points)
		if count < 2 {
			continue
		}
		normals := make([]engine.Vec, count)
		for i := range normals {
			normals[i] = engine.RoadNormalAt(r.Points, i)
		}
		heightBias := float64(routeID) * .035
		type surface struct {
			texName           string
			width, lift, tile float64
			red, green, blue  float64
		}
		surfaces := []surface{
			{"sand.jpg", 9.8, .135, 6.5, .70, .64, .55},
			{"asphalt.jpg", 7.8, .235, 7.0, .91, .94, .98},
		}
		for _, sf := range surfaces {
			en(GL_TEXTURE_2D)
			bind(sf.texName)
			col(sf.red, sf.green, sf.blue, 1)
			begin(GL_QUADS)
			for i := 0; i < count-1; i++ {
				a, b := r.Points[i], r.Points[i+1]
				na, nb := normals[i], normals[i+1]
				pa, pb := a.Add(na.Mul(sf.width)), a.Add(na.Mul(-sf.width))
				pc, pd := b.Add(nb.Mul(-sf.width)), b.Add(nb.Mul(sf.width))
				ua, ub := r.Cumulative[i]/sf.tile, r.Cumulative[i+1]/sf.tile
				lift := sf.lift + heightBias
				tq(pos3{pa.X, engine.Terrain(pa.X, pa.Z) + lift, pa.Z}, pos3{pb.X, engine.Terrain(pb.X, pb.Z) + lift, pb.Z}, pos3{pc.X, engine.Terrain(pc.X, pc.Z) + lift, pc.Z}, pos3{pd.X, engine.Terrain(pd.X, pd.Z) + lift, pd.Z}, 0, ua, 4, ub)
			}
			end()
		}
		dis(GL_TEXTURE_2D)
		// Dashed centreline floats above the road layer, not coplanar with it.
		col(.96, .86, .57, 1)
		begin(GL_QUADS)
		for i := 0; i < count-1; i += 4 {
			a, b := r.Points[i], r.Points[i+1]
			na, nb := normals[i], normals[i+1]
			p1, p2 := a.Add(na.Mul(.11)), a.Add(na.Mul(-.11))
			p3, p4 := b.Add(nb.Mul(-.11)), b.Add(nb.Mul(.11))
			lift := .269 + heightBias
			tq(pos3{p1.X, engine.Terrain(p1.X, p1.Z) + lift, p1.Z}, pos3{p2.X, engine.Terrain(p2.X, p2.Z) + lift, p2.Z}, pos3{p3.X, engine.Terrain(p3.X, p3.Z) + lift, p3.Z}, pos3{p4.X, engine.Terrain(p4.X, p4.Z) + lift, p4.Z}, 0, 0, 1, 1)
		}
		end()
	}
}

func renderTree(t engine.Tree) {
	push()
	y := engine.Terrain(t.Pos.X, t.Pos.Z) - .06
	move(t.Pos.X, y, t.Pos.Z)
	s := t.Size
	en(GL_TEXTURE_2D)
	bind("bark.jpg")
	col(.77, .74, .69, 1)
	begin(GL_QUADS)
	for i := 0; i < 8; i++ {
		a, b := float64(i)*2*math.Pi/8, float64(i+1)*2*math.Pi/8
		r := s * .20
		tq(pos3{math.Cos(a) * r, 0, math.Sin(a) * r}, pos3{math.Cos(b) * r, 0, math.Sin(b) * r}, pos3{math.Cos(b) * r, 2.8 * s, math.Sin(b) * r}, pos3{math.Cos(a) * r, 2.8 * s, math.Sin(a) * r}, float64(i)/8, 1, float64(i+1)/8, 0)
	}
	end()
	bind("foliage.jpg")
	// Many smaller overlapping textured leaf clumps instead of three cones.
	// Slightly varied light/dark greens avoid flat geometric pyramid foliage.
	for layer := 0; layer < 4; layer++ {
		level := (2.35 + float64(layer)*.43) * s
		rings := 14
		for j := 0; j < rings; j++ {
			a := float64(j)*2*math.Pi/float64(rings) + float64(layer)*.31
			radius := (1.65 - float64(layer)*.22) * s
			cx, cz := math.Cos(a)*radius*.63, math.Sin(a)*radius*.63
			leafSize := (.62 + .20*math.Sin(float64(j*3+layer*7))) * s
			cshade := .72 + .18*math.Sin(a*2+float64(layer)*.5)
			col(.40*cshade, .78*cshade, .43*cshade, 1)
			begin(GL_QUADS)
			tq(pos3{cx - leafSize, level - .18*s, cz}, pos3{cx + leafSize, level - .18*s, cz}, pos3{cx + leafSize*.6, level + leafSize*.96, cz + .22*s}, pos3{cx - leafSize*.6, level + leafSize*.96, cz + .22*s}, 0, 1, 1, 0)
			end()
			col(.29*cshade, .62*cshade, .32*cshade, 1)
			begin(GL_QUADS)
			tq(pos3{cx, level - .17*s, cz - leafSize}, pos3{cx, level - .17*s, cz + leafSize}, pos3{cx + .22*s, level + leafSize*.9, cz + leafSize*.6}, pos3{cx + .22*s, level + leafSize*.9, cz - leafSize*.6}, 0, 1, 1, 0)
			end()
		}
	}
	pop()
}
func renderBuilding(x, z float64, id int) {
	push()
	move(x, engine.Terrain(x, z)-.04, z)
	h := float64(11 + (id*13)%42)
	bind("building.jpg")
	en(GL_TEXTURE_2D)
	col(.70, .79, .84, 1)
	texturedBox(13, 15, h, 3, 7)
	dis(GL_TEXTURE_2D)
	col(.18, .25, .37, 1)
	for yy := 4.; yy < h-1; yy += 4.8 {
		for xx := -5.; xx < 5; xx += 3.5 {
			box(xx, yy, 7.54, xx+1.8, yy+2.4, 7.61)
		}
	}
	if app.mapID == 1 {
		// Decorative luminous frontage and side trim, no invisible collider.
		dis(GL_TEXTURE_2D)
		for yy := 3.5; yy < h-2; yy += 6.8 {
			if id%2 == 0 {
				col(.06, .78, 1, .96)
			} else {
				col(1, .21, .55, .96)
			}
			box(-6.51, yy, -7.55, -6.40, yy+.14, 7.55)
			box(6.40, yy, -7.55, 6.51, yy+.14, 7.55)
		}
	}
	pop()
}

// Lightweight native foliage geometry sits on continuous photo-textured terrain.
// Does not create vehicle colliders or large imported GLB mesh spikes.
func renderGrassTufts() {
	dis(GL_TEXTURE_2D)
	count := []int{700, 1550, 2500}[engine.ClampInt(app.quality, 0, 2)]
	begin(GL_TRIANGLES)
	for i := 0; i < count; i++ {
		x := math.Mod(float64(i)*31.137+37, 870) - 435
		z := math.Mod(float64(i)*97.713+91, 840) - 420
		if engine.DistRoad(engine.Vec{X: x, Z: z}, app.world.Routes) < 12.5 || app.world.IsWater(engine.Vec{X: x, Z: z}, 2) || app.world.RampClearance(engine.Vec{X: x, Z: z}, 2) {
			continue
		}
		y := engine.Terrain(x, z) - .65
		height := .15 + math.Mod(float64(i)*.0713, .34)
		radius := .09 + math.Mod(float64(i)*.047, .11)
		shade := .55 + math.Mod(float64(i)*.013, .25)
		col(.18*shade, .48*shade, .16*shade, 1)
		vtx(x-radius, y, z)
		vtx(x+radius, y, z)
		vtx(x, y+height, z+radius*.2)
		vtx(x, y, z-radius)
		vtx(x, y, z+radius)
		vtx(x+radius*.2, y+height*.82, z)
	}
	end()
}

func invalidateWorldGeometry() {
	if remaster.Enabled {
		gpuWorld.ready = remaster.BufferMode >= 0
		return
	}
	releaseBufferedWorld()
	releaseRenderChunks()
	if app.staticList != 0 {
		gl("glDeleteLists", uintptr(app.staticList), 1)
		app.staticList = 0
	}
}

func buildStatic() {
	if buildBufferedWorld() {
		return
	}
	if app.staticList > 0 {
		gl("glDeleteLists", uintptr(app.staticList), 1)
	}
	app.staticList = uint32(gl("glGenLists", 1))
	gl("glNewList", uintptr(app.staticList), GL_COMPILE)
	renderStaticDetails()
	// Imported skyline mesh disabled to prevent stretched spike artifacts on some GPUs.
	gl("glEndList")
	buildWorldChunks()
}
func renderPoles() {
	for _, p := range app.world.Poles {
		push()
		move(p.Pos.X, engine.Terrain(p.Pos.X, p.Pos.Z)+.10, p.Pos.Z)
		if p.Falling {
			rotate(p.Angle*180/math.Pi, p.FallDirection.Z, 0, -p.FallDirection.X)
		}
		rotate(p.Heading*180/math.Pi, 0, 1, 0)
		if !drawPBRAsset("streetlight", 0) {
			dis(GL_TEXTURE_2D)
			col(.28, .32, .35, 1)
			box(-.12, 0, -.12, .12, 9.8, .12)
			box(0, 9.5, -.12, 3.9, 9.7, .12)
		}
		pop()
	}
}

func renderLightPools() {
	if app.mapID != 1 {
		return
	}
	dis(GL_TEXTURE_2D)
	dis(GL_FOG)
	en(GL_BLEND)
	gl("glDepthMask", 0)
	for _, pole := range app.world.Poles {
		if pole.Falling || pole.Pos.Dist(app.world.Car.Pos) > 160 {
			continue
		}
		baseY := engine.Terrain(pole.Pos.X, pole.Pos.Z) + .35
		const slices = 18
		for i := 0; i < slices; i++ {
			a := 2 * math.Pi * float64(i) / slices
			b := 2 * math.Pi * float64(i+1) / slices
			begin(GL_TRIANGLES)
			col(1, .73, .34, .16)
			vtx(pole.Pos.X, baseY, pole.Pos.Z)
			col(1, .60, .26, 0)
			vtx(pole.Pos.X+math.Cos(a)*7, baseY, pole.Pos.Z+math.Sin(a)*7)
			vtx(pole.Pos.X+math.Cos(b)*7, baseY, pole.Pos.Z+math.Sin(b)*7)
			end()
		}
		// A small emissive-looking lamp head, cheaper than dynamic lights.
		push()
		move(pole.Pos.X, engine.Terrain(pole.Pos.X, pole.Pos.Z)+9.7, pole.Pos.Z)
		col(1, .82, .47, .9)
		box(-.22, -.10, -.22, .22, .10, .22)
		pop()
	}
	gl("glDepthMask", 1)
	en(GL_FOG)
}

func renderPickups() {
	dis(GL_TEXTURE_2D)
	for _, p := range app.world.Pickups {
		if p.Cooldown > 0 {
			continue
		}
		push()
		move(p.Pos.X, engine.Terrain(p.Pos.X, p.Pos.Z)+2+math.Sin(app.t*3)*.45, p.Pos.Z)
		rotate(app.t*80, 0, 1, 0)
		col(.06, .96, 1, 1)
		box(-.50, -.65, -.13, .5, .65, .13)
		col(.8, 1, 1, 1)
		box(-.20, -.4, -.16, .2, .4, .16)
		pop()
	}
}
func drawCar(pos engine.Vec, angle float64, variant int, player bool, braking bool) {
	id := 2
	if player {
		id = beta.carSelect
	} else if variant%4 == 3 {
		id = 3
	}
	if beta.mesh[id] != 0 || (id < 2 && bufferedCars[id].available) {
		renderNativeCar(id, pos, angle, func() float64 {
			if player {
				return app.renderCar.WheelSpin
			}
			return app.t * 27 / .34
		}(), func() float64 {
			if player {
				return app.renderCar.VisualSteer
			}
			return 0
		}(), variant)
		return
	}

	push()
	move(pos.X, engine.Terrain(pos.X, pos.Z)+engine.CarElevation(engine.DistRoad(pos, app.world.Routes), id), pos.Z)
	rotate(angle*180/math.Pi, 0, 1, 0)
	dis(GL_TEXTURE_2D)
	cols := [7][3]float64{{.97, .31, .10}, {.15, .39, .83}, {.90, .92, .96}, {.19, .22, .25}, {.92, .76, .16}, {.14, .62, .46}, {.6, .25, .63}}
	c := cols[variant%7]
	w, d, h := 2.05, 4.3, 1.2
	switch variant % 7 {
	case 2:
		w = 2.5
		d = 5.8
		h = 1.7
	case 3:
		w = 2.7
		d = 7.3
		h = 2.45
	case 4:
		w = 2.7
		d = 5.7
		h = 2
	case 5:
		w = 2.5
		d = 9.0
		h = 3.1
	case 6:
		w = 2.7
		d = 7.8
		h = 2.8
	}
	if player {
		w = 2.15
		d = 4.7
		h = 1.12
		c = [3]float64{.95, .26, .08}
	}
	col(c[0]*.70, c[1]*.70, c[2]*.70, 1)
	box(-w/2, .2, -d/2, w/2, .70, d/2)
	col(c[0], c[1], c[2], 1)
	box(-w/2, .7, -d/2+.4, w/2, h+.28, d/2-.80)
	col(.07, .12, .20, 1)
	box(-w*.44, h*.70, -d*.23, w*.44, h+.32, d*.21)
	col(.17, .24, .30, 1)
	box(-w/2-.01, h*.7, -d*.17, -w/2+.02, h+.26, d*.20)
	box(w/2-.02, h*.7, -d*.17, w/2+.01, h+.26, d*.20)
	col(c[0]*1.13, c[1]*1.13, c[2]*1.13, 1)
	box(-w/2, .38, d/2-.76, w/2, .70, d/2+.06)
	col(.04, .045, .055, 1)
	box(-w/2, .21, -d/2-.1, w/2, .48, -d/2+.17)
	for _, xx := range []float64{-1, 1} {
		for _, zz := range []float64{-1, 1} {
			push()
			move(xx*w*.49, .39, zz*d*.32)
			rotate(app.t*360*math.Abs(app.world.Car.Speed)/15, 1, 0, 0)
			col(.055, .055, .062, 1)
			box(-.27, -.45, -.33, .27, .45, .33)
			col(.69, .73, .76, 1)
			box(xx*.26-.05, -.21, -.22, xx*.26+.05, .21, .22)
			pop()
		}
	}
	// headlights at nose, red rear lamps at back
	col(1, 1, .76, 1)
	box(-w*.38, .48, d/2+.06, -w*.14, .64, d/2+.11)
	box(w*.14, .48, d/2+.06, w*.38, .64, d/2+.11)
	if braking {
		col(1, .07, .02, 1)
	} else {
		col(.74, .07, .06, 1)
	}
	box(-w*.43, .53, -d/2-.13, -w*.16, .71, -d/2-.07)
	box(w*.16, .53, -d/2-.13, w*.43, .71, -d/2-.07)
	if player {
		col(.05, .05, .06, 1)
		box(-w*.56, .2, -d/2-.20, w*.56, .35, -d/2-.02)
		col(c[0]*.8, c[1]*.8, c[2]*.8, 1)
		box(-w*.48, h+.15, -d/2+.30, w*.48, h+.23, -d/2+.05)
		if app.world.Car.Boosting {
			col(.15, .60, 1, .75)
			box(-.61, .2, -d/2-2, -.27, .48, -d/2-.25)
			box(.27, .2, -d/2-2, .61, .48, -d/2-.25)
		}
	}
	pop()
}
func hudRect(x, y, w, h, r, g, b, a float64) {
	dis(GL_TEXTURE_2D)
	col(r, g, b, a)
	begin(GL_QUADS)
	vtx(x, y, 0)
	vtx(x+w, y, 0)
	vtx(x+w, y+h, 0)
	vtx(x, y+h, 0)
	end()
}
func hudText(s string, x, y, size float64, color [4]float64) {
	if usingFont035() {
		if beta.screen != screenDrive {
			size *= uiFontScale()
		}
		drawFont035(s, x, y, size*1.43, color)
		return
	}
	if beta.screen != screenDrive {
		size *= uiFontScale()
	}
	en(GL_TEXTURE_2D)
	bind("b037_font_fallback.png")
	col(color[0], color[1], color[2], color[3])
	begin(GL_QUADS)
	for _, ch := range s {
		if ch < 32 || ch >= 127 {
			ch = '?'
		}
		i := int(ch) - 32
		if ch == ' ' {
			x += size * .37
			continue
		}
		tx, ty := float64(i%16*80)/1280, float64(i/16*96)/576
		u0, v0 := tx, ty
		u1, v1 := tx+80./1280, ty+96./576
		width, height := size*.94, size*1.43
		tex(u0, v0)
		vtx(x, y, 0)
		tex(u1, v0)
		vtx(x+width, y, 0)
		tex(u1, v1)
		vtx(x+width, y+height, 0)
		tex(u0, v1)
		vtx(x, y+height, 0)
		x += size * (fontInkWidths[i]/80*.94 + .075)
	}
	end()
}

var white = [4]float64{1, 1, 1, 1}
var cyan = [4]float64{.28, .85, .86, 1}

func drawMinimap() {
	w := float64(app.width)
	x, y := w-245., 32.
	hudRect(x, y, 216, 202, .024, .055, .072, .73)
	dis(GL_TEXTURE_2D)
	centerX, centerY := x+108., y+101.
	scale := .21
	// Heading-up minimap: the player remains centered and forward is always up.
	// This fixes the previous mirrored / opposite-feeling GPS movement.
	project := func(p engine.Vec) (float64, float64) {
		right, forward := engine.MapRelative(app.world.Car.Pos, p, app.world.Car.Heading)
		return centerX + right*scale, centerY - forward*scale
	}
	// Keep roads and traffic pins inside the panel, not across the screen.
	en(GL_SCISSOR_TEST)
	gl("glScissor", uintptr(int(x+6)), uintptr(int(float64(app.height)-(y+196))), 204, 190)
	for _, r := range app.world.Routes {
		col(.44, .68, .66, .83)
		begin(GL_LINE_STRIP)
		for _, p := range r.Points {
			mx, my := project(p)
			vtx(mx, my, 0)
		}
		end()
	}
	for _, r := range app.world.Ramps {
		mx, my := project(r.Start)
		_, f := r.Sample(0)
		nx, ny := project(r.Start.Add(f.Mul(18)))
		dx, dy := nx-mx, ny-my
		length := math.Hypot(dx, dy)
		if length > 0 {
			dx /= length
			dy /= length
			col(1, .45, .1, 1)
			begin(GL_TRIANGLES)
			vtx(mx+dx*7, my+dy*7, 0)
			vtx(mx-dx*4-dy*4, my-dy*4+dx*4, 0)
			vtx(mx-dx*4+dy*4, my-dy*4-dx*4, 0)
			end()
		}
	}
	for _, t := range app.world.Traffic {
		if t.Pos.Dist(app.world.Car.Pos) > 120 {
			continue
		}
		mx, my := project(t.Pos)
		col(1, .55, .18, .95)
		begin(GL_QUADS)
		vtx(mx-2, my-2, 0)
		vtx(mx+2, my-2, 0)
		vtx(mx+2, my+2, 0)
		vtx(mx-2, my+2, 0)
		end()
	}
	col(.12, .94, 1, 1)
	begin(GL_TRIANGLES)
	vtx(centerX, centerY-14, 0)
	vtx(centerX-8, centerY+9, 0)
	vtx(centerX+8, centerY+9, 0)
	end()
	dis(GL_SCISSOR_TEST)
	hudText("MAP", x+15, y+10, 14, cyan)
}
func drawHUD() {
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	ow := float32(app.width)
	oh := float32(app.height)
	ortho := [16]float32{2 / ow, 0, 0, 0, 0, -2 / oh, 0, 0, 0, 0, -1, 0, -1, 1, 0, 1}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&ortho[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	en(GL_BLEND)
	w, h := float64(app.width), float64(app.height)
	hudRect(20, 18, 365, 77, .02, .045, .06, .67)
	hudText("SPEED IN NEED", 34, 25, 29, white)
	hudText("ALPHA BUILD v0.5.3  |  NATIVE WINDOWS", 36, 64, 13, cyan)
	drawMinimap()
	hudRect(w-320, h-172, 297, 140, .01, .04, .055, .82)
	speed := math.Abs(app.world.Car.Speed) * 3.6
	hudText(fmt.Sprintf("%03.0f", speed), w-302, h-167, 62, white)
	hudText("KM/H", w-144, h-145, 20, cyan)
	hudText(fmt.Sprintf("NITRO  %03.0f%%", app.world.Car.Nitro), w-295, h-83, 17, white)
	hudRect(w-295, h-52, 247, 15, .1, .23, .24, .9)
	hudRect(w-295, h-52, 247*app.world.Car.Nitro/100, 15, .05, .82, .98, 1)
	hudRect(22, h-143, 340, 117, .01, .03, .046, .75)
	hudText(fmt.Sprintf("DRIFT   %06.0f", app.world.Car.DriftScore), 35, h-132, 20, white)
	hudText(fmt.Sprintf("DAMAGE  %02.0f%%", app.world.Car.Damage), 35, h-104, 17, cyan)
	hudText(fmt.Sprintf("%s | %s", func() string {
		if app.boolMouse {
			return "MOUSE STEER"
		}
		return "WASD STEER"
	}(), func() string {
		if app.boolFPP {
			return "FPP"
		}
		return "TPP"
	}()), 35, h-75, 17, white)
	scene := []string{"COASTAL SUNSET", "NEON CITY", "CANYON ROADS"}[app.mapID]
	hudText(scene, 26, 109, 20, white)
	hudText(fmt.Sprintf("FPS %.0f", app.floatFPS), w-135, 247, 18, cyan)
	if app.world.Car.Boosting {
		hudRect(0, h*.55, w, 4, .08, .7, 1, .24)
		hudText("NITRO BOOST", w*.44, h*.25, 30, cyan)
	}
	if app.showHelp {
		hudRect(w*.19, h*.14, w*.62, h*.70, .02, .03, .045, .90)
		msgs := []string{"CONTROLS  -  H TO HIDE", "W / S   ACCELERATE / BRAKE", "A / D   KEYBOARD STEERING", "M       TOGGLE MOUSE STEERING", "RIGHT MOUSE  FREE-LOOK", "F       TPP / FPP CAMERA", "SPACE   HANDBRAKE + SMOKE", "SHIFT   LIMITED NITRO", "1 / 2 / 3  MAP SKY / WORLD", "V       VIDEO PRESET", "I       INVERT MOUSE X", "[ / ]   MOUSE SENSITIVITY", "UP / DOWN  CAMERA DISTANCE", "R       RESET    P  PAUSE"}
		for i, s := range msgs {
			hudText(s, w*.23, h*.18+float64(i)*31, 18, white)
		}
	}
	if !app.started || app.paused {
		hudRect(0, 0, w, h, .018, .027, .046, .59)
		hudRect(w*.18, h*.16, w*.64, h*.66, .02, .045, .07, .94)
		hudText("SPEED IN NEED", w*.23, h*.22, 53, white)
		hudText("BETA v0.1 - NATIVE FREE DRIVE", w*.24, h*.33, 25, cyan)
		if app.paused {
			hudText("PAUSED", w*.25, h*.43, 30, white)
			hudText("PRESS P TO CONTINUE", w*.25, h*.52, 25, cyan)
		} else {
			hudText("PRESS ENTER TO DRIVE", w*.25, h*.43, 32, white)
		}
		hudText("WASD DRIVE    SPACE HANDBRAKE    SHIFT NITRO", w*.24, h*.56, 20, white)
		hudText("V CAMERA / F RADIO    M MOUSE STEERING    H HELP", w*.24, h*.62, 20, cyan)
		hudText("1 COAST    2 NEON    3 CANYON", w*.24, h*.68, 20, white)
	}
	en(GL_DEPTH_TEST)
}
func drawFrame(dt float64) {
	qaLifecycle038()
	beginFrame038()
	updateAudio030()
	if remaster.Enabled {
		drawRemasteredFrame(dt)
		return
	}
	frameRenderStart := time.Now()
	if beta.screen != screenDrive {
		drawBetaMenuScene()
		presentFrame()
		return
	}
	if app.width <= 0 || app.height <= 0 {
		return
	}
	if app.staticList == 0 && !gpuWorld.ready {
		buildStatic()
	}
	c := engine.InterpolateCar(app.prevCar, app.world.Car, app.renderAlpha)
	app.renderCar = c
	h := c.Body.Height - .29
	// Follow a blend of vehicle heading and actual travel direction when drifting.
	velocityYaw := math.Atan2(math.Sin(c.Heading)*c.Speed+math.Cos(c.Heading)*c.Lateral, math.Cos(c.Heading)*c.Speed-math.Sin(c.Heading)*c.Lateral)
	blend := engine.Clamp(math.Abs(c.Lateral)/18, 0, .75)
	delta := math.Atan2(math.Sin(velocityYaw-c.Heading), math.Cos(velocityYaw-c.Heading))
	followYaw := c.Heading + delta*blend
	beta.cameraHeading += math.Atan2(math.Sin(followYaw-beta.cameraHeading), math.Cos(followYaw-beta.cameraHeading)) * (1 - math.Exp(-dt*6.6))
	hAngle := beta.cameraHeading + app.camYaw
	var target, want pos3
	if app.boolFPP {
		target = pos3{c.Pos.X + math.Sin(hAngle)*12, h + 1.45 + app.camPitch*2, c.Pos.Z + math.Cos(hAngle)*12}
		want = pos3{c.Pos.X + math.Sin(c.Heading)*.95, h + 1.6, c.Pos.Z + math.Cos(c.Heading)*.95}
	} else {
		lookAhead := engine.Clamp(math.Abs(c.Speed)*.072, 0, 2.8)
		target = pos3{c.Pos.X + math.Sin(followYaw)*lookAhead, h + 1.25, c.Pos.Z + math.Cos(followYaw)*lookAhead}
		dist := app.camDist + math.Abs(c.Speed)*.015
		if c.Boosting {
			dist += .6
		}
		want = pos3{c.Pos.X - math.Sin(hAngle)*dist, h + 1.9 + app.camPitch*4 + dist*.12, c.Pos.Z - math.Cos(hAngle)*dist}
	}
	// Exponential spring response avoids framerate-dependent camera jerks.
	if !app.boolFPP {
		targetXZ := engine.Vec{X: target.x, Z: target.z}
		desiredXZ := engine.Vec{X: want.x, Z: want.z}
		clearance := app.world.CameraObstructionFraction(targetXZ, desiredXZ)
		if clearance < 1 {
			want = add(target, scale(sub(want, target), clearance))
		}
	}
	app.cam = add(app.cam, scale(sub(want, app.cam), 1-math.Exp(-dt*10.8)))
	minY := engine.Terrain(app.cam.x, app.cam.z) + 1.25
	if app.cam.y < minY {
		app.cam.y = minY
	}
	fov := 65. + math.Abs(c.Speed)*.12
	boostTarget := 0.0
	if c.Boosting {
		boostTarget = 9
	}
	app.nitroFOV += (boostTarget - app.nitroFOV) * (1 - math.Exp(-dt*4.5))
	fov += app.nitroFOV
	beginGPUTiming()
	renderSunShadows(c)
	beginHDRScene()
	perspective(fov)
	gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	gl("glLoadIdentity")
	applyCamera(app.cam, target)
	drawAtmosphere(app.cam)
	// Atmosphere uses smooth gradients without tiled cloud artwork.
	gl("glFogi", GL_FOG_MODE, GL_LINEAR_FOG)
	fogStart := [1]float32{150}
	fogEnd := [1]float32{660}
	gl("glFogfv", GL_FOG_START, uintptr(unsafe.Pointer(&fogStart[0])))
	gl("glFogfv", GL_FOG_END, uintptr(unsafe.Pointer(&fogEnd[0])))
	fog := [4]float32{.43, .58, .62, 1}
	if app.mapID == 1 {
		fog = [4]float32{.08, .12, .25, 1}
	}
	if app.mapID == 2 {
		fog = [4]float32{.58, .54, .48, 1}
	}
	gl("glFogfv", GL_FOG_COLOR, uintptr(unsafe.Pointer(&fog[0])))
	en(GL_FOG)
	useLighting(app.lightingOn)
	updateSceneLights(c)
	if !drawBufferedWorld(app.cam, target) {
		gl("glCallList", uintptr(app.staticList))
		renderVisibleWorldChunks(app.cam, target)
	}
	drawPBRBuildings()
	renderPoles()
	renderVehicleLampMeshes()
	renderPickups()
	for _, t := range app.world.Traffic {
		if t.Pos.Dist(c.Pos) < 275 {
			drawCar(t.Pos, t.Heading, t.Variant, false, false)
		}
	}
	if !app.boolFPP {
		drawCar(c.Pos, c.Heading, 0, true, key('S'))
	} else { // view of vehicle hood
		push()
		move(c.Pos.X, c.Body.Height-.49, c.Pos.Z)
		rotate(c.Heading*180/math.Pi, 0, 1, 0)
		dis(GL_TEXTURE_2D)
		col(.83, .20, .07, 1)
		box(-1.03, .35, 1.25, 1.03, .68, 2.45)
		pop()
	}
	renderRearSmoke(c)
	useLighting(false)
	drawRemasterDebug(c)
	finishHDRScene()
	drawBetaHUD()
	endGPUTiming()
	app.renderMS = float64(time.Since(frameRenderStart).Microseconds()) / 1000
	presentFrame()
}
func main() {
	defer leaveLAN050()
	beginStartupProfile041()
	defer crashReport023()
	defer padRumble033(0)
	assetInit023.Do(initAssets023)
	if assetError023 != nil {
		message023("Keep SPEED_IN_NEED_v0.5.3.pak beside this EXE. Extract the complete game folder before playing.")
		return
	}
	initQA()
	app.world = engine.NewWorld(0)
	app.quality = 1
	app.sensitivity = 1
	app.camDist = 5.8
	app.boolFPP = false
	beta.bindWaiting = -1
	readCfg()
	initializeProfiles040(filepath.Join(filepath.Dir(cfgPath()), "profiles.json"))
	defer saveProfileSettings040()
	initializeRemasterSettings()
	configureUnityHost040()
	initUnityFrame040()
	defer closeUnityFrame040()
	for _, arg := range os.Args[1:] {
		if arg == "--balanced" {
			applyAokawaPreset(1)
		}
		if arg == "--rtx3060" {
			applyHigh035()
		}
		if arg == "--cinematic" || arg == "--ultra" {
			applyUltra035()
		}
		if arg == "--safe" || arg == "--performance" || arg == "--integrated" {
			applyAokawaPreset(0)
			beta.cfg.Resolution = 0
			beta.cfg.Fullscreen = false
			beta.cfg.FrameCap = 60
			beta.cfg.VSync = false
			beta.cfg.AdaptiveQuality = true
			beta.cfg.Lightweight = arg == "--integrated"
			if arg == "--integrated" {
				applyIntegrated035()
			}
		}
	}
	if qa.enabled {
		beta.cfg = defaultCfg()
		app.quality = 1
		app.boolMouse = false
		app.sensitivity = 1
		app.camDist = 6
	}
	if qa.enabled {
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_PRESET")); e == nil && n >= 0 && n <= 3 {
			applyAokawaPreset(n)
		}
	}
	if qa.enabled {
		beta.cfg.Lightweight = os.Getenv("SPEED_QA_LIGHTWEIGHT") == "1"
		if n, e := strconv.Atoi(os.Getenv("SPEED_QA_TEXTURE_BUDGET")); e == nil && n >= 0 && n <= 4 {
			beta.cfg.TextureBudget = n
		}
	}
	app.cam = pos3{0, 5, -232}
	if qa.enabled && os.Getenv("SPEED_QA_RTX3060") == "1" {
		applyHigh035()
	}
	app.run = true
	if qa.enabled {
		qaProfile035()
	}
	queryGPU038()
	launchOptions038()
	rendererRequested038 = beta.cfg.RendererBackend
	for _, arg := range os.Args[1:] {
		if arg == "--dx12" {
			rendererRequested038 = 1
			beta.cfg.RendererBackend = 1
		}
		if arg == "--opengl" {
			rendererRequested038 = 0
			beta.cfg.RendererBackend = 0
		}
	}
	unityStage040("creating window")
	initWindow()
	unityStage040("OpenGL window ready")
	if beta.cfg.Fullscreen && !qa.enabled {
		toggleFullscreen()
	}
	setupGL()
	unityStage040("OpenGL setup ready")
	initBackend038()
	unityStage040("backend ready")
	createWhiteTexture()
	initLighting()
	initPBRAssets()
	initRemasterGraphics()
	unityStage040("graphics assets ready")
	startAudio030()
	defer stopAudio030()
	defer writeAudioReport030()
	installBetaWorldTextures()
	if !remaster.Enabled {
		initBufferedCars()
		initBufferedWheels()
	}
	initGPUTiming()
	applyAokawaVSync()
	if qa.enabled {
		glExt("wglSwapIntervalEXT", 0)
		app.vsync = false
	}
	app.lightingOn = beta.cfg.Lighting
	for i := 0; i < 8; i++ {
		if remaster.Enabled {
			continue
		}
		if i < 2 && bufferedCars[i].available {
			continue
		}
		compileMesh(i)
	}
	if !remaster.Enabled {
		compileWheelMeshes()
		compileTrafficFinishes()
	}
	beta.screen = screenTitle
	app.prevCar = app.world.Car
	app.showHelp = false
	last := time.Now()
	var physicsClock038 engine.FixedClock038
	frames := 0
	frameTimer := last
	unityStage040("render loop ready")
	for app.run {
		if !app.minimized || qa.enabled {
			latencyBegin038()
		}
		beta020SyncPointer()
		if qa.enabled && !qa.started {
			qaSelectStage()
			qa.started = true
		}
		var m msg
		for win(user32, "PeekMessageW", uintptr(unsafe.Pointer(&m)), 0, 0, 0, PM_REMOVE) != 0 {
			if m.message == 0x12 {
				app.run = false
				break
			}
			win(user32, "TranslateMessage", uintptr(unsafe.Pointer(&m)))
			win(user32, "DispatchMessageW", uintptr(unsafe.Pointer(&m)))
		}
		if app.minimized && !qa.enabled {
			invalidateTemporal038()
			time.Sleep(16 * time.Millisecond)
			last = time.Now()
			continue
		}
		now := time.Now()
		dt := now.Sub(last).Seconds()
		physicsElapsed038 := dt
		app.frameMS = dt * 1000
		recordFrame(app.frameMS)
		last = now
		if dt > .075 {
			dt = .075
		}
		if dt < 0 {
			dt = 0
		}
		updateLAN050(dt)
		updateGamepad033(dt)
		updateNavigation033(dt)
		if !photo041.Active {
			advanceCycle034(dt)
		}
		if app.started && beta.screen == screenDrive && !remaster.FullMap && !photo041.Active {
			// Vehicle-local steering is independent of camera yaw and mouse settings.
			steer := aokawaKeyboardSteer(pressed(2), pressed(3))
			beta013CaptureLook()
			in := engine.Input{Steer: steer, Boost: pressed(4), Handbrake: pressed(5)}
			if pressed(0) {
				in.Throttle = 1
			}
			if pressed(1) {
				in.Brake = 1
			}
			in = mergeGamepad033(in)
			in = mergeAutopilot034(in)
			if multiplayer050.Open || garage050.Open || beta.screen != screenDrive || remaster.FullMap || photo041.Active {
				in = engine.Input{Brake: 1, Handbrake: true}
			}
			if qa.enabled {
				in = qaInput()
				if qa.stage == 138 {
					in.Steer = .65
				}
			}
			// Physics matches the visible tree budget at the selected graphics preset.
			treeBudget := []int{120, 230, 360}[engine.ClampInt(app.quality, 0, 2)]
			app.world.VisibleTrees = engine.ClampInt(treeBudget, 0, len(app.world.Trees))
			simStart := time.Now()
			physicsClock038.Advance(physicsElapsed038, func(fixedDT float64) {
				app.prevCar = app.world.Car
				applyGarage050()
				stepPlayer051(in, fixedDT)
				resolveLANContacts050()
				if multiplayer050.Session == nil && !workshop051.Active {
					advanceProgress040(fixedDT)
				}
			})
			app.simMS = float64(time.Since(simStart).Microseconds()) / 1000
			app.renderAlpha = physicsClock038.Alpha()
			app.t += dt
			beta.idleMouse += dt
			if beta.cfg.CameraAutoReturnRefine && beta.idleMouse > 4 && !app.camHeld {
				app.camYaw *= math.Exp(-dt * 1.3)
			}
			checkFinish()
		} else {
			physicsClock038.Pause()
			app.prevCar = app.world.Car
			app.renderAlpha = 1
			if !remasterSceneOverlay() {
				app.t += dt * .3
			}
		}
		updateLANRace050(dt)
		publishLAN050()
		latencyMarker038(1)
		latencyMarker038(2)
		drawFrame(dt)
		unityFirstFrame040()
		updateAdaptive026(dt, float64(time.Since(now).Microseconds())/1000)
		beta.menuTime += dt
		if beta.screen == screenCars || beta.screen == screenTitle {
			beta.carTurn += dt
		}
		frames++
		if time.Since(frameTimer) > time.Second {
			app.floatFPS = float64(frames) / time.Since(frameTimer).Seconds()
			frames = 0
			frameTimer = time.Now()
		}
		// Honor the cap even with VSync enabled, without a busy wait.
		paceFrame026(now)
	}
	backend038.Close()
	releaseRemasterBuffers()
	shutdownGPUTiming()
	finishBenchmark()
	invalidateWorldGeometry()
	releaseBufferedCars()
	releaseSceneResources()
	win(ogl, "wglMakeCurrent", 0, 0)
	win(ogl, "wglDeleteContext", app.rc)
	win(user32, "ReleaseDC", app.hwnd, app.dc)
}

func renderStaticDetails() {
	renderRoundedHills()
	if app.mapID == 0 {
		renderCoastalWater()
	}
	renderRoads()
	renderRamps()
	// Solid roadside boulders share exact radius/position with physics.
	bind("rock.jpg")
	en(GL_TEXTURE_2D)
	for _, rock := range app.world.Rocks {
		push()
		move(rock.Pos.X, engine.Terrain(rock.Pos.X, rock.Pos.Z)-.18, rock.Pos.Z)
		for j := 0; j < 9; j++ {
			a, b := 2*math.Pi*float64(j)/9, 2*math.Pi*float64(j+1)/9
			r := rock.Radius
			col(.60+float64(j%3)*.07, .62+float64(j%3)*.06, .60+float64(j%3)*.05, 1)
			begin(GL_QUADS)
			tq(pos3{math.Cos(a) * r, 0, math.Sin(a) * r}, pos3{math.Cos(b) * r, 0, math.Sin(b) * r}, pos3{math.Cos(b) * r * .49, rock.Height, math.Sin(b) * r * .49}, pos3{math.Cos(a) * r * .49, rock.Height, math.Sin(a) * r * .49}, float64(j)/9, 1, float64(j+1)/9, 0)
			end()
		}
		pop()
	}
	if pbrAssets["building0"] == nil || app.quality == 0 {
		for i, p := range app.world.Buildings {
			renderBuilding(p.X, p.Z, i)
		}
	}

}
