package orangepi

import (
	"time"

	"github.com/mmalessa/orio"
)

type OrangePi struct {
	ChannelHook chan bool
	ChannelStop chan int
}

var (
	ledRed           = orio.Pin(orio.PH7)
	ledGreen         = orio.Pin(orio.PH8)
	hookSwitch       = orio.Pin(orio.PH6)
	hookOnState      = orio.High
	hookCurrentState = hookOnState
	powerOffSwitch   = orio.Pin(orio.PC8)
	powerOffActive   = orio.Low
)

func (op *OrangePi) Start() error {
	orio.DebugMode = false
	op.initialize()
	ledRed.Low()
	ledGreen.High()
	op.loop()
	return nil
}

func (op *OrangePi) Stop() {
	ledRed.Low()
	ledGreen.Low()
	orio.Close()
}

func (op *OrangePi) initialize() {
	ledRed.Output()
	ledGreen.Output()
	hookSwitch.Input()
	powerOffSwitch.Input()
}

func (op *OrangePi) loop() {
	go func() {
		for {
			if hookSwitch.State() != hookCurrentState {
				time.Sleep(3 * time.Millisecond)
				hookState := hookSwitch.State()
				if hookState != hookCurrentState {
					hookCurrentState = hookState
					op.onHookEdgeDetected()
				}
			}

			if powerOffSwitch.State() == powerOffActive {
				time.Sleep(2 * time.Second)
				if powerOffSwitch.State() == powerOffActive {
					op.onPowerOff()
				}
			}
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
	ledRed.High()
	op.ChannelHook <- true
}

func (op *OrangePi) onHandsetOn() {
	ledRed.Low()
	op.ChannelHook <- false
}

func (op *OrangePi) onPowerOff() {
	op.ChannelStop <- 2
}
