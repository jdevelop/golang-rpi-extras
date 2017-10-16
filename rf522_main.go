package main

import (
	"fmt"
	"log"
	"github.com/ecc1/spi"
)

func main() {

	spiDev, err := spi.Open("/dev/spidev0.0", 1000000, 0)

	spiDev.SetMode(0)
	spiDev.SetBitsPerWord(8)
	spiDev.SetLSBFirst(false)
	spiDev.SetMaxSpeed(1000000)

	if err != nil {
		log.Fatal(err)
	}

	writeSpiData := func(dataIn []byte) (err error) {
		err = spiDev.Transfer(dataIn)
		return
	}

	devWrite := func(address int, data byte) (err error) {
		newData := [2]byte{(byte(address) << 1) & 0x7E, data}
		fmt.Print("<< ", newData, " ")
		err = writeSpiData(newData[0:])
		fmt.Println(">>", newData)
		return
	}

	if err != nil {
		log.Fatal(err)
	}

	devWrite(0x01, 0x0F)

	if err != nil {
		log.Fatal(err)
	}

}
