package main

// #cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #include <mini-os/experimental.h>
import "C"
import "fmt"

func main() {}

//export Main
func Main(unused int) {
	C.test()

	fmt.Println("Hello World!")
}
