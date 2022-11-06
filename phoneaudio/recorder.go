package phoneaudio

/*
https://github.com/viert/lame
https://github.com/mewkiz/flac
*/

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/sirupsen/logrus"
	"github.com/sunicy/go-lame"
	wave "github.com/zenwerk/go-wave"
	//"github.com/sunicy/go-lame"
)

func (pa *PhoneAudio) Record(fileName string) error {
	if !pa.active {
		return nil
	}
	switch {
	case strings.HasSuffix(fileName, ".wav"):
		return pa.recordWav(fileName)
	case strings.HasSuffix(fileName, ".aiff"):
		return pa.recordAiff(fileName)
	case strings.HasSuffix(fileName, ".mp3"):
		return pa.recordMp3(fileName)
	default:
		return fmt.Errorf("unknown format for file: %s", fileName)
	}
}

func (pa *PhoneAudio) recordWav(fileName string) error {
	logrus.Infof("Start recording (%s), max time: %ds", fileName, maxRecordTime)

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
EndRecording:
	for {
		if !pa.active {
			break EndRecording
		}
		select {
		case <-ctx.Done():
			break EndRecording
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

func (pa *PhoneAudio) recordMp3(fileName string) error {
	logrus.Info("Record MP3")
	re := regexp.MustCompile(`.mp3$`)
	wavFile := re.ReplaceAllString(fileName, ".wav")
	logrus.Infof("Wav file: %s", wavFile)
	if err := pa.recordWav(wavFile); err != nil {
		return err
	}
	go pa.wav2mp3(wavFile, fileName)
	return nil
}

func (pa *PhoneAudio) wav2mp3(wavFileName string, mp3FileName string) error {
	logrus.Infof("Convert wav2mp3 %s->%s", wavFileName, mp3FileName)

	wavShortName := pa.getShortFileName(wavFileName)
	mp3ShortName := pa.getShortFileName(mp3FileName)

	wavFile, err := os.OpenFile(wavFileName, os.O_RDONLY, 0555)
	if err != nil {
		logrus.Errorf("Cannot open file: %s", wavShortName)
		logrus.Error(err)
	}
	defer wavFile.Close()

	mp3File, err := os.OpenFile(mp3FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		logrus.Errorf("Cannot open file: %s", mp3ShortName)
		logrus.Error(err)
	}
	defer mp3File.Close()

	wavHdr, err := lame.ReadWavHeader(wavFile)
	if err != nil {
		logrus.Errorf("Cannot read wav header: %s", wavShortName)
		logrus.Error(err)
	}

	wr, err := lame.NewWriter(mp3File)
	wr.EncodeOptions = wavHdr.ToEncodeOptions()
	defer wr.Close()

	if _, err := io.Copy(wr, wavFile); err != nil {
		logrus.Errorf("Cannot io.Copy: %s->%s", wavShortName, mp3ShortName)
		logrus.Error(err)
	}

	logrus.Infof("Converted: %s->%s", wavShortName, mp3ShortName)
	if err := os.Remove(wavFileName); err != nil {
		logrus.Error("Cannot remove wav file: 5s", wavShortName)
		logrus.Error(err)
	}
	return nil
}

func (pa *PhoneAudio) getShortFileName(fileName string) string {
	re := regexp.MustCompile(`/[^/]$`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) != 2 {
		return fileName
	}
	return matches[1]
}

func (pa *PhoneAudio) recordAiff(fileName string) error {
	logrus.Infof("Recording Aiff: %s", fileName)

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	logrus.Info("prepareAudioFile")
	if err := pa.prepareAiffAudioFile(f); err != nil {
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

func (pa *PhoneAudio) prepareAiffAudioFile(f *os.File) error {
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
