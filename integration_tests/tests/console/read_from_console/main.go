package main

import "C"
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	Main(0)
}

//export Main
func Main(unused int) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Hello, what's your name? ")
	name, _ := reader.ReadString('\n')
	fmt.Printf("Hello, %s\n", strings.TrimSpace(name))
}
