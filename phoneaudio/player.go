package phoneaudio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
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
	case strings.HasSuffix(fileName, ".aiff"):
		return pa.playAiff(fileName)
	default:
		return fmt.Errorf("unknown format for file: %s", fileName)
	}

	return nil
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
	decoder.Format(rate, channels, mpg123.ENC_SIGNED_16)

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

func (pa *PhoneAudio) playAiff(fileName string) error {
	if !pa.active {
		return nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	id, data, err := readChunk(f)
	if err != nil {
		return err
	}
	if id.String() != "FORM" {
		return errors.New(fmt.Sprintf("bad file format (%s) should be FORM", id.String()))
	}
	_, err = data.Read(id[:])
	if err != nil {
		return err
	}
	if id.String() != "AIFF" {
		return errors.New(fmt.Sprintf("bad file format (%s) should be AIFF", id.String()))
	}

	var c commonChunk
	var audio io.Reader
	for {
		id, chunk, err := readChunk(data)
		if err == io.EOF {
			break
		}
		switch id.String() {
		case "COMM":
			if err := binary.Read(chunk, binary.BigEndian, &c); err != nil {
				return err
			}
		case "SSND":
			chunk.Seek(8, 1) //ignore offset and block
			audio = chunk
		case "NAME":
			break
		default:
			// FIXME
			fmt.Printf("ignoring unknown chunk '%s'\n", id)
		}
	}

	// out := make([]int32, 8192)
	out := make([]int32, 1024)
	stream, err := portaudio.OpenDefaultStream(0, numOutputChannels, float64(streamSampleRate), len(out), &out)

	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	for remaining := int(c.NumSamples); remaining > 0; remaining -= len(out) {
		if !pa.active {
			return nil
		}

		if len(out) > remaining {
			out = out[:remaining]
		}
		err := binary.Read(audio, binary.BigEndian, out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := stream.Write(); err != nil {
			return err
		}
	}

	return nil
}

func readChunk(r readerAtSeeker) (id ID, data *io.SectionReader, err error) {
	_, err = r.Read(id[:])
	if err != nil {
		return
	}
	var n int32
	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}
	off, _ := r.Seek(0, 1)
	data = io.NewSectionReader(r, off, int64(n))
	_, err = r.Seek(int64(n), 1)
	return
}
