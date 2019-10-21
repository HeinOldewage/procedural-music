package main

import (
	"github.com/faiface/beep"
)

type Concenated struct {
	notes    []Clip
	streamer beep.Streamer
}

func Concat(notes ...Clip) *Mixer {
	streamers := make([]beep.Streamer, 0, len(notes))
	for _, note := range notes {
		streamers = append(streamers, note)
	}
	return &Mixer{
		notes:    notes,
		streamer: beep.Seq(streamers...),
	}
}

func (m *Concenated) Reset() {
	for _, note := range m.notes {
		note.Reset()
	}
}

func (m *Concenated) Stream(samples [][2]float64) (n int, ok bool) {
	return m.streamer.Stream(samples)
}

func (n *Concenated) Err() error {
	return nil
}
