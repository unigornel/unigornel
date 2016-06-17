package main

import (
	"github.ugent.be/unigornel/integration_tests/tests"
	"github.ugent.be/unigornel/integration_tests/tests/console"
)

var allTests = []tests.Test{
	&console.SimpleTest,
	&console.SleepAndTimeTest,
	&console.ReadFromConsoleTest,
}
