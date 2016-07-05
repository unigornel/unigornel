package main

import (
	"github.com/unigornel/unigornel/integration_tests/tests"
	"github.com/unigornel/unigornel/integration_tests/tests/console"
	"github.com/unigornel/unigornel/integration_tests/tests/network"
)

var allTests = []tests.Test{
	&console.SimpleTest,
	&console.SleepAndTimeTest,
	&console.ReadFromConsoleTest,

	&network.PingTest{},
	&network.PingAddressTest{},
}
