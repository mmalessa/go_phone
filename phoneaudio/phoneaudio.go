package phoneaudio

import (
	"github.com/gordonklaus/portaudio"
)

var (
	streamSampleRate  int = 44100 // don't change yet
	numInputChannels  int = 1
	numOutputChannels int = 1
	maxRecordTime     int = 10 // seconds
	// greetings_file    string = "greetings/greetings.aiff"
	// //FIXME
	// recordings_directory string = "recordings"
	// recording_file       string = "recordings/0000.aiff"
)

type PhoneAudio struct {
	AnnouncementFile    string
	RecordingsDirectory string
	active              bool
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
	recordingFile, err := pa.findRecordingFileName()
	if err != nil {
		pa.BusyTone(6000)
		return err
	}

	pa.RingingTone(4000)
	if err := pa.Play(pa.AnnouncementFile); err != nil {
		pa.BusyTone(6000)
		return err
	}
	pa.Beep(700)
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

func (pa *PhoneAudio) findRecordingFileName() (string, error) {
	return "recordings/0000.aif", nil
}
