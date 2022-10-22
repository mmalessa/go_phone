package main

import (
	// "fmt"

	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmalessa/go_phone/orangepi"
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

	channelHook := make(chan bool)
	var hookState bool = false

	catchEscape(channel_stop)
	defer catchEscapeCleanUp()

	opi := orangepi.OrangePi{
		ChannelHook: channelHook,
	}
	if err := opi.Start(); err != nil {
		panic(err)
	}
	defer opi.Stop()

	for {
		select {
		case <-channel_stop:
			break
		case hookCurrentState := <-channelHook:
			if hookCurrentState != hookState {
				hookState = hookCurrentState
				fmt.Printf("Hook state: %t\n", hookState)
			}
		}
	}

	// pha := phoneaudio.PhoneAudio{}
	// go func() {
	// 	<-channel_stop
	// 	pha.Stop()
	// }()
	// pha.Initialize()
	// // loop
	// pha.Start()
	// defer pha.Terminate()
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
