package raspberrypi

import (
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type RaspberryPi struct {
}

var (
	ledRed      = rpio.Pin(6)
	ledGreen    = rpio.Pin(5)
	hookSwitch  = rpio.Pin(21)
	hookOnState = rpio.High
	powerSwitch = rpio.Pin(17)

	hookCurrentState = hookOnState
)

func (rp *RaspberryPi) Start() error {
	if err := rpio.Open(); err != nil {
		return err
	}
	rp.initialize()
	rp.loop()
	return nil
}

func (rp *RaspberryPi) Stop() {
	ledRed.Low()
	ledGreen.Low()
	hookSwitch.Detect(rpio.NoEdge)
	powerSwitch.Detect(rpio.NoEdge)
	rpio.Close()
}

func (rp *RaspberryPi) initialize() {
	ledRed.Output()
	ledGreen.Output()
	hookSwitch.Input()
	powerSwitch.Input()

	if hookOnState == rpio.High {
		hookSwitch.PullUp()
	} else {
		hookSwitch.PullDown()
	}
	hookSwitch.Detect(rpio.AnyEdge)
}

func (rp *RaspberryPi) loop() {
	go func() {
		for {
			// ledRed.Toggle()
			if hookSwitch.EdgeDetected() {
				rp.onHookEdgeDetected()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func (rp *RaspberryPi) onHookEdgeDetected() {
	time.Sleep(10 * time.Millisecond)

	hookState := hookSwitch.Read()
	if hookState == hookCurrentState {
		return
	}
	hookCurrentState = hookState
	if hookCurrentState == hookOnState {
		rp.onHandsetOn()
	} else {
		rp.onHandsetOff()
	}
}

func (rp *RaspberryPi) onHandsetOff() {
	ledGreen.High()
}

func (rp *RaspberryPi) onHandsetOn() {
	ledGreen.Low()
}
