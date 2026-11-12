package main

func drawTitle038Legacy(w, h float64) {
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.008, .018, .035, .99)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.014, .037, .058, .06)
	vtx(w*.78, h, 0)
	vtx(w*.78, 0, 0)
	end()
	// Original display lettering has its own coloured bevel; UI text retains
	// the readable distance-field alphabet at every window size.
	hudRect(w*.068, h*.14, w*.055, 3, .17, .86, .96, 1)
	remasterText("OPEN ROAD SERIES  /  BETA 038", w*.068, h*.17, h*.013, cyan)
	beta020TexturedPanel("b038_title_logo.png", w*.055, h*.235, w*.405, w*.405*840/2560)
	hudRect(w*.068, h*.505, w*.274, 2, .99, .46, .27, .95)
	remasterText("CITY STREETS. COAST ROADS. HIGHER GROUND.", w*.068, h*.54, h*.012, white)
	for i, item := range [][2]string{{"start", "START DRIVING"}, {"settings", "SETTINGS"}, {"quit", "EXIT GAME"}} {
		beta020Button(item[0], item[1], w*.068, h*(.61+float64(i)*.092), w*.275, h*.065, i == 0)
	}
	remasterText("V0.5.3  /  NATIVE EDITION", w*.068, h*.93, h*.012, cyan)
}
