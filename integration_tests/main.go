package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/unigornel/unigornel/integration_tests/junit"
	"github.com/unigornel/unigornel/integration_tests/tests"
)

type options struct {
	ShowHelp  bool
	ListTests bool
	TestName  string
	JUnit     string
}

func main() {
	o := parseOptions()

	if o.ListTests {
		listTests()
		return
	}

	var ts []tests.Test
	if o.TestName != "" {
		t := testWithName(allTests, o.TestName)
		if t == nil {
			log.Fatalf("error: no test with name '%s'\n", o.TestName)
		}
		ts = []tests.Test{t}
	} else {
		ts = allTests
	}

	log.Printf("Running %d tests\n", len(ts))
	results := runTests(ts)
	report, failures := reportFromResults(results)
	log.Printf("Ran %d tests with %d failures\n", len(ts), failures)

	if o.JUnit != "" {
		log.Printf("Writing junit report to %s\n", o.JUnit)
		fh, err := os.Create(o.JUnit)
		if err != nil {
			log.Fatalln(err)
		}
		defer fh.Close()
		enc := xml.NewEncoder(fh)
		enc.Indent("", "  ")
		if err := enc.Encode(report); err != nil {
			log.Fatalln(err)
		}
	}
}

func listTests() {
	for _, test := range allTests {
		fmt.Println(test.GetName())
	}
}

func testWithName(ts []tests.Test, name string) tests.Test {
	for _, test := range ts {
		if test.GetName() == name {
			return test
		}
	}
	return nil
}

func runTests(ts []tests.Test) []tests.Result {
	results := make([]tests.Result, len(ts))
	for i, test := range ts {
		fmt.Printf("Running test %s (%d/%d)\n", test.GetName(), i, len(ts))
		results[i] = tests.Run(test)
	}
	return results
}

func reportFromResults(rs []tests.Result) (junit.Report, int) {
	failures := 0
	for _, r := range rs {
		if r.Error != nil {
			failures++
		}
	}

	ts := junit.TestSuite{
		Tests:    len(rs),
		Failures: failures,
		Name:     "integration tests",
	}

	for _, r := range rs {
		ts.TestCases = append(ts.TestCases, tests.JUnit(r))
	}

	return junit.Report{
		Suites: []junit.TestSuite{ts},
	}, failures
}

func parseOptions() options {
	var o options

	flag.BoolVar(&o.ShowHelp, "help", false, "show this help")
	flag.BoolVar(&o.ListTests, "list", false, "list all known tests and exit")
	flag.StringVar(&o.TestName, "test", "", "run a specific test")
	flag.StringVar(&o.JUnit, "junit", "", "write a JUnit report to the specified file")

	flag.Parse()

	printHelp := func() {
		fmt.Println("test [options]")
		flag.PrintDefaults()
	}

	if o.ShowHelp {
		printHelp()
		os.Exit(0)
	}

	return o
}
