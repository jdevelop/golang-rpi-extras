package sensor_hcsr04

import (
	"github.com/stianeikeland/go-rpio"
	"time"
)

const HardStop = 1000000

type HCSR04 struct {
	EchoPin rpio.Pin

	PingPin rpio.Pin
}

func NewHCSR04(echo int, // Echo pin
	ping int, // Trigger pin
) (result HCSR04) {
	if err := rpio.Open(); err != nil {
		panic(err.Error())
	}

	result.EchoPin = rpio.Pin(echo)
	result.PingPin = rpio.Pin(ping)

	return
}

func (hcsr *HCSR04) MeasureDistance() float32 {

	hcsr.EchoPin.Output()
	hcsr.PingPin.Output()

	hcsr.EchoPin.Low()
	hcsr.PingPin.Low()

	hcsr.EchoPin.Input()

	strobeZero := 0
	strobeOne := 0

	// strobe
	delayUs(200)
	hcsr.PingPin.High()
	delayUs(15)
	hcsr.PingPin.Low()

	// wait until strobe back

	for strobeZero = 0; strobeZero < HardStop && hcsr.EchoPin.Read() != rpio.High; strobeZero++ {
	}
	start := time.Now()
	for strobeOne = 0; strobeOne < HardStop && hcsr.EchoPin.Read() != rpio.Low; strobeOne++ {
		delayUs(1)
	}
	end := time.Now()

	return float32(end.UnixNano()-start.UnixNano()) / (58.0 * 1000)
}

func delayUs(ms int) {
	time.Sleep(time.Duration(ms) * time.Microsecond)
}
