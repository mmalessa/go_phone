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
	stateActive      = false
	ledRed           = orio.Pin(orio.PA15) // PH7 PA15
	ledGreen         = orio.Pin(orio.PA16) // PH8 PA16
	hookSwitch       = orio.Pin(orio.PA14) // PH6 PA14
	hookOnState      = orio.High
	hookCurrentState = hookOnState
	powerOffSwitch   = orio.Pin(orio.PA3) // PC8 PA3
	powerOffActive   = orio.Low
)

func (op *OrangePi) Start() error {
	orio.DebugMode = false
	op.initialize()
	stateActive = true
	ledRed.Low()
	ledGreen.High()
	op.loop()
	return nil
}

func (op *OrangePi) Stop() {
	ledRed.Low()
	ledGreen.Low()
	stateActive = false
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
			if stateActive && hookSwitch.State() != hookCurrentState {
				time.Sleep(3 * time.Millisecond)
				hookState := hookSwitch.State()
				if hookState != hookCurrentState {
					hookCurrentState = hookState
					op.onHookEdgeDetected()
				}
			}

			if stateActive && powerOffSwitch.State() == powerOffActive {
				time.Sleep(500 * time.Millisecond)
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
