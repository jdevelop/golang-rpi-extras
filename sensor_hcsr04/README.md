# Example

```go
package main

import (
	hc "github.com/jdevelop/golang-rpi-extras/sensor_hcsr04"
	lcd "github.com/jdevelop/golang-rpi-extras/lcd_hd44780"
	"fmt"
	"time"
)

func main() {

        // Use BCM pin numbering
        // Echo pin
        // Trigger pin
	h := hc.NewHCSR04(4, 25)

	myLcd, err := lcd.NewLCD4([]int{27, 22, 23, 24}, 17, 18)

	if err != nil {
		panic(err.Error())
	}

	myLcd.Init()
	myLcd.Cls()

	for true {
		distance := h.MeasureDistance()
		fmt.Println(distance)
		myLcd.SetCursor(0, 0)
		myLcd.Print(fmt.Sprintf("Distance: %5.2f", distance))
		time.Sleep(time.Duration(1) * time.Second)
	}

}
```