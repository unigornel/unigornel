package brctl

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBrctlShow(t *testing.T) {
	cases := []struct {
		Input       string
		ErrorRegexp *regexp.Regexp
		Bridges     []Bridge
	}{
		{
			Input:       brctl1,
			ErrorRegexp: nil,
			Bridges: []Bridge{
				{"lan", "8000.00215ec286be", false, []string{"eth0", "vif1.0", "vif2.1", "vif3.0", "vif7.0"}},
				{"lxdbr0", "8000.000000000000", true, nil},
				{"wan", "8000.000e0ccf5e7c", false, []string{"eth1", "vif1.1", "vif2.0"}},
			},
		},
		{
			Input:       brctl2,
			ErrorRegexp: nil,
			Bridges: []Bridge{
				{"lxdbr0", "8000.000000000000", false, nil},
			},
		},
		{
			Input:       brctl3,
			ErrorRegexp: nil,
			Bridges:     []Bridge{},
		},
		{
			Input:       brctl4,
			ErrorRegexp: regexp.MustCompile("empty line"),
			Bridges:     nil,
		},
		{
			Input:       brctl5,
			ErrorRegexp: regexp.MustCompile("not.*valid.*bridge.*decl"),
			Bridges:     nil,
		},
		{
			Input:       brctl6,
			ErrorRegexp: regexp.MustCompile("no.*decl.*before"),
			Bridges:     nil,
		},
	}

	for i, c := range cases {
		bridges, err := parseBrctlShow(bytes.NewBuffer([]byte(c.Input)))
		if c.ErrorRegexp == nil && err != nil {
			assert.Nil(t, err, "for test %d", i)
		} else if c.ErrorRegexp != nil && err == nil {
			assert.NotNil(t, err, "for test %d", i)
		} else if c.ErrorRegexp != nil && err != nil {
			assert.True(t, c.ErrorRegexp.MatchString(err.Error()), "for test %d", i)
		}

		assert.True(
			t, reflect.DeepEqual(bridges, c.Bridges),
			"expected %v != %v (for test %d)",
			c.Bridges, bridges, i,
		)
	}
}

var brctl1 = `bridge name	bridge id		STP enabled	interfaces
lan		8000.00215ec286be	no		eth0
							vif1.0
							vif2.1
							vif3.0
							vif7.0
lxdbr0		8000.000000000000	yes		
wan		8000.000e0ccf5e7c	no		eth1
							vif1.1
							vif2.0
`

var brctl2 = `bridge name	bridge id		STP enabled	interfaces
lxdbr0		8000.000000000000	no		
`

var brctl3 = "bridge name	bridge id		STP enabled	interfaces\n"

var brctl4 = `bridge name	bridge id		STP enabled	interfaces

lxdbr0		8000.000000000000	no		
`

var brctl5 = `bridge name	bridge id		STP enabled	interfaces
lxdbr0		
`

var brctl6 = `bridge name	bridge id		STP enabled	interfaces
							vif1.0
`
