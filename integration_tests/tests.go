package main

import (
	"github.ugent.be/unigornel/integration_tests/tests"
	"github.ugent.be/unigornel/integration_tests/tests/helloworld"
)

var allTests = []tests.Test{
	&helloworld.Test,
}
