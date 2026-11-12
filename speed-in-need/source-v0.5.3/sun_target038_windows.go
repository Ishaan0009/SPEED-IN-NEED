package main

import "unsafe"

func sunTexture038(width, height int) {
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&hdrScene.sun)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(hdrScene.sun))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x881A, uintptr(width), uintptr(height), 0, GL_RGBA, glFloat, 0)
	glExt("glFramebufferTexture2D", 0x8D40, 0x8CE3, GL_TEXTURE_2D, uintptr(hdrScene.sun), 0)
}
