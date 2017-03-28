package lcd

import (
	"time"
	"gobot.io/x/gobot/platforms/raspi"
)

const LineOne = 0x00 // start of line 1
const LineTwo = 0x40 // start of line 2

const lcd_Clear = 1            // 0b00000001 replace all characters with ASCII 'space'
const lcd_Home = 2             // 0b00000010 return cursor to first position on first line
const lcd_EntryMode = 6        // 0b00000110 shift cursor from left to right on read/write
const lcd_DisplayOff = 8       // 0b00001000 turn display off
const lcd_DisplayOn = 12       // 0b00001100 display on, cursor off, don't blink character
const lcd_FunctionReset = 48   // 0b00110000 reset the LCD
const lcd_FunctionSet4bit = 40 // 0b00101000 4-bit data, 2-line display, 5 x 7 font
const lcd_SetCursor = 128      // 0b10000000 set cursor position

type Pin struct {
	physId string
}

type PiLCD4 struct {
	adapterRef *raspi.Adaptor

	// Data pins, 1 - lowest
	DataPin1 Pin
	DataPin2 Pin
	DataPin3 Pin
	DataPin4 Pin

	// register select pin
	RsPin Pin
	// enable pin
	EnablePin Pin
}

type PiLCD interface {
	Init()

	Cls()

	Print(data string)
}

func (r *PiLCD4) Init() {
	ref := raspi.NewAdaptor()

	ref.DigitalWrite(r.RsPin.physId, 0)     // set RS to low
	ref.DigitalWrite(r.EnablePin.physId, 0) // set E to low

	// repeat the Init command 3 times
	write4Bits(r, lcd_FunctionReset)
	delay(10)
	write4Bits(r, lcd_FunctionReset)
	delay(200)
	write4Bits(r, lcd_FunctionReset)
	delay(200)

	// setup 4 bits
	write4Bits(r, lcd_FunctionSet4bit)
	delay(80)
	writeInstruction(r, lcd_FunctionSet4bit)
	delay(80)

	// turn screen off
	writeInstruction(r, lcd_DisplayOff)
	delay(80)

	writeInstruction(r, lcd_Clear)
	delay(4)

	writeInstruction(r, lcd_EntryMode)
	delay(80)

	writeInstruction(r, lcd_DisplayOn) // turn the display ON
	delay(80)
}

func (r *PiLCD4) Cls() {
	writeInstruction(r, lcd_Clear)
}

func (r *PiLCD4) Print(data string) {
	writeString(r, data)
}

// =============================================== service methods

func writeString(ref *PiLCD4, data string) {
	for _, v := range []byte(data) {
		writeChar(ref, v)
	}
}

func writeChar(ref *PiLCD4, data uint8) {
	sendData(ref)
	write4Bits(ref, data>>4)
	write4Bits(ref, data)
}

func write4Bits(ref *PiLCD4, data uint8) {
	ref.adapterRef.DigitalWrite(ref.DataPin1.physId, data&1)
	ref.adapterRef.DigitalWrite(ref.DataPin2.physId, data&2)
	ref.adapterRef.DigitalWrite(ref.DataPin2.physId, data&4)
	ref.adapterRef.DigitalWrite(ref.DataPin2.physId, data&8)
	ref.adapterRef.DigitalWrite(ref.EnablePin.physId, 1)
	delay(1)
	ref.adapterRef.DigitalWrite(ref.EnablePin.physId, 0)
}

func sendInstruction(ref *PiLCD4) {
	ref.adapterRef.DigitalWrite(ref.RsPin.physId, 0)
	ref.adapterRef.DigitalWrite(ref.EnablePin.physId, 0)
}

func sendData(ref *PiLCD4) {
	ref.adapterRef.DigitalWrite(ref.RsPin.physId, 1)
	ref.adapterRef.DigitalWrite(ref.EnablePin.physId, 0)
}

func writeInstruction(ref *PiLCD4, data uint8) {
	sendInstruction(ref)
	// write high 4 bits
	write4Bits(ref, data>>4)
	// write low  bits
	write4Bits(ref, data)

}

func delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Microsecond)
}
