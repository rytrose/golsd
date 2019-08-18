package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type audio struct {
	buffers []*beep.Buffer
	formats []beep.Format
	beats   [][]*beep.Buffer
}

func (a *audio) SampleRate() beep.SampleRate {
	return a.formats[0].SampleRate
}

func (a *audio) Speech(p int) beep.StreamSeeker {
	buffer := a.buffers[p]
	return buffer.Streamer(0, buffer.Len())
}

func (a *audio) Beat(p int) beep.StreamSeeker {
	buffer := a.beats[p][rand.Intn(4)]
	return buffer.Streamer(0, buffer.Len())
}

func loadLSDAudio() *audio {
	moot := &sync.Mutex{}
	wg := sync.WaitGroup{}

	s := &audio{
		buffers: make([]*beep.Buffer, 0),
		formats: make([]beep.Format, 0),
		beats:   make([][]*beep.Buffer, 0),
	}

	for _, name := range lsdFilenames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/%s.wav", lsdBaseURL, name)
			reader := readerFromURL(url)
			streamer, format, err := wav.Decode(reader)
			if err != nil {
				panic(fmt.Sprintf("unable to open wav stream: %s", err))
			}

			buffer := beep.NewBuffer(format)
			buffer.Append(streamer)
			streamer.Close()

			bufferBeats := make([]*beep.Buffer, 0)

			bufferMoot := &sync.Mutex{}
			bufferWG := sync.WaitGroup{}
			for i := 1; i < 5; i++ {
				bufferWG.Add(1)
				go func(name string, i int) {
					defer bufferWG.Done()
					url := fmt.Sprintf("%s/%s%d.wav", lsdBaseURL, name, i)
					reader := readerFromURL(url)
					streamer, _, err := wav.Decode(reader)
					if err != nil {
						panic(fmt.Sprintf("unable to open wav stream: %s", err))
					}

					buffer := beep.NewBuffer(format)
					buffer.Append(streamer)
					bufferMoot.Lock()
					bufferBeats = append(bufferBeats, buffer)
					bufferMoot.Unlock()
					streamer.Close()
				}(name, i)
			}
			bufferWG.Wait()

			// Use mutex to ensure buffers, formats, and beats are in the same order
			moot.Lock()
			s.buffers = append(s.buffers, buffer)
			s.formats = append(s.formats, format)
			s.beats = append(s.beats, bufferBeats)
			moot.Unlock()
		}(name)
	}
	wg.Wait()

	return s
}

var (
	ctrl         *beep.Ctrl // Used to play/pause
	startTime    time.Time  // Used in the volume sinusoids
	speechPeriod float64    // Speech volume sinusoid
	beatPeriod   float64    // Beat volume sinusoid
	t            float64    // Time for sinusoid calculations
)

func main() {
	// Seed RNG
	rand.Seed(time.Now().Unix())

	// Fetch audio from s3
	l := loadLSDAudio()

	// Initialize audio settings
	speaker.Init(l.SampleRate(), l.SampleRate().N(time.Second/30))

	// Speech period between 30s and 1m
	speechPeriod = float64(time.Duration(30.0)*time.Second) / float64(time.Second)
	speechPeriod = speechPeriod + (30.0 * rand.Float64())

	// Beat period between 15s and 30s
	beatPeriod = float64(time.Duration(15.0)*time.Second) / float64(time.Second)
	beatPeriod = speechPeriod + (15.0 * rand.Float64())

	// Play speech and beat together
	mixer := &beep.Mixer{}

	// Used to signal end of playback
	done := make(chan bool)

	// Play.pause
	ctrl = &beep.Ctrl{Streamer: mixer, Paused: false}

	// Send to output
	speaker.Play(ctrl)

	// Used for volume sinusoid calc
	startTime = time.Now()

	for {
		// Pick a random speech sample
		fileIndex := rand.Intn(len(l.buffers))

		// Get current time
		t = float64(time.Since(startTime)) / float64(time.Second)

		// Wrap speech sample in volume
		speech := &volume{
			Streamer: beep.Seq(l.Speech(fileIndex), beep.Callback(func() {
				done <- true
			})),
			Base: 2,
			VolumeFunc: func() float64 {
				// Set speech volume (-3 to 2)
				return -0.5 + 2.5*math.Cos(2*math.Pi*t*(1/speechPeriod))
			},
			Silent: false,
		}

		// Wrap beat in volume
		beat := &volume{
			Streamer: l.Beat(fileIndex),
			Base:     2,
			VolumeFunc: func() float64 {
				// Set beat volume (-6 to 0)
				return -3.0 - 3.0*math.Cos(2*math.Pi*t*(1/beatPeriod))
			},
			Silent: false,
		}

		// Plays the Streamers by adding them to the playing mixer
		mixer.Add(speech, beat)

		// Wait for playback to be done
		<-done
	}
}
