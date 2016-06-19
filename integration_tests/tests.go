package main

import (
	"github.com/unigornel/unigornel/integration_tests/tests"
	"github.com/unigornel/unigornel/integration_tests/tests/console"
)

var allTests = []tests.Test{
	&console.SimpleTest,
	&console.SleepAndTimeTest,
	&console.ReadFromConsoleTest,
}
