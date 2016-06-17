package xen

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	OnCrashPreserve = "preserve"
)

type Kernel struct {
	Binary  string
	Memory  int
	Name    string
	OnCrash string
}

func (kernel *Kernel) CreatePausedUniqueName(f func(cmd *exec.Cmd)) (*Domain, error) {
	fh, err := ioutil.TempFile("", "kernel-"+kernel.Name+"-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(fh.Name())
	defer fh.Close()

	kernel.Name = path.Base(fh.Name())
	kernel.WriteConfiguration(fh)

	cmd := Create(true, fh.Name())
	if f != nil {
		f(cmd)
	}

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	dom, err := DomainWithName(kernel.Name)
	if err != nil {
		return nil, err
	} else if dom == nil {
		return nil, fmt.Errorf("could not create the paused kernel: %s", kernel)
	}
	return dom, err
}

func (k Kernel) WriteConfiguration(w io.Writer) {
	fmt.Fprintf(w, "kernel = \"%s\"\n", k.Binary)
	fmt.Fprintf(w, "memory = %d\n", k.Memory)
	fmt.Fprintf(w, "name = \"%s\"\n", k.Name)
	fmt.Fprintf(w, "on_crash = \"%s\"\n", k.OnCrash)
}

type DomainState int

const (
	DomainStateUnknown  DomainState = 0x00
	DomainStateRunning  DomainState = 0x01
	DomainStateBlocked  DomainState = 0x02
	DomainStatePaused   DomainState = 0x04
	DomainStateShutdown DomainState = 0x08
	DomainStateCrashed  DomainState = 0x10
	DomainStateDying    DomainState = 0x20
)

func (state DomainState) Check(mask DomainState) bool {
	return (state & mask) == mask
}

type Domain struct {
	ID     int
	Name   string
	Memory int
	VCPUs  int
	State  DomainState
	Time   float64
}

func (domain Domain) Update() (*Domain, error) {
	return DomainWithID(domain.ID)
}

func (domain Domain) Destroy() *exec.Cmd {
	return Destroy(domain.ID)
}

func (domain Domain) Unpause() *exec.Cmd {
	return Unpause(domain.ID)
}

func (domain Domain) Console() *exec.Cmd {
	return Console(domain.ID)
}

func ListDomains() ([]Domain, error) {
	out, err := List().Output()
	if err != nil {
		return nil, err
	}

	domains := make([]Domain, 0)
	for _, line := range strings.Split(string(out), "\n")[1:] {
		if line != "" {
			domain, err := domainFromListLine(line)
			if err != nil {
				return nil, err
			}
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

func DomainWithID(id int) (*Domain, error) {
	return DomainWith(func(domain Domain) bool {
		return domain.ID == id
	})
}

func DomainWithName(name string) (*Domain, error) {
	return DomainWith(func(domain Domain) bool {
		return domain.Name == name
	})
}

func DomainWith(f func(Domain) bool) (*Domain, error) {
	domains, err := ListDomains()
	if err != nil {
		return nil, err
	}

	for _, dom := range domains {
		if f(dom) {
			return &dom, nil
		}
	}
	return nil, nil
}

func domainFromListLine(str string) (domain Domain, err error) {
	s := regexp.MustCompile("\\s+").Split(str, -1)
	if len(s) != 6 {
		err = fmt.Errorf("could not parse xl list output: invalid line: %s", str)
		return
	}

	name := s[0]
	id := s[1]
	mem := s[2]
	vcpus := s[3]
	state := s[4]
	time := s[5]

	domain.Name = name

	domain.ID, err = strconv.Atoi(id)
	if err != nil {
		return
	}

	domain.Memory, err = strconv.Atoi(mem)
	if err != nil {
		return
	}

	domain.VCPUs, err = strconv.Atoi(vcpus)
	if err != nil {
		return
	}

	domain.State, err = parseDomainState(state)
	if err != nil {
		return
	}

	domain.Time, err = strconv.ParseFloat(time, 64)
	return
}

func parseDomainState(str string) (DomainState, error) {
	var mapping = []struct {
		Symbol byte
		State  DomainState
	}{
		{'r', DomainStateRunning},
		{'b', DomainStateBlocked},
		{'p', DomainStatePaused},
		{'s', DomainStateShutdown},
		{'c', DomainStateCrashed},
		{'d', DomainStateDying},
	}

	var state DomainState
	if len(str) != len(mapping) {
		return state, fmt.Errorf("domain state string '%s' should have length 6", str)
	}

	for i, m := range mapping {
		if str[i] == m.Symbol {
			state |= m.State
		}
	}

	return state, nil
}
