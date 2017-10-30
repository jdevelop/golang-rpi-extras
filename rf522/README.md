# RF522 wireless RFID reader

Ported the [Python](https://github.com/mxgxw/MFRC522-python) library to access RF522 cards to golang.

Example usage:

```go
package main

import (
	"github.com/jdevelop/golang-rpi-extras/rf522"
	"log"
	"github.com/sirupsen/logrus"
	"os"
	"flag"
	"fmt"
)

func main() {

	sector := flag.Int("sector", 1, "card sector")
	block := flag.Int("block", 0, "card block")

	flag.Parse()

	// use BCM numbering here
	logrus.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
	rfid, err := rf522.MakeRFID(0, 0, 1000000, 25, 24)
	if err != nil {
		log.Fatal(err)
	}

	data, err := rfid.ReadCard(byte(*sector), byte(*block), rf522.DefaultKey)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RFID sector %d, block %d : %v", *sector, *block, data)

}

```