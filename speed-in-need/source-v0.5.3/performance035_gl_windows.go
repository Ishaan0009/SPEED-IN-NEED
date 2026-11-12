package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type nativeProcKey035 struct {
	DLL  *syscall.LazyDLL
	Name string
}

// Audio and the render loop both use win(). A LazyProc resolves once and is
// safe to reuse across goroutines, while DLL identity keeps exports isolated.
type nativeProcCache035 struct{ entries sync.Map }

func (c *nativeProcCache035) lookup(d *syscall.LazyDLL, name string) *syscall.LazyProc {
	key := nativeProcKey035{d, name}
	if proc, ok := c.entries.Load(key); ok {
		return proc.(*syscall.LazyProc)
	}
	proc, _ := c.entries.LoadOrStore(key, d.NewProc(name))
	return proc.(*syscall.LazyProc)
}

var nativeProcedures035 nativeProcCache035

type textureCallCache035 struct {
	context                       uintptr
	maxUnits                      int
	unitLimitSource               string
	active                        int
	activeKnown                   bool
	bindings                      [32]uintptr
	known                         [32]bool
	compilingList                 bool
	insideBegin                   bool
	activeSkipped, bindingSkipped uint64
	materialPairsSkipped          uint64
}

func (c *textureCallCache035) invalidate() {
	c.activeKnown = false
	c.known = [32]bool{}
}

func (c *textureCallCache035) resetContext(context uintptr) {
	c.context, c.maxUnits, c.compilingList, c.insideBegin, c.unitLimitSource = context, 0, false, false, ""
	c.invalidate()
}

func (c *textureCallCache035) canElideBinding(unit int, texture uintptr) bool {
	return c.context != 0 && !c.compilingList && !c.insideBegin && unit >= 0 && unit < c.maxUnits && unit < len(c.bindings) && c.known[unit] && c.bindings[unit] == texture
}

// Called only for available functions, immediately before their real driver
// call. Unknown state is always passed through until it is explicitly set.
func (c *textureCallCache035) skip(name string, args []uintptr) bool {
	if c.context == 0 {
		return false
	}
	switch name {
	case "glBegin":
		c.insideBegin = true
	case "glEnd":
		c.insideBegin = false
	case "glNewList":
		c.invalidate()
		c.compilingList = true
	case "glEndList":
		c.compilingList = false
		c.invalidate()
	case "glCallList", "glCallLists", "glPushAttrib", "glPopAttrib", "glBindTextures", "glBindTextureUnit", "glBindMultiTextureEXT":
		// A list or restored attribute stack can change active units/bindings
		// without executing individual commands through our wrappers.
		c.invalidate()
	case "glDeleteTextures":
		// Invalidate all units rather than inspecting a caller-owned pointer.
		// The driver may immediately recycle any deleted texture name.
		c.known = [32]bool{}
	case "glActiveTexture", "glActiveTextureARB":
		if c.compilingList {
			return false
		}
		if c.insideBegin {
			c.invalidate()
			return false
		}
		if len(args) != 1 || args[0] < 0x84C0 || c.maxUnits <= 0 || args[0]-0x84C0 >= uintptr(c.maxUnits) || args[0]-0x84C0 >= uintptr(len(c.bindings)) {
			// Invalid/unsupported unit calls can leave the driver's old unit
			// selected. Subsequent binds must not become known for that unit.
			c.invalidate()
			return false
		}
		unit := int(args[0] - 0x84C0)
		if c.activeKnown && c.active == unit {
			c.activeSkipped++
			return true
		}
		c.active, c.activeKnown = unit, true
	case "glBindTexture":
		if c.compilingList {
			return false
		}
		if c.insideBegin {
			c.invalidate()
			return false
		}
		if len(args) != 2 {
			c.invalidate()
			return false
		}
		if args[0] != GL_TEXTURE_2D || !c.activeKnown {
			return false
		}
		if args[1] > 0xffffffff {
			c.invalidate()
			return false
		}
		if c.known[c.active] && c.bindings[c.active] == args[1] {
			c.bindingSkipped++
			return true
		}
		c.bindings[c.active], c.known[c.active] = args[1], true
	}
	return false
}

var textureCalls035 textureCallCache035
var textureOwnerThread035 atomic.Uint32
var textureCacheConcurrent035 atomic.Bool

func nativeContextChange035(d *syscall.LazyDLL, name string, args []uintptr, result uintptr) {
	if d != ogl || (name != "wglMakeCurrent" && name != "wglDeleteContext" && name != "wglShareLists") {
		return
	}
	id, _, _ := nativeProcedures035.lookup(kernel32, "GetCurrentThreadId").Call()
	thread := uint32(id)
	textureOwnerThread035.CompareAndSwap(0, thread)
	if textureOwnerThread035.Load() != thread || name == "wglShareLists" {
		// Current contexts are thread-local. Conservatively stop suppression
		// if a second GL thread or shared context could mutate tracked state.
		textureCacheConcurrent035.Store(true)
		return
	}
	if textureCacheConcurrent035.Load() {
		return
	}
	context := uintptr(0)
	if name == "wglMakeCurrent" && result != 0 && len(args) == 2 {
		context = args[1]
	}
	// Reset even on failure, release, deletion, or switching back to an old
	// context: its state may have been changed since it was last observed.
	textureCalls035.resetContext(context)
}

func skipNativeGL035(name string, args []uintptr) bool {
	if textureCacheConcurrent035.Load() || textureCalls035.context == 0 {
		return false
	}
	if (name == "glActiveTexture" || name == "glActiveTextureARB") && textureCalls035.maxUnits == 0 && !textureCalls035.compilingList && !textureCalls035.insideBegin {
		var units int32
		// The GLSL 1.20 renderer uses shader texture units beyond the fixed
		// function limit (four on this GPU). ActiveTexture accepts the
		// combined image-unit limit, so those bindings can be tracked too.
		nativeProcedures035.lookup(ogl, "glGetIntegerv").Call(0x8B4D, uintptr(unsafe.Pointer(&units)))
		textureCalls035.unitLimitSource = "GL_MAX_COMBINED_TEXTURE_IMAGE_UNITS"
		if units <= 0 {
			nativeProcedures035.lookup(ogl, "glGetIntegerv").Call(0x84E2, uintptr(unsafe.Pointer(&units)))
			textureCalls035.unitLimitSource = "GL_MAX_TEXTURE_UNITS fallback"
		}
		runtime.KeepAlive(&units)
		textureCalls035.maxUnits = textureUnitLimit035(int(units), len(textureCalls035.bindings))
	}
	return textureCalls035.skip(name, args)
}

func textureUnitLimit035(driverUnits, capacity int) int {
	if driverUnits <= 0 || capacity <= 0 {
		return -1
	}
	return min(driverUnits, capacity)
}

// Material binding always restores unit zero afterwards. A texture already
// bound to its destination unit needs neither an active-unit switch nor a bind.
// On a cache miss, select the destination BEFORE lazy texture decoding/upload:
// pbrTexture binds the new texture as part of uploading it.
func bindPBRTexture035(unit int, name string) {
	texture, cached := pbrTextures[name]
	if name == "" {
		texture, cached = app.whiteTexture, true
	}
	if cached && capture == nil && !textureCacheConcurrent035.Load() && textureCalls035.canElideBinding(unit, uintptr(texture)) {
		textureCalls035.materialPairsSkipped++
		return
	}
	glExt("glActiveTexture", uintptr(0x84C0+unit))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture(name)))
}

func nativeGLStats035() map[string]any {
	return map[string]any{
		"active_texture_calls_skipped":      textureCalls035.activeSkipped,
		"texture_binding_calls_skipped":     textureCalls035.bindingSkipped,
		"tracked_texture_units":             textureCalls035.maxUnits,
		"texture_unit_limit_source":         textureCalls035.unitLimitSource,
		"material_texture_pairs_skipped":    textureCalls035.materialPairsSkipped,
		"concurrent_context_cache_disabled": textureCacheConcurrent035.Load(),
	}
}
