package orangepi

import (
	"time"

	"github.com/mmalessa/orio"
)

type OrangePi struct {
	ChannelHook chan bool
}

var (
	ledRed           = orio.Pin(orio.PH7)
	ledGreen         = orio.Pin(orio.PH8)
	hookSwitch       = orio.Pin(orio.PH6)
	hookOnState      = orio.High
	hookCurrentState = hookOnState
)

func (op *OrangePi) Start() error {
	orio.DebugMode = false
	op.initialize()
	op.loop()
	return nil
}

func (op *OrangePi) Stop() {
	ledGreen.Low()
	ledGreen.Low()
	orio.Close()
}

func (op *OrangePi) initialize() {
	ledRed.Output()
	ledGreen.Output()
	hookSwitch.Input()
}

func (op *OrangePi) loop() {
	go func() {
		for {
			hookState := hookSwitch.State()
			if hookState == hookCurrentState {
				continue
			}
			time.Sleep(3 * time.Millisecond)
			if hookState == hookCurrentState {
				continue
			}
			hookCurrentState = hookState
			op.onHookEdgeDetected()
		}
	}()
}

func (op *OrangePi) onHookEdgeDetected() {
	if hookCurrentState == hookOnState {
		op.onHandsetOn()
	} else {
		op.onHandsetOff()
	}
}

func (op *OrangePi) onHandsetOff() {
	ledGreen.High()
	op.ChannelHook <- true
}

func (op *OrangePi) onHandsetOn() {
	ledGreen.Low()
	op.ChannelHook <- false
}
