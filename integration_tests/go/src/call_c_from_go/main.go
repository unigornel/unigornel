package main

// #cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #include <mini-os/experimental.h>
import "C"
import "fmt"

//go:cgo_import_static test
//go:linkname test test
var test byte

func main() {}

//export Main
func Main(unused int) {
	C.test()

	fmt.Println("Hello World!")
}
