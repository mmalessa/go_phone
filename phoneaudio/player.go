package phoneaudio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/bobertlo/go-mpg123/mpg123"
	"github.com/gordonklaus/portaudio"
	"github.com/sirupsen/logrus"
)

type readerAtSeeker interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type ID [4]byte

func (id ID) String() string {
	return string(id[:])
}

type commonChunk struct {
	NumChans      int16
	NumSamples    int32
	BitsPerSample int16
	SampleRate    [10]byte
}

func (pa *PhoneAudio) Play(fileName string) error {
	if !pa.active {
		return nil
	}
	switch {
	case strings.HasSuffix(fileName, ".mp3"):
		return pa.playMp3(fileName)
	default:
		return fmt.Errorf("unknown format for file: %s", fileName)
	}
}

func (pa *PhoneAudio) playMp3(fileName string) error {
	decoder, err := mpg123.NewDecoder("")
	if err != nil {
		return err
	}
	if err := decoder.Open(fileName); err != nil {
		return err
	}
	defer decoder.Close()

	// get audio format information
	rate, channels, _ := decoder.GetFormat()
	logrus.Infof("Play MP3 ENC_SIGNED_16, Sample Rate: %d, Channels: %d", rate, channels)
	decoder.FormatNone()
	decoder.Format(rate, channels, int(mpg123.ENC_SIGNED_16))

	portaudio.Initialize()
	defer portaudio.Terminate()
	out := make([]int16, 8192)
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(rate), len(out), &out)
	if err != nil {
		return err
	}
	defer stream.Close()
	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	for {
		if !pa.active {
			break
		}
		audio := make([]byte, 2*len(out))
		_, err = decoder.Read(audio)
		if err == mpg123.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := binary.Read(bytes.NewBuffer(audio), binary.LittleEndian, out); err != nil {
			return err
		}
		if err := stream.Write(); err != nil {
			return err
		}
	}

	return nil
}
