package commands

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