//go:build windows

package main

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"unsafe"
)

var lightPipeline struct {
	program                           uint32
	active                            bool
	vehicleLoc                        int32
	samplerLoc                        int32
	pbrLoc, metallicLoc, roughnessLoc int32
	errorText                         string
	ext                               map[string]uintptr
}

// Native arguments may contain Go pointers. Keep their backing storage alive
// and off the movable stack until the driver call returns.
//
//go:uintptrescapes
func glExt(name string, args ...uintptr) uintptr {
	if outputFBO038 != 0 && name == "glBindFramebuffer" && len(args) == 2 && args[0] == 0x8D40 && args[1] == 0 {
		args[1] = uintptr(outputFBO038)
	}
	if lightPipeline.ext == nil {
		lightPipeline.ext = map[string]uintptr{}
	}
	ptr, cached := lightPipeline.ext[name]
	if !cached {
		text := append([]byte(name), 0)
		ptr = win(ogl, "wglGetProcAddress", uintptr(unsafe.Pointer(&text[0])))
		if ptr <= 3 || ptr == ^uintptr(0) {
			ptr = 0
		}
		lightPipeline.ext[name] = ptr
	}
	if ptr == 0 {
		return 0
	}
	if skipNativeGL035(name, args) {
		return 0
	}
	r, _, lastError038 := syscall.SyscallN(ptr, args...)
	if strings.HasPrefix(name, "wglDX") {
		interopError038 = lastError038
	}
	return r
}

func glExtAvailable(name string) bool {
	if lightPipeline.ext == nil {
		lightPipeline.ext = map[string]uintptr{}
	}
	if ptr, ok := lightPipeline.ext[name]; ok {
		return ptr != 0
	}
	text := append([]byte(name), 0)
	ptr := win(ogl, "wglGetProcAddress", uintptr(unsafe.Pointer(&text[0])))
	if ptr <= 3 || ptr == ^uintptr(0) {
		ptr = 0
	}
	lightPipeline.ext[name] = ptr
	return ptr != 0
}

func shaderLog(shader uint32, program bool) string {
	var n int32
	if program {
		glExt("glGetProgramiv", uintptr(shader), 0x8B84, uintptr(unsafe.Pointer(&n)))
	} else {
		glExt("glGetShaderiv", uintptr(shader), 0x8B84, uintptr(unsafe.Pointer(&n)))
	}
	if n <= 1 {
		return ""
	}
	if n > 16384 {
		n = 16384
	}
	buf := make([]byte, n+1)
	var used int32
	if program {
		glExt("glGetProgramInfoLog", uintptr(shader), uintptr(n), uintptr(unsafe.Pointer(&used)), uintptr(unsafe.Pointer(&buf[0])))
	} else {
		glExt("glGetShaderInfoLog", uintptr(shader), uintptr(n), uintptr(unsafe.Pointer(&used)), uintptr(unsafe.Pointer(&buf[0])))
	}
	if used < 0 {
		used = 0
	}
	if used > n {
		used = n
	}
	return strings.TrimSpace(string(buf[:used]))
}

func compileGLSL(kind uintptr, source string) (uint32, error) {
	id := uint32(glExt("glCreateShader", kind))
	if id == 0 {
		return 0, fmt.Errorf("glCreateShader failed")
	}
	bytes := append([]byte(source), 0)
	ptr := uintptr(unsafe.Pointer(&bytes[0]))
	glExt("glShaderSource", uintptr(id), 1, uintptr(unsafe.Pointer(&ptr)), 0)
	glExt("glCompileShader", uintptr(id))
	var status int32
	glExt("glGetShaderiv", uintptr(id), 0x8B81, uintptr(unsafe.Pointer(&status)))
	if status == 0 {
		err := fmt.Errorf("GLSL compile failed: %s", shaderLog(id, false))
		glExt("glDeleteShader", uintptr(id))
		return 0, err
	}
	return id, nil
}

func initLighting() {
	required := []string{"glCreateShader", "glShaderSource", "glCompileShader", "glGetShaderiv", "glCreateProgram", "glAttachShader", "glLinkProgram", "glGetProgramiv", "glUseProgram", "glGetUniformLocation", "glUniform1i", "glUniform1f", "glUniform3fv"}
	for _, name := range required {
		if !glExtAvailable(name) {
			lightPipeline.errorText = "GLSL unsupported: " + name
			return
		}
	}
	vraw, ev := assets.ReadFile("shaders/lighting.vert")
	fraw, ef := assets.ReadFile("shaders/lighting.frag")
	if ev != nil || ef != nil {
		lightPipeline.errorText = "Shader assets unavailable"
		return
	}
	v, err := compileGLSL(0x8B31, string(vraw))
	if err != nil {
		lightPipeline.errorText = err.Error()
		return
	}
	f, err := compileGLSL(0x8B30, string(fraw))
	if err != nil {
		lightPipeline.errorText = err.Error()
		glExt("glDeleteShader", uintptr(v))
		return
	}
	program := uint32(glExt("glCreateProgram"))
	glExt("glAttachShader", uintptr(program), uintptr(v))
	glExt("glAttachShader", uintptr(program), uintptr(f))
	// Keep the instance stream away from legacy position/color/UV aliases.
	if glExtAvailable("glBindAttribLocation") {
		previous := []byte("aPreviousPosition038\x00")
		glExt("glBindAttribLocation", uintptr(program), 11, uintptr(unsafe.Pointer(&previous[0])))
		yaw := []byte("aInstanceYaw\x00")
		glExt("glBindAttribLocation", uintptr(program), 5, uintptr(unsafe.Pointer(&yaw[0])))
		tangent := []byte("aTangent\x00")
		glExt("glBindAttribLocation", uintptr(program), 7, uintptr(unsafe.Pointer(&tangent[0])))
		name := []byte("aInstance\x00")
		glExt("glBindAttribLocation", uintptr(program), 6, uintptr(unsafe.Pointer(&name[0])))
	}
	glExt("glLinkProgram", uintptr(program))
	var status int32
	glExt("glGetProgramiv", uintptr(program), 0x8B82, uintptr(unsafe.Pointer(&status)))
	glExt("glDeleteShader", uintptr(v))
	glExt("glDeleteShader", uintptr(f))
	if status == 0 {
		lightPipeline.errorText = "GLSL link failed: " + shaderLog(program, true)
		glExt("glDeleteProgram", uintptr(program))
		return
	}
	lightPipeline.program = program
	vName := []byte("uVehicle\x00")
	tName := []byte("uDiffuse\x00")
	lightPipeline.vehicleLoc = int32(glExt("glGetUniformLocation", uintptr(program), uintptr(unsafe.Pointer(&vName[0]))))
	lightPipeline.samplerLoc = int32(glExt("glGetUniformLocation", uintptr(program), uintptr(unsafe.Pointer(&tName[0]))))
	for i, name := range []string{"uPBR", "uMetallic", "uRoughness"} {
		data := append([]byte(name), 0)
		loc := int32(glExt("glGetUniformLocation", uintptr(program), uintptr(unsafe.Pointer(&data[0]))))
		switch i {
		case 0:
			lightPipeline.pbrLoc = loc
		case 1:
			lightPipeline.metallicLoc = loc
		case 2:
			lightPipeline.roughnessLoc = loc
		}
	}
	glExt("glUseProgram", uintptr(program))
	if lightPipeline.samplerLoc >= 0 {
		glExt("glUniform1i", uintptr(lightPipeline.samplerLoc), 0)
	}
	for i, name := range []string{"uDiffuse", "uMRMap", "uNormalMap", "uAOMap", "uEnvironment", "uShadowMap", "uShadowMap2", "uEmissiveMap"} {
		data := append([]byte(name), 0)
		loc := int32(glExt("glGetUniformLocation", uintptr(program), uintptr(unsafe.Pointer(&data[0]))))
		if loc >= 0 {
			glExt("glUniform1i", uintptr(loc), uintptr(i))
		}
	}
	glExt("glUseProgram", 0)
}

func useLighting(enabled bool) {
	if !enabled || lightPipeline.program == 0 {
		if lightPipeline.active {
			glExt("glUseProgram", 0)
		}
		lightPipeline.active = false
		return
	}
	glExt("glUseProgram", uintptr(lightPipeline.program))
	lightPipeline.active = true
	setWorldFade(1)
	setInstancedWorld(false)
	setVehicleGloss(false)
	setPBRMaterial(false, 0, .8)
}

func setVehicleGloss(vehicle bool) {
	if !lightPipeline.active || lightPipeline.vehicleLoc < 0 {
		return
	}
	value := uintptr(0)
	if vehicle {
		value = 1
	}
	glExt("glUniform1i", uintptr(lightPipeline.vehicleLoc), value)
}

// Keep normal direction per face for the GLSL fallback-friendly legacy meshes.
func meshFaceNormal(a, b, c [3]float64) {
	x1, y1, z1 := b[0]-a[0], b[1]-a[1], b[2]-a[2]
	x2, y2, z2 := c[0]-a[0], c[1]-a[1], c[2]-a[2]
	x, y, z := y1*z2-z1*y2, z1*x2-x1*z2, x1*y2-y1*x2
	n := math.Sqrt(x*x + y*y + z*z)
	if n < 1e-7 {
		x, y, z, n = 0, 1, 0, 1
	}
	if capture != nil {
		capture.normal = pos3{x / n, y / n, z / n}
		return
	}
	data := [3]float32{float32(x / n), float32(y / n), float32(z / n)}
	gl("glNormal3fv", uintptr(unsafe.Pointer(&data[0])))
}

// Set material category without reconstructing geometry or GPU buffers.
func setPBRMaterial(enabled bool, metallic, roughness float64) {
	sceneInt("uPBR", boolInt(enabled))
	sceneFloat("uMetallic", metallic)
	sceneFloat("uRoughness", roughness)
}
