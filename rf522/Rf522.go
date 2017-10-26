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

	err = dev.Init()

	device = dev

	return
}

func (r *RFID) Init() (err error) {
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
	logrus.Debug("Init done")
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
		res = res + fmt.Sprintf("%02x, ", byte(v))
	}
	res = res + fmt.Sprintf("%02x", data[len(data)-1])
	res = res + "]"
	return
}

/*
    def dev_write(self, address, value):
        self.spi_transfer([(address << 1) & 0x7E, value])
 */

func (r *RFID) devWrite(address int, data byte) (err error) {
	newData := [2]byte{(byte(address) << 1) & 0x7E, data}
	readBuf, err := r.writeSpiData(newData[:])
	if logrus.GetLevel() == logrus.DebugLevel {
		newData[0] = newData[0] >> 1
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
	rb, err := r.writeSpiData(data[:])
	if logrus.GetLevel() == logrus.DebugLevel {
		data[0] = (data[0] >> 1) & 0x7f
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
	logrus.Debug("Set mask ", address, mask)
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	logrus.Debug("Set mask value ", address, current|byte(mask))
	err = r.devWrite(address, current|byte(mask))
	return
}

/*
    def clear_bitmask(self, address, mask):
        current = self.dev_read(address)
        self.dev_write(address, current & (~mask))
 */
func (r *RFID) clearBitmask(address, mask int) (err error) {
	logrus.Debug("Clear mask ", address, mask)
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	logrus.Debug("Set mask value ", address, current&^byte(mask))
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
		logrus.Debug("Antenna", current)
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

func (r *RFID) cardWrite(command byte, data []byte) (error bool, backData []byte, backLength int, err error) {
	backData = make([]byte, 0)
	backLength = -1
	error = false
	irq := byte(0x00)
	irqWait := byte(0x00)

	switch command {
	case commands.ModeAuth:
		irq = 0x12
		irqWait = 0x10
	case commands.ModeTransrec:
		irq = 0x77
		irqWait = 0x30
	}

	r.devWrite(0x02, irq|0x80)
	r.clearBitmask(0x04, 0x80)
	r.setBitmask(0x0A, 0x80)
	r.devWrite(0x01, commands.ModeIdle)

	for _, v := range data {
		r.devWrite(0x09, v)
	}

	r.devWrite(0x01, command)

	if command == commands.ModeTransrec {
		r.setBitmask(0x0D, 0x80)
	}

	i := 2000
	n := byte(0)

	for ; i > 0; i-- {
		n, err = r.devRead(0x04)
		if err != nil {
			return
		}
		if n&0x01 == 0 || n&irqWait == 0 {
			break
		}
	}

	r.clearBitmask(0x0D, 0x80)

	if i == 0 {
		error = true
		return
	}

	if d, err1 := r.devRead(0x06); err1 != nil || d&0x1B != 0 {
		err = err1
		error = true
		logrus.Error("E2")
		return
	}

	if n&irq&0x01 == 1 {
		logrus.Error("E1")
		error = true
	}

	if command == commands.ModeTransrec {
		n, err = r.devRead(0x0A)
		logrus.Info("N is ", n)
		if err != nil {
			return
		}
		lastBits, err1 := r.devRead(0x0C)
		logrus.Info("lastBits is ", lastBits)
		if err1 != nil {
			err = err1
			return
		}
		lastBits = lastBits & 0x07
		if lastBits != 0 {
			backLength = (int(n)-1)*8 + int(lastBits)
		} else {
			backLength = int(n) * 8
		}

		if n == 0 {
			n = 1
		}

		if n > 16 {
			n = 16
		}

		for i := byte(0); i < n; i++ {
			byteVal, err1 := r.devRead(0x09)
			if err1 != nil {
				err = err1
				return
			}
			backData = append(backData, byteVal)
		}

	}

	return
}

func (r *RFID) Request() (error bool, backBits int, err error) {
	error = true
	backBits = 0
	err = r.devWrite(0x0D, 0x07)
	if err != nil {
		return
	}

	error, _, backBits, err = r.cardWrite(commands.ModeTransrec, []byte{0x26}[:])

	logrus.Info(error, err, backBits)

	error = err != nil || error || backBits != 0x10

	return
}

func (r *RFID) Wait() (err error) {
	irqChannel := make(chan bool)
	r.IrqPin.BeginWatch(gpio.EdgeFalling, func() {
		irqChannel <- true
	})

	defer func() {
		r.IrqPin.EndWatch()
		close(irqChannel)
	}()

	err = r.Init()
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
	logrus.SetLevel(logrus.ErrorLevel)

interruptLoop:
	for {
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
		case _ = <-irqChannel:
			break interruptLoop
		case <-time.After(100 * time.Millisecond):
			// do nothing
		}
	}
	return
}
