package ip

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
)

type Inet struct {
	IP        net.IP
	Net       *net.IPNet
	Scope     string
	Interface string
}

type Addr struct {
	Interface string
	Inet      *Inet
}

func parseDevLine(line string) (string, error) {
	r := regexp.MustCompile("^\\d+: (\\S+):")
	matches := r.FindStringSubmatch(line)
	if matches == nil {
		return "", fmt.Errorf("invalid device declaration")
	}
	return matches[1], nil
}

func parseInetLine(line string) (*Inet, error) {
	r := regexp.MustCompile("^\\s+inet (\\S+) (?:brd \\S+ )?scope (\\S+) (\\S+)")
	matches := r.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid inet line")
	}

	ip, net, err := net.ParseCIDR(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid cidr: %v", err)
	}

	inet := Inet{
		IP:        ip,
		Net:       net,
		Scope:     matches[2],
		Interface: matches[3],
	}

	return &inet, nil
}

func parseAddrShow(reader io.Reader) ([]Addr, error) {
	isDecl := regexp.MustCompile("^\\S")
	isInet := regexp.MustCompile("^\\s+inet ")
	s := bufio.NewScanner(reader)

	var current *Addr
	addrs := make([]Addr, 0)
	for i := 1; s.Scan(); i++ {
		line := s.Text()

		if isInet.MatchString(line) {
			if current == nil {
				return nil, fmt.Errorf("line %d: no device declaration before inet", i)
			}
			inet, err := parseInetLine(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i, err)
			}
			current.Inet = inet

		} else if isDecl.MatchString(line) {
			if current != nil {
				addrs = append(addrs, *current)
			}
			iface, err := parseDevLine(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i, err)
			}
			current = &Addr{Interface: iface}
		}
	}

	if current != nil {
		addrs = append(addrs, *current)
	}

	return addrs, nil
}
