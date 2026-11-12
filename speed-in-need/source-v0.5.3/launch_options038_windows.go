package main

import "os"

// Launch profiles change this session's existing settings. Capability checks
// remain authoritative; a command-line label never unlocks unsupported SDKs.
func launchOptions038() {
	if qa.enabled {
		return
	}
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--uncapped":
			beta.cfg.VSync = false
			beta.cfg.FrameCap = 0
		case "--dlss-quality", "--dlss-balanced", "--dlss-performance", "--dlaa":
			if gpu038.DLSS == 0 && gpu038.SLInit == 0 {
				beta.cfg.RendererBackend = 1
				beta.cfg.Upscaling038 = map[string]int{"--dlss-quality": 1, "--dlss-balanced": 2, "--dlss-performance": 3, "--dlaa": 4}[arg]
			}
		case "--fsr-fg":
			if gpu038.FSRSupported {
				beta.cfg.RendererBackend = 1
				beta.cfg.FrameGeneration038 = 1
			}
		case "--rt-shadows", "--rt-reflections":
			if gpu038.RayTier >= 11 {
				beta.cfg.RendererBackend = 1
				beta.cfg.RayShadows038 = 1
				if arg == "--rt-reflections" {
					beta.cfg.RayShadows038 = 2
				}
			}
		case "--rt-low", "--rt-medium", "--rt-high", "--rt-ultra":
			beta.cfg.RayQuality038 = map[string]int{"--rt-low": 0, "--rt-medium": 1, "--rt-high": 2, "--rt-ultra": 3}[arg]
		}
	}
}
