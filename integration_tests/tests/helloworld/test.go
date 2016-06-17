package helloworld

import (
	"fmt"
	"strings"
	"time"

	"github.ugent.be/unigornel/integration_tests/tests"
)

var Test = tests.SimpleTest{
	Name:        "hello_world",
	Package:     tests.SimpleTestPackage("helloworld", "main"),
	Memory:      256,
	Timeout:     10 * time.Second,
	CanCrash:    true,
	CanShutdown: true,
	CheckRun: func(out string) error {
		if !strings.Contains(out, "Hello World!") {
			return fmt.Errorf("'Hello World!' substring not in output.")
		}
		return nil
	},
}
