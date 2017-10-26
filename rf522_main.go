package main

import (
	"github.com/jdevelop/golang-rpi-extras/rf522"
	"log"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	// use BCM numbering here
	logrus.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	rfid, err := rf522.MakeRFID(0, 0, 1000000, 25, 24)
	if err != nil {
		log.Fatal(err)
	}
	err = rfid.Wait()
	if err != nil {
		log.Fatal(err)
	}
	logrus.SetLevel(logrus.DebugLevel)
	rfid.Init()
	errorF, backBits, err := rfid.Request()
	logrus.Info(errorF, backBits, err)
}