#Example

```go
package main

import lcd "github.com/jdevelop/golang-rpi-extras/lcd_hd44780"

func main() {
        // Use BCM PIN numbering
        // 4 pins for data transfer
        // RS pin
        // E pin
	rpi, err := lcd.NewLCD4([]int{27, 22, 23, 24}, 17, 18)

	if err != nil {
		panic(err.Error())
	}

	rpi.Cls()
	rpi.Print("-=   HELLO   =-")
	rpi.SetCursor(1, 0)
	rpi.Print("-=   WORLD  =-")

}

```