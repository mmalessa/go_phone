package main

import (
	// "fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmalessa/go_phone/phoneaudio"
	"github.com/mmalessa/go_phone/phonepi"
	"github.com/sirupsen/logrus"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	channel_stop := make(chan int)
	go func() {
		sig := <-sigs
		logrus.Infof("SIGNAL RECEIVED: %s", sig)
		channel_stop <- 1
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}()

	catchEscape(channel_stop)

	phpi := phonepi.PhonePi {}
	if err := phpi.Start(); err != nil {
		panic(err)
	}
	defer phpi.Stop()

	pha := phoneaudio.PhoneAudio{
		SampleRate:        44100, // don't change yet
		NumInputChannels:  1,
		NumOutputChannels: 1,
		MaxRecordTime:     10, // sec
	}
	go func() {
		<-channel_stop
		pha.Stop()
	}()


	pha.Initialize()
	defer pha.Terminate()
	pha.Start()
	pha.RingingTone(4000)
	// greetings_file := "greetings/greetings.aiff"
	// if err := pha.Play(greetings_file); err != nil {
	// 	panic(err)
	// }
	// pha.Beep(700)
	// recording_file := "recordings/0000.aiff"
	// err := pha.Record(recording_file)
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err)
	// }
	pha.BusyTone(3000)

}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
