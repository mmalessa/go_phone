package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mmalessa/go_phone/filemanager"
	"github.com/mmalessa/go_phone/orangepi"
	"github.com/mmalessa/go_phone/phoneaudio"
	"github.com/sirupsen/logrus"
)

var opi orangepi.OrangePi

var storageDir string = "/media/usb/" //var storageDir string = "/root/go_phone/" // tests only
var greetingsFileName string = "greetings.mp3"
var recordingsSubDir string = "recordings"
var recordingsFileExtension string = "wav"
var greetingsSubDir string = "greetings"
var maxRecordTime int = 120

func main() {
	configLogs()

	logrus.Info("GoPhone start")

	storageDir = strings.TrimRight(storageDir, "/")
	if err := checkStorageDirectory(); err != nil {
		logrus.Fatal(err)
	}

	channelStop := make(chan int)
	channelHook := make(chan bool)
	var hookState bool = false

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logrus.Debugf("SIGNAL RECEIVED: %s", sig)
		channelHook <- false
		channelStop <- 1
	}()

	opi = orangepi.OrangePi{
		ChannelHook: channelHook,
		ChannelStop: channelStop,
	}
	if err := opi.Start(); err != nil {
		panic(err)
	}
	defer opi.Stop()

	pha := phoneaudio.PhoneAudio{
		GreetingsFile: filepath.Join(storageDir, greetingsSubDir, greetingsFileName),
		FileManager: filemanager.FileManager{
			RecordingsDirectory: filepath.Join(storageDir, recordingsSubDir),
			RecordingsExtention: recordingsFileExtension,
		},
	}
	pha.SetMaxRecordTime(maxRecordTime)

	pha.Initialize()
	defer pha.Terminate()

	for {
		select {
		case hookCurrentState := <-channelHook:
			if hookCurrentState != hookState {
				hookState = hookCurrentState
				logrus.Debugf("Hook state: %t", hookState)
				if hookState {
					go func() {
						if err := pha.Start(); err != nil {
							logrus.Error(err)
						}
					}()
				} else {
					pha.Stop()
				}
			}
		case chval := <-channelStop:
			logrus.Debugf("PowerOff (%d)", chval)
			stopPhone()
			break
		}
	}
}

func stopPhone() {
	logrus.Info("GoPhone stop")
	opi.Stop()
	os.Exit(0)
}

func configLogs() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
}
