//go:build windows

package main

import (
	"bufio"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

const skyRowsPerChunk040 = 64
const skyRetireGrace040 = 2.0

type skySnapshot040 struct {
	Texture       uint32
	Size          [2]int
	Maximum       float32
	PeakDirection pos3
	Celestial     skyCelestial040
}

type skyChunk040 struct {
	Width, Height, Y, Rows int
	Pixels                 []float32
	Header, Finished       bool
	Maximum                float32
	PeakX, PeakY           int
	Celestial              skyCelestial040
	Error                  error
}

// Only compact HDR energy at least one degree above the horizon is eligible.
// The supplied evening/night panoramas were also visually inspected: these
// clusters are their sun/moon, rather than lamps in the captured ground.
type skyCelestial040 struct {
	Maximum      float32
	X, Y, Pixels int
	Weight       float64
	Vector       pos3
	Coherence    float64
	Verified     bool
}

func (p *skyCelestial040) observe(pixels []float32, row, width, height int) {
	// Most panorama pixels are dim. Avoid coordinate divisions for those pixels,
	// while retaining the exact horizon exclusion used by celestial calibration.
	rows := int(math.Ceil(float64(height)*(0.5-1./180)-.5)) - row
	if rows <= 0 {
		return
	}
	for i := 0; i < min(len(pixels)/3, rows*width); i++ {
		value := max(pixels[i*3], pixels[i*3+1], pixels[i*3+2])
		if value <= 100 {
			continue
		}
		y := row + i/width
		x := i % width
		if value > p.Maximum {
			p.Maximum, p.X, p.Y = value, x, y
		}
		weight := float64(value)
		p.Vector = add(p.Vector, scale(skyPeakDirection040(x, y, width, height), weight))
		p.Weight += weight
		p.Pixels++
	}
}

func (p *skyCelestial040) finish(width, height int) {
	if p.Weight <= 0 {
		return
	}
	v := scale(p.Vector, 1/p.Weight)
	p.Coherence = math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	peak := skyPeakDirection040(p.X, p.Y, width, height)
	// Broad clouds, ground lights and dispersed bright stars cannot become the
	// directional light anchor. Day's cloudy maximum is below the HDR threshold.
	p.Verified = p.Maximum > 100 && peak.y > math.Sin(math.Pi/180) && p.Coherence > .995 && p.Pixels < width*height/100
}

type skyJob040 struct {
	Label, Width, Height, Uploaded int
	Filename051                    string
	Texture                        uint32
	Chunks                         chan skyChunk040
	Recycled                       chan []float32
	Cancel, Done                   chan struct{}
	Stop                           sync.Once
	Cancelled                      bool
}

var sky040 struct {
	TrueHDR           [3]bool
	Maximum           [3]float32
	PeakDirection     [3]pos3
	Celestial         [3]skyCelestial040
	Fallback          [3]skySnapshot040
	Wanted            [3]bool
	LastWanted        [3]float64
	Retired           uint64
	CancelledInactive uint64
	Job               *skyJob040
	Requested         int
	MaxTexture        int
	LastPump          float64
	Pumped            bool
	Closed            bool
	Failed            map[[2]int]bool
	LastError         string
	RowsUploaded      uint64
	UpgradeCount      uint64
}

func skyHeader040(reader io.Reader) (int, int, error) {
	var header [12]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, 0, err
	}
	w, h := int(binary.LittleEndian.Uint32(header[4:8])), int(binary.LittleEndian.Uint32(header[8:]))
	if string(header[:4]) != "SKY4" || w < 2 || w > 16384 || w&(w-1) != 0 || h != w/2 {
		return 0, 0, fmt.Errorf("invalid streamed HDR sky dimensions %dx%d", w, h)
	}
	return w, h, nil
}

func decodeRGBE040(rgbe []byte, rgb []float32) (float32, int, error) {
	if len(rgbe)%4 != 0 || len(rgb) != len(rgbe)/4*3 {
		return 0, 0, fmt.Errorf("invalid RGBE row buffer")
	}
	maximum, peak := float32(0), 0
	for i := 0; i < len(rgbe)/4; i++ {
		scale := rgbeScale050[rgbe[i*4+3]]
		for channel := 0; channel < 3; channel++ {
			value := float32(rgbe[i*4+channel]) * scale
			// Byte mantissas and the finite exponent table cannot produce NaN/Inf.
			if value > 65504 {
				return 0, 0, fmt.Errorf("HDR radiance exceeds finite RGB16F range")
			}
			rgb[i*3+channel] = value
			if value > maximum {
				maximum, peak = value, i
			}
		}
	}
	return maximum, peak, nil
}

func skyPeakDirection040(x, y, width, height int) pos3 {
	longitude := ((float64(x)+.5)/float64(width) - .5) * math.Pi * 2
	polar := (float64(y) + .5) / float64(height) * math.Pi
	return pos3{math.Cos(longitude) * math.Sin(polar), math.Cos(polar), math.Sin(longitude) * math.Sin(polar)}
}

func newSkyJob040(label, width int) *skyJob040 {
	job := &skyJob040{Label: label, Width: width, Chunks: make(chan skyChunk040, 1), Recycled: make(chan []float32, 2), Cancel: make(chan struct{}), Done: make(chan struct{})}
	job.Filename051 = fmt.Sprintf("assets/sky040_%s_%d.sky040", []string{"day", "afternoon", "night"}[label], width)
	if weather050.Rain && !workshop051.Active {
		rainLabel := "day"
		if label == 2 {
			rainLabel = "night"
		}
		job.Filename051 = fmt.Sprintf("assets/sky052_%s_2048.sky040", rainLabel)
	}

	if workshop051.Active || garage033 {
		job.Filename051 = "assets/sky051_garage_2048.sky040"
	}
	go job.decode()
	return job
}

func (j *skyJob040) send(chunk skyChunk040) bool {
	select {
	case j.Chunks <- chunk:
		return true
	case <-j.Cancel:
		return false
	}
}

func (j *skyJob040) decode() {
	defer close(j.Done)
	defer close(j.Chunks)
	file, err := assets.Open(j.Filename051)
	if err != nil {
		j.send(skyChunk040{Error: err})
		return
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64<<10)
	w, h, err := skyHeader040(reader)
	if err != nil || w != j.Width {
		if err == nil {
			err = fmt.Errorf("HDR sky tier/header mismatch")
		}
		j.send(skyChunk040{Error: err})
		return
	}
	decoder, err := zlib.NewReader(reader)
	if err != nil {
		j.send(skyChunk040{Error: err})
		return
	}
	defer decoder.Close()
	if !j.send(skyChunk040{Header: true, Width: w, Height: h}) {
		return
	}
	// Exactly two recyclable float buffers: at 16K this is 24 MiB, plus one
	// 4 MiB RGBE row buffer. No full decoded panorama exists on the CPU.
	for i := 0; i < 2; i++ {
		j.Recycled <- make([]float32, w*skyRowsPerChunk040*3)
	}
	packed := make([]byte, w*skyRowsPerChunk040*4)
	maximum, peakX, peakY := float32(0), 0, 0
	var celestial skyCelestial040
	for y := 0; y < h; y += skyRowsPerChunk040 {
		rows := min(skyRowsPerChunk040, h-y)
		var pixels []float32
		select {
		case pixels = <-j.Recycled:
		case <-j.Cancel:
			return
		}
		raw := packed[:w*rows*4]
		if _, err := io.ReadFull(decoder, raw); err != nil {
			j.send(skyChunk040{Error: err})
			return
		}
		pixels = pixels[:w*rows*3]
		value, peak, err := decodeRGBE040(raw, pixels)
		if err != nil {
			j.send(skyChunk040{Error: err})
			return
		}
		if value > maximum {
			maximum, peakX, peakY = value, peak%w, y+peak/w
		}
		celestial.observe(pixels, y, w, h)
		if !j.send(skyChunk040{Width: w, Height: h, Y: y, Rows: rows, Pixels: pixels}) {
			return
		}
	}
	var extra [1]byte
	if n, err := decoder.Read(extra[:]); n != 0 || err != io.EOF {
		j.send(skyChunk040{Error: fmt.Errorf("HDR expanded size/checksum mismatch: %v", err)})
		return
	}
	if err := decoder.Close(); err != nil {
		j.send(skyChunk040{Error: err})
		return
	}
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		j.send(skyChunk040{Error: fmt.Errorf("HDR stream trailing data/checksum mismatch: %v", err)})
		return
	}
	celestial.finish(w, h)
	j.send(skyChunk040{Finished: true, Width: w, Height: h, Maximum: maximum, PeakX: peakX, PeakY: peakY, Celestial: celestial})
}

func skyTargetSize040(quality, setting int, vram uint64, maxTexture int) int {
	width := 4096
	if quality == 2 {
		width = 8192
	} else if quality >= 3 {
		width = 16384
	}
	if setting >= 1 && setting <= 4 {
		width = []int{2048, 4096, 8192, 16384}[setting-1]
	}
	cap := 8192
	if vram == 0 || vram <= 6<<30 {
		cap = 4096
	} else if vram >= 10<<30 {
		cap = 16384
	}
	width = min(width, cap)
	for width > maxTexture && width > 512 {
		width /= 2
	}
	return width
}

func skyMaxTexture040() int {
	if sky040.MaxTexture == 0 {
		var maximum int32
		gl("glGetIntegerv", 0x0d33, uintptr(unsafe.Pointer(&maximum)))
		sky040.MaxTexture = int(maximum)
	}
	return sky040.MaxTexture
}

func allocateSky040(j *skyJob040, width, height int) error {
	j.Height = height
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&j.Texture)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(j.Texture))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812f)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881b, uintptr(width), uintptr(height), 0, 0x1907, glFloat, 0)
	if code := gl("glGetError"); code != 0 {
		return fmt.Errorf("HDR %dx%d RGB16F allocation failed: GL0x%x", width, height, code)
	}
	return nil
}

func stopSkyJob040(j *skyJob040) {
	if j == nil {
		return
	}
	j.Cancelled = true
	j.Stop.Do(func() { close(j.Cancel) })
	if j.Texture != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&j.Texture)))
		j.Texture = 0
	}
}

func installSkySnapshot040(label int, s skySnapshot040) {
	sky020.Textures[label], sky020.Sizes[label] = s.Texture, s.Size
	sky040.TrueHDR[label], sky040.Maximum[label] = true, s.Maximum
	sky040.PeakDirection[label], sky040.Celestial[label] = s.PeakDirection, s.Celestial
}

// Keep the complete bootstrap image and its measured calibration metadata.
// The caller deletes only the returned superseded texture, never the fallback.
func commitSkySnapshot040(label int, s skySnapshot040) uint32 {
	old := sky020.Textures[label]
	if s.Size[0] == 2048 {
		sky040.Fallback[label] = s
	}
	installSkySnapshot040(label, s)
	if old == sky040.Fallback[label].Texture || old == s.Texture {
		return 0
	}
	return old
}

func restoreSkyFallback040(label int) uint32 {
	fallback := sky040.Fallback[label]
	old := sky020.Textures[label]
	if fallback.Texture == 0 || old == fallback.Texture {
		return 0
	}
	installSkySnapshot040(label, fallback)
	sky040.Retired++
	return old
}

// The dominant source is always wanted. A partner enters at 8% weight and
// stays until below 4%, so time/blend noise cannot churn large texture jobs.
func skyResidencyWants040(current, next int, blend float64, previous [3]bool) ([3]bool, [3]float64) {
	weights := [3]float64{}
	blend = math.Max(0, math.Min(1, blend))
	weights[current] += 1 - blend
	weights[next] += blend
	dominant := 0
	for i := 1; i < 3; i++ {
		if weights[i] > weights[dominant] {
			dominant = i
		}
	}
	wanted := [3]bool{}
	for i, w := range weights {
		threshold := .08
		if previous[i] {
			threshold = .04
		}
		wanted[i] = w >= threshold || i == dominant
	}
	return wanted, weights
}

func skyRetireMask040(wanted [3]bool, lastWanted [3]float64, high [3]bool, now float64, incoming int) [3]bool {
	retire := [3]bool{}
	count := 0
	for i, resident := range high {
		if resident && !wanted[i] && (now < lastWanted[i] || now-lastWanted[i] >= skyRetireGrace040) {
			retire[i] = true
		}
		if resident && !retire[i] {
			count++
		}
	}
	if incoming >= 0 && (!high[incoming] || retire[incoming]) {
		count++
	}
	// Two distinct high-detail labels maximum, including a pending upload.
	// Capacity pressure can retire an unused source before its grace expires.
	for count > 2 {
		victim := -1
		for i, resident := range high {
			if resident && !wanted[i] && !retire[i] && (victim < 0 || lastWanted[i] < lastWanted[victim]) {
				victim = i
			}
		}
		if victim < 0 {
			break
		}
		retire[victim] = true
		count--
	}
	return retire
}

func consumeSkyChunk040(j *skyJob040, c skyChunk040) error {
	if c.Error != nil {
		return c.Error
	}
	if c.Header {
		return allocateSky040(j, c.Width, c.Height)
	}
	if len(c.Pixels) != 0 {
		if j.Texture == 0 || c.Y != j.Uploaded || c.Width != j.Width || c.Height != j.Height {
			return fmt.Errorf("non-contiguous HDR upload rows")
		}
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(j.Texture))
		gl("glTexSubImage2D", GL_TEXTURE_2D, 0, 0, uintptr(c.Y), uintptr(c.Width), uintptr(c.Rows), 0x1907, glFloat, uintptr(unsafe.Pointer(&c.Pixels[0])))
		runtime.KeepAlive(c.Pixels)
		j.Uploaded += c.Rows
		sky040.RowsUploaded += uint64(c.Rows)
		j.Recycled <- c.Pixels[:cap(c.Pixels)]
	}
	if !c.Finished {
		return nil
	}
	if j.Texture == 0 || j.Uploaded != j.Height {
		return fmt.Errorf("incomplete HDR panorama")
	}
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(j.Texture))
	if glExtAvailable("glGenerateMipmap") {
		glExt("glGenerateMipmap", GL_TEXTURE_2D)
		gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2703)
	}
	if code := gl("glGetError"); code != 0 {
		return fmt.Errorf("HDR mip generation failed: GL0x%x", code)
	}
	old := commitSkySnapshot040(j.Label, skySnapshot040{Texture: j.Texture, Size: [2]int{j.Width, j.Height}, Maximum: c.Maximum, PeakDirection: skyPeakDirection040(c.PeakX, c.PeakY, c.Width, c.Height), Celestial: c.Celestial})
	j.Texture = 0
	if old != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&old)))
	}
	return nil
}

// Bootstrap all three time-of-day textures at 2K without allocating a decoded
// image. Higher-resolution tiers are never a title-screen startup dependency.
func bootstrapSky040(label int) bool {
	if skyMaxTexture040() < 2048 {
		return false
	}
	sky040.Closed = false
	j := newSkyJob040(label, 2048)
	complete := false
	for chunk := range j.Chunks {
		if err := consumeSkyChunk040(j, chunk); err != nil {
			sky040.LastError = err.Error()
			stopSkyJob040(j)
			break
		}
		complete = complete || chunk.Finished
	}
	stopSkyJob040(j)
	<-j.Done
	return complete
}

func pumpSky040() {
	if sky040.Closed || sunShadow.drawing || (sky040.Pumped && sky040.LastPump == app.t) {
		return
	}
	sky040.Pumped, sky040.LastPump = true, app.t
	target := skyViewportTarget050(app.quality, beta.cfg.SkyDetail040, gpu038.Dedicated, skyMaxTexture040(), app.width)
	sky040.Requested = target
	if sky040.Failed == nil {
		sky040.Failed = make(map[[2]int]bool)
	}
	i, next, blend := skyBlend034()
	wanted, weights := skyResidencyWants040(i, next, blend, sky040.Wanted)
	sky040.Wanted = wanted
	for label, needed := range wanted {
		if needed {
			sky040.LastWanted[label] = app.t
		}
	}
	newLabel := -1
	if target > 2048 {
		for label, needed := range wanted {
			if needed && sky040.TrueHDR[label] && sky020.Sizes[label][0] != target && !sky040.Failed[[2]int{label, target}] && (newLabel < 0 || weights[label] > weights[newLabel]) {
				newLabel = label
			}
		}
	}
	incoming := newLabel
	if j := sky040.Job; j != nil && !j.Cancelled && wanted[j.Label] && j.Width == target {
		incoming = j.Label
	}
	high := [3]bool{}
	for label, size := range sky020.Sizes {
		high[label] = size[0] > 2048 && sky040.Fallback[label].Texture != 0
	}
	retire := skyRetireMask040(wanted, sky040.LastWanted, high, app.t, incoming)
	if target <= 2048 {
		retire = high
		newLabel = -1
	}
	if sky040.Job == nil && newLabel < 0 && !retire[0] && !retire[1] && !retire[2] {
		return
	}
	var active, previous int32
	gl("glGetIntegerv", 0x84e0, uintptr(unsafe.Pointer(&active)))
	glExt("glActiveTexture", 0x84c0)
	gl("glGetIntegerv", 0x8069, uintptr(unsafe.Pointer(&previous)))
	defer func() {
		if previous != 0 && gl("glIsTexture", uintptr(previous)) == 0 {
			previous = 0
		}
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(previous))
		glExt("glActiveTexture", uintptr(active))
	}()
	for label, evict := range retire {
		if evict {
			if old := restoreSkyFallback040(label); old != 0 {
				gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&old)))
			}
		}
	}
	if j := sky040.Job; j != nil {
		if (j.Width != target || !wanted[j.Label] || target <= 2048) && !j.Cancelled {
			if !wanted[j.Label] {
				sky040.CancelledInactive++
			}
			stopSkyJob040(j)
		}
		if j.Cancelled {
			select {
			case <-j.Done:
				sky040.Job = nil
			default:
			}
			return
		}
		// Header/completion packets do not count against the two64-row budget.
		rows := 0
		for packet := 0; packet < 4 && rows < 2; packet++ {
			select {
			case chunk, ok := <-j.Chunks:
				if !ok {
					sky040.Job = nil
					return
				}
				if err := consumeSkyChunk040(j, chunk); err != nil {
					sky040.LastError = err.Error()
					sky040.Failed[[2]int{j.Label, j.Width}] = true
					stopSkyJob040(j)
					return
				}
				if len(chunk.Pixels) > 0 {
					rows++
				}
				if chunk.Finished {
					sky040.UpgradeCount++
					sky040.Job = nil
					return
				}
			default:
				return
			}
		}
		return
	}
	if newLabel >= 0 {
		sky040.Job = newSkyJob040(newLabel, target)
	}
}

func skyDisplayGains040() [2]float32 {
	i, next, _ := skyBlend034()
	values := [3]float32{.80, 1, .12}
	gains := [2]float32{1, 1}
	for j, label := range []int{i, next} {
		if sky040.TrueHDR[label] {
			gains[j] = values[label]
		}
	}
	return gains
}

func environmentSkyGains040() [2]float32 {
	i, next, _ := skyBlend034()
	values := [3]float32{.50, .65, .04}
	gains := [2]float32{1, 1}
	for j, label := range []int{i, next} {
		if sky040.TrueHDR[label] {
			gains[j] = values[label]
		}
	}
	return gains
}

func skyDirectionUV040(direction pos3) (float64, float64) {
	direction = norm(direction)
	return math.Atan2(direction.z, direction.x)/(2*math.Pi) + .5, math.Acos(math.Max(-1, math.Min(1, direction.y))) / math.Pi
}

// Artistic panorama calibration, not a rigid rotation of the physical capture.
// Yaw is shared with the ground; the vertical map fixes both poles and the
// horizon while bringing the verified upper-sky sun/moon onto the scene light.
func skyAlignmentFor040(label int, target pos3) [4]float32 {
	p := sky040.Celestial[label]
	if !sky040.TrueHDR[label] || !p.Verified || target.y <= .001 {
		return [4]float32{}
	}
	w, h := sky020.Sizes[label][0], sky020.Sizes[label][1]
	if w < 1 || h < 1 {
		return [4]float32{}
	}
	u, v := skyDirectionUV040(skyPeakDirection040(p.X, p.Y, w, h))
	tu, tv := skyDirectionUV040(target)
	tv = math.Max(.0001, math.Min(.4999, tv))
	return [4]float32{float32(u - tu), float32(v), float32(tv), 1}
}

func skyAlignment040() [2][4]float32 {
	i, next, _ := skyBlend034()
	direction := beta011CelestialDirection() // Includes the game's moon at night.
	return [2][4]float32{skyAlignmentFor040(i, direction), skyAlignmentFor040(next, direction)}
}

// Call while the scene lighting program is active, after its environment gain.
func setEnvironmentAlignment040() {
	if !lightPipeline.active {
		return
	}
	values := skyAlignment040()
	for i, name := range []string{"uPanoramaAlign040", "uPanoramaAlignNext040"} {
		if location := sceneLocation(name); location >= 0 {
			glExt("glUniform4fv", uintptr(location), 1, uintptr(unsafe.Pointer(&values[i][0])))
		}
	}
}

func mapPanoramaUV040(u, v float64, p [4]float32) (float64, float64) {
	if p[3] < .5 {
		return u, v
	}
	u += float64(p[0])
	u -= math.Floor(u)
	source, target := float64(p[1]), float64(p[2])
	if v < target {
		v = v * source / target
	} else if v < .5 {
		v = source + (v-target)*(.5-source)/(.5-target)
	}
	return u, v
}

// Only verified, aligned HDR celestial energy replaces the procedural disc.
func proceduralCelestialWeight040() float32 {
	_, _, blend := skyBlend034()
	p := skyAlignment040()
	return float32(1 - float64(p[0][3])*(1-blend) - float64(p[1][3])*blend)
}

func skyTextureStorage040(width, height, bytesPerPixel int, mips bool) uint64 {
	if width < 1 || height < 1 {
		return 0
	}
	total := uint64(0)
	for {
		total += uint64(width) * uint64(height) * uint64(bytesPerPixel)
		if !mips || (width == 1 && height == 1) {
			return total
		}
		width, height = max(1, width/2), max(1, height/2)
	}
}

// Logical allocation estimate, not a driver memory query. RGB16F can be padded
// to RGBA16F; pending textures have only their base level until completion.
func skyResidentStorage040() (logical, padded uint64, highCount int) {
	seen := map[uint32]bool{}
	addTexture := func(texture uint32, size [2]int, mips bool) {
		if texture == 0 || seen[texture] {
			return
		}
		seen[texture] = true
		logical += skyTextureStorage040(size[0], size[1], 6, mips)
		padded += skyTextureStorage040(size[0], size[1], 8, mips)
	}
	for _, fallback := range sky040.Fallback {
		addTexture(fallback.Texture, fallback.Size, true)
	}
	for label, texture := range sky020.Textures {
		addTexture(texture, sky020.Sizes[label], true)
		if texture != 0 && sky020.Sizes[label][0] > 2048 {
			highCount++
		}
	}
	if j := sky040.Job; j != nil {
		addTexture(j.Texture, [2]int{j.Width, j.Height}, false)
	}
	return logical, padded, highCount
}

func sky040Report() map[string]any {
	var directions [3][3]float64
	var complete16K [3]bool
	var anchors [3]map[string]any
	target := beta011CelestialDirection()
	for i, p := range sky040.PeakDirection {
		directions[i] = [3]float64{p.x, p.y, p.z}
		complete16K[i] = sky040.TrueHDR[i] && sky020.Sizes[i][0] == 16384 && sky020.Sizes[i][1] == 8192
		anchor := sky040.Celestial[i]
		alignment := skyAlignmentFor040(i, target)
		mapped := p
		if anchor.Verified && sky020.Sizes[i][0] > 0 {
			mapped = skyPeakDirection040(anchor.X, anchor.Y, sky020.Sizes[i][0], sky020.Sizes[i][1])
			if alignment[3] > .5 {
				u, v := skyDirectionUV040(mapped)
				u -= float64(alignment[0])
				u -= math.Floor(u)
				source, destination := float64(alignment[1]), float64(alignment[2])
				if v < source {
					v = v * destination / source
				} else if v < .5 {
					v = destination + (v-source)*(.5-destination)/(.5-source)
				}
				longitude, polar := (u-.5)*2*math.Pi, v*math.Pi
				mapped = pos3{math.Cos(longitude) * math.Sin(polar), math.Cos(polar), math.Sin(longitude) * math.Sin(polar)}
			}
		}
		errorDegrees := math.Acos(math.Max(-1, math.Min(1, mapped.x*target.x+mapped.y*target.y+mapped.z*target.z))) * 180 / math.Pi
		anchors[i] = map[string]any{"verified_upper_sky": anchor.Verified, "pixels_above_100": anchor.Pixels, "coherence": anchor.Coherence, "source_pixel": [2]int{anchor.X, anchor.Y}, "maximum_radiance": anchor.Maximum, "alignment": alignment, "mapped_peak_direction": [3]float64{mapped.x, mapped.y, mapped.z}, "scene_direction_error_degrees": errorDegrees, "aligned": alignment[3] > .5}
	}
	report := map[string]any{"true_hdr": sky040.TrueHDR, "true_16k_complete": complete16K, "celestial_calibration": anchors, "calibration_kind": "artistic horizon-preserving UV remap; not a rigid capture rotation", "dimensions": sky020.Sizes, "requested_width": sky040.Requested, "maximum_radiance": sky040.Maximum, "source_peak_direction": directions, "rows_uploaded": sky040.RowsUploaded, "completed_upgrades": sky040.UpgradeCount, "cpu_row_buffer_limit_bytes_16k": 28 << 20, "row_upload_limit_per_frame": 128, "last_error": sky040.LastError, "display_gain": skyDisplayGains040(), "environment_gain": environmentSkyGains040(), "procedural_celestial_weight": proceduralCelestialWeight040()}
	logical, padded, residentCount := skyResidentStorage040()
	fallbacks := [3]bool{}
	for label, fallback := range sky040.Fallback {
		fallbacks[label] = fallback.Texture != 0
	}
	report["residency"] = map[string]any{"policy": "dominant source plus meaningful blend partner; permanent 2K fallback per source", "wanted_high_detail": sky040.Wanted, "fallback_2k_available": fallbacks, "completed_high_detail_count": residentCount, "maximum_high_detail_labels_including_pending": 2, "estimated_rgb16f_bytes": logical, "estimated_padded_rgba16f_bytes": padded, "allocation_estimate_not_driver_query": true, "retired_to_2k": sky040.Retired, "inactive_jobs_cancelled": sky040.CancelledInactive, "blend_partner_enter_weight": .08, "blend_partner_exit_weight": .04, "retire_grace_seconds": skyRetireGrace040}
	if j := sky040.Job; j != nil {
		report["pending"] = map[string]any{"label": j.Label, "width": j.Width, "uploaded_rows": j.Uploaded, "height": j.Height, "cancelled": j.Cancelled}
	}
	return report
}

func releaseSky040() {
	sky040.Closed = true
	if j := sky040.Job; j != nil {
		stopSkyJob040(j)
		select {
		case <-j.Done:
		case <-time.After(2 * time.Second):
			sky040.LastError = "HDR decoder cancellation exceeded shutdown wait"
		}
		sky040.Job = nil
	}
	// releaseBeta020Sky deletes the three currently bound source textures next.
	// Delete only retained fallback images that are distinct from those sources.
	active := map[uint32]bool{}
	for _, texture := range sky020.Textures {
		active[texture] = true
	}
	for label, fallback := range sky040.Fallback {
		if fallback.Texture != 0 && !active[fallback.Texture] {
			texture := fallback.Texture
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&texture)))
			active[texture] = true
		}
		sky040.Fallback[label] = skySnapshot040{}
	}
}
