package phoneaudio

/*
https://github.com/viert/lame
https://github.com/mewkiz/flac
*/

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/sirupsen/logrus"
	"github.com/viert/go-lame"
	wave "github.com/zenwerk/go-wave"
	//"github.com/sunicy/go-lame"
)

func (pa *PhoneAudio) Record(fileName string) error {
	if !pa.active {
		return nil
	}
	switch {
	// case strings.HasSuffix(fileName, ".flac"):
	// 	return pa.RecordFlac(fileName)
	case strings.HasSuffix(fileName, ".wav"):
		return pa.RecordWav(fileName)
	case strings.HasSuffix(fileName, ".aiff"):
		return pa.RecordAiff(fileName)
	// case strings.HasSuffix(fileName, ".mp3"):
	// 	return pa.RecordMp3(fileName)
	default:
		return fmt.Errorf("unknown format for file: %s", fileName)
	}
}

func (pa *PhoneAudio) RecordWav(fileName string) error {
	logrus.Infof("Start recording (%s)", fileName)

	of, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer of.Close()

	param := wave.WriterParam{
		Out:           of,
		Channel:       numInputChannels,
		SampleRate:    streamSampleRate,
		BitsPerSample: 16, // 8/16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	if err != nil {
		logrus.Debug("error on NewWriter")
		return err
	}
	defer waveWriter.Close()

	in := make([]int16, 1024)
	portaudio.Initialize()
	stream, err := portaudio.OpenDefaultStream(numInputChannels, 0, float64(streamSampleRate), len(in), in)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	ctxBg := context.Background()
	ctx, cancel := context.WithTimeout(ctxBg, time.Duration(maxRecordTime)*time.Second)
	defer cancel()
	for {
		if !pa.active {
			break
		}
		select {
		case <-ctx.Done():
			break
		default:
		}

		if err := stream.Read(); err != nil {
			logrus.Debug("error on stream read")
			return err
		}
		if _, err := waveWriter.WriteSample16([]int16(in)); err != nil {
			logrus.Debug("error on writer")
			return err
		}

	}
	logrus.Infof("End recording")
	return nil
}

func (pa *PhoneAudio) RecordFlac(fileName string) error {
	logrus.Infof("Recording FLAC: %s", fileName)
	return nil
}

// doesn't work :-(
func (pa *PhoneAudio) RecordMp3(fileName string) error {
	logrus.Infof("Recording MP3: %s", fileName)

	of, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer of.Close()

	enc := lame.NewEncoder(of)
	enc.SetInSamplerate(streamSampleRate)
	enc.SetHighPassFrequency(10000)
	enc.SetLowPassFrequency(100)
	enc.SetNumChannels(numOutputChannels)
	enc.SetQuality(5)
	defer enc.Close()

	in := make([]int32, 1024)
	stream, err := portaudio.OpenDefaultStream(numInputChannels, 0, float64(streamSampleRate), len(in), in)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	ctxBg := context.Background()
	ctx, cancel := context.WithTimeout(ctxBg, time.Duration(maxRecordTime)*time.Second)
	defer cancel()
	for {
		if !pa.active {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := stream.Read(); err != nil {
			return err
		}

		if err := binary.Write(enc, binary.BigEndian, in); err != nil {
			return err
		}
	}

	return nil
}

func (pa *PhoneAudio) RecordAiff(fileName string) error {
	logrus.Infof("Recording Aiff: %s", fileName)

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	logrus.Info("prepareAudioFile")
	if err := pa.prepareAudioFile(f); err != nil {
		return err
	}

	nSamples := 0

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
	stream, err := portaudio.OpenDefaultStream(numInputChannels, 0, float64(streamSampleRate), len(in), in)
	if err != nil {
		return err
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		return err
	}

	// FIXME
	ctxBg := context.Background()
	ctx, cancel := context.WithTimeout(ctxBg, time.Duration(maxRecordTime)*time.Second)
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
	if err := binary.Write(f, binary.BigEndian, int16(numInputChannels)); err != nil { //channels
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
