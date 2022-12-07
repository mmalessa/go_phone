package phoneaudio

import (
	"context"
	"math"
	"time"

	"github.com/gordonklaus/portaudio"
)

// https://en.wikipedia.org/wiki/Busy_signal
func (pa *PhoneAudio) BusyTone(play_time int) error {
	s := newStereoSine(425, 425, float64(streamSampleRate), 0.5)
	defer s.Close()
	tone_time := 500
	pause_time := 500
	return pa.playTone(s, play_time, tone_time, pause_time)
}

// https://en.wikipedia.org/wiki/Ringing_tone
func (pa *PhoneAudio) RingingTone(play_time int) error {
	s := newStereoSine(425, 425, float64(streamSampleRate), 0.5)
	defer s.Close()
	tone_time := 1000
	pause_time := 2000
	return pa.playTone(s, play_time, tone_time, pause_time)
}

func (pa *PhoneAudio) Beep(tone_time int) error {
	s := newStereoSine(1000, 1000, float64(streamSampleRate), 0.2)
	defer s.Close()
	pause_time := 200
	return pa.playTone(s, tone_time+pause_time, tone_time, pause_time)
}

func (pa *PhoneAudio) ErrorTone(tone_time int) error {
	s := newStereoSine(550, 580, float64(streamSampleRate), 0.5)
	defer s.Close()
	return pa.playTone(s, tone_time, 250, 250)
}

func (pa *PhoneAudio) playTone(s *stereoSine, play_time int, tone_time int, pause_time int) error {
	if !pa.active {
		return nil
	}
	// FIXME
	ctxBg := context.Background()
	ctx, cancel := context.WithTimeout(ctxBg, time.Duration(play_time)*time.Millisecond)
	defer cancel()

	i := 0
	for {
		if !pa.active {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		default:
			i++
			switch i {
			case 1:
				if err := s.Start(); err != nil {
					return err
				}
			case 2:
				time.Sleep(time.Duration(tone_time) * time.Millisecond)
			case 3:
				if err := s.Stop(); err != nil {
					return err
				}
			case 4:
				time.Sleep(time.Duration(pause_time) * time.Millisecond)
				i = 0
			}
		}
	}
}

type stereoSine struct {
	*portaudio.Stream
	stepL, phaseL float64
	stepR, phaseR float64
	vol           float64
}

func newStereoSine(freqL, freqR, sampleRate float64, volume float64) *stereoSine {
	s := &stereoSine{nil, freqL / sampleRate, 0, freqR / sampleRate, 0, volume}
	var err error
	s.Stream, err = portaudio.OpenDefaultStream(0, 2, sampleRate, 0, s.processAudio)
	if err != nil {
		panic(err)
	}
	return s
}

func (g *stereoSine) processAudio(out [][]float32) {
	for i := range out[0] {
		out[0][i] = float32(math.Sin(2*math.Pi*g.phaseL)) * float32(g.vol)
		_, g.phaseL = math.Modf(g.phaseL + g.stepL)
		out[1][i] = float32(math.Sin(2*math.Pi*g.phaseR)) * float32(g.vol)
		_, g.phaseR = math.Modf(g.phaseR + g.stepR)
	}
}
