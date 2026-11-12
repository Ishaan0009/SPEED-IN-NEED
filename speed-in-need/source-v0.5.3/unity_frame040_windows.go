package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	unityFrameHeader040       = 64
	unityFrameMaxWidth040     = 3840
	unityFrameMaxHeight040    = 2160
	unityFrameSlotBytes040    = unityFrameMaxWidth040 * unityFrameMaxHeight040 * 4
	unityFrameMappingBytes040 = unityFrameHeader040 + 2*unityFrameSlotBytes040
	unityFrameInterval040     = time.Second / 30
)

var unityFrame040 struct {
	File             *os.File
	Mapping, Address uintptr
	Bytes            []byte
	Path, Error      string
	Last             time.Time
}

// Standard portable layouts are GameRuntime/../Logs and EXE/Logs. Never let a
// preview command line truncate an arbitrary document or other project file.
func unityFramePath040(request, executable string) (string, error) {
	if !filepath.IsAbs(request) || !strings.EqualFold(filepath.Base(request), "frame040.bin") {
		return "", errors.New("Unity frame path must be an absolute Logs/frame040.bin path")
	}
	path := filepath.Clean(request)
	exeDir := filepath.Dir(executable)
	for _, base := range []string{exeDir, filepath.Dir(exeDir)} {
		expected := filepath.Join(base, "Logs", "frame040.bin")
		if strings.EqualFold(path, expected) {
			return path, nil
		}
	}
	return "", errors.New("Unity frame path must be in the game's or Unity project's Logs folder")
}

func unityFrameArguments040(args []string) (string, error) {
	path := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "--unity-frame=") {
			if path != "" {
				return "", errors.New("duplicate Unity frame path")
			}
			path = strings.TrimPrefix(arg, "--unity-frame=")
			if path == "" {
				return "", errors.New("empty Unity frame path")
			}
		}
	}
	return path, nil
}

func unityFrameSize040(width, height int) (int, bool) {
	if width < 1 || height < 1 || width > unityFrameMaxWidth040 || height > unityFrameMaxHeight040 {
		return 0, false
	}
	return width * height * 4, true
}

func unityFrameOffset040(slot int) (int, bool) {
	if slot < 0 || slot > 1 {
		return 0, false
	}
	return unityFrameHeader040 + slot*unityFrameSlotBytes040, true
}

func unityFrameDue040(last, now time.Time) bool {
	return last.IsZero() || now.Sub(last) >= unityFrameInterval040
}

func unityFrameSequence040(mapped []byte) *uint32 {
	return (*uint32)(unsafe.Pointer(&mapped[12]))
}

// Sequence is a seqlock: odd means a write is in progress, nonzero even is a
// complete publication. Reader copies only a stable matching even sequence.
func beginUnityFrame040(mapped []byte) uint32 {
	next := atomic.LoadUint32(unityFrameSequence040(mapped))
	if next&1 != 0 {
		next++
	}
	next++
	atomic.StoreUint32(unityFrameSequence040(mapped), next)
	return next
}

func publishUnityFrame040(mapped []byte, width, height, slot int, writing uint32) bool {
	bytes, valid := unityFrameSize040(width, height)
	if !valid || len(mapped) < unityFrameHeader040 || slot < 0 || slot > 1 || writing&1 == 0 {
		return false
	}
	binary.LittleEndian.PutUint32(mapped[4:8], uint32(width))
	binary.LittleEndian.PutUint32(mapped[8:12], uint32(height))
	binary.LittleEndian.PutUint32(mapped[16:20], uint32(slot))
	binary.LittleEndian.PutUint32(mapped[20:24], uint32(bytes))
	next := writing + 1
	if next == 0 {
		next = 2
	}
	atomic.StoreUint32(unityFrameSequence040(mapped), next)
	return true
}

// Call after configureUnityHost040; GL is not needed until the export hook.
func initUnityFrame040() {
	path, err := unityFrameArguments040(os.Args[1:])
	if path == "" && err == nil {
		return
	}
	if err == nil && unityHost040.Parent == 0 {
		err = errors.New("Unity frame preview requires a Unity parent window")
	}
	var exe string
	if err == nil {
		exe, err = os.Executable()
	}
	if err == nil {
		path, err = unityFramePath040(path, exe)
	}
	if err == nil {
		err = openUnityFrame040(path)
	}
	if err != nil {
		unityFrame040.Error = err.Error()
		unityStage040("frame mapping unavailable: " + err.Error())
	}
}

func openUnityFrame040(path string) error {
	closeUnityFrame040()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err = file.Truncate(unityFrameMappingBytes040); err != nil {
		file.Close()
		return err
	}
	mapping, _, callErr := kernel32.NewProc("CreateFileMappingW").Call(file.Fd(), 0, 4, 0, unityFrameMappingBytes040, 0)
	if mapping == 0 {
		file.Close()
		return fmt.Errorf("CreateFileMapping: %v", callErr)
	}
	address, _, callErr := kernel32.NewProc("MapViewOfFile").Call(mapping, 2, 0, 0, unityFrameMappingBytes040)
	if address == 0 {
		kernel32.NewProc("CloseHandle").Call(mapping)
		file.Close()
		return fmt.Errorf("MapViewOfFile: %v", callErr)
	}
	unityFrame040.File, unityFrame040.Mapping, unityFrame040.Address = file, mapping, address
	unityFrame040.Bytes = unsafe.Slice((*byte)(unsafe.Pointer(address)), unityFrameMappingBytes040)
	unityFrame040.Path, unityFrame040.Error, unityFrame040.Last = path, "", time.Time{}
	clear(unityFrame040.Bytes[:unityFrameHeader040])
	copy(unityFrame040.Bytes[:4], "SIN4")
	unityStage040(fmt.Sprintf("frame mapping ready %s bytes=%d", path, unityFrameMappingBytes040))
	return nil
}

// Call on the rendering OS thread after the final scene + HUD, immediately
// before present038/SwapBuffers. Normal native play never enables this path.
// Pixels are RGBA8, tightly packed and bottom row first, as glReadPixels emits.
func exportUnityFrame040() {
	mapped := unityFrame040.Bytes
	if len(mapped) == 0 || app.minimized || capture != nil {
		return
	}
	if outputFBO038 != 0 {
		unityFrame040.Error = "Unity frame preview requires the OpenGL presentation path"
		return
	}
	now := time.Now()
	if !unityFrameDue040(unityFrame040.Last, now) {
		return
	}
	width, height := app.width, app.height
	if _, valid := unityFrameSize040(width, height); !valid {
		unityFrame040.Error = "Unity preview viewport exceeds 3840 x 2160"
		return
	}
	unityFrame040.Last = now
	slot := 1 - int(binary.LittleEndian.Uint32(mapped[16:20])&1)
	offset, _ := unityFrameOffset040(slot)
	write := beginUnityFrame040(mapped)
	// Preserve pixel-pack/read state so screenshots and native rendering remain
	// unchanged. The preview explicitly reads the final default back buffer.
	var state [7]int32
	keys := [...]uintptr{0x0C02, 0x0D05, 0x0D02, 0x0D03, 0x0D04, 0x88ED, 0x8CAA}
	for i, key := range keys {
		gl("glGetIntegerv", key, uintptr(unsafe.Pointer(&state[i])))
	}
	glExt("glBindFramebuffer", 0x8CA8, 0)
	glExt("glBindBuffer", 0x88EB, 0)
	gl("glReadBuffer", 0x405)
	gl("glPixelStorei", 0x0D05, 1)
	for _, key := range []uintptr{0x0D02, 0x0D03, 0x0D04} {
		gl("glPixelStorei", key, 0)
	}
	gl("glReadPixels", 0, 0, uintptr(width), uintptr(height), GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&mapped[offset])))
	glExt("glBindFramebuffer", 0x8CA8, uintptr(state[6]))
	gl("glReadBuffer", uintptr(state[0]))
	glExt("glBindBuffer", 0x88EB, uintptr(state[5]))
	for i, key := range keys[1:5] {
		gl("glPixelStorei", key, uintptr(state[i+1]))
	}
	publishUnityFrame040(mapped, width, height, slot, write)
}

func closeUnityFrame040() {
	if len(unityFrame040.Bytes) >= unityFrameHeader040 {
		atomic.StoreUint32(unityFrameSequence040(unityFrame040.Bytes), 0)
	}
	unityFrame040.Bytes = nil
	if unityFrame040.Address != 0 {
		kernel32.NewProc("UnmapViewOfFile").Call(unityFrame040.Address)
		unityFrame040.Address = 0
	}
	if unityFrame040.Mapping != 0 {
		kernel32.NewProc("CloseHandle").Call(unityFrame040.Mapping)
		unityFrame040.Mapping = 0
	}
	if unityFrame040.File != nil {
		unityFrame040.File.Close()
		unityFrame040.File = nil
	}
	unityFrame040.Last = time.Time{}
}
