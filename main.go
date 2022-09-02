package main

import (
	// "fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmalessa/go_phone/phoneaudio"
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
	defer catchEscapeCleanUp()

	// phpi := phonepi.PhonePi {}
	// if err := phpi.Start(); err != nil {
	// 	panic(err)
	// }
	// defer phpi.Stop()

	pha := phoneaudio.PhoneAudio{}
	go func() {
		<-channel_stop
		pha.Stop()
	}()

	pha.Initialize()

	// loop
	pha.Start()

	defer pha.Terminate()
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
