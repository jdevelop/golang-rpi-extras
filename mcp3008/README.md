# Read value from ADT [MCP3008](https://www.adafruit.com/products/856)

# Example

```go
package main

import (
        "time"
        "fmt"
        mcp "github.com/jdevelop/golang-rpi-extras/mcp3008"
        )

func main() {

	const channel = 0

	// /dev/spidev0.0
	mcp, err := mcp.NewMCP3008(0, 0, mcp.Mode0, 500000)
	
	if err != nil {
		panic(err.Error())
	}

	for {
		fmt.Println(mcp.Measure(channel))
		time.Sleep(time.Second)
	}

}
```