//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"unsafe"
)

var debugCallback023 uintptr

func initGLDebug023() {
	if debugCallback023 != 0 || !qa.enabled || os.Getenv("SPEED_QA_GL_DEBUG") != "1" || !glExtAvailable("glDebugMessageCallback") {
		return
	}
	debugCallback023 = syscall.NewCallback(func(source, kind, id, severity, length uintptr, message *byte, user uintptr) uintptr {
		if kind != 0x824C {
			return 0
		}
		n := int(length)
		if n < 0 || n > 4096 {
			return 0
		}
		text := string(unsafe.Slice(message, n))
		f, err := os.OpenFile(filepath.Join(qa.dir, "GL_DIAGNOSTIC_BETA_023.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			fmt.Fprintf(f, "stage=%d frame=%d id=%d: %s\n%s\n", qa.stage, qa.frame, id, text, debug.Stack())
			f.Close()
		}
		return 0
	})
	gl("glEnable", 0x92E0)
	gl("glEnable", 0x8242)
	glExt("glDebugMessageCallback", debugCallback023, 0)
}
