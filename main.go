package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type lsd struct {
	buffers []*beep.Buffer
	formats []beep.Format
	beats   [][]*beep.Buffer
}

func (l *lsd) SampleRate() beep.SampleRate {
	return l.formats[0].SampleRate
}

func (l *lsd) Speech(p int) beep.StreamSeeker {
	buffer := l.buffers[p]
	return buffer.Streamer(0, buffer.Len())
}

func (l *lsd) Beat(p int) beep.StreamSeeker {
	buffer := l.beats[p][rand.Intn(4)]
	return buffer.Streamer(0, buffer.Len())
}

func loadLSD() *lsd {
	sounds := &lsd{
		buffers: make([]*beep.Buffer, 0),
		formats: make([]beep.Format, 0),
		beats:   make([][]*beep.Buffer, 0),
	}

	for _, name := range lsdFilenames {
		url := fmt.Sprintf("%s/%s.wav", lsdBaseURL, name)
		reader := readerFromURL(url)
		streamer, format, err := wav.Decode(reader)
		if err != nil {
			panic(fmt.Sprintf("unable to open wav stream: %s", err))
		}

		sounds.formats = append(sounds.formats, format)
		buffer := beep.NewBuffer(format)
		buffer.Append(streamer)
		sounds.buffers = append(sounds.buffers, buffer)
		streamer.Close()

		if stringInSlice(name, beatFilenames) {
			bufferBeats := make([]*beep.Buffer, 0)
			for i := 1; i < 5; i++ {
				url := fmt.Sprintf("%s/%s%d.wav", lsdBaseURL, name, i)
				reader := readerFromURL(url)
				streamer, _, err := wav.Decode(reader)
				if err != nil {
					panic(fmt.Sprintf("unable to open wav stream: %s", err))
				}

				buffer := beep.NewBuffer(format)
				buffer.Append(streamer)
				bufferBeats = append(bufferBeats, buffer)
				streamer.Close()
			}
			sounds.beats = append(sounds.beats, bufferBeats)
		}
	}

	return sounds
}

func readerFromURL(url string) io.ReadCloser {
	fmt.Println(fmt.Sprintf("Getting: %s", url))
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Sprintf("Couldn't download file %s: %s", url, err))
	}

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Unable to read file to bytes %s: %s", url, err))
	}
	resp.Body.Close()

	return ioutil.NopCloser(bytes.NewReader(fileBytes))
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func main() {
	l := loadLSD()

	speaker.Init(l.SampleRate(), l.SampleRate().N(time.Second/30))

	done := make(chan bool)

	mixer := beep.Mixer{}
	speaker.Play(&mixer)

	for {
		fileIndex := rand.Intn(len(l.buffers))
		if fileIndex < len(l.beats) {
			mixer.Add(beep.Seq(l.Speech(fileIndex), beep.Callback(func() {
				done <- true
			})), l.Beat(fileIndex))
		} else {
			speaker.Play(beep.Seq(l.Speech(fileIndex), beep.Callback(func() {
				done <- true
			})))
		}
		<-done
	}
}
