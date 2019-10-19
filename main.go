package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
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
	Volume     float64
	Duration   time.Duration
	position   int64
}

func (no *CNote) Stream(samples [][2]float64) (n int, ok bool) {
	var fadeVolume float64 = 0
	totalSamples := int64(no.SampleRate.N(no.Duration))
	if no.position >= totalSamples {
		return 0, false
	}
	var numCylcesToFade float64 = 4
	oneCycle := float64(no.SampleRate.N(time.Duration(1 / no.Frequency * float64(time.Second))))
	totalFadeCycles := oneCycle * numCylcesToFade

	for i := range samples {
		x := math.Sin(no.Frequency * 2 * math.Pi * float64(no.position) / float64(no.SampleRate))

		if float64(totalSamples-no.position) < totalFadeCycles {
			fadeVolume = -(totalFadeCycles - float64(totalSamples-no.position)) / totalFadeCycles
		}
		if float64(no.position) < totalFadeCycles {
			fadeVolume = -(totalFadeCycles - float64(no.position)) / totalFadeCycles
		}

		if no.position >= totalSamples {
			return i, false

		}
		no.position++
		samples[i][0] = x * math.Pow(2, no.Volume+fadeVolume*4)
		samples[i][1] = x * math.Pow(2, no.Volume+fadeVolume*4)
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
	extraTime := dur % time.Duration(1/no.Frequency*float64(time.Second))
	no.Duration = dur - extraTime
	return no
}

func (no *CNote) AtVolume(vol float64) *CNote {
	no.Volume = vol
	return no
}

func (no *CNote) String() string {
	return fmt.Sprint(no.Frequency)
}

func (no *CNote) Copy() *CNote {
	res := *no
	return &res
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

func UpDown(notes []*CNote) chan Note {
	res := make(chan Note)
	go func() {
		for {
			for _, note := range notes {
				res <- note.For(time.Millisecond * 500)
			}
		}
	}()
	return res
}

func Notes(scale []int, numNotes int) chan Note {
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
		noteCounter := 0
		for noteCounter < numNotes {

			var volume float64 = -3

			switch (noteCounter) % 4 {
			case 0:
				volume = -3
			case 2:
				volume = -3
			}

			noteCounter++

			note := ourNotes[currentNote].Copy().For(time.Millisecond * 100).AtVolume(volume)

			notes <- note
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
				log.Println("Playing song")
				for i := 0; i < phraseCount; i++ {
					log.Println("Playing phrase")
					(<-phrases)(song)
				}
				song <- &Rest{SampleRate: beep.SampleRate(441000), Duration: time.Millisecond * 1000}
			}
		}()
	}
}

var debug = flag.Bool("debug", false, "debug")
var save = flag.Bool("save", false, "save")

func main() {

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	sr := beep.SampleRate(441000)
	format := beep.Format{SampleRate: sr, NumChannels: 2, Precision: 2}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*50))

	song := make(chan beep.Streamer)

	var noteSource chan Note
	noteSource = Notes(scaleMelodicMinorDown, 8*4*1)
	if *debug {
		noteSource = UpDown([]*CNote{C4, C4})
	}

	Song(10, Phrases(2, 2, 4, Bars(4, noteSource)))(song)

	var currentStreamer beep.Streamer = &Rest{}
	streamer := beep.StreamerFunc(func(samples [][2]float64) (int, bool) {
		total := 0
		for i := range samples {
			samples[i][0] = 0
			samples[i][1] = 0
		}
		for total < len(samples) {
			n, ok := currentStreamer.Stream(samples[total:])
			total += n
			if !ok {
				select {
				case currentStreamer = <-song:
				case <-time.NewTimer(time.Second).C:
					return total, false
				}

			}

		}
		return total, true
	})
	if !*save {
		speaker.Play(streamer)
		select {}
	} else {
		file, err := os.Create("out.wav")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		wav.Encode(file, streamer, format)
	}

}
