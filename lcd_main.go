package main

import (
	lcd "github.com/jdevelop/golang-rpi-extras/lcd_hd44780"
	"flag"
	"strings"
	"strconv"
)

func main() {

	rsPin := flag.Int("rs", 20, "RS pin")
	ePin := flag.Int("e", 21, "E pin")
	data := flag.String("data", "6,13,19,26", "data pins, comma-separated")

	pinsStr := strings.Split(*data, ",")
	pins := make([]int, 4)

	for i, pin := range pinsStr {
		pins[i], _ = strconv.Atoi(pin)
	}

	//rpi, err := lcd.NewLCD4([]int{6, 13, 19, 26}, 20, 21)
	rpi, err := lcd.NewLCD4(pins, *rsPin, *ePin)

	if err != nil {
		panic(err.Error())
	}

	rpi.Init()
	rpi.Cls()
	rpi.Print("-=  HELLO  =-")
	rpi.SetCursor(1, 0)
	rpi.Print("-=  WORLD  =-")

}
