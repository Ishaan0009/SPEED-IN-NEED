//go:build windows

package main

func aokawaHover(x, y int) {
	fx, fy := float64(x)/float64(app.width), float64(y)/float64(app.height)
	if beta.screen == screenPause && fx > .37 && fx < .63 {
		for i := 0; i < 4; i++ {
			a := .4 + float64(i)*.073
			if fy > a && fy < a+.054 {
				beta.selection = i
				return
			}
		}
	}
	if beta.screen == screenSettings && fx > .095 && fx < .905 && fy >= .257 && fy < .752 {
		row := int((fy-.257)/.055) + settingFirst033()
		if row < len(settingNames[remaster.SettingsTab]) {
			beta.selection = row
		}
	}
}
