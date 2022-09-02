package phoneaudio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
)

var (
	streamSampleRate  int    = 44100 // don't change yet
	numInputChannels  int    = 1
	numOutputChannels int    = 1
	maxRecordTime     int    = 10 // seconds
	greetings_file    string = "greetings/greetings.aiff"
	//FIXME
	recordings_directory string = "recordings"
	recording_file       string = "recordings/0000.aiff"
)

type PhoneAudio struct {
	active bool
}

func (pa *PhoneAudio) Initialize() error {
	return portaudio.Initialize()
}

func (pa *PhoneAudio) Terminate() error {
	return portaudio.Terminate()
}

func (pa *PhoneAudio) SetMaxRecordTime(mrt int) {
	maxRecordTime = mrt
}

func (pa *PhoneAudio) Start() {
	pa.active = true
	pa.RingingTone(4000)
	if err := pa.Play(greetings_file); err != nil {
		panic(err)
	}
	pa.Beep(700)
	err := pa.Record(recording_file)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	pa.BusyTone(3000)
	pa.active = false
}

func (pa *PhoneAudio) Stop() {
	pa.active = false
}
