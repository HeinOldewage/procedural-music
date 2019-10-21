package main

import (
	"testing"
)

func TestFade(t *testing.T) {
	fader := &Fader{
		StartPosition: 0,
		EndPosition:   2,
		StartVolume:   -4,
		EndVolume:     0,
	}
	volume := fader.Fade(0)
	if volume != -4 {
		t.Error("Volume not correct at start")
	}
	volume = fader.Fade(2)
	if volume != 0 {
		t.Error("Volume not correct at start")
	}
	volume = fader.Fade(1)
	if volume != -2 {
		t.Error("Volume not correct at start")
	}
}
