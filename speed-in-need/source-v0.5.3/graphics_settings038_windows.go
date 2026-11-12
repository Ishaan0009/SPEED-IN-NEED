package main

import "fmt"

var graphicsLabels038 = []string{"RENDERER / RESTART", "DETECTED GPU", "DEDICATED VIDEO MEMORY", "UPSCALING / ANTI-ALIASING", "FRAME GENERATION / RESTART TO ENABLE", "NVIDIA FRAME GENERATION", "RAY TRACING", "RAY-TRACING QUALITY", "AUTO VRAM BUDGET", "ROAD RECOVERY"}

func graphicsValues038() []string {
	device := gpu038.Adapter
	if device == "" {
		device = "GPU QUERY UNAVAILABLE"
	}
	sr, fg, rt := "UNAVAILABLE / DEVICE OR RUNTIME", "UNAVAILABLE / DEVICE OR DRIVER", "UNAVAILABLE / DEVICE"
	if gpu038.DLSS == 0 && gpu038.SLInit == 0 {
		sr = []string{"NATIVE", "DLSS QUALITY", "DLSS BALANCED", "DLSS PERFORMANCE", "DLAA / NATIVE", "DLSS ULTRA PERFORMANCE / 4K"}[beta.cfg.Upscaling038]
		if beta.cfg.RendererBackend != 1 {
			sr = "SELECT DX12 AND RESTART"
		}
	}
	if gpu038.NVIDIAFG == 0 && gpu038.SLInit == 0 {
		fg = "SDK SUPPORTED / HARDWARE TEST PENDING"
	}
	if gpu038.RayTier >= 11 {
		rt = []string{"OFF / RASTER SHADOWS", "SOFT SUN SHADOWS", "SOFT SHADOWS + LOCAL PAINT REFLECTIONS"}[beta.cfg.RayShadows038]
		if beta.cfg.RendererBackend != 1 {
			rt = "SELECT DX12 AND RESTART"
		}
	}
	backend := "OPENGL / STABLE"
	if beta.cfg.RendererBackend == 1 {
		backend = "DX12 HYBRID / EXPERIMENTAL"
	}
	framegen := "UNAVAILABLE / RUNTIME"
	if gpu038.FSRSupported {
		framegen = []string{"OFF", "AMD FSR " + gpu038.FSRProvider + " / AUTO LOW-FPS GUARD"}[beta.cfg.FrameGeneration038]
		if beta.cfg.RendererBackend != 1 {
			framegen = "SELECT DX12 AND RESTART"
		}
	}
	if rendererError038 != "" {
		backend = "OPENGL FALLBACK / RESTART TO RETRY"
		sr = "UNAVAILABLE IN THIS SESSION"
		framegen = sr
		rt = sr
	}
	if temporalTransport038.Error != "" {
		sr = "DISABLED AFTER A RENDERING ERROR"
		rt = sr
	}
	return []string{backend, device, fmt.Sprintf("%.1f GiB / %dK TEXTURE CAP", float64(gpu038.Dedicated)/(1<<30), gpuTextureCap038()/1024), sr, framegen, fg, rt, []string{"LOW", "MEDIUM", "HIGH", "ULTRA"}[rayQuality038()], onoff(beta.cfg.GPUTextureAuto), "HOLD D-PAD DOWN / 1 SECOND"}
}
func adjustGraphics038(delta int) bool {
	if remaster.SettingsTab != 8 {
		return false
	}
	switch beta.selection {
	case 0:
		if gpu038.DX12 {
			beta.cfg.RendererBackend = (beta.cfg.RendererBackend + 1) % 2
			saveCfg()
		}
	case 8:
		beta.cfg.GPUTextureAuto = !beta.cfg.GPUTextureAuto
		saveCfg()
	case 3:
		if gpu038.DLSS == 0 && gpu038.SLInit == 0 && beta.cfg.RendererBackend == 1 {
			count := 5
			if app.width >= 3840 {
				count = 6
			}
			beta.cfg.Upscaling038 = (beta.cfg.Upscaling038 + delta + count) % count
			destroyHDR()
			saveCfg()
		}
	case 4:
		if gpu038.FSRSupported && beta.cfg.RendererBackend == 1 {
			beta.cfg.FrameGeneration038 = 1 - beta.cfg.FrameGeneration038
			invalidateTemporal038()
			saveCfg()
		}
	case 6:
		if gpu038.RayTier >= 11 && beta.cfg.RendererBackend == 1 {
			beta.cfg.RayShadows038 = (beta.cfg.RayShadows038 + delta + 3) % 3
			destroyHDR()
			saveCfg()
		}
	case 7:
		if gpu038.RayTier >= 11 && beta.cfg.RendererBackend == 1 {
			beta.cfg.RayQuality038 = (beta.cfg.RayQuality038 + delta + 4) % 4
			resetRayScene038()
			destroyHDR()
			saveCfg()
		}
	}
	return true
}
func graphicsRowDisabled038(row int) bool {
	if remaster.SettingsTab != 8 {
		return false
	}
	if row == 0 {
		return !gpu038.DX12
	}
	if row == 8 {
		return false
	}
	if beta.cfg.RendererBackend != 1 {
		return true
	}
	if rendererError038 != "" {
		return true
	}
	switch row {
	case 3:
		return gpu038.DLSS != 0 || gpu038.SLInit != 0
	case 4:
		return !gpu038.FSRSupported
	case 6, 7:
		return gpu038.RayTier < 11
	}
	return true
}

func gpuTextureCapFor038(dedicated, budget uint64) int {
	if dedicated == 0 || budget == 0 {
		return 8192
	}
	if budget < dedicated {
		dedicated = budget
	}
	if dedicated < 3<<30 {
		return 1024
	}
	if dedicated < 7<<30 {
		return 2048
	}
	if dedicated < 10<<30 {
		return 4096
	}
	return 8192
}
func gpuTextureCap038() int { return gpuTextureCapFor038(gpu038.Dedicated, gpu038.Budget) }
