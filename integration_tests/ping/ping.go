package ping

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"
	"strconv"
)

type Response struct {
	Error   string
	ICMPSeq int
	TTL     int
	Time    string
}

func Ping(host string, args ...string) *exec.Cmd {
	argv := append(args, host)
	return exec.Command("ping", argv...)
}

func Parse(r io.Reader) ([]Response, error) {
	okRegexp := regexp.MustCompile(`icmp_seq=(\d+) ttl=(\d+) time=(.*)$`)
	errRegexp := regexp.MustCompile(`icmp_seq=(\d+) (.*)$`)
	scanner := bufio.NewScanner(r)

	responses := make([]Response, 0)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := okRegexp.FindStringSubmatch(line); matches != nil {
			seq, _ := strconv.Atoi(matches[1])
			ttl, _ := strconv.Atoi(matches[2])
			time := matches[3]
			responses = append(responses, Response{
				ICMPSeq: seq,
				TTL:     ttl,
				Time:    time,
			})

		} else if matches := errRegexp.FindStringSubmatch(line); matches != nil {
			seq, _ := strconv.Atoi(matches[1])
			err := matches[2]
			responses = append(responses, Response{
				ICMPSeq: seq,
				Error:   err,
			})
		}
	}

	return responses, scanner.Err()
}
