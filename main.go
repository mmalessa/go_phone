package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mmalessa/go_phone/orangepi"
	"github.com/sirupsen/logrus"
)

var opi orangepi.OrangePi

func main() {
	configLogs()

	logrus.Info("GoPhone start")

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
		// stopPhone()
	}()

	termiosStart(channelStop)

	opi = orangepi.OrangePi{
		ChannelHook: channelHook,
		ChannelStop: channelStop,
	}
	if err := opi.Start(); err != nil {
		panic(err)
	}
	defer opi.Stop()

	for {
		select {
		case hookCurrentState := <-channelHook:
			if hookCurrentState != hookState {
				hookState = hookCurrentState
				logrus.Debugf("Hook state: %t", hookState)
			}
		case chval := <-channelStop:
			logrus.Debugf("PowerOff (%d)", chval)
			stopPhone()
			break
		}
	}

	// pha := phoneaudio.PhoneAudio{}
	// go func() {
	// 	<-channelHook
	// 	pha.Stop()
	// }()
	// pha.Initialize()
	// // loop
	// pha.Start()
	// defer pha.Terminate()
}

func stopPhone() {
	termiosCleanUp()
	logrus.Info("GoPhone stop")
	opi.Stop()
	os.Exit(0)
}

func configLogs() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{})
}
