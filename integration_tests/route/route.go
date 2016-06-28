package route

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

type Route struct {
	Destination string
	Gateway     string
	Netmask     string
	Interface   string
}

func parseRouteNLine(line string) (*Route, error) {
	r := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+.*?(\S+)$`)
	matches := r.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid route line")
	}
	return &Route{
		Destination: matches[1],
		Gateway:     matches[2],
		Netmask:     matches[3],
		Interface:   matches[4],
	}, nil
}

func parseRouteN(reader io.Reader) ([]Route, error) {
	s := bufio.NewScanner(reader)

	// Discard header
	if !s.Scan() || !s.Scan() {
		return nil, s.Err()
	}

	routes := make([]Route, 0)
	for i := 3; s.Scan(); i++ {
		line := s.Text()
		r, err := parseRouteNLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", i, err)
		}
		routes = append(routes, *r)
	}
	return routes, s.Err()
}
