package main

import (
	"github.com/jdevelop/golang-rpi-extras/rf522"
	"log"
	"github.com/sirupsen/logrus"
	"os"
	"github.com/jdevelop/golang-rpi-extras/rf522/commands"
)

func main() {
	// use BCM numbering here
	logrus.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
	rfid, err := rf522.MakeRFID(0, 0, 1000000, 25, 24)
	if err != nil {
		log.Fatal(err)
	}
	err = rfid.Wait()
	if err != nil {
		log.Fatal(err)
	}
	logrus.SetLevel(logrus.InfoLevel)
	rfid.Init()
	status, backBits, err := rfid.Request()
	logrus.Info(status, backBits, err)
	if !status || err != nil {
		log.Fatal("No request ", status, err)
	}
	status, uuid, err := rfid.AntiColl()
	if !status || err != nil {
		logrus.Fatal("No anticol ", status, err)
	}
	logrus.Info("UUID is ", uuid)

	blocks, err := rfid.SelectTag(uuid)
	if err != nil {
		logrus.Fatal("Error selecting tag ", err)
	}

	logrus.Info("Blocks found ", blocks)

	blockAddress := byte(5)

	state, err := rfid.Auth(commands.PICC_AUTHENT1B, blockAddress, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, uuid)
	if err != nil || state != rf522.AuthOk {
		logrus.Fatal("Can not authenticate ", err, " => ", state)
	}

	status, data, err := rfid.Read(blockAddress)

	if !status || err != nil {
		logrus.Fatal("Can not read ", blocks, " => ", status, err)
	}

	logrus.Info("Read data ", data)

	logrus.Info("Write data")

	err = rfid.Write(blockAddress, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F})
	if err != nil {
		logrus.Fatal(err)
	}

	err = rfid.StopCrypto()
	if err != nil {
		logrus.Fatal(err)
	}

}
