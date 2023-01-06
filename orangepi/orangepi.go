package orangepi

import (
	"time"

	"github.com/mmalessa/orio"
)

type OrangePi struct {
	ChannelHookState chan bool
}

var (
	stateActive      = false
	ledGreen         = orio.Pin(orio.PA10)
	hookSwitch       = orio.Pin(orio.PA2)
	hookOnState      = orio.Low
	hookCurrentState = hookOnState
)

func (op *OrangePi) Start() error {
	orio.DebugMode = false
	op.initialize()
	stateActive = true
	ledGreen.High()
	op.loop()
	return nil
}

func (op *OrangePi) Stop() {
	ledGreen.Low()
	stateActive = false
	orio.Close()
}

func (op *OrangePi) initialize() {
	ledGreen.Output()
	hookSwitch.Input()
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
		}
	}()
}

func (op *OrangePi) onHookEdgeDetected() {
	if hookCurrentState == hookOnState {
		op.ChannelHookState <- false
	} else {
		op.ChannelHookState <- true
	}
}
