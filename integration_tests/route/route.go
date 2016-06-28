package route

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
)

type Route4 struct {
	Destination string
	Gateway     string
	Netmask     string
	Interface   string
}

func (route Route4) IPNet() (*net.IPNet, error) {
	ip := net.ParseIP(route.Destination)
	if ip == nil {
		return nil, fmt.Errorf("destination is not a valid IPv4 address")
	}

	mask := net.ParseIP(route.Netmask)
	if mask == nil {
		return nil, fmt.Errorf("netmask is not a valid IPv4 address")
	}

	return &net.IPNet{
		IP:   ip,
		Mask: net.IPMask(mask),
	}, nil
}

func AllIPv4Routes() ([]Route4, error) {
	out, err := exec.Command("route", "-n", "-4").Output()
	if err != nil {
		return nil, err
	}
	return parseRouteN4(bytes.NewBuffer(out))
}

func parseRouteN4Line(line string) (*Route4, error) {
	r := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+.*?(\S+)$`)
	matches := r.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid route line")
	}
	return &Route4{
		Destination: matches[1],
		Gateway:     matches[2],
		Netmask:     matches[3],
		Interface:   matches[4],
	}, nil
}

func parseRouteN4(reader io.Reader) ([]Route4, error) {
	s := bufio.NewScanner(reader)

	// Discard header
	if !s.Scan() || !s.Scan() {
		return nil, s.Err()
	}

	routes := make([]Route4, 0)
	for i := 3; s.Scan(); i++ {
		line := s.Text()
		r, err := parseRouteN4Line(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", i, err)
		}
		routes = append(routes, *r)
	}
	return routes, s.Err()
}
