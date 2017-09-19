package rf522

import (
	"golang.org/x/exp/io/spi"
	"fmt"
	"github.com/davecheney/gpio"
)

const (
	ModeIdle     = 0x00
	ModeAuth     = 0x0E
	ModeReceive  = 0x08
	ModeTransmit = 0x04
	ModeTransrec = 0x0C
	ModeReset    = 0x0F
	ModeCrc      = 0x03

	AuthA = 0x60
	AuthB = 0x61

	ActRead      = 0x30
	ActWrite     = 0xA0
	ActIncrement = 0xC1
	ActDecrement = 0xC0
	ActRestore   = 0xC2
	ActTransfer  = 0xB0

	ActReqIdl = 0x26
	ActReqAll = 0x52
	ActRntIcl = 0x93
	ActSelect = 0x93
	ActEnd    = 0x50

	RegTxControl = 0x14
	Length       = 16

	AntennaGain = 0x04
)

type RFID struct {
	ResetPin      gpio.Pin
	IrqPin        gpio.Pin
	CePin         gpio.Pin
	Authenticated bool
	antennaGain   int
	MaxSpeedHz    int
	spiDev        *spi.Device
	irqChannel    chan bool
}

func MakeRFID(busId, deviceId, maxSpeed, resetPin, irqPin, cePin int) (device *RFID, err error) {
	spiDev, err := spi.Open(&spi.Devfs{
		Dev:      fmt.Sprintf("/dev/spidev%1d.%2d", busId, deviceId),
		Mode:     spi.Mode3,
		MaxSpeed: int64(maxSpeed),
	})
	if err != nil {
		return
	}

	dev := &RFID{
		MaxSpeedHz: maxSpeed,
		spiDev:     spiDev,
	}

	pin, err := gpio.OpenPin(resetPin, gpio.ModeOutput)
	if err != nil {
		return
	}
	pin.Set()
	dev.ResetPin = pin

	pin, err = gpio.OpenPin(irqPin, gpio.ModeInput)
	if err != nil {
		return
	}
	dev.IrqPin = pin

	if cePin != 0 {
		pin, err = gpio.OpenPin(cePin, gpio.ModeOutput)
		if err != nil {
			return
		}
		pin.Set()
		dev.CePin = pin
	}

	pin.BeginWatch(gpio.EdgeFalling, func() {})

	dev.irqChannel = make(chan bool)

	err = dev.Reset()
	if err != nil {
		return
	}
	dev.devWrite(0x2A, 0x8D)
	dev.devWrite(0x2B, 0x3E)
	dev.devWrite(0x2D, 30)
	dev.devWrite(0x2C, 0)
	dev.devWrite(0x15, 0x40)
	dev.devWrite(0x11, 0x3D)
	dev.devWrite(0x26, 5<<4)
	dev.SetAntenna(true)

	device = dev

	return
}

func (r *RFID) writeSpiData(dataIn, dataOut []byte) (err error) {
	if r.CePin != nil {
		r.CePin.Clear()
	}
	err = r.spiDev.Tx(dataIn, dataOut)
	if r.CePin != nil {
		r.CePin.Set()
	}
	return
}

/*
    def dev_write(self, address, value):
        self.spi_transfer([(address << 1) & 0x7E, value])
 */

func (r *RFID) devWrite(address int, data byte) (err error) {
	newData := [2]byte{(byte(address) << 1) & 0x7E, data}
	err = r.writeSpiData(newData[0:], []byte{})
	return
}

/*
    def dev_read(self, address):
        return self.spi_transfer([((address << 1) & 0x7E) | 0x80, 0])[1]
 */
func (r *RFID) devRead(address int) (result byte, err error) {
	data := [2]byte{(byte(address)<<1)&0x7E | 0x80, 0}
	rb := make([]byte, 2)
	err = r.writeSpiData(data[0:], rb)
	result = rb[1]
	return
}

/*
    def set_bitmask(self, address, mask):
        current = self.dev_read(address)
        self.dev_write(address, current | mask)
 */
func (r *RFID) setBitmask(address, mask int) (err error) {
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	err = r.devWrite(address, current|byte(mask))
	return
}

/*
    def clear_bitmask(self, address, mask):
        current = self.dev_read(address)
        self.dev_write(address, current & (~mask))
 */
func (r *RFID) clearBitmask(address, mask int) (err error) {
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	err = r.devWrite(address, current&^byte(mask))
	return

}

/*
    def set_antenna_gain(self, gain):
        """
        Sets antenna gain from a value from 0 to 7.
        """
        if 0 <= gain <= 7:
            self.antenna_gain = gain
 */
 func (r *RFID) SetAntennaGain(gain int) {
 	if 0 <= gain && gain <= 7 {
 		r.antennaGain = gain
	}
 }

func (r *RFID) Reset() (err error) {
	r.Authenticated = false
	err = r.devWrite(0x01, ModeReset)
	return
}

/*
    def set_antenna(self, state):
        if state == True:
            current = self.dev_read(self.reg_tx_control)
            if ~(current & 0x03):
                self.set_bitmask(self.reg_tx_control, 0x03)
        else:
            self.clear_bitmask(self.reg_tx_control, 0x03)

 */
func (r *RFID) SetAntenna(state bool) (err error) {
	if state {
		current, err := r.devRead(RegTxControl)
		if err != nil {
			return
		}
		if current&0x03 == 0 {
			err = r.setBitmask(RegTxControl, 0x03)
		}
	} else {
		err = r.clearBitmask(RegTxControl, 0x03)
	}
	return
}
