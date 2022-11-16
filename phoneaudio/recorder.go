package phoneaudio

/*
https://github.com/viert/lame
https://github.com/mewkiz/flac
*/

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		logrus.Debug("Error on NewWriter")
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
		if err := stream.Read(); err != nil {
			logrus.Debug("Error on stream read")
			return err
		}
		if _, err := waveWriter.WriteSample16([]int16(in)); err != nil {
			logrus.Debug("Error on writer")
			return err
		}

		select {
		case <-ctx.Done():
			logrus.Debug("Time is up")
			break EndRecording
		default:
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
	go func() {
		if err := pa.wav2mp3(wavFile, fileName); err != nil {
			logrus.Error(err)
		}
	}()
	return nil
}

func (pa *PhoneAudio) wav2mp3(wavFileName string, mp3FileName string) error {
	wavShortName := filepath.Base(wavFileName)
	mp3ShortName := filepath.Base(mp3FileName)
	logrus.Infof("Convert: %s->%s (assync)", wavShortName, mp3ShortName)

	wavFile, err := os.OpenFile(wavFileName, os.O_RDONLY, 0555)
	if err != nil {
		logrus.Errorf("Cannot open file: %s", wavShortName)
		return err
	}
	defer wavFile.Close()

	mp3File, err := os.OpenFile(mp3FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		logrus.Errorf("Cannot open file: %s", mp3ShortName)
		return err
	}
	defer mp3File.Close()

	wavHdr, err := lame.ReadWavHeader(wavFile)
	if err != nil {
		logrus.Errorf("Cannot read wav header: %s", wavShortName)
		return err
	}

	wr, err := lame.NewWriter(mp3File)
	wr.EncodeOptions = wavHdr.ToEncodeOptions()
	defer wr.Close()

	if _, err := io.Copy(wr, wavFile); err != nil {
		logrus.Errorf("Cannot io.Copy: %s->%s", wavShortName, mp3ShortName)
		return err
	}

	logrus.Infof("Converted: %s->%s (assync)", wavShortName, mp3ShortName)
	if err := os.Remove(wavFileName); err != nil {
		logrus.Error("Cannot remove wav file: 5s", wavFileName)
		return err
	}
	return nil
}
