package route

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRouteNLine(t *testing.T) {
	cases := []struct {
		Input       string
		Result      *Route4
		ErrorRegexp *regexp.Regexp
	}{
		{invalidRouteLine, nil, regexp.MustCompile("invalid")},
		{
			Input:       validRouteLine,
			Result:      &Route4{"0.0.0.0", "10.0.2.2", "0.0.0.0", "enp0s3"},
			ErrorRegexp: nil,
		},
	}

	for i, c := range cases {
		result, err := parseRouteN4Line(c.Input)
		if c.ErrorRegexp != nil && err == nil {
			assert.NotNil(t, err, "for case %d", i)
		} else if c.ErrorRegexp == nil && err != nil {
			assert.Nil(t, err, "for case %d", i)
		} else if c.ErrorRegexp != nil && err != nil {
			assert.True(
				t, c.ErrorRegexp.MatchString(err.Error()),
				"unexpected error: %v (for case %d)", err, i,
			)
		}

		assert.True(
			t, reflect.DeepEqual(c.Result, result),
			"expected %v != %v (for case %d)",
			c.Result, result, i,
		)
	}
}

func TestParseRouteN(t *testing.T) {
	cases := []struct {
		Input       string
		Result      []Route4
		ErrorRegexp *regexp.Regexp
	}{
		{invalidRoute, nil, regexp.MustCompile("invalid")},
		{
			Input:       validRoute,
			ErrorRegexp: nil,
			Result: []Route4{
				{"0.0.0.0", "10.0.2.2", "0.0.0.0", "enp0s3"},
				{"10.0.2.0", "0.0.0.0", "255.255.255.0", "enp0s3"},
				{"10.0.200.0", "0.0.0.0", "255.255.255.0", "test0"},
				{"10.87.130.0", "0.0.0.0", "255.255.255.0", "lxdbr0"},
			},
		},
	}

	for i, c := range cases {
		result, err := parseRouteN4(bytes.NewBuffer([]byte(c.Input)))
		if c.ErrorRegexp != nil && err == nil {
			assert.NotNil(t, err, "for case %d", i)
		} else if c.ErrorRegexp == nil && err != nil {
			assert.Nil(t, err, "for case %d", i)
		} else if c.ErrorRegexp != nil && err != nil {
			assert.True(
				t, c.ErrorRegexp.MatchString(err.Error()),
				"unexpected error: %v (for case %d)", err, i,
			)
		}

		assert.True(
			t, reflect.DeepEqual(c.Result, result),
			"expected %v != %v (for case %d)",
			c.Result, result, i,
		)
	}
}

var validRouteLine = "0.0.0.0         10.0.2.2        0.0.0.0         UG    0      0        0 enp0s3"
var invalidRouteLine = "invalid"

var validRoute = `Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         10.0.2.2        0.0.0.0         UG    0      0        0 enp0s3
10.0.2.0        0.0.0.0         255.255.255.0   U     0      0        0 enp0s3
10.0.200.0      0.0.0.0         255.255.255.0   U     0      0        0 test0
10.87.130.0     0.0.0.0         255.255.255.0   U     0      0        0 lxdbr0
`

var invalidRoute = `Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface

10.0.2.0        0.0.0.0         255.255.255.0   U     0      0        0 enp0s3
10.0.200.0      0.0.0.0         255.255.255.0   U     0      0        0 test0
10.87.130.0     0.0.0.0         255.255.255.0   U     0      0        0 lxdbr0
`
