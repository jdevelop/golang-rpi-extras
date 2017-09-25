package main

import (
	"github.com/jdevelop/golang-rpi-extras/rf522"
	"log"
	"fmt"
)

func rfid() {
	// use BCM numbering here
	rfid, err := rf522.MakeRFID(0, 0, 1000000, 25, 24)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting")
	err = rfid.Wait()
	fmt.Println("Something happened")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Hooray!")
	}
}

func main() {

	rfid()

}
