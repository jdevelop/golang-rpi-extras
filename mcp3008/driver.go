package mcp3008

import (
	"fmt"
	"golang.org/x/exp/io/spi"
)

type Mode int

const (
	Mode0 = Mode(0)
	Mode1 = Mode(1)
	Mode2 = Mode(2)
	Mode3 = Mode(3)
)

type MCP3008 struct {
	bus     int
	channel int
	dev     *spi.Device
}

type MeasureAnalog interface {
	Measure(channel int) (result int, err error)
}

func NewMCP3008(bus int, busChannel int, mode Mode, maxSpeed int64) (mcp MCP3008, resErr error) {
	var dev, err = spi.Open(&spi.Devfs{
		Dev:      fmt.Sprintf("/dev/spidev%d.%d", bus, busChannel),
		Mode:     spi.Mode(mode),
		MaxSpeed: maxSpeed,
	})
	if err != nil {
		resErr = err
		return
	}
	mcp.dev = dev
	return
}

func (r MCP3008) Measure(channel int) (result int, err error) {
	resp := make([]byte, 3)
	if err := r.dev.Tx([]byte{1, byte((8 + channel) << 4), 0}, resp); err == nil {
		result = int(resp[1]&3)<<8 + int(resp[2])
	}
	return
}
