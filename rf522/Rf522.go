package rf522

import (
	"fmt"
	"github.com/jdevelop/golang-rpi-extras/rf522/commands"
	"time"
	"github.com/ecc1/spi"
	"github.com/davecheney/gpio"
	rpio "github.com/davecheney/gpio/rpi"
	"github.com/sirupsen/logrus"
	"errors"
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
	err = r.devWrite(commands.CommandReg, commands.PCD_RESETPHASE)
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
		current, err := r.devRead(commands.TxControlReg)
		logrus.Debug("Antenna", current)
		if err != nil {
			return err
		}
		if current&0x03 == 0 {
			err = r.setBitmask(commands.TxControlReg, 0x03)
		}
	} else {
		err = r.clearBitmask(commands.TxControlReg, 0x03)
	}
	return
}

func (r *RFID) cardWrite(command byte, data []byte) (error bool, backData []byte, backLength int, err error) {
	backData = make([]byte, 0)
	backLength = -1
	error = false
	irqEn := byte(0x00)
	irqWait := byte(0x00)

	switch command {
	case commands.PCD_AUTHENT:
		irqEn = 0x12
		irqWait = 0x10
	case commands.PCD_TRANSCEIVE:
		irqEn = 0x77
		irqWait = 0x30
	}

	r.devWrite(commands.CommIEnReg, irqEn|0x80)
	r.clearBitmask(commands.CommIrqReg, 0x80)
	r.setBitmask(commands.FIFOLevelReg, 0x80)
	r.devWrite(commands.CommandReg, commands.PCD_IDLE)

	for _, v := range data {
		r.devWrite(commands.FIFODataReg, v)
	}

	r.devWrite(commands.CommandReg, command)

	if command == commands.PCD_TRANSCEIVE {
		r.setBitmask(commands.BitFramingReg, 0x80)
	}

	i := 2000
	n := byte(0)

	for ; i > 0; i-- {
		n, err = r.devRead(commands.CommIrqReg)
		if err != nil {
			return
		}
		if n&(irqWait|1) != 0 {
			break
		}
	}

	r.clearBitmask(commands.BitFramingReg, 0x80)

	if i == 0 {
		error = true
		return
	}

	if d, err1 := r.devRead(commands.ErrorReg); err1 != nil || d&0x1B != 0 {
		err = err1
		error = true
		logrus.Error("E2")
		return
	}

	if n&irqEn&0x01 == 1 {
		logrus.Error("E1")
		error = true
	}

	if command == commands.PCD_TRANSCEIVE {
		n, err = r.devRead(commands.FIFOLevelReg)
		if err != nil {
			return
		}
		lastBits, err1 := r.devRead(commands.ControlReg)
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
			byteVal, err1 := r.devRead(commands.FIFODataReg)
			if err1 != nil {
				err = err1
				return
			}
			backData = append(backData, byteVal)
		}

	}

	return
}

func (r *RFID) Request() (status bool, backBits int, err error) {
	status = true
	backBits = 0
	err = r.devWrite(commands.BitFramingReg, 0x07)
	if err != nil {
		return
	}

	status, _, backBits, err = r.cardWrite(commands.PCD_TRANSCEIVE, []byte{0x26}[:])

	logrus.Info(status, err, backBits)

	status = err == nil && !status && backBits == 0x10

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
	err = r.devWrite(commands.CommIrqReg, 0x00)
	if err != nil {
		return
	}
	err = r.devWrite(commands.CommIEnReg, 0xA0)
	if err != nil {
		return
	}
	logrus.SetLevel(logrus.ErrorLevel)

interruptLoop:
	for {
		err = r.devWrite(commands.FIFODataReg, 0x26)
		if err != nil {
			return
		}
		err = r.devWrite(commands.CommandReg, 0x0C)
		if err != nil {
			return
		}
		err = r.devWrite(commands.BitFramingReg, 0x87)
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

func (r *RFID) AntiColl() (status bool, backData []byte, err error) {

	err = r.devWrite(commands.BitFramingReg, 0x00)

	status, backData, _, err = r.cardWrite(commands.PCD_TRANSCEIVE, []byte{commands.PICC_ANTICOLL, 0x20}[:])

	if status || err != nil {
		logrus.Error("Card write ", status, err)
		return
	}

	if len(backData) != 5 {
		logrus.Error("Back data expected 5, actual ", len(backData))
		status = false
		return
	}

	crc := byte(0)

	for _, v := range backData[:4] {
		crc = crc ^ v
	}

	logrus.Debug("Back data ", printBytes(backData), ", CRC ", printBytes([]byte{crc}))

	status = crc == backData[4]

	return
}

func (r *RFID) CRC(inData []byte) (res []byte, err error) {
	res = []byte{0, 0}
	err = r.clearBitmask(commands.DivIrqReg, 0x04)
	if err != nil {
		return
	}
	err = r.setBitmask(commands.FIFOLevelReg, 0x80)
	if err != nil {
		return
	}
	for _, v := range inData {
		r.devWrite(commands.FIFODataReg, v)
	}
	err = r.devWrite(commands.CommandReg, commands.PCD_CALCCRC)
	if err != nil {
		return
	}
	for i := byte(0xFF); i > 0; i-- {
		n, err1 := r.devRead(commands.DivIrqReg)
		if err1 != nil {
			err = err1
			return
		}
		if n&0x04 > 0 {
			break
		}
	}
	lsb, err := r.devRead(commands.CRCResultRegL)
	if err != nil {
		return
	}
	res[0] = lsb

	msb, err := r.devRead(commands.CRCResultRegM)
	if err != nil {
		return
	}
	res[1] = msb
	return
}

func (r *RFID) SelectTag(serial []byte) (blocks byte, err error) {
	dataBuf := make([]byte, len(serial)+2)
	dataBuf[0] = commands.PICC_SElECTTAG
	dataBuf[1] = 0x70
	copy(dataBuf[2:], serial)
	crc, err := r.CRC(dataBuf)
	if err != nil {
		return
	}
	dataBuf = append(dataBuf, crc[0], crc[1])
	status, backData, backLen, err := r.cardWrite(commands.PCD_TRANSCEIVE, dataBuf)
	if status || err != nil {
		logrus.Warn("Can't select tag ", status, backData, backLen, err)
		return
	}

	logrus.Info("Tag info : ", status, backData, backLen, err)

	if !status && backLen == 0x18 {
		blocks = backData[0]
	} else {
		blocks = 0
	}
	return
}

type AuthStatus byte

const (
	AuthOk          AuthStatus = iota
	AuthReadFailure
	AuthFailure
)

func (r *RFID) Auth(mode byte, blockAddress byte, sectorKey []byte, serial []byte) (authS AuthStatus, err error) {
	buffer := make([]byte, 2)
	buffer[0] = mode
	buffer[1] = blockAddress
	buffer = append(buffer, sectorKey...)
	buffer = append(buffer, serial[:4]...)
	logrus.Info("CARD Auth: ", printBytes(buffer))
	status, _, _, err := r.cardWrite(commands.PCD_AUTHENT, buffer)
	if err != nil {
		logrus.Error(err)
		return
	}
	if status {
		logrus.Warn("Can not read the card data")
		authS = AuthReadFailure
		return
	}
	n, err := r.devRead(commands.Status2Reg)
	if err != nil {
		logrus.Warn("Can not read device status register")
		return
	}
	if n&0x08 != 0 {
		logrus.Debug("N is ", n)
		authS = AuthFailure
	}
	authS = AuthOk
	return
}

func (r *RFID) StopCrypto() (err error) {
	err = r.clearBitmask(commands.Status2Reg, 0x08)
	return
}

func (r *RFID) preAccess(blockAddr byte, cmd byte) (status bool, data []byte, backLen int, err error) {
	send := make([]byte, 4)
	send[0] = cmd
	send[1] = blockAddr

	crc, err := r.CRC(send[:2])
	if err != nil {
		return
	}
	send[2] = crc[0]
	send[3] = crc[1]
	logrus.Info("Send access data ", printBytes(send))
	status, data, backLen, err = r.cardWrite(commands.PCD_TRANSCEIVE, send)
	return
}

func (r *RFID) Read(blockAddr byte) (status bool, data []byte, err error) {
	status, data, backLen, err := r.preAccess(blockAddr, commands.PICC_READ)
	logrus.Info("Read data:  ", backLen, status, printBytes(data), err)
	status = len(data) == 16
	return
}

func (r *RFID) Write(blockAddr byte, data []byte) (err error) {
	status, read, backLen, err := r.preAccess(blockAddr, commands.PICC_WRITE)
	if status || err != nil || backLen != 4 {
		logrus.Warn("Can not grant Write to block ", status, read, backLen, err)
		return
	}
	if read[0]&0x0F != 0x0A {
		err = errors.New("can't authorize write")
		return
	}
	newData := make([]byte, 18)
	copy(newData, data[:16])
	crc, err := r.CRC(newData[:16])
	if err != nil {
		logrus.Warn("Can't calculate CRC")
		return
	}
	newData[16] = crc[0]
	newData[17] = crc[1]
	status, read, backLen, err = r.cardWrite(commands.PCD_TRANSCEIVE, newData)
	if status {
		err = errors.New("can not write card data")
		return
	}
	status = backLen == 4 && read[0]&0x0F == 0x0A
	return
}
