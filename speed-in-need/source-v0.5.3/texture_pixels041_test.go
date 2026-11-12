package main

import (
	"image"
	"image/color"
	"testing"
)

func TestTexturePixelsSubimageAndStraightAlpha041(t *testing.T) {
	im := image.NewNRGBA(image.Rect(10, 20, 14, 23))
	im.SetNRGBA(11, 21, color.NRGBA{200, 100, 50, 128})
	sub := im.SubImage(image.Rect(11, 21, 13, 23))
	got := texturePixels041(sub)
	if len(got) != 16 || got[0] != 200 || got[1] != 100 || got[2] != 50 || got[3] != 128 || got[7] != 0 {
		t.Fatalf("subimage conversion changed pixels or included stride padding: %v", got)
	}
	premultiplied := image.NewRGBA(image.Rect(0, 0, 1, 1))
	premultiplied.Set(0, 0, color.NRGBA{200, 100, 50, 128})
	got = texturePixels041(premultiplied)
	if got[0] < 198 || got[0] > 201 || got[3] != 128 {
		t.Fatalf("premultiplied input was not converted to straight alpha: %v", got)
	}
}
