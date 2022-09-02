package phoneaudio

import (
	"context"
	"encoding/binary"
	"os"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/sirupsen/logrus"
)

func (pa *PhoneAudio) Record(fileName string) error {
	if !pa.active {
		return nil
	}
	if !strings.HasSuffix(fileName, ".aiff") {
		fileName += ".aiff"
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	logrus.Info("prepareAudioFile")
	if err := pa.prepareAudioFile(f); err != nil {
		return err
	}

	nSamples := 0


	// FIXME (panic!!!)
	defer func() {
		logrus.Info("fillInMissingSizes")
		// fill in missing sizes
		totalBytes := 4 + 8 + 18 + 8 + 8 + 4*nSamples
		if _, err = f.Seek(4, 0); err != nil {
			panic(err)
		}
		if err := binary.Write(f, binary.BigEndian, int32(totalBytes)); err != nil {
			panic(err)
		}
		if _, err := f.Seek(22, 0); err != nil {
			panic(err)
		}
		if err := binary.Write(f, binary.BigEndian, int32(nSamples)); err != nil {
			panic(err)
		}
		if _, err = f.Seek(42, 0); err != nil {
			panic(err)
		}
		if err := binary.Write(f, binary.BigEndian, int32(4*nSamples+8)); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()


	logrus.Info("openDefaultStream")
	in := make([]int32, 64)
	stream, err := portaudio.OpenDefaultStream(pa.NumInputChannels, 0, float64(pa.SampleRate), len(in), in)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}

	// FIXME
	ctxBg := context.Background()
	ctx, cancel := context.WithTimeout(ctxBg, time.Duration(pa.MaxRecordTime)*time.Second)
	defer cancel()
	for {
		if !pa.active {
			return nil
		}
		if err := stream.Read(); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, in); err != nil {
			return err
		}
		nSamples += len(in)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	if err := stream.Stop(); err != nil {
		return err
	}

	return nil
}

func (pa *PhoneAudio) prepareAudioFile(f *os.File) error {
	// form chunk
	if _, err := f.WriteString("FORM"); err != nil {
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(0)); err != nil {
		return err
	}
	if _, err := f.WriteString("AIFF"); err != nil {
		return err
	}

	// common chunk
	if _, err := f.WriteString("COMM"); err != nil {
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(18)); err != nil { //size
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int16(pa.NumInputChannels)); err != nil { //channels
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(0)); err != nil { //number of samples
		return err
	}
	// FIXME - by parameter
	if err := binary.Write(f, binary.BigEndian, int16(32)); err != nil { //bits per sample
		return err
	}
	// FIXME - by parameter
	if _, err := f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}); err != nil { //80-bit sample rate 44100
		return err
	}

	// sound chunk
	if _, err := f.WriteString("SSND"); err != nil {
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(0)); err != nil { //size
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(0)); err != nil { //offset
		return err
	}
	if err := binary.Write(f, binary.BigEndian, int32(0)); err != nil { //block
		return err
	}

	return nil
}
