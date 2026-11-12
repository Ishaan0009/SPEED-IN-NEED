package main

import "os"

// Offline renderer auditing must not initialize vendor SDKs, native capability
// probes or a network session. It uses local assets and the OS OpenGL driver.
func offlineAudit053() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--offline-audit" {
			return true
		}
	}
	return false
}

func nativeSDKDisabled053() bool {
	if offlineAudit053() {
		return true
	}
	for _, arg := range os.Args[1:] {
		if arg == "--opengl" {
			return true
		}
	}
	return false
}
