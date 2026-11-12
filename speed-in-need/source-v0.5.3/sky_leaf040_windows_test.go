//go:build windows

package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"speedinneed/engine"
	"testing"
	"time"
)

func Test040SkyRGBERetainsLinearHDR(t *testing.T) {
	encoded := []byte{128, 64, 32, 129, 160, 80, 40, 132, 0, 0, 0, 0}
	decoded := make([]float32, 9)
	maximum, peak, err := decodeRGBE040(encoded, decoded)
	if err != nil || maximum != 10 || peak != 1 {
		t.Fatal(maximum, peak, err)
	}
	if decoded[0] != 1 || decoded[1] != .5 || decoded[2] != .25 || decoded[3] != 10 || decoded[8] != 0 {
		t.Fatal("RGBE was tone mapped, gamma converted or decoded incorrectly", decoded)
	}
	if _, _, err := decodeRGBE040([]byte{255, 255, 255, 255}, make([]float32, 3)); err == nil {
		t.Fatal("non-RGB16F radiance accepted")
	}
}

func Test040SkyHeaderAndHardwareBudget(t *testing.T) {
	var header bytes.Buffer
	header.WriteString("SKY4")
	binary.Write(&header, binary.LittleEndian, uint32(16384))
	binary.Write(&header, binary.LittleEndian, uint32(8192))
	if w, h, err := skyHeader040(&header); err != nil || w != 16384 || h != 8192 {
		t.Fatal(w, h, err)
	}
	for _, c := range []struct {
		quality, setting int
		vram             uint64
		maximum, want    int
	}{{3, 0, 12 << 30, 32768, 16384}, {2, 0, 12 << 30, 32768, 8192}, {1, 0, 12 << 30, 32768, 4096}, {3, 4, 6 << 30, 32768, 4096}, {3, 4, 8 << 30, 32768, 8192}, {3, 4, 0, 32768, 4096}, {3, 4, 12 << 30, 8192, 8192}, {3, 1, 12 << 30, 32768, 2048}} {
		if got := skyTargetSize040(c.quality, c.setting, c.vram, c.maximum); got != c.want {
			t.Fatal(c, got)
		}
	}
}

func nominalSkyLight040(hour float64) pos3 {
	angle := (hour - 6) / 24 * engine.Tau
	sun := norm(pos3{math.Cos(angle) * .83, math.Max(.035, math.Sin(angle)), .38})
	moon := norm(pos3{-math.Cos(angle) * .83, math.Max(.035, -math.Sin(angle)), .38})
	f := engine.Clamp((engine.NightAmount(hour)-.40)/.20, 0, 1)
	f = f * f * (3 - 2*f)
	return norm(add(scale(sun, 1-f), scale(moon, f)))
}

func Test040SuppliedSkyRowsStreamAndCancel(t *testing.T) {
	for label, name := range []string{"day", "afternoon", "night"} {
		job := newSkyJob040(label, 2048)
		rows, complete := 0, false
		for chunk := range job.Chunks {
			if chunk.Error != nil {
				t.Fatal(name, chunk.Error)
			}
			if len(chunk.Pixels) != 0 {
				if chunk.Y != rows || chunk.Rows > 64 || len(chunk.Pixels) > 2048*64*3 {
					t.Fatal("unbounded or non-contiguous sky rows")
				}
				rows += chunk.Rows
				job.Recycled <- chunk.Pixels[:cap(chunk.Pixels)]
			}
			if chunk.Finished {
				// The supplied Castel cloudy panorama has diffuse HDR peaks around 3.47.
                if chunk.Maximum <= 1 || rows != 1024 {
					t.Fatal("genuine source HDR range/rows not preserved", name, chunk.Maximum, rows)
				}
				peak := skyPeakDirection040(chunk.PeakX, chunk.PeakY, chunk.Width, chunk.Height)
				if label!=2 && chunk.Celestial.Verified != (label == 1) {
					t.Fatal("supplied sun/moon was not correctly separated from the cloudy day maximum", name, chunk.Celestial)
				}
				t.Logf("%s upper-sky HDR cluster: pixels %d, coherence %.7f, verified %t", name, chunk.Celestial.Pixels, chunk.Celestial.Coherence, chunk.Celestial.Verified)
				d := nominalSkyLight040([]float64{12, 17.8, 22}[label])
				angle := math.Acos(engine.Clamp(peak.x*d.x+peak.y*d.y+peak.z*d.z, -1, 1)) * 180 / math.Pi
				t.Logf("%s: 2K peak %.3f, source direction %.4f/%.4f/%.4f, nominal scene separation %.2fdeg", name, chunk.Maximum, peak.x, peak.y, peak.z, angle)
				complete = true
			}
		}
		if !complete {
			t.Fatal("stream ended without verified completion")
		}
		<-job.Done
	}
	job := newSkyJob040(0, 16384)
	select {
	case packet := <-job.Chunks:
		if packet.Error != nil {
			t.Fatal(packet.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sky worker did not begin")
	}
	job.Stop.Do(func() { close(job.Cancel) })
	select {
	case <-job.Done:
	case <-time.After(2 * time.Second):
		t.Fatal("sky cancellation blocked on a full row queue")
	}
}

func Test040SkyCalibrationKeepsHorizonPolesAndLight(t *testing.T) {
	previous := sky040
	previousSky := sky020
	defer func() { sky040 = previous; sky020 = previousSky }()
	sky040.TrueHDR[1] = true
	sky020.Sizes[1] = [2]int{2048, 1024}
	sky040.Celestial[1] = skyCelestial040{Verified: true, X: 1024, Y: 478, Maximum: 29952}
	for _, hour := range []float64{6, 12, 17.8, 18.5, 22} {
		target := nominalSkyLight040(hour)
		p := skyAlignmentFor040(1, target)
		if p[3] != 1 {
			t.Fatal("verified celestial calibration disabled", hour, p)
		}
		u, v := skyDirectionUV040(target)
		mu, mv := mapPanoramaUV040(u, v, p)
		wantU, wantV := skyDirectionUV040(skyPeakDirection040(1024, 478, 2048, 1024))
		du := mu - wantU
		du -= math.Floor(du + .5)
		if math.Abs(du) > 1e-6 || math.Abs(mv-wantV) > 1e-6 {
			t.Fatal("calibrated light is not at measured source peak", hour, mu, mv, wantU, wantV)
		}
		for _, horizonOrPole := range []float64{0, .5, 1} {
			_, mapped := mapPanoramaUV040(.3, horizonOrPole, p)
			if math.Abs(mapped-horizonOrPole) > 1e-7 {
				t.Fatal("panorama horizon/pole moved", hour, horizonOrPole, mapped)
			}
		}
		last := -1.
		for y := 0; y <= 10000; y++ {
			_, mapped := mapPanoramaUV040(.3, float64(y)/10000, p)
			if mapped < last || mapped < 0 || mapped > 1 {
				t.Fatal("panorama map folded", hour, y, mapped, last)
			}
			last = mapped
		}
	}
	var p skyCelestial040
	ground := make([]float32, 16*3)
	ground[0] = 10000
	p.observe(ground, 12, 16, 16)
	p.finish(16, 16)
	if p.Verified || p.Pixels != 0 {
		t.Fatal("captured ground lamp was used as a sun", p)
	}
	if u, v := mapPanoramaUV040(.7, .2, [4]float32{}); u != .7 || v != .2 {
		t.Fatal("legacy sky changed without genuine HDR calibration")
	}
}

func Test040SkyResidencyUsesOnlyMeaningfulBlendPartners(t *testing.T) {
	want, _ := skyResidencyWants040(0, 0, 0, [3]bool{})
	if want != [3]bool{true, false, false} {
		t.Fatal("static day loaded unused skies", want)
	}
	want, _ = skyResidencyWants040(0, 1, .01, want)
	if want != [3]bool{true, false, false} {
		t.Fatal("insignificant blend partner loaded", want)
	}
	want, _ = skyResidencyWants040(0, 1, .08, want)
	if want != [3]bool{true, true, false} {
		t.Fatal("visible blend partner was not loaded", want)
	}
	want, _ = skyResidencyWants040(0, 1, .06, want)
	if want != [3]bool{true, true, false} {
		t.Fatal("blend entry/exit thresholds lack hysteresis", want)
	}
	want, _ = skyResidencyWants040(0, 1, .03, want)
	if want != [3]bool{true, false, false} {
		t.Fatal("negligible blend partner stayed wanted", want)
	}
	want, _ = skyResidencyWants040(0, 1, .99, want)
	if want != [3]bool{false, true, false} {
		t.Fatal("dominant next source was not prioritized", want)
	}
	want, weights := skyResidencyWants040(2, 2, .5, [3]bool{true, true, true})
	if want != [3]bool{false, false, true} || weights != [3]float64{0, 0, 1} {
		t.Fatal("one source counted twice", want, weights)
	}
}

func Test040SkyResidencyGraceAndCapacity(t *testing.T) {
	wanted := [3]bool{false, true, false}
	high := [3]bool{true, true, false}
	last := [3]float64{9, 10, 0}
	if got := skyRetireMask040(wanted, last, high, 10, -1); got != [3]bool{} {
		t.Fatal("retired too early", got)
	}
	if got := skyRetireMask040(wanted, last, high, 11, -1); got != [3]bool{true, false, false} {
		t.Fatal("unused source did not retire", got)
	}
	wanted[2] = true
	if got := skyRetireMask040(wanted, last, high, 10, 2); got != [3]bool{true, false, false} {
		t.Fatal("third high-detail label exceeded capacity", got)
	}
	for current := 0; current < 3; current++ {
		for next := 0; next < 3; next++ {
			for _, blend := range []float64{0, .01, .03, .06, .1, .5, .94, .99, 1} {
				want, _ := skyResidencyWants040(current, next, blend, [3]bool{})
				count := 0
				for _, w := range want {
					if w {
						count++
					}
				}
				if count < 1 || count > 2 {
					t.Fatal("invalid active source count", current, next, blend, want)
				}
			}
		}
	}
}

func Test040SkySnapshotRetainsFallbackAndRestoresCalibration(t *testing.T) {
	previous, previousSky := sky040, sky020
	defer func() { sky040 = previous; sky020 = previousSky }()
	sky040.Fallback = [3]skySnapshot040{}
	sky020.Textures = [3]uint32{}
	sky020.Sizes = [3][2]int{}
	low := skySnapshot040{Texture: 10, Size: [2]int{2048, 1024}, Maximum: 2688, PeakDirection: pos3{.7, .7, 0}, Celestial: skyCelestial040{X: 1023, Y: 255, Verified: true}}
	if old := commitSkySnapshot040(2, low); old != 0 {
		t.Fatal("new fallback returned a texture for deletion", old)
	}
	high := skySnapshot040{Texture: 11, Size: [2]int{16384, 8192}, Maximum: 3936, PeakDirection: pos3{.7, .7, 0}, Celestial: skyCelestial040{X: 8184, Y: 2040, Verified: true}}
	if old := commitSkySnapshot040(2, high); old != 0 || sky040.Fallback[2] != low {
		t.Fatal("upgrade would delete permanent fallback", old, sky040.Fallback[2])
	}
	high.Texture = 12
	if old := commitSkySnapshot040(2, high); old != 11 {
		t.Fatal("superseded high texture leaked", old)
	}
	if old := restoreSkyFallback040(2); old != 12 {
		t.Fatal("retired high texture leaked", old)
	}
	if sky020.Textures[2] != 10 || sky020.Sizes[2] != low.Size || sky040.Celestial[2] != low.Celestial || sky040.Maximum[2] != low.Maximum {
		t.Fatal("2K calibration metadata was not restored")
	}
	if old := restoreSkyFallback040(2); old != 0 {
		t.Fatal("fallback would be deleted twice", old)
	}
}

func Test040SkyResidentStorageCountsAliasesAndPendingBase(t *testing.T) {
	previous, previousSky := sky040, sky020
	defer func() { sky040 = previous; sky020 = previousSky }()
	sky040.Fallback = [3]skySnapshot040{}
	sky040.Job = nil
	sky020.Textures = [3]uint32{}
	sky020.Sizes = [3][2]int{}
	for label := 0; label < 3; label++ {
		commitSkySnapshot040(label, skySnapshot040{Texture: uint32(label + 1), Size: [2]int{2048, 1024}})
	}
	logical, padded, count := skyResidentStorage040()
	base := skyTextureStorage040(2048, 1024, 6, true)
	if logical != 3*base || padded != 3*skyTextureStorage040(2048, 1024, 8, true) || count != 0 {
		t.Fatal("fallback aliases counted twice", logical, padded, count)
	}
	commitSkySnapshot040(0, skySnapshot040{Texture: 4, Size: [2]int{16384, 8192}})
	logical, padded, count = skyResidentStorage040()
	if logical != 3*base+skyTextureStorage040(16384, 8192, 6, true) || count != 1 {
		t.Fatal("single active 16K storage incorrect", logical, count)
	}
	if logical >= 1200<<20 || padded >= 1500<<20 {
		t.Fatal("static 16K sky still uses multiple GiB", logical, padded)
	}
	sky040.Job = &skyJob040{Texture: 5, Width: 16384, Height: 8192}
	withPending, _, _ := skyResidentStorage040()
	if withPending-logical != 16384*8192*6 {
		t.Fatal("pending texture mip levels not actually allocated were counted", withPending-logical)
	}
	t.Logf("static active 16K plus all 2K fallbacks: logical %.2fMiB, padded estimate %.2fMiB", float64(logical)/(1<<20), float64(padded)/(1<<20))
}

func makeLeafStream040(levels [][2]int) []byte {
	var output bytes.Buffer
	output.WriteString("LMP4")
	binary.Write(&output, binary.LittleEndian, uint32(len(levels)))
	for _, size := range levels {
		pixels := make([]byte, size[0]*size[1]*4)
		for i := 0; i < len(pixels)/4; i++ {
			pixels[i*4], pixels[i*4+1], pixels[i*4+2] = 30, 130, 50
			if i%2 == 0 {
				pixels[i*4+3] = 255
			}
		}
		var compressed bytes.Buffer
		codec := zlib.NewWriter(&compressed)
		codec.Write(pixels)
		codec.Close()
		for _, v := range []uint32{uint32(size[0]), uint32(size[1]), uint32(compressed.Len())} {
			binary.Write(&output, binary.LittleEndian, v)
		}
		output.Write(compressed.Bytes())
	}
	return output.Bytes()
}

func Test040LeafMipChainUsesAuthoredLevels(t *testing.T) {
	data := makeLeafStream040([][2]int{{4, 4}, {2, 2}, {1, 1}})
	sizes := []int{}
	levels, byteCount, err := readLeafMipChain040(bytes.NewReader(data), 2, func(level, w, h int, pixels []byte) error {
		if level != len(sizes) || len(pixels) != w*h*4 {
			t.Fatal("leaf level mismatch")
		}
		sizes = append(sizes, w)
		return nil
	})
	if err != nil || levels != 2 || byteCount != 20 || len(sizes) != 2 || sizes[0] != 2 || sizes[1] != 1 {
		t.Fatal(levels, byteCount, sizes, err)
	}
	bad := makeLeafStream040([][2]int{{4, 4}, {1, 1}})
	if _, _, err := readLeafMipChain040(bytes.NewReader(bad), 4, func(int, int, int, []byte) error { return nil }); err == nil {
		t.Fatal("missing mip level accepted")
	}
	if _, _, err := readLeafMipChain040(bytes.NewReader(data[:len(data)-2]), 4, func(int, int, int, []byte) error { return nil }); err == nil {
		t.Fatal("truncated leaf stream accepted")
	}
}

func Test040SuppliedLeavesPreserveAuthoredCoverage(t *testing.T) {
	raw, err := assets.ReadFile("assets/pbr/b040_leaf_manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Records []struct {
			Mips   string
			Levels []struct {
				Width, Height int
				Coverage      float64
			}
		}
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, record := range manifest.Records {
		file, err := assets.Open("assets/pbr/" + record.Mips)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = readLeafMipChain040(file, 1024, func(_ int, w, h int, pixels []byte) error {
			coverage := 0
			for i := 3; i < len(pixels); i += 4 {
				if pixels[i] >= 90 {
					coverage++
				}
			}
			for _, source := range record.Levels {
				if source.Width == w && source.Height == h && math.Abs(float64(coverage)/float64(w*h)-source.Coverage) > 1e-8 {
					t.Fatal(record.Mips, "authored alpha coverage changed")
				}
			}
			return nil
		})
		file.Close()
		if err != nil && err != io.EOF {
			t.Fatal(record.Mips, err)
		}
	}
}
