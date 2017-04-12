package lcd_hd44780

import (
	"github.com/stianeikeland/go-rpio"
	"time"
)

const LineTwo = 0x40 // start of line 2

type PiLCD4 struct {
	// Data pins
	dataPins []rpio.Pin

	// register select pin
	rsPin rpio.Pin

	// enable pin
	enablePin rpio.Pin
}

func NewLCD4(data []int, rs int, e int) (pLcd PiLCD4, err error) {
	if err = rpio.Open(); err != nil {
		return
	}
	pLcd.rsPin = rpio.Pin(rs)
	pLcd.enablePin = rpio.Pin(e)
	pLcd.dataPins = make([]rpio.Pin, len(data))
	for i, v := range data {
		pLcd.dataPins[i] = rpio.Pin(v)
	}
	return
}

type PiLCD interface {
	Init()

	Cls()

	Print(data string)

	WriteChar(data uint8)

	SetCursor(line uint8, column uint8)
}

func clearBits(r *PiLCD4) {
	for _, v := range r.dataPins {
		v.Low()
	}
}

func (r *PiLCD4) Init() {

	r.rsPin.Output()
	r.enablePin.Output()

	for _, v := range r.dataPins {
		v.Output()
	}

	clearBits(r)

	delayMs(15)

	r.rsPin.Low()     // set RS to low
	r.enablePin.Low() // set E to low

	writeDelay := func(data uint8, delay int) {
		write4Bits(r, data)
		delayUs(delay)
	}

	// repeat the Init command 3 times
	writeDelay(3, 50)
	writeDelay(3, 10)
	writeDelay(3, 10)

	writeDelay(2, 10)

	// set 4bit mode. 2 lines, 5x7
	writeInstruction(r, 0x14)

	// disable display
	writeInstruction(r, 0x10)

	// clear display
	writeInstruction(r, 0x1)
	delayMs(2)

	// cursor shift right, no display move
	writeInstruction(r, 0x06)

	// enable display no cursor
	writeInstruction(r, 0x0c)

	// clear screen
	writeInstruction(r, 0x1)
	delayMs(2)

	// home
	writeInstruction(r, 2)
	delayMs(2)
}

func (r *PiLCD4) Cls() {
	writeInstruction(r, 0x01)
	delayMs(2)
}

func (r *PiLCD4) SetCursor(line uint8, column uint8) {
	writeInstruction(r, 0x80|(line*LineTwo+column))
}

func (r *PiLCD4) Print(data string) {
	for _, v := range []byte(data) {
		r.WriteChar(v)
	}
}

func (r *PiLCD4) WriteChar(data uint8) {
	sendData(r)
	write4Bits(r, data>>4)
	write4Bits(r, data)
	delayUs(10)
}

// =============================================== service methods ============================================

func write4Bits(ref *PiLCD4, data uint8) {
	for i, v := range ref.dataPins {
		if data&(1<<uint(i)) > 0 {
			v.High()
		} else {
			v.Low()
		}
	}
	strobe(ref)
}

func sendInstruction(ref *PiLCD4) {
	ref.rsPin.Low()
	ref.enablePin.Low()
}

func sendData(ref *PiLCD4) {
	ref.rsPin.High()
	ref.enablePin.Low()
}

func writeInstruction(ref *PiLCD4, data uint8) {
	sendInstruction(ref)
	// write high 4 bits
	write4Bits(ref, data>>4)
	// write low  bits
	write4Bits(ref, data)
	delayUs(50)
}

func strobe(ref *PiLCD4) {
	ref.enablePin.High()
	delayUs(2)
	ref.enablePin.Low()
}

func delayUs(ms int) {
	time.Sleep(time.Duration(ms) * time.Microsecond)
}

func delayMs(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
