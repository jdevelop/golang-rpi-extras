package lcd

import (
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

const LineTwo = 0x40 // start of line 2

type Pin struct {
	PhysId string
}

type PiLCD4 struct {
	adapterRef *raspi.Adaptor

	// Data pins
	DataPins []Pin

	// register select pin
	RsPin Pin

	// enable pin
	EnablePin Pin
}

type PiLCD interface {
	Init()

	Cls()

	Print(data string)

	WriteChar(data uint8)

	SetCursor(line uint8, column uint8)
}

func clearBits(r *PiLCD4) {
	for _, v := range r.DataPins {
		r.adapterRef.DigitalWrite(v.PhysId, 0)
	}
}

func (r *PiLCD4) Init() {
	ref := raspi.NewAdaptor()
	ref.Connect()
	r.adapterRef = ref

	clearBits(r)

	delayMs(15)

	ref.DigitalWrite(r.RsPin.PhysId, 0)     // set RS to low
	ref.DigitalWrite(r.EnablePin.PhysId, 0) // set E to low

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
	for i, v := range ref.DataPins {
		ref.adapterRef.DigitalWrite(v.PhysId, data&(1<<uint(i)))
	}
	strobe(ref)
}

func sendInstruction(ref *PiLCD4) {
	ref.adapterRef.DigitalWrite(ref.RsPin.PhysId, 0)
	ref.adapterRef.DigitalWrite(ref.EnablePin.PhysId, 0)
}

func sendData(ref *PiLCD4) {
	ref.adapterRef.DigitalWrite(ref.RsPin.PhysId, 1)
	ref.adapterRef.DigitalWrite(ref.EnablePin.PhysId, 0)
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
	ref.adapterRef.DigitalWrite(ref.EnablePin.PhysId, 1)
	delayUs(2)
	ref.adapterRef.DigitalWrite(ref.EnablePin.PhysId, 0)
}

func delayUs(ms int) {
	time.Sleep(time.Duration(ms) * time.Microsecond)
}

func delayMs(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
