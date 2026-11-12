package main

func titleButton035(id, label string, x, y, w, h float64) {
	syncMenuFocus033()
	index := len(ui020.Buttons)
	ui020.Buttons = append(ui020.Buttons, button020{id, x, y, w, h})
	hover := float64(remaster.MouseX) >= x && float64(remaster.MouseX) <= x+w && float64(remaster.MouseY) >= y && float64(remaster.MouseY) <= y+h
	if hover {
		ui020.Hover = index
	}
	active := hover || focus033.ID == id || (focus033.ID == "" && id == "start")
	background := [4]float64{.035, .065, .075, .92}
	text := white
	if active {
		background = [4]float64{.055, .24, .28, .98}
		text = white
	}
	roundRect032(x+2, y+4, w, h, 6, [4]float64{0, 0, 0, .22})
	roundRect032(x, y, w, h, 5, background)
	size := h * .33
	if size > float64(app.height)*.021 {
		size = float64(app.height) * .021
	}
	remasterText(label, x+w*.055, y+(h-size)*.5, size, text)
	if active {
		remasterText(">", x+w*.90, y+(h-size)*.5, size, text)
	}
}

func headline035(text string, x, y, h, width float64, c [4]float64) {
	if w := remasterTextWidth(text, h); w > width {
		h *= width / w
	}
	remasterText(text, x, y, h, c)
}
func drawTitle035(w, h float64) {
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.008, .015, .020, .99)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.008, .015, .020, .05)
	vtx(w*.64, h, 0)
	vtx(w*.64, 0, 0)
	end()
	hudRect(w*.065, h*.10, 34, 3, .45, .88, .80, 1)
	remasterText("SPEED IN NEED  /  OPEN ROAD SERIES", w*.065, h*.135, h*.012, cyan)
	headline035("SPEED", w*.06, h*.225, h*.104, w*.39, white)
	headline035("IN NEED", w*.06, h*.345, h*.104, w*.39, white)
	remasterText("THE CITY. THE COAST. THE OPEN ROAD.", w*.065, h*.505, h*.012, [4]float64{.71, .78, .80, 1})
	beta020Button("start", "START DRIVING", w*.065, h*.60, w*.265, h*.075, true)
	beta020Button("settings", "SETTINGS", w*.065, h*.70, w*.265, h*.052, false)
	beta020Button("quit", "EXIT GAME", w*.065, h*.775, w*.265, h*.052, false)
	hudRect(w*.065, h*.89, w*.29, 1, .30, .39, .42, .8)
	remasterText("V0.5.3  /  NATIVE EDITION", w*.065, h*.925, h*.011, [4]float64{.64, .73, .76, 1})
	remasterText("ROAM FREELY. FIND YOUR OWN ROAD.", w*.65, h*.925, h*.011, white)
}
