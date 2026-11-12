package main

import (
	"encoding/json"
	"strings"
	"unsafe"
)

var font035 struct {
	Program, VBO uint32
	Kern         map[string]float64
	Vertices     [][4]float32
}

func initFont035() {
	raw, err := assets.ReadFile("assets/b037_ui_font_sdf.json")
	if err != nil {
		return
	}
	var data struct {
		Glyphs  map[string]remasterGlyph
		Kerning map[string]float64
	}
	if json.Unmarshal(raw, &data) != nil || app.tex["b037_ui_font_sdf.png"] == 0 {
		return
	}
	vraw, _ := assets.ReadFile("shaders/font035.vert")
	fraw, _ := assets.ReadFile("shaders/font035.frag")
	v, err := compileGLSL(0x8B31, string(vraw))
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		return
	}
	f, err := compileGLSL(0x8B30, string(fraw))
	if err != nil {
		pbrErrors = append(pbrErrors, err.Error())
		glExt("glDeleteShader", uintptr(v))
		return
	}
	p := uint32(glExt("glCreateProgram"))
	glExt("glAttachShader", uintptr(p), uintptr(v))
	glExt("glAttachShader", uintptr(p), uintptr(f))
	glExt("glLinkProgram", uintptr(p))
	glExt("glDeleteShader", uintptr(v))
	glExt("glDeleteShader", uintptr(f))
	var ok int32
	glExt("glGetProgramiv", uintptr(p), 0x8B82, uintptr(unsafe.Pointer(&ok)))
	if ok == 0 {
		pbrErrors = append(pbrErrors, "Font035: "+shaderLog(p, true))
		glExt("glDeleteProgram", uintptr(p))
		return
	}
	font035.Program = p
	font035.Kern = data.Kerning
	remaster.Glyphs = data.Glyphs
	remaster.Atlas = app.tex["b037_ui_font_sdf.png"]
}
func usingFont035() bool {
	return font035.Program != 0 && remaster.Atlas == app.tex["b037_ui_font_sdf.png"]
}
func fontKerning035(prev, r rune) float64 {
	if !usingFont035() || prev == 0 {
		return 0
	}
	return font035.Kern[string([]rune{prev, r})]
}
func drawFont035(s string, x, y, h float64, c [4]float64) {
	font035.Vertices = font035.Vertices[:0]
	var prev rune
	for _, r := range strings.ToUpper(s) {
		g, ok := remaster.Glyphs[string(r)]
		if !ok {
			g = remaster.Glyphs["?"]
		}
		x += fontKerning035(prev, r) * h
		if r != ' ' {
			size := float32(g.Cell * h)
			pad := h * g.Pad
			xx, yy := float32(x-pad), float32(y-pad)
			u, v, du, dv := float32(g.U), float32(g.V), float32(g.DU), float32(g.DV)
			font035.Vertices = append(font035.Vertices, [4]float32{xx, yy, u, v}, [4]float32{xx + size, yy, u + du, v}, [4]float32{xx + size, yy + size, u + du, v + dv}, [4]float32{xx, yy, u, v}, [4]float32{xx + size, yy + size, u + du, v + dv}, [4]float32{xx, yy + size, u, v + dv})
		}
		x += g.Advance * h
		prev = r
	}
	if len(font035.Vertices) == 0 {
		return
	}
	en(GL_TEXTURE_2D)
	glExt("glActiveTexture", 0x84C0)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(remaster.Atlas))
	glExt("glUseProgram", uintptr(font035.Program))
	col(c[0], c[1], c[2], c[3])
	if font035.VBO == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&font035.VBO)))
	}
	glExt("glBindVertexArray", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(font035.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(font035.Vertices))*16, uintptr(unsafe.Pointer(&font035.Vertices[0])), 0x88E0)
	gl("glDisableClientState", glNormalArray)
	gl("glDisableClientState", glColorArray)
	gl("glEnableClientState", glVertexArray)
	gl("glEnableClientState", 0x8078)
	gl("glVertexPointer", 2, glFloat, 16, 0)
	gl("glTexCoordPointer", 2, glFloat, 16, 8)
	gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(len(font035.Vertices)))
	gl("glDisableClientState", glVertexArray)
	gl("glDisableClientState", 0x8078)
	glExt("glBindBuffer", glArrayBuffer, 0)
	glExt("glUseProgram", 0)
	dis(GL_TEXTURE_2D)
}
