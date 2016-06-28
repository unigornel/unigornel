package ip

import (
	"bytes"
	"net"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDevLine(t *testing.T) {
	cases := []struct {
		Input       string
		Result      string
		ErrorRegexp *regexp.Regexp
	}{
		{validDev, "lo", nil},
		{invalidDev, "", regexp.MustCompile("invalid.*decl")},
	}

	for i, c := range cases {
		result, err := parseDevLine(c.Input)
		if c.ErrorRegexp != nil && err == nil {
			assert.NotNil(t, err, "for case %d", i)
		} else if c.ErrorRegexp == nil && err != nil {
			assert.Nil(t, err, "for case %d", i)
		} else if c.ErrorRegexp != nil && err != nil {
			assert.True(t, c.ErrorRegexp.MatchString(err.Error()), "for case %d", i)
		}

		assert.True(
			t, reflect.DeepEqual(c.Result, result),
			"expected %v != %v (for case %d)",
			c.Result, result, i,
		)
	}
}

func TestParseInetLine(t *testing.T) {
	cases := []struct {
		Input       string
		Result      *Inet
		ErrorRegexp *regexp.Regexp
	}{
		{invalidInet, nil, regexp.MustCompile("invalid")},
		{invalidCIDR, nil, regexp.MustCompile("invalid.*cidr")},
		{validInet6, nil, regexp.MustCompile("invalid")},
		{
			Input: validInet,
			Result: &Inet{
				IP: net.ParseIP("127.0.0.1"),
				Net: &net.IPNet{
					IP:   net.ParseIP("127.0.0.0"),
					Mask: net.CIDRMask(8, 32),
				},
				Scope:     "host",
				Interface: "lo",
			},
			ErrorRegexp: nil,
		},
		{
			Input: validInetWithBrd,
			Result: &Inet{
				IP: net.ParseIP("10.0.2.15"),
				Net: &net.IPNet{
					IP:   net.ParseIP("10.0.2.0"),
					Mask: net.CIDRMask(24, 32),
				},
				Scope:     "global",
				Interface: "enp0s3",
			},
			ErrorRegexp: nil,
		},
	}

	for i, c := range cases {
		result, err := parseInetLine(c.Input)
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

		if c.Result != nil {
			assert.True(
				t, ipEqual(c.Result.IP, result.IP),
				"expected %v != %v (for case %d)",
				c.Result.IP, result.IP, i,
			)
			assert.True(
				t, ipEqual(c.Result.Net.IP, result.Net.IP),
				"expected %v != %v (for case %d)",
				c.Result.Net.IP, result.Net.IP, i,
			)

			result.IP = c.Result.IP
			result.Net.IP = c.Result.Net.IP
		} else {
			assert.Nil(t, result, "(for case %d)")
		}

		assert.True(
			t, reflect.DeepEqual(c.Result, result),
			"expected %v != %v (for case %d)",
			c.Result, result, i,
		)
	}
}

func TestParseAddrShow(t *testing.T) {
	cases := []struct {
		Input       string
		ErrorRegexp *regexp.Regexp
	}{
		{invalidAddrShowNoDecl, regexp.MustCompile("no.*decl")},
		{invalidAddrShowInetDecl, regexp.MustCompile(".*")},
		{invalidAddrShowIfaceDecl, regexp.MustCompile(".*")},
		{validAddrShow, nil},
	}

	for i, c := range cases {
		_, err := parseAddrShow(bytes.NewBuffer([]byte(c.Input)))
		if c.ErrorRegexp != nil && err == nil {
			assert.NotNil(t, err, "for case %d", i)
		} else if c.ErrorRegexp == nil && err != nil {
			assert.Nil(t, err, "for case %d", i)
		} else if c.ErrorRegexp != nil && err != nil {
			assert.True(t, c.ErrorRegexp.MatchString(err.Error()), "for case %d", i)
		}
	}
}

func ipEqual(a, b net.IP) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b != nil {
		return a.Equal(b)
	} else {
		return false
	}
}

var validDev = "1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1"
var invalidDev = "not a number: lo:"

var validInet = "\tinet 127.0.0.1/8 scope host lo"
var validInetWithBrd = "\tinet 10.0.2.15/24 brd 10.0.2.255 scope global enp0s3"
var invalidCIDR = "\tinet 127.0.0.1/-4 scope host lo"
var invalidInet = "\tlink/ether 08:00:27:46:86:26 brd ff:ff:ff:ff:ff:ff"

var validInet6 = "\tinet6 ::1/128 scope host "

var validAddrShow = `1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: enp0s3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 08:00:27:46:86:26 brd ff:ff:ff:ff:ff:ff
    inet 10.0.2.15/24 brd 10.0.2.255 scope global enp0s3
       valid_lft forever preferred_lft forever
    inet6 fe80::a00:27ff:fe46:8626/64 scope link 
       valid_lft forever preferred_lft forever
3: enp0s8: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 08:00:27:ed:ec:08 brd ff:ff:ff:ff:ff:ff
4: enp0s9: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 08:00:27:ea:ef:62 brd ff:ff:ff:ff:ff:ff
5: lxdbr0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ether 82:bd:91:e4:81:6e brd ff:ff:ff:ff:ff:ff
    inet 10.87.130.1/24 scope global lxdbr0
       valid_lft forever preferred_lft forever
    inet6 fe80::80bd:91ff:fee4:816e/64 scope link 
       valid_lft forever preferred_lft forever
`

var invalidAddrShowIfaceDecl = "not a valid number: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1\n"
var invalidAddrShowInetDecl = `2: enp0s3: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    inet no-ip/24 brd 10.0.2.255 scope global enp0s3
`
var invalidAddrShowNoDecl = " inet no-ip/24 brd 10.0.2.255 scope global enp0s3\n"
