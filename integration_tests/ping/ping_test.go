package ping

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	r, err := Parse(bytes.NewBuffer([]byte(pingOut)))
	require.Nil(t, err)

	expected := []Response{
		{ICMPSeq: 1, TTL: 64, Time: "0.047 ms"},
		{ICMPSeq: 2, TTL: 64, Time: "0.047 ms"},
		{ICMPSeq: 3, Error: "Destination Host Unreachable"},
		{ICMPSeq: 4, TTL: 64, Time: "0.055 ms"},
	}
	assert.True(
		t, reflect.DeepEqual(r, expected),
		"expected %v != %v",
		expected, r,
	)
}

var pingOut = `PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.
64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.047 ms
64 bytes from 127.0.0.1: icmp_seq=2 ttl=64 time=0.047 ms
From 127.0.0.1 icmp_seq=3 Destination Host Unreachable
64 bytes from 127.0.0.1: icmp_seq=4 ttl=64 time=0.055 ms

--- 127.0.0.1 ping statistics ---
10 packets transmitted, 10 received, 0% packet loss, time 4499ms
rtt min/avg/max/mdev = 0.044/0.057/0.072/0.010 ms
`
