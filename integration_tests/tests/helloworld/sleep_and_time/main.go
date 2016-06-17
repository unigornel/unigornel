package main

import "C"
import (
	"fmt"
	"time"
)

func main() {}

//export Main
func Main(unused int) {
	for {
		t := time.Now()
		fmt.Printf("%v [%v] Hello World!\n", t.UnixNano(), t.UTC())
		time.Sleep(100 * time.Millisecond)
	}
}
