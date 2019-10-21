package main

type Fader struct {
	StartPosition int64
	EndPosition   int64
	StartVolume   float64
	EndVolume     float64
}

//Returns a Db adjustement
func (f *Fader) Fade(position int64) float64 {
	if position < f.StartPosition || position > f.EndPosition {
		return 0
	}
	return f.StartVolume + (f.EndVolume-f.StartVolume)*(float64(position-f.StartPosition)/float64(f.EndPosition-f.StartPosition))
}
