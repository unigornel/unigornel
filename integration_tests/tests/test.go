package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.ugent.be/unigornel/integration_tests/junit"
)

type Test interface {
	GetName() string
	Build(io.Writer) error
	Setup(io.Writer) error
	Run(io.Writer) error
	Check(io.Writer) error
	Clean(w io.Writer, success bool) error
}

type Result struct {
	Test   Test
	Error  error
	Output string
}

func JUnit(r Result) junit.TestCase {
	tc := junit.TestCase{
		ClassName: r.Test.GetName(),
		Name:      r.Test.GetName(),
		Output:    r.Output,
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
	result.Test = t
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
