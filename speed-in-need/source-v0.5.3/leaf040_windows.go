//go:build windows

package main

import (
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"unsafe"
)

type leafRecord040 struct {
	OldAlbedo string  `json:"old_albedo"`
	NewAlbedo string  `json:"new_albedo"`
	NewNormal string  `json:"new_normal"`
	Mips      string  `json:"mips"`
	Size      [2]int  `json:"size"`
	Cutoff    float64 `json:"cutoff"`
}

var leafManifestOnce040 sync.Once
var leafRecords040 map[string]leafRecord040
var leafStatus040 struct {
	Textures, Levels int
	Bytes            uint64
	Error            string
}

func initLeafManifest040() {
	leafRecords040 = make(map[string]leafRecord040)
	raw, err := assets.ReadFile("assets/pbr/b040_leaf_manifest.json")
	if err != nil {
		return // Older compatible packs retain their original textures.
	}
	var manifest struct {
		Format  string
		Records []leafRecord040
	}
	if err := json.Unmarshal(raw, &manifest); err != nil || manifest.Format != "LMP4" {
		leafStatus040.Error = "Invalid full-source leaf manifest"
		return
	}
	for _, record := range manifest.Records {
		if record.OldAlbedo == "" || record.NewAlbedo == "" || record.Mips == "" {
			continue
		}
		leafRecords040[record.OldAlbedo] = record
		leafRecords040[record.NewAlbedo] = record
	}
}

func fullLeafMaterial040(m *pbrMaterial) {
	leafManifestOnce040.Do(initLeafManifest040)
	record, ok := leafRecords040[m.Albedo]
	if !ok {
		return
	}
	m.Albedo = record.NewAlbedo
	if record.NewNormal != "" {
		m.Normal = record.NewNormal
		if m.NormalScale <= 0 {
			m.NormalScale = 1
		}
	}
	m.Alpha, m.Surface, m.Double = "MASK", "foliage", true
	m.Cutoff = record.Cutoff
	if m.Cutoff <= 0 || m.Cutoff >= 1 {
		m.Cutoff = .35
	}
}

// Each mip was authored in premultiplied alpha, unpremultiplied with coloured
// transparent padding and rescaled for the same alpha-cut coverage. Never run
// glGenerateMipmap on this texture, which would discard that coverage work.
func readLeafMipChain040(reader io.Reader, cap int, upload func(level, width, height int, rgba []byte) error) (int, uint64, error) {
	var header [8]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, 0, err
	}
	count := int(binary.LittleEndian.Uint32(header[4:]))
	if string(header[:4]) != "LMP4" || count < 1 || count > 15 {
		return 0, 0, fmt.Errorf("invalid leaf mip header")
	}
	previousW, previousH, uploaded := 0, 0, 0
	var total uint64
	for sourceLevel := 0; sourceLevel < count; sourceLevel++ {
		var description [12]byte
		if _, err := io.ReadFull(reader, description[:]); err != nil {
			return uploaded, total, err
		}
		w, h := int(binary.LittleEndian.Uint32(description[:4])), int(binary.LittleEndian.Uint32(description[4:8]))
		compressed := int64(binary.LittleEndian.Uint32(description[8:]))
		bytes := uint64(w) * uint64(h) * 4
		if w < 1 || h < 1 || w > 8192 || h > 8192 || bytes > 256<<20 || compressed < 1 || compressed > int64(bytes*2+1024) {
			return uploaded, total, fmt.Errorf("leaf mip exceeds budget")
		}
		if sourceLevel > 0 && (w != max(1, previousW/2) || h != max(1, previousH/2)) {
			return uploaded, total, fmt.Errorf("non-contiguous leaf mip chain")
		}
		previousW, previousH = w, h
		if w > cap || h > cap {
			if _, err := io.CopyN(io.Discard, reader, compressed); err != nil {
				return uploaded, total, err
			}
			continue
		}
		limited := &io.LimitedReader{R: reader, N: compressed}
		decoder, err := zlib.NewReader(limited)
		if err != nil {
			return uploaded, total, err
		}
		pixels := make([]byte, int(bytes))
		_, err = io.ReadFull(decoder, pixels)
		if err == nil {
			var extra [1]byte
			if n, end := decoder.Read(extra[:]); n != 0 || end != io.EOF {
				err = fmt.Errorf("leaf mip expanded size/checksum mismatch: %v", end)
			}
		}
		closeErr := decoder.Close()
		if err == nil {
			err = closeErr
		}
		if err == nil && limited.N != 0 {
			err = fmt.Errorf("leaf mip compressed trailing bytes")
		}
		if err != nil {
			return uploaded, total, err
		}
		if err := upload(uploaded, w, h, pixels); err != nil {
			return uploaded, total, err
		}
		uploaded++
		total += bytes
	}
	if uploaded == 0 || previousW != 1 || previousH != 1 {
		return uploaded, total, fmt.Errorf("incomplete leaf mip pyramid")
	}
	var extra [1]byte
	if n, err := reader.Read(extra[:]); n != 0 || err != io.EOF {
		return uploaded, total, fmt.Errorf("leaf stream trailing bytes/checksum mismatch: %v", err)
	}
	return uploaded, total, nil
}

func pbrLeafTexture040(name string) (uint32, bool) {
	leafManifestOnce040.Do(initLeafManifest040)
	record, match := leafRecords040[name]
	if !match {
		return 0, false
	}
	file, err := assets.Open("assets/pbr/" + record.Mips)
	if err != nil {
		leafStatus040.Error = err.Error()
		return 0, false
	}
	defer file.Close()
	var texture uint32
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&texture)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2703)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT)
	levels, bytes, err := readLeafMipChain040(file, textureCap024(), func(level, w, h int, pixels []byte) error {
		gl("glTexImage2D", GL_TEXTURE_2D, uintptr(level), GL_RGBA, uintptr(w), uintptr(h), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&pixels[0])))
		runtime.KeepAlive(pixels)
		return nil
	})
	if err != nil {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&texture)))
		leafStatus040.Error = record.Mips + ": " + err.Error()
		pbrErrors = append(pbrErrors, leafStatus040.Error)
		return 0, false
	}
	gl("glTexParameteri", GL_TEXTURE_2D, 0x813d, uintptr(levels-1))
	if supportsAnisotropy() {
		gl("glTexParameterf", GL_TEXTURE_2D, 0x84fe, b32(anisotropy026()))
	}
	gl("glTexParameterf", GL_TEXTURE_2D, 0x8501, b32(float64(beta.cfg.TextureBias)))
	leafStatus040.Textures++
	leafStatus040.Levels += levels
	leafStatus040.Bytes += bytes
	return texture, true
}
