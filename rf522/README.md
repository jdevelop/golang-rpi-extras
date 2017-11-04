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
	"strconv"
	"github.com/jdevelop/golang-rpi-extras/rf522/commands"
)

func main() {

	//currentAccessKey := [6]byte{1, 2, 3, 4, 5, 6}
	currentAccessKey := [6]byte{6, 5, 4, 3, 2, 1}
	currentAccessMethod := byte(commands.PICC_AUTHENT1B)

	sector := flag.Int("s", 1, "card sector")
	block := flag.Int("b", 0, "card block")
	overwriteKey := flag.Bool("wa", false, "Overwrite keys")
	overwriteBlock := flag.Bool("wb", false, "Overwrite block with 0-15")

	flag.Parse()

	// use BCM numbering here
	logrus.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
	rfid, err := rf522.MakeRFID(0, 0, 1000000, 25, 24)
	if err != nil {
		log.Fatal(err)
	}

	data, err := rfid.ReadCard(currentAccessMethod, *sector, *block, currentAccessKey[:])
	auth, err := rfid.ReadAuth(currentAccessMethod, *sector, currentAccessKey[:])

	access := rf522.ParseBlockAccess(auth[6:10])

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("RFID sector %d, block %d : %v, auth: %v\n", *sector, *block, data, auth)
	fmt.Printf("Permissions: B0: %s, B1: %s, B2: %s, B3/A: %s\n",
		strconv.FormatUint(uint64(access.B0), 2),
		strconv.FormatUint(uint64(access.B1), 2),
		strconv.FormatUint(uint64(access.B2), 2),
		strconv.FormatUint(uint64(access.B3), 2),
	)

	if *overwriteBlock {
		err = rfid.WriteBlock(currentAccessMethod,
			*sector,
			*block,
			[16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			currentAccessKey[:])
		if err != nil {
			log.Fatal(err)
		}
	}

	if *overwriteKey {
		err = rfid.WriteSectorTrail(commands.PICC_AUTHENT1A,
			*sector,
			[6]byte{1, 2, 3, 4, 5, 6},
			[6]byte{6, 5, 4, 3, 2, 1},
			&rf522.BlocksAccess{
				B0: rf522.RAB_WB_IB_DAB,
				B1: rf522.RB_WB_IN_DN,
				B2: rf522.AnyKeyRWID,
				B3: rf522.KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN,
			},
			currentAccessKey[:],
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Write successful")
	}

}
```