package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/unigornel/unigornel/integration_tests/junit"
)

type Test interface {
	GetName() string
	GetCategory() string
	GetInfo() string
	Build(io.Writer) error
	Setup(io.Writer) error
	Run(io.Writer) error
	Check(io.Writer) error
	Clean(w io.Writer, success bool) error
}

type Result struct {
	Test     Test
	Error    error
	Output   string
	Duration time.Duration
}

func JUnit(r Result) junit.TestCase {
	tc := junit.TestCase{
		ClassName: r.Test.GetCategory() + "." + r.Test.GetName(),
		Name:      r.Test.GetName(),
		Output:    r.Output,
		Time:      fmt.Sprintf("%f", r.Duration.Seconds()),
	}
	if r.Error != nil {
		tc.Failure = &junit.Failure{
			Message:  "integration test failure",
			Contents: r.Error.Error(),
		}
	}
	return tc
}

func Run(t Test) (result Result) {
	start := time.Now()
	result.Test = t
	defer func() {
		result.Duration = time.Now().Sub(start)
		fmt.Println("[+] test ended in", result.Duration)
	}()
	defer func() {
		if result.Error == nil {
			fmt.Println("[+] successfully ran test")
		} else {
			fmt.Println("[-] test error:", result.Error)
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("panic occurred while running test: %v", r)
			}
			result.Error = err
		}
	}()

	buffer := bytes.NewBuffer(nil)
	w := io.MultiWriter(os.Stdout, buffer)
	defer func() {
		result.Output = string(buffer.Bytes())
	}()

	result.Error = t.Build(w)
	if result.Error != nil {
		return
	}

	result.Error = t.Setup(w)
	if result.Error != nil {
		return
	}

	result.Error = t.Run(w)
	if result.Error != nil {
		t.Clean(buffer, false)
		return
	}

	result.Error = t.Check(w)
	if result.Error != nil {
		t.Clean(buffer, true)
		return
	}

	result.Error = t.Clean(w, true)
	return
}
