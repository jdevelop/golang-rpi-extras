package oled_spi

import (
	"errors"
	"fmt"
	"github.com/ecc1/spi"
	ssd "github.com/jdevelop/golang-rpi-extras/ssd1306"
	"github.com/jdevelop/gpio"
	"image"
	"image/color"
	"time"
)

type SSD1306 struct {
	spiDev               *spi.Device
	Width, Height, Pages uint
	reset, dc            gpio.Pin
	buffer               []byte
	cmd                  []byte
}

type SSD1306Setup struct {
	BusId, DevId, ResetPin, DcPin uint
	Width, Height                 uint
}

func defaultSetup() *SSD1306Setup {
	return &SSD1306Setup{
		BusId:    0,
		DevId:    0,
		DcPin:    17,
		ResetPin: 27,
		Width:    128,
		Height:   96,
	}
}

type Setup func(*SSD1306Setup)

func Width(w uint) Setup {
	return func(c *SSD1306Setup) {
		c.Width = w
	}
}

func Height(w uint) Setup {
	return func(c *SSD1306Setup) {
		c.Height = w
	}
}

func MakeSSD1306(funcs ...Setup) (display *SSD1306, err error) {
	ssd := defaultSetup()
	for _, s := range funcs {
		s(ssd)
	}
	spiDev, err := spi.Open(fmt.Sprintf("/dev/spidev%d.%d", ssd.BusId, ssd.DevId), 8000000, 0)
	if err != nil {
		return
	}
	spiDev.SetLSBFirst(false)
	spiDev.SetBitsPerWord(8)

	dsp := SSD1306{
		Height: ssd.Height,
		Width:  ssd.Width,
	}

	display = &dsp

	dsp.Pages = dsp.Height / 8

	dsp.buffer = make([]byte, dsp.Width*dsp.Pages)

	dsp.spiDev = spiDev

	dsp.cmd = make([]byte, 1)

	dsp.reset, err = gpio.OpenPin(int(ssd.ResetPin), gpio.ModeOutput)
	if err != nil {
		return
	}

	dsp.dc, err = gpio.OpenPin(int(ssd.DcPin), gpio.ModeOutput)
	if err != nil {
		return
	}

	return
}

func (s *SSD1306) Command(cmd byte) (err error) {
	s.dc.Clear()
	s.cmd[0] = cmd
	err = s.spiDev.Write(s.cmd)
	return
}

func (s *SSD1306) Commands(cmd []byte) (err error) {
	s.dc.Clear()
	err = s.spiDev.Write(cmd)
	return
}

func (s *SSD1306) Data(data []byte) (err error) {
	s.dc.Set()
	err = s.spiDev.Write(data)
	return
}

func (s *SSD1306) Reset() (err error) {
	s.reset.Set()
	time.Sleep(time.Millisecond)
	s.reset.Clear()
	time.Sleep(10 * time.Millisecond)
	s.reset.Set()
	return
}

func (s *SSD1306) Start() (err error) {
	err = s.Reset()
	if err != nil {
		return
	}
	s.Init()
	err = s.Command(ssd.SSD1306_DISPLAYON)
	return
}

var initCmds = []byte{ssd.SSD1306_DISPLAYOFF, ssd.SSD1306_SETDISPLAYCLOCKDIV, 0x80,
	ssd.SSD1306_SETMULTIPLEX, 0x3F, ssd.SSD1306_SETDISPLAYOFFSET,
	0x0, ssd.SSD1306_SETSTARTLINE | 0x0, ssd.SSD1306_CHARGEPUMP, 0x10,
	ssd.SSD1306_MEMORYMODE, 0x00, ssd.SSD1306_SEGREMAP | 0x1, ssd.SSD1306_COMSCANDEC, ssd.SSD1306_SETCOMPINS,
	0x12, ssd.SSD1306_SETCONTRAST, 0x9F, ssd.SSD1306_SETPRECHARGE, 0x22,
	ssd.SSD1306_SETVCOMDETECT, 0x40, ssd.SSD1306_DISPLAYALLON_RESUME,
	ssd.SSD1306_NORMALDISPLAY}

func (s *SSD1306) Init() (err error) {
	s.Commands(initCmds)
	return
}

var refreshCmdBuf = []byte{ssd.SSD1306_COLUMNADDR, 0, 0, ssd.SSD1306_PAGEADDR, 0, 0}

func (s *SSD1306) Refresh() (err error) {
	refreshCmdBuf[2] = byte(s.Width - 1)
	refreshCmdBuf[5] = byte(s.Pages - 1)
	for _, cmdByte := range refreshCmdBuf {
		err = s.Command(cmdByte)
		if err != nil {
			return
		}
	}
	s.dc.Set()
	err = s.spiDev.Write(s.buffer)
	return
}

func (s *SSD1306) Clear() {
	for i := range s.buffer {
		s.buffer[i] = 0
	}
}

func (s *SSD1306) Image(img image.Image) (err error) {
	dim := img.Bounds()
	if uint(dim.Max.X) > s.Width || uint(dim.Max.Y) > s.Height {
		err = errors.New("Image dimensions are not correct")
		return
	}
	byteIdx := 0
	for page := uint(0); page < s.Pages; page++ {
		for column := uint(0); column < s.Width; column++ {
			bits := byte(0)
			for bit := uint(0); bit < 8; bit++ {
				if c := img.At(int(column), int(page*8+7-bit)); c != color.Black {
					bits = bits | (1 << bit)
				}
			}
			s.buffer[byteIdx] = bits
			byteIdx = byteIdx + 1
		}
	}
	return
}

func (s *SSD1306) Contrast(contrast byte) (err error) {
	err = s.Command(ssd.SSD1306_SETCONTRAST)
	if err != nil {
		return
	}
	err = s.Command(contrast)
	return
}
