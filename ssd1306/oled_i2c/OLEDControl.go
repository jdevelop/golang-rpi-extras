package oled_i2c

// Modified from the original code at https://github.com/goiot/devices/tree/master/monochromeoled
// Works with 128x62 and 128x32

import (
	"fmt"
	"image"

	"github.com/jdevelop/golang-rpi-extras/ssd1306"
	"golang.org/x/exp/io/i2c"
	"golang.org/x/exp/io/i2c/driver"
)

// OLED represents an SSD1306 OLED display.
type OLED struct {
	dev *i2c.Device

	w   int    // width of the display
	h   int    // height of the display
	buf []byte // each pixel is represented by a bit
}

// Open opens an SSD1306 OLED display. Once not in use, it needs to
// be close by calling Close.
// The default width is 128, height is 64 if zero values are given.
func Open(o driver.Opener, addr, w, h int) (*OLED, error) {
	dev, err := i2c.Open(o, addr)
	buf := make([]byte, w*(h/8)+1)
	buf[0] = ssd1306.SSD1306_SETSTARTLINE // start frame of pixel data
	oled := &OLED{dev: dev, w: w, h: h, buf: buf}
	err = oled.Init()
	if err != nil {
		return nil, err
	}
	return oled, nil
}

var initSeq = []byte{
	0xAE, 0xA8, 0x3F, 0xD3, 0x00, 0x40, 0xA1, 0xC8,
	0xA6, 0xD5, 0x80, 0xDA, 0x12, 0x81, 0x00, 0xB0,
	0xA4, 0xDB, 0x40, 0x20, 0x00, 0x00, 0x10, 0x8D,
	0x14, 0x2E, 0xA6, 0xAF,
}

// Init sets up the display for writing
func (o *OLED) Init() (err error) {
	err = o.writeI2CCmd(initSeq)
	return
}

// On turns on the display if it is off.
func (o *OLED) On() error {
	return o.dev.Write([]byte{0x0, ssd1306.SSD1306_DISPLAYON})
}

// Off turns off the display if it is on.
func (o *OLED) Off() error {
	return o.dev.Write([]byte{0x0, ssd1306.SSD1306_DISPLAYOFF})
}

// Clear clears the entire display.
func (o *OLED) Clear() error {
	for i := 1; i < len(o.buf); i++ {
		o.buf[i] = 0
	}
	return o.Draw()
}

// SetPixel set and x,y pixel to on or off
func (o *OLED) SetPixel(x, y int, v byte) error {
	if x >= o.w || y >= o.h {
		return fmt.Errorf("(x=%v, y=%v) is out of bounds on this %vx%v display", x, y, o.w, o.h)
	}
	if v > 1 {
		return fmt.Errorf("value needs to be either 0 or 1; given %v", v)
	}
	i := 1 + x + (y/8)*o.w
	if v == 0 {
		o.buf[i] &= ^(1 << uint((y & 7)))
	} else {
		o.buf[i] |= 1 << uint((y & 7))
	}
	return nil
}

// SetImage draws an image on the display buffer starting from x, y.
// A call to Draw is required to display it on the OLED display.
func (o *OLED) SetImage(x, y int, img image.Image) error {
	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	endX := x + imgW
	endY := y + imgH

	if endX >= o.w {
		endX = o.w
	}
	if endY >= o.h {
		endY = o.h
	}

	var imgI, imgY int
	for i := x; i < endX; i++ {
		imgY = 0
		for j := y; j < endY; j++ {
			r, g, b, _ := img.At(imgI, imgY).RGBA()
			var v byte
			if r+g+b > 0 {
				v = 0x1
			}
			if err := o.SetPixel(i, j, v); err != nil {
				return err
			}
			imgY++
		}
		imgI++
	}
	return nil
}

// Draw draws the intermediate pixel buffer on the display.
// See SetPixel and SetImage to mutate the buffer.
func (o *OLED) Draw() error {
	if err := o.writeI2CCmd([]byte{
		ssd1306.SSD1306_DISPLAYALLON_RESUME, // write mode
		ssd1306.SSD1306_SETSTARTLINE | 0,    // start line = 0
		ssd1306.SSD1306_COLUMNADDR, 0, uint8(o.w),
		ssd1306.SSD1306_PAGEADDR, 0, 7,
	}); err != nil { // the write mode
		return err
	}
	return o.dev.Write(o.buf)
}

// EnableScroll starts scrolling in the horizontal direction starting from
// startY column to endY column.
func (o *OLED) EnableScroll(startY, endY int) error {
	panic("not implemented")
}

// DisableScroll stops the scrolling on the display.
func (o *OLED) DisableScroll() error {
	panic("not implemented")
}

// Width returns the display width.
func (o *OLED) Width() int { return o.w }

// Height returns the display height.
func (o *OLED) Height() int { return o.h }

// Close closes the display.
func (o *OLED) Close() error {
	return o.dev.Close()
}

func (o *OLED) writeI2CCmd(commands []byte) (err error) {
	data := make([]byte, 2)
	for _, v := range initSeq {
		data[1] = v
		err = o.dev.Write(data)
		if err != nil {
			return
		}
	}
	return
}
