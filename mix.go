package main

import (
	"math"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
)

type Mixer struct {
	notes    []Clip
	streamer beep.Streamer
}

func Mix(notes ...Clip) *Mixer {
	streamers := make([]beep.Streamer, 0, len(notes))
	for _, note := range notes {
		streamers = append(streamers, &effects.Volume{Streamer: note, Base: 2, Volume: -math.Log2(float64(len(notes)))})
	}
	return &Mixer{
		notes:    notes,
		streamer: beep.Mix(streamers...),
	}
}

func (m *Mixer) Reset() {
	for _, note := range m.notes {
		note.Reset()
	}
}

func (m *Mixer) Stream(samples [][2]float64) (n int, ok bool) {
	return m.streamer.Stream(samples)
}

func (n *Mixer) Err() error {
	return nil
}
