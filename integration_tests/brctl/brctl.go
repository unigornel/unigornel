package brctl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

type Bridge struct {
	Name       string
	ID         string
	STP        bool
	Interfaces []string
}

func Show() ([]Bridge, error) {
	out, err := exec.Command("brctl", "show").Output()
	if err != nil {
		return nil, err
	}

	return parseBrctlShow(bytes.NewBuffer(out))
}

func Create(name string) error {
	out, err := exec.Command("brctl", "addbr", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: brctl addbr: %v", strings.TrimSpace(string(out)))
	}
	return nil
}

func Delete(name string) error {
	out, err := exec.Command("brctl", "delbr", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: brctl delbr: %v", strings.TrimSpace(string(out)))
	}
	return nil
}

func CreateNumbered(prefix string) (string, error) {
	var err error
	var name string

	for i := 0; i < 5; i++ {
		bridges, err := Show()
		if err != nil {
			return "", err
		}

		next := 0
		for {
			name = fmt.Sprintf("%s%d", prefix, next)

			found := false
			for _, b := range bridges {
				if b.Name == name {
					found = true
					break
				}
			}

			if !found {
				break
			}

			next++
		}

		name = fmt.Sprintf("%s%d", prefix, next)
		err = Create(name)
		if err == nil {
			break
		}
	}

	return name, err
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
