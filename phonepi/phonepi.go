package phonepi

import (
	"github.com/stianeikeland/go-rpio/v4"
	"time"
)

type PhonePi struct {
}

var (
	ledRed = rpio.Pin(6)	
	ledGreen = rpio.Pin(5)
	hookSwitch = rpio.Pin(21)
	hookOnState = rpio.High
	powerSwitch = rpio.Pin(17)

	hookCurrentState = hookOnState
)

func (pp *PhonePi) Start() error {
	if err := rpio.Open(); err != nil {
		return err
	}
	pp.initialize()
	pp.loop()
	return nil
}

func (pp *PhonePi) Stop() {
	ledRed.Low()
	ledGreen.Low()
	hookSwitch.Detect(rpio.NoEdge)
	powerSwitch.Detect(rpio.NoEdge)
	rpio.Close()
}

func (pp *PhonePi) initialize() {
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

func (pp *PhonePi) loop() {
	go func() {
		for {
			// ledRed.Toggle()
			if hookSwitch.EdgeDetected() {
				pp.onHookEdgeDetected()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func (pp *PhonePi) onHookEdgeDetected() {
	time.Sleep(10 * time.Millisecond)

	hookState := hookSwitch.Read()
	if hookState == hookCurrentState {
		return
	}
	hookCurrentState = hookState
	if hookCurrentState == hookOnState {
		pp.onHandsetOn()
	} else {
		pp.onHandsetOff()
	}
}

func (pp *PhonePi) onHandsetOff() {
	ledGreen.High()
}

func (pp *PhonePi) onHandsetOn() {
	ledGreen.Low()
}