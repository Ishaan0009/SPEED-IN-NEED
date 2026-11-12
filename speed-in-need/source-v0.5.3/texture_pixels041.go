package main

import (
	"image"
	"image/draw"
)

// OpenGL expects tightly packed, straight-alpha pixels. JPEG and PNG decoders
// expose bulk conversion paths; per-pixel interface calls made menu boot slow.
func texturePixels041(im image.Image) []byte {
	b := im.Bounds()
	if n, ok := im.(*image.NRGBA); ok && n.Stride == b.Dx()*4 {
		return n.Pix[:b.Dx()*b.Dy()*4]
	}
	n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(n, n.Bounds(), im, b.Min, draw.Src)
	return n.Pix
}
