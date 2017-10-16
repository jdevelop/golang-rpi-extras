package rf522

import (
	"fmt"
	"github.com/jdevelop/golang-rpi-extras/rf522/commands"
	"time"
	"github.com/ecc1/spi"
	"github.com/davecheney/gpio"
	rpio "github.com/davecheney/gpio/rpi"
	"github.com/sirupsen/logrus"
)

type RFID struct {
	ResetPin      gpio.Pin
	IrqPin        gpio.Pin
	Authenticated bool
	antennaGain   int
	MaxSpeedHz    int
	spiDev        *spi.Device
	irqChannel    chan bool
}

func MakeRFID(busId, deviceId, maxSpeed, resetPin, irqPin int) (device *RFID, err error) {

	spiDev, err := spi.Open(fmt.Sprintf("/dev/spidev%d.%d", busId, deviceId), maxSpeed, 0)

	spiDev.SetLSBFirst(false)
	spiDev.SetBitsPerWord(8)

	if err != nil {
		return
	}

	dev := &RFID{
		spiDev:     spiDev,
		MaxSpeedHz: maxSpeed,
	}

	pin, err := rpio.OpenPin(resetPin, gpio.ModeOutput)
	if err != nil {
		return
	}
	dev.ResetPin = pin
	dev.ResetPin.Set()

	pin, err = rpio.OpenPin(irqPin, gpio.ModeInput)
	if err != nil {
		return
	}
	dev.IrqPin = pin
	dev.IrqPin.PullUp()

	dev.irqChannel = make(chan bool)

	dev.IrqPin.BeginWatch(gpio.EdgeFalling, func() {
		logrus.Debug("Interrupt")
		dev.irqChannel <- true
	})

	err = dev.init()

	device = dev

	return
}

func (r *RFID) init() (err error) {
	err = r.Reset()
	if err != nil {
		return
	}
	err = r.devWrite(0x2A, 0x8D)
	if err != nil {
		return
	}
	err = r.devWrite(0x2B, 0x3E)
	if err != nil {
		return
	}
	err = r.devWrite(0x2D, 30)
	if err != nil {
		return
	}
	err = r.devWrite(0x2C, 0)
	if err != nil {
		return
	}
	err = r.devWrite(0x15, 0x40)
	if err != nil {
		return
	}
	err = r.devWrite(0x11, 0x3D)
	if err != nil {
		return
	}
	err = r.SetAntenna(true)
	if err != nil {
		return
	}
	return
}

func (r *RFID) writeSpiData(dataIn []byte) (out []byte, err error) {
	out = make([]byte, len(dataIn))
	copy(out, dataIn)
	err = r.spiDev.Transfer(out)
	return
}

func printBytes(data []byte) (res string) {
	res = "["
	for _, v := range data[0:len(data)-1] {
		res = res + fmt.Sprintf("%d, ", byte(v))
	}
	res = res + fmt.Sprintf("%d", data[len(data)-1])
	res = res + "]"
	return
}

/*
    def dev_write(self, address, value):
        self.spi_transfer([(address << 1) & 0x7E, value])
 */

func (r *RFID) devWrite(address int, data byte) (err error) {
	newData := [2]byte{(byte(address) << 1) & 0x7E, data}
	readBuf, err := r.writeSpiData(newData[0:])
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debug(">>" + printBytes(newData[:]) + " " + printBytes(readBuf))
	}
	return
}

/*
    def dev_read(self, address):
        return self.spi_transfer([((address << 1) & 0x7E) | 0x80, 0])[1]
 */
func (r *RFID) devRead(address int) (result byte, err error) {
	data := [2]byte{((byte(address) << 1) & 0x7E) | 0x80, 0}
	rb, err := r.writeSpiData(data[0:])
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debug("<<" + printBytes(data[:]) + " " + printBytes(rb))
	}
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
	err = r.devWrite(0x01, commands.ModeReset)
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
		current, err := r.devRead(commands.RegTxControl)
		fmt.Println("Antenna", current)
		if err != nil {
			return err
		}
		if current&0x03 == 0 {
			err = r.setBitmask(commands.RegTxControl, 0x03)
		}
	} else {
		err = r.clearBitmask(commands.RegTxControl, 0x03)
	}
	return
}

func (r *RFID) Wait() (err error) {
	err = r.init()
	if err != nil {
		return
	}
	err = r.devWrite(0x04, 0x00)
	if err != nil {
		return
	}
	err = r.devWrite(0x02, 0xA0)
	if err != nil {
		return
	}
	waiting := true
	for waiting {
		err = r.devWrite(0x09, 0x26)
		if err != nil {
			return
		}
		err = r.devWrite(0x01, 0x0C)
		if err != nil {
			return
		}
		err = r.devWrite(0x0D, 0x87)
		if err != nil {
			return
		}
		select {
		case _ = <-r.irqChannel:
			waiting = false
		case <-time.After(100 * time.Millisecond):
			// do nothing
		}
	}
	return
}
