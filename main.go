package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

var (
	C4 = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 261.6255653006}
	Db = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 277.1826309769}
	D  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 293.6647679174}
	Eb = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 311.1269837221}
	E  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 329.6275569129}
	F  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 349.2282314330}
	Gb = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 369.9944227116}
	G  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 391.9954359817}
	Ab = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 415.3046975799}
	A  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 440.0000000000}
	Bb = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 466.1637615181}
	B  = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 493.8833012561}
	C7 = &CNote{SampleRate: beep.SampleRate(441000), Frequency: 523.2511306012}
)

type CNote struct {
	SampleRate beep.SampleRate
	Frequency  float64
	Duration   time.Duration
	position   int64
}

func (no *CNote) Stream(samples [][2]float64) (n int, ok bool) {
	if no.position >= int64(no.SampleRate.N(no.Duration)) {
		return 0, false
	}
	for i := range samples {
		x := math.Sin(no.Frequency * 2 * math.Pi * float64(no.position) / float64(no.SampleRate))

		if no.position >= int64(no.SampleRate.N(no.Duration)) {
			return i, false
		}
		no.position++
		samples[i][0] = x
		samples[i][1] = x
	}
	return len(samples), true
}

func (n *CNote) Err() error {
	return nil
}

func (n *CNote) Reset() {
	n.position = 0
}

func (no *CNote) For(dur time.Duration) *CNote {
	res := *no

	extraTime := dur % time.Duration(1/no.Frequency*float64(time.Second))
	res.Duration = dur - extraTime
	return &res
}

func (no *CNote) String() string {
	return fmt.Sprint(no.Frequency)
}

type Rest struct {
	SampleRate beep.SampleRate
	Duration   time.Duration
	position   int64
}

func (no *Rest) Reset() {
	no.position = 0
}

func (no *Rest) Stream(samples [][2]float64) (n int, ok bool) {
	if no.position >= int64(no.SampleRate.N(no.Duration)) {
		return 0, false
	}
	for i := range samples {
		if no.position >= int64(no.SampleRate.N(no.Duration)) {
			return i, false
		}
		no.position++
		samples[i][0] = 0
		samples[i][1] = 0
	}
	return len(samples), true
}

func (no *Rest) For(dur time.Duration) *Rest {
	res := *no
	res.Duration = dur
	return &res
}

func (n *Rest) Err() error {
	return nil
}

var freq = flag.Float64("f", 261.6255653006, "frequency")

var (
	allNotes              = []*CNote{C4, Db, D, Eb, E, F, Gb, G, Ab, A, Bb, B, C7}
	scaleChromatic        = []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1} // (random, atonal: all twelve notes)
	scaleMajor            = []int{2, 2, 1, 2, 2, 2, 1}             // (classic, happy)
	scaleHarmonicMinor    = []int{2, 1, 2, 2, 1, 3, 1}             // (haunting, creepy)
	scaleMinorPentatonic  = []int{3, 2, 2, 3, 2}                   // (blues, rock)
	scaleNaturalMinor     = []int{2, 1, 2, 2, 1, 2, 2}             // (scary, epic)
	scaleMelodicMinorUp   = []int{2, 1, 2, 2, 2, 2, 1}             // (wistful, mysterious)
	scaleMelodicMinorDown = []int{2, 2, 1, 2, 2, 1, 2}             // (sombre, soulful)
	scaleDorian           = []int{2, 1, 2, 2, 2, 1, 2}             // (cool, jazzy)
	scaleMixolydian       = []int{2, 2, 1, 2, 2, 1, 2}             // (progressive, complex)
	scaleAhavaRaba        = []int{1, 3, 1, 2, 1, 2, 2}             // (exotic, unfamiliar)
	scaleMajorPentatonic  = []int{2, 2, 3, 2, 3}                   // (country, gleeful)
	scaleDiatonic         = []int{2, 2, 2, 2, 2, 2}                // (bizarre, symmetrical)
)

type Note interface {
	beep.Streamer
	Reset()
}

func Notes(scale []int) chan Note {
	notes := make(chan Note)
	rest := &Rest{SampleRate: beep.SampleRate(441000)}
	go func() {
		ourNotes := make([]*CNote, 0, len(scale))
		nextNote := 0
		for i := range scale {
			ourNotes = append(ourNotes, allNotes[nextNote])
			nextNote += scale[i]
		}
		ourNotes = append(ourNotes, allNotes[nextNote])
		currentNote := rand.Intn(len(ourNotes))
		for {
			notes <- ourNotes[currentNote].For(time.Millisecond * 100)
			notes <- rest.For(time.Millisecond * 150)

			noteRange := 3
			if rand.Float64() < 0.10 {
				noteRange = 5
			}

			currentNote = (currentNote + rand.Intn(noteRange) - 1) % len(ourNotes)
			if currentNote < 0 {
				currentNote = len(ourNotes) - 1
			}

		}
	}()

	return notes
}

type Bar func(bars chan beep.Streamer)

func Bars(size int, sourceNotes chan Note) chan Bar {
	res := make(chan Bar)
	go func() {
		for {
			bar := make([]Note, 0, size)
			for k := 0; k < size; k++ {
				bar = append(bar, <-sourceNotes)
			}
			res <- func(song chan beep.Streamer) {
				for _, note := range bar {
					note.Reset()
					song <- note
				}
			}
		}

	}()
	return res
}

func Phrases(numBarsToRepeat, repeatCount, newCount int, bars chan Bar) chan func(chan beep.Streamer) {
	res := make(chan func(chan beep.Streamer))

	go func() {
		for {
			res <- func(song chan beep.Streamer) {
				theme := make([]Bar, 0, numBarsToRepeat)

				for i := 0; i < numBarsToRepeat; i++ {
					theme = append(theme, <-bars)
				}

				for k := 0; k < repeatCount; k++ {
					//Theme
					log.Println("Playing theme")
					for i := 0; i < numBarsToRepeat; i++ {
						theme[i](song)
					}
					//improv
					log.Println("Playing improv")
					for k := 0; k < newCount; k++ {
						(<-bars)(song)
					}
				}
				//theme
				log.Println("Playing theme")
				for i := 0; i < numBarsToRepeat; i++ {
					theme[i](song)
				}

			}
		}
	}()

	return res
}

func Song(phraseCount int, phrases chan func(chan beep.Streamer)) func(chan beep.Streamer) {
	return func(song chan beep.Streamer) {
		go func() {
			for {
				for i := 0; i < phraseCount; i++ {
					log.Println("Playing prase")
					(<-phrases)(song)
				}
				song <- &Rest{SampleRate: beep.SampleRate(441000), Duration: time.Millisecond * 100}
			}
		}()
	}
}

func main() {

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	sr := beep.SampleRate(441000)
	format := beep.Format{SampleRate: sr, NumChannels: 2, Precision: 6}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*50))

	song := make(chan beep.Streamer)

	Song(10, Phrases(2, 2, 4, Bars(8, Notes(scaleMajor))))(song)

	noteCounter := 0
	var currentStreamer beep.Streamer = &Rest{}
	speaker.Play(beep.StreamerFunc(func(samples [][2]float64) (int, bool) {
		total := 0
		for i := range samples {
			samples[i][0] = 0
			samples[i][1] = 0
		}
		if true {
			for total < len(samples) {
				n, ok := currentStreamer.Stream(samples[total:])
				total += n
				if !ok {
					note := <-song

					var volume float64 = -10

					switch (noteCounter / 2) % 4 {
					case 0:
						volume = 0.1
					case 2:
						volume = 0.2
					}

					noteCounter++

					currentStreamer = note
					if false {
						currentStreamer = &effects.Volume{Streamer: note, Base: 2, Volume: volume}
					}
				}
			}
		}
		return total, true
	}))

	select {}
}
