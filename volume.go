package main

import (
	"math"

	"github.com/faiface/beep"
)

type volume struct {
	Streamer   beep.Streamer
	Base       float64
	VolumeFunc func() float64
	Silent     bool
}

// Sets the gain according to value returned from func
func (v *volume) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)
	gain := 0.0
	if !v.Silent {
		gain = math.Pow(v.Base, v.VolumeFunc())
	}
	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (v *volume) Err() error {
	return v.Streamer.Err()
}
