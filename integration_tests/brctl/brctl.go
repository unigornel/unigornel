package brctl

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Bridge struct {
	Name       string
	ID         string
	STP        bool
	Interfaces []string
}

func parseBrctlShowDeclLine(line string) (*Bridge, error) {
	line = strings.TrimSpace(line)
	parts := regexp.MustCompile("\\s+").Split(line, -1)
	if len(parts) != 4 && len(parts) != 3 {
		return nil, fmt.Errorf("not a valid brctl bridge declaration: '%s'", line)
	}

	b := Bridge{
		Name: parts[0],
		ID:   parts[1],
		STP:  false,
	}
	if parts[2] == "yes" {
		b.STP = true
	}
	if len(parts) == 4 {
		b.Interfaces = []string{parts[3]}
	}
	return &b, nil
}

func parseBrctlShowIntefaceLine(line string) string {
	return strings.TrimSpace(line)
}

func parseBrctlShow(reader io.Reader) ([]Bridge, error) {
	whitespace := regexp.MustCompile("^\\s+.*$")
	s := bufio.NewScanner(reader)

	// Discard header
	if !s.Scan() {
		return nil, s.Err()
	}

	// Parse bridges
	var current *Bridge
	bridges := make([]Bridge, 0)
	for i := 2; s.Scan(); i++ {
		line := s.Text()
		if len(line) == 0 {
			return nil, fmt.Errorf("line %d: unexpected empty line", i)
		}

		if whitespace.MatchString(line) {
			if current == nil {
				return nil, fmt.Errorf("line %d: no bridge declaration before interface spec", i)
			}
			current.Interfaces = append(current.Interfaces, parseBrctlShowIntefaceLine(line))
		} else {
			if current != nil {
				bridges = append(bridges, *current)
			}
			b, err := parseBrctlShowDeclLine(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i, err)
			}
			current = b
		}
	}

	if current != nil {
		bridges = append(bridges, *current)
	}

	return bridges, s.Err()
}
