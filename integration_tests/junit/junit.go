package junit

import "encoding/xml"

type Report struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []TestSuite
}

type TestSuite struct {
	XMLName    xml.Name   `xml:"testsuite"`
	Tests      int        `xml:"tests,attr"`
	Failures   int        `xml:"failures,attr"`
	Time       string     `xml:"time,attr"`
	Name       string     `xml:"name,attr"`
	Properties []Property `xml:"properties>property,omitempty"`
	TestCases  []TestCase
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	ClassName string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Time      string   `xml:"time,attr"`
	Output    string   `xml:"system-out,omitempty"`
	Failure   *Failure `xml:"failure,omitempty"`
}

type Failure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}
