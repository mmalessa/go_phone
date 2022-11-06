package phoneaudio

import (
	"github.com/gordonklaus/portaudio"
	"github.com/mmalessa/go_phone/filemanager"
)

var (
	streamSampleRate  int = 44100 // don't change yet
	numInputChannels  int = 1
	numOutputChannels int = 1
	maxRecordTime     int = 180 // seconds
)

type PhoneAudio struct {
	GreetingsFile string
	FileManager   filemanager.FileManager
	active        bool
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

func (pa *PhoneAudio) Start() error {
	pa.active = true
	recordingFile, err := pa.FileManager.FindRecordingFileName()
	if err != nil {
		pa.BusyTone(6000)
		return err
	}

	pa.RingingTone(500)
	// if err := pa.Play(pa.GreetingsFile); err != nil {
	// 	pa.ErrorTone(2000)
	// 	return err
	// }

	pa.Beep(800)
	if err := pa.Record(recordingFile); err != nil {
		pa.BusyTone(6000)
		return err
	}

	pa.BusyTone(3000)
	pa.active = false
	return nil
}

func (pa *PhoneAudio) Stop() {
	pa.active = false
}
