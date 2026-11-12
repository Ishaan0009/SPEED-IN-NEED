package main

import (
	"math"
	"testing"
)

func TestRGBE050AllExponentsPreserveRadianceAndBounds(t *testing.T) {
	rgb := make([]float32, 3)
	for exponent := 0; exponent < 256; exponent++ {
		for m := 0; m < 256; m++ {
			scale := float32(0)
			if exponent > 0 {
				scale = float32(math.Ldexp(1, exponent-136))
			}
			want := float32(m) * scale
			maximum, _, err := decodeRGBE040([]byte{byte(m), byte(m), byte(m), byte(exponent)}, rgb)
			if want > 65504 {
				if err == nil {
					t.Fatalf("unsafe exponent %d mantissa %d accepted", exponent, m)
				}
				continue
			}
			if err != nil || maximum != want || rgb[0] != want || rgb[1] != want || rgb[2] != want {
				t.Fatalf("RGBE %d/%d changed: %v, %v", exponent, m, rgb, err)
			}
		}
	}
}
func TestSkyViewport050HonorsExplicitDetailAndHardware(t *testing.T) {
	for _, c := range []struct {
		quality, setting, width, maxTexture int
		vram                                uint64
		want                                int
	}{
		{0, 0, 1920, 16384, 12 << 30, 2048}, {1, 0, 1920, 16384, 12 << 30, 4096},
		{2, 0, 1920, 16384, 12 << 30, 4096}, {3, 0, 2560, 16384, 12 << 30, 4096},
		{3, 0, 3840, 16384, 12 << 30, 16384}, {3, 4, 1920, 16384, 12 << 30, 16384},
		{3, 4, 3840, 16384, 4 << 30, 4096}, {3, 4, 3840, 2048, 12 << 30, 2048},
	} {
		if got := skyViewportTarget050(c.quality, c.setting, c.vram, c.maxTexture, c.width); got != c.want {
			t.Fatalf("%+v: got %d", c, got)
		}
	}
}
func BenchmarkRGBEDecode050(b *testing.B) {
	pixels := 16384 * 64
	raw := make([]byte, pixels*4)
	rgb := make([]float32, pixels*3)
	for i := 0; i < pixels; i++ {
		copy(raw[i*4:], []byte{100, 80, 60, 128})
	}
	b.SetBytes(int64(len(raw)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := decodeRGBE040(raw, rgb); err != nil {
			b.Fatal(err)
		}
	}
}
