package helloworld

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.ugent.be/unigornel/integration_tests/tests"
)

const category = "console"

var SimpleTest = tests.SimpleTest{
	Name:        "hello_world",
	Category:    category,
	Package:     tests.SimpleTestPackage("helloworld", "simple"),
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

var SleepAndTimeTest = tests.SimpleTest{
	Name:       "sleep_and_time",
	Category:   category,
	Package:    tests.SimpleTestPackage("helloworld", "sleep_and_time"),
	Memory:     256,
	Timeout:    2 * time.Second,
	CanTimeout: true,
	CheckRun: func(out string) error {
		sleepInterval := int64(100e6)
		minTime := int64(1451606400000000000) // 2016-1-1 0:0:0.0 UTC

		r := regexp.MustCompile("^(\\d+) \\[.*\\] Hello World!")
		lines := strings.Split(out, "\n")

		minMatches := int(1 * time.Second / (time.Duration(sleepInterval) * time.Nanosecond))
		numMatches := 0

		prev := int64(0)
		for _, l := range lines {
			matches := r.FindStringSubmatch(l)
			if matches == nil {
				continue
			}
			numMatches++

			t, _ := strconv.ParseInt(matches[1], 10, 64)
			if t < minTime {
				return fmt.Errorf("time must be after %d\n", minTime)
			}

			if t-prev < sleepInterval {
				return fmt.Errorf("minimum sleep interval is %d ns\n", sleepInterval)
			}
			prev = t
		}

		if numMatches < minMatches {
			return fmt.Errorf("minimum number of hello worlds is %d\n", minMatches)
		}

		return nil
	},
}

var ReadFromConsoleTest = tests.SimpleTest{
	Name:        "read_from_console",
	Category:    category,
	Package:     tests.SimpleTestPackage("helloworld", "read_from_console"),
	Memory:      256,
	Timeout:     10 * time.Second,
	CanCrash:    true,
	CanShutdown: true,
	Stdin:       []byte("Unigornel\n"),
	CheckRun: func(out string) error {
		if !strings.Contains(out, "Hello, what's your name? Hello, Unigornel") {
			return fmt.Errorf("console output did not match")
		}
		return nil
	},
}
