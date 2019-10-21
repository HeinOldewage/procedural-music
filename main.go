package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type Note struct {
	SampleRate beep.SampleRate
	Frequency  float64
	Volume     float64
	Duration   time.Duration
	position   int64
	ADSR       struct {
		Attack  *Fader
		Decay   *Fader
		Sustain *Fader
		Release *Fader
	}
}

func NewNote(Frequency float64) *Note {
	return &Note{
		SampleRate: beep.SampleRate(44100),
		Frequency:  Frequency,
		ADSR: struct {
			Attack  *Fader
			Decay   *Fader
			Sustain *Fader
			Release *Fader
		}{},
	}
}

func (no *Note) Stream(samples [][2]float64) (n int, ok bool) {

	totalSamples := int64(no.SampleRate.N(no.Duration))
	if no.position >= totalSamples {
		return 0, false
	}

	for i := range samples {
		x := math.Sin(no.Frequency * 2 * math.Pi * float64(no.position) / float64(no.SampleRate))

		if no.position >= totalSamples {
			return i, false
		}
		fadeVolume := no.ADSR.Attack.Fade(no.position) +
			no.ADSR.Decay.Fade(no.position) +
			no.ADSR.Sustain.Fade(no.position) +
			no.ADSR.Release.Fade(no.position)
		no.position++
		samples[i][0] = x * math.Pow(2, no.Volume+fadeVolume)
		samples[i][1] = x * math.Pow(2, no.Volume+fadeVolume)
	}
	return len(samples), true
}

func (n *Note) Err() error {
	return nil
}

func (n *Note) Reset() {
	n.position = 0
}

func (no *Note) For(dur time.Duration) *Note {
	extraTime := dur % time.Duration(1/no.Frequency*float64(time.Second))
	no.Duration = dur - extraTime

	totalSamples := int64(no.SampleRate.N(no.Duration))

	no.ADSR.Attack = &Fader{StartPosition: 0, EndPosition: totalSamples / 10, StartVolume: -5, EndVolume: 0}
	no.ADSR.Decay = &Fader{StartPosition: totalSamples/10 + 1, EndPosition: (totalSamples / 5) * 2, StartVolume: 0, EndVolume: -1}
	no.ADSR.Sustain = &Fader{StartPosition: (totalSamples/5)*2 + 1, EndPosition: (totalSamples / 5) * 4, StartVolume: -1, EndVolume: -1}
	no.ADSR.Release = &Fader{StartPosition: (totalSamples/5)*4 + 1, EndPosition: totalSamples, StartVolume: -1, EndVolume: -10}
	return no
}

func (no *Note) AtVolume(vol float64) *Note {
	no.Volume = vol
	return no
}

func (no *Note) Freq(freq float64) *Note {
	no.Frequency = freq
	return no
}

func (no *Note) String() string {
	return fmt.Sprint(no.Frequency)
}

func (no *Note) Copy() *Note {
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
	allNotes              = GenNotes()
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

func GenNotes() []*Note {
	res := make([]*Note, 0, 127)
	var a float64 = 440 // a is 440 hz...
	for x := 0; x < 127; x++ {
		res = append(res, NewNote(
			(a/32.0)*math.Pow(2, (float64(x)-9)/12.0),
		))
	}
	return res
}

type Clip interface {
	beep.Streamer
	Reset()
}

func UsingScale(scale []int, allNotes []*Note) []*Note {
	ourNotes := make([]*Note, 0, len(scale))
	nextNote := 0
	for i := range scale {
		ourNotes = append(ourNotes, allNotes[nextNote])
		nextNote += scale[i]
	}
	ourNotes = append(ourNotes, allNotes[nextNote])
	return ourNotes
}

func Notes(ourNotes []*Note) func() *Note {
	noteCounter := 0
	currentNote := rand.Intn(len(ourNotes))
	return func() *Note {
		var volume float64 = -1
		switch (noteCounter) % 4 {
		case 0:
			volume = 0
		case 2:
			volume = -0.5
		}
		noteCounter++

		note := ourNotes[currentNote].Copy().For(time.Millisecond * 100).AtVolume(volume)

		//Choose next note
		noteRange := 3
		if rand.Float64() < 0.10 {
			noteRange = 5
		}

		currentNote = (currentNote + rand.Intn(noteRange) - 1) % len(ourNotes)
		if currentNote < 0 {
			currentNote = len(ourNotes) - 1
		}

		return note

	}

}

type BaseMelody []*Note

func BaseMelodies(size int, sourceNotes func() *Note) func() BaseMelody {
	return func() BaseMelody {
		voice := make([]*Note, 0, size)
		for k := 0; k < size; k++ {
			voice = append(voice, sourceNotes())
		}
		return voice
	}
}

type Voice Clip

func Harmonizers(octaveShift int, active float64) func(melody BaseMelody) Voice {
	return func(melody BaseMelody) Voice {
		voice := []Clip{}
		for _, note := range melody {
			note.Reset()
			harmonies := []*Note{
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(octaveShift))),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(octaveShift)) * math.Pow(2, 4.0/12.0)),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(octaveShift)) * math.Pow(2, 7.0/12.0)),
			}
			interval := rand.Intn(len(harmonies))
			chosen := harmonies[interval]

			if rand.Float64() < active {

				log.Println("Harmonizing", note.Frequency, "with", chosen.Frequency, "interval", interval)
				voice = append(voice,
					Concat(Mix(note.Copy(), chosen),
						&Rest{SampleRate: beep.SampleRate(44100), Duration: time.Millisecond * 150}),
				)
			} else {
				log.Println("Not Harmonizing", note.Frequency)
				voice = append(voice,
					Concat(note.Copy(),
						&Rest{SampleRate: beep.SampleRate(44100), Duration: time.Millisecond * 150}),
				)
			}
		}
		return Concat(voice...)
	}
}

func Improvizer(active float64) func(melody BaseMelody) Voice {
	return func(melody BaseMelody) Voice {
		voice := []Clip{}
		for _, note := range melody {
			note.Reset()
			harmoniesDown := []*Note{
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(-1))),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(-1)) * math.Pow(2, 4.0/12.0)),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(-1)) * math.Pow(2, 7.0/12.0)),
			}
			harmoniesUp := []*Note{
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(1))),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(1)) * math.Pow(2, 4.0/12.0)),
				note.Copy().Freq(note.Frequency * math.Pow(2, float64(1)) * math.Pow(2, 7.0/12.0)),
			}

			intervalDown := rand.Intn(len(harmoniesDown))
			chosenDown := harmoniesDown[intervalDown]
			intervalUp := rand.Intn(len(harmoniesDown))
			chosenUp := harmoniesUp[intervalUp]

			notes := make([]Clip, 0)
			notes = append(notes, note.Copy())
			if rand.Float64() < active {

				log.Println("Harmonizing", note.Frequency, "with", chosenDown.Frequency, "interval", intervalDown)
				notes = append(notes, chosenDown)
			}
			if rand.Float64() < active {
				log.Println("Harmonizing", note.Frequency, "with", chosenUp.Frequency, "interval", intervalUp)
				notes = append(notes, chosenUp)
			}
			voice = append(voice,
				Concat(Mix(notes...), &Rest{SampleRate: beep.SampleRate(44100), Duration: time.Millisecond * 150}),
			)
		}
		return Concat(voice...)
	}
}

type Phrase Clip

type Harmonizer func(melody BaseMelody) Voice

func Phrases(numBarsToRepeat, repeatCount, newCount int, melodies func() BaseMelody, harmonizers []Harmonizer, improvizer func(melody BaseMelody) Voice) func() Phrase {
	return func() Phrase {
		phrase := make([]Clip, 0)
		themes := make([]BaseMelody, 0, numBarsToRepeat)
		for i := 0; i < numBarsToRepeat; i++ {
			themes = append(themes, melodies())
		}

		for k := 0; k < repeatCount; k++ {
			//Theme

			log.Println("Playing theme")
			for _, theme := range themes {
				voices := make([]Clip, 0)
				for _, harmonizer := range harmonizers {
					voices = append(voices, harmonizer(theme))
				}
				phrase = append(phrase, Mix(voices...))

			}

			//improv
			log.Println("Playing improv")
			for k := 0; k < newCount; k++ {
				phrase = append(phrase, improvizer(melodies()))
			}
		}
		/*//theme
		for _, clip := range theme {
			phrase = append(phrase,clip)
		}*/

		return Concat(phrase...)
	}

}

func Song(phraseCount int, phrases func() Phrase) func(func(beep.Streamer)) {
	return func(noteHandler func(beep.Streamer)) {
		log.Println("Playing song")
		for i := 0; i < phraseCount; i++ {
			log.Println("Playing phrase")
			noteHandler(phrases())
		}
		noteHandler(&Rest{SampleRate: beep.SampleRate(44100), Duration: time.Millisecond * 1000})
	}
}

var debug = flag.Bool("debug", false, "debug")
var save = flag.Bool("save", false, "save")

func main() {

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	sr := beep.SampleRate(44100)
	format := beep.Format{SampleRate: sr, NumChannels: 2, Precision: 2}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*50))

	song := make(chan beep.Streamer)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(song)
		defer wg.Done()
		foundationScale := UsingScale(scaleMajor, allNotes[72:72+13])
		Song(10, Phrases(2, 2, 1,
			BaseMelodies(4, Notes(foundationScale)),
			[]Harmonizer{
				Harmonizers(-1, 0.8),
				Harmonizers(-2, 0.3),
			},
			Improvizer(1), //Harmonizer(1, 0.5)
		))(func(note beep.Streamer) {
			song <- note
		})
	}()

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
				case currentStreamer, ok = <-song:
					if !ok {
						return total, false
					}
				}
			}

		}
		return total, true
	})
	if !*save {
		speaker.Play(streamer)
		wg.Wait()
	} else {
		file, err := os.Create("out.wav")
		if err != nil {
			panic(err)
		}
		defer file.Close()
		wav.Encode(file, streamer, format)
	}

}
