package phoneaudio

import "github.com/gordonklaus/portaudio"

type PhoneAudio struct {
	active            bool
	SampleRate        int
	NumInputChannels  int
	NumOutputChannels int
	MaxRecordTime     int
}

func (pa *PhoneAudio) Initialize() error {
	return portaudio.Initialize()
}

func (pa *PhoneAudio) Terminate() error {
	return portaudio.Terminate()
}

func (pa *PhoneAudio) Start() {
	pa.active = true
}

func (pa *PhoneAudio) Stop() {
	pa.active = false
}
