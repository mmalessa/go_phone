package phoneaudio

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gordonklaus/portaudio"
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
		default:
			// FIXME
			fmt.Printf("ignoring unknown chunk '%s'\n", id)
		}
	}

	out := make([]int32, 8192)
	stream, err := portaudio.OpenDefaultStream(0, pa.NumOutputChannels, float64(pa.SampleRate), len(out), &out)
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
