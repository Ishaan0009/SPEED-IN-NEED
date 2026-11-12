package main

import (
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
)

func testTextureCache035() textureCallCache035 {
	var cache textureCallCache035
	cache.resetContext(1)
	cache.maxUnits = 8
	return cache
}

func TestGLCache035TracksIndependentTextureUnits(t *testing.T) {
	c := testTextureCache035()
	call := func(name string, args ...uintptr) bool { return c.skip(name, args) }
	if call("glBindTexture", GL_TEXTURE_2D, 9) {
		t.Fatal("unknown active unit cannot be suppressed")
	}
	if call("glActiveTexture", 0x84C0) || !call("glActiveTexture", 0x84C0) {
		t.Fatal("active unit first/repeat")
	}
	if call("glBindTexture", GL_TEXTURE_2D, 9) || !call("glBindTexture", GL_TEXTURE_2D, 9) {
		t.Fatal("2D bind first/repeat")
	}
	if call("glActiveTexture", 0x84C1) || call("glBindTexture", GL_TEXTURE_2D, 9) || !call("glBindTexture", GL_TEXTURE_2D, 9) {
		t.Fatal("unit1 binding is independent")
	}
	if call("glActiveTexture", 0x84C0) || !call("glBindTexture", GL_TEXTURE_2D, 9) {
		t.Fatal("unit0 binding was not retained")
	}
	if call("glBindTexture", 0x806F, 9) {
		t.Fatal("only 2D bindings are tracked")
	}
	if c.activeSkipped != 1 || c.bindingSkipped != 3 {
		t.Fatal("incorrect instrumentation", c.activeSkipped, c.bindingSkipped)
	}
}

func TestGLCache035ShaderUnitsUseCombinedLimit(t *testing.T) {
	for _, test := range [][3]int{{192, 32, 32}, {16, 32, 16}, {4, 32, 4}, {0, 32, -1}, {-1, 32, -1}, {16, 0, -1}} {
		if got := textureUnitLimit035(test[0], test[1]); got != test[2] {
			t.Fatal("driver limit must be bounded", test, got)
		}
	}
	c := testTextureCache035()
	c.maxUnits = textureUnitLimit035(192, len(c.bindings))
	for _, unit := range []uintptr{0x84C7, 0x84CE, 0x84CF} {
		if c.skip("glActiveTexture", []uintptr{unit}) {
			t.Fatal("first shader-unit activation skipped")
		}
		if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 49}) || !c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 49}) {
			t.Fatal("shader unit outside fixed limit was not cached", unit)
		}
	}
	if c.skip("glActiveTexture", []uintptr{0x84C0 + 32}) {
		t.Fatal("unit outside array must pass through")
	}
}

func TestGLCache035MaterialBindingElisionPreservesActiveUnit(t *testing.T) {
	c := testTextureCache035()
	c.skip("glActiveTexture", []uintptr{0x84C0})
	c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 17})
	c.skip("glActiveTexture", []uintptr{0x84C1})
	c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 23})
	if !c.canElideBinding(0, 17) || !c.canElideBinding(1, 23) || c.active != 1 {
		t.Fatal("known destination binding must not switch the active unit")
	}
	for _, unit := range []int{-1, 2, 8, 32} {
		if c.canElideBinding(unit, 17) {
			t.Fatal("untracked material unit was elided", unit)
		}
	}
	if c.canElideBinding(0, 23) {
		t.Fatal("different texture must be bound")
	}
	c.compilingList = true
	if c.canElideBinding(0, 17) {
		t.Fatal("material binding omitted from display list")
	}
	c.compilingList = false
	c.insideBegin = true
	if c.canElideBinding(0, 17) {
		t.Fatal("invalid begin/end command was hidden")
	}
	c.insideBegin = false
	c.skip("glDeleteTextures", nil)
	if c.canElideBinding(0, 17) {
		t.Fatal("deleted or recycled material texture was elided")
	}
}

func TestGLCache035DeletionAndContextReuse(t *testing.T) {
	c := testTextureCache035()
	c.skip("glActiveTexture", []uintptr{0x84C0})
	c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 91})
	c.skip("glDeleteTextures", []uintptr{1, 0}) // no caller pointer inspection
	if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 91}) {
		t.Fatal("deleted/recycled texture name was skipped")
	}
	for _, context := range []uintptr{2, 1, 0, 1} {
		c.resetContext(context)
		c.maxUnits = 8
		if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 91}) {
			t.Fatal("stale binding across context switch")
		}
		if c.skip("glActiveTexture", []uintptr{0x84C0}) {
			t.Fatal("stale active unit across context switch")
		}
		if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 91}) {
			t.Fatal("first binding in context must execute")
		}
	}
}

func TestGLCache035InvalidUnitsCannotMakeStaleBindings(t *testing.T) {
	for _, unit := range []uintptr{0, 0x84BF, 0x84C8, 0x84C0 + 32, ^uintptr(0)} {
		c := testTextureCache035()
		c.skip("glActiveTexture", []uintptr{0x84C0})
		c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 3})
		if c.skip("glActiveTexture", []uintptr{unit}) {
			t.Fatal("invalid/untracked unit was suppressed", unit)
		}
		c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 11})
		c.skip("glActiveTexture", []uintptr{0x84C0})
		if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 3}) {
			t.Fatal("unknown-unit bind polluted cache", unit)
		}
	}
	c := testTextureCache035()
	c.skip("glActiveTexture", []uintptr{0x84C0})
	for _, malformed := range [][]uintptr{nil, {GL_TEXTURE_2D}, {GL_TEXTURE_2D, 1, 2}} {
		if c.skip("glBindTexture", malformed) {
			t.Fatal("malformed call suppressed")
		}
	}
}

func TestGLCache035DisplayListsAndExternalStateRestore(t *testing.T) {
	c := testTextureCache035()
	c.skip("glActiveTexture", []uintptr{0x84C0})
	c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5})
	c.skip("glNewList", []uintptr{99, GL_COMPILE})
	for i := 0; i < 3; i++ {
		if c.skip("glActiveTexture", []uintptr{0x84C0}) || c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5}) {
			t.Fatal("recorded list command was omitted")
		}
	}
	c.skip("glEndList", nil)
	for _, boundary := range []string{"glCallList", "glCallLists", "glPopAttrib", "glPushAttrib", "glBindTextures", "glBindTextureUnit", "glBindMultiTextureEXT"} {
		c.skip("glActiveTexture", []uintptr{0x84C0})
		c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5})
		c.skip(boundary, nil)
		if c.skip("glActiveTexture", []uintptr{0x84C0}) || c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5}) {
			t.Fatal("stale cache after hidden state change", boundary)
		}
	}
}

func TestGLCache035BeginEndNeverSuppressesInvalidStateCalls(t *testing.T) {
	c := testTextureCache035()
	c.skip("glActiveTexture", []uintptr{0x84C0})
	c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5})
	c.skip("glBegin", []uintptr{GL_TRIANGLES})
	if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5}) || c.skip("glActiveTexture", []uintptr{0x84C0}) {
		t.Fatal("invalid begin/end state call was hidden")
	}
	c.skip("glEnd", nil)
	c.skip("glActiveTexture", []uintptr{0x84C0})
	if c.skip("glBindTexture", []uintptr{GL_TEXTURE_2D, 5}) {
		t.Fatal("unknown state after invalid call")
	}
}

func TestGLCache035ContextHooksDisableAcrossRenderThreads(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldCache, oldOwner, oldDisabled := textureCalls035, textureOwnerThread035.Load(), textureCacheConcurrent035.Load()
	defer func() {
		textureCalls035 = oldCache
		textureOwnerThread035.Store(oldOwner)
		textureCacheConcurrent035.Store(oldDisabled)
	}()
	textureCalls035 = textureCallCache035{}
	textureOwnerThread035.Store(0)
	textureCacheConcurrent035.Store(false)
	nativeContextChange035(ogl, "wglMakeCurrent", []uintptr{1, 11}, 1)
	if textureCalls035.context != 11 {
		t.Fatal("context was not captured")
	}
	textureCalls035.maxUnits = 8
	textureCalls035.skip("glActiveTexture", []uintptr{0x84C0})
	nativeContextChange035(ogl, "wglMakeCurrent", []uintptr{1, 12}, 1)
	if textureCalls035.context != 12 || textureCalls035.activeKnown || textureCalls035.maxUnits != 0 {
		t.Fatal("context switch kept stale state")
	}
	nativeContextChange035(ogl, "wglMakeCurrent", []uintptr{1, 11}, 0)
	if textureCalls035.context != 0 {
		t.Fatal("failed context transition retained a cache")
	}
	nativeContextChange035(ogl, "wglMakeCurrent", []uintptr{1, 11}, 1)
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		nativeContextChange035(ogl, "wglMakeCurrent", []uintptr{2, 22}, 1)
		close(done)
	}()
	<-done
	if !textureCacheConcurrent035.Load() || skipNativeGL035("glActiveTexture", []uintptr{0x84C0}) {
		t.Fatal("multiple GL threads must disable shared suppression")
	}
}

func TestNativeProc035ConcurrentLookupAndDLLIsolation(t *testing.T) {
	var cache nativeProcCache035
	const workers = 24
	results := make(chan *syscall.LazyProc, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- cache.lookup(kernel32, "GetCurrentProcessId") }()
	}
	wg.Wait()
	close(results)
	first := cache.lookup(kernel32, "GetCurrentProcessId")
	for result := range results {
		if result != first {
			t.Fatal("concurrent lookup created separate live entries")
		}
	}
	otherDLL := syscall.NewLazyDLL("kernel32.dll")
	if cache.lookup(otherDLL, "GetCurrentProcessId") == first {
		t.Fatal("different DLL objects share lookup key")
	}
	process, _, _ := first.Call()
	if int(process) != os.Getpid() {
		t.Fatal("cached export called wrong procedure", process)
	}
}

var benchmarkProcResult035 uintptr

func BenchmarkNativeProc035(b *testing.B) {
	for _, cached := range []bool{false, true} {
		name := "ResolveEveryCall"
		if cached {
			name = "CachedLazyProc"
		}
		b.Run(name, func(b *testing.B) {
			var cache nativeProcCache035
			for i := 0; i < b.N; i++ {
				var proc *syscall.LazyProc
				if cached {
					proc = cache.lookup(kernel32, "GetCurrentProcessId")
				} else {
					proc = kernel32.NewProc("GetCurrentProcessId")
				}
				benchmarkProcResult035, _, _ = proc.Call()
			}
		})
	}
}
