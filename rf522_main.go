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

	state, err := rfid.Auth(commands.AuthB, 0, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, uuid)
	if err != nil || state != rf522.AuthOk {
		logrus.Fatal("Can not authenticate ", err, " => ", state)
	}

	status, data, err := rfid.Read(8)

	if !status || err != nil {
		logrus.Fatal("Can not read ", blocks, " => ", status, err)
	}

	logrus.Info("Read data ", data)

	err = rfid.StopCrypto()
	if err != nil {
		logrus.Fatal(err)
	}

}
