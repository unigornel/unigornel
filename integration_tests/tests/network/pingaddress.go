package network

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/unigornel/unigornel/integration_tests/brctl"
	"github.com/unigornel/unigornel/integration_tests/ifconfig"
	"github.com/unigornel/unigornel/integration_tests/ip"
	"github.com/unigornel/unigornel/integration_tests/tests"
	"github.com/unigornel/unigornel/integration_tests/xen"
)

type PingAddressTest struct {
	PingTest
	output string
}

func (t *PingAddressTest) GetName() string {
	return "ping_address"
}

func (t *PingAddressTest) GetCategory() string {
	return t.PingTest.GetCategory()
}

func (t *PingAddressTest) GetInfo() string {
	return t.PingTest.GetInfo()
}

func (t *PingAddressTest) Build(w io.Writer) error {
	if err := t.PingTest.setupNetwork(w); err != nil {
		return err
	}

	p := tests.SimpleTestPackage("network", "ping_address")

	if err := tests.GoGet(w, p); err != nil {
		return err
	}

	if err := tests.UpdateLibs(w); err != nil {
		return err
	}

	addr := t.network.unikernelIP.String()
	netmask := net.IP(t.network.netmask).To4().String()
	dest := t.network.xenIP.String()
	ldflags := fmt.Sprintf(
		"-X main.ipAddress=%s -X main.ipNetmask=%s -X main.ipDestination=%s",
		addr, netmask, dest,
	)
	file, err := tests.Build(w, "ping", p, "--ldflags", ldflags)
	t.unikernel = file
	return err
}

func (t *PingAddressTest) Setup(w io.Writer) error {
	// Setup the bridge.
	fmt.Fprintln(w, "[+] creating a bridge")
	bridge, err := brctl.CreateNumbered("unigornel")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, "[+] using bridge", bridge)
	t.bridge = bridge

	fmt.Fprintln(w, "[+] using ip address", t.network.xenIP, "netmask", t.network.netmask)
	if err := ifconfig.SetIP(bridge, t.network.xenIP, t.network.netmask); err != nil {
		return err
	}

	// Create the unikernel
	kernel := xen.Kernel{
		Binary:  t.unikernel,
		Memory:  256,
		Name:    t.GetName(),
		OnCrash: xen.OnCrashPreserve,
		VIF:     fmt.Sprintf("['bridge=%s']", bridge),
	}

	fmt.Fprintln(w, "[+] creating paused kernel")
	dom, err := kernel.CreatePausedUniqueName(func(cmd *exec.Cmd) {
		cmd.Stdout = w
		cmd.Stderr = w
	})
	fmt.Fprintln(w, "[+] domain created:", dom)
	t.domain = dom
	return err
}

func (t *PingAddressTest) Run(w io.Writer) error {
	out := bytes.NewBuffer(nil)
	console := t.domain.Console()
	console.Stdout = io.MultiWriter(w, out)
	console.Stderr = console.Stdout
	_, err := console.StdinPipe()
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "[+] attaching to the console")
	if err := console.Start(); err != nil {
		return err
	}

	done := make(chan struct{})
	timeout := make(chan struct{})

	go func() {
		console.Wait()
		close(done)
	}()
	go func() {
		time.Sleep(5 * time.Second)
		close(timeout)
	}()

	fmt.Fprintln(w, "[+] unpausing unikernel domain")
	if err := t.domain.Unpause().Run(); err != nil {
		console.Process.Kill()
		return err
	}

	select {
	case <-done:
		return fmt.Errorf("console exited unexpectedly")
	case <-timeout:
		console.Process.Kill()
	}

	t.output = string(out.Bytes())

	return nil
}

func (t *PingAddressTest) Check(w io.Writer) error {
	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(t.output)))

	replyRegex := regexp.MustCompile("got.*reply.*(\\d+)")

	matchIndex := 1
	for scanner.Scan() {
		matches := replyRegex.FindStringSubmatch(scanner.Text())
		if matches == nil {
			continue
		}

		seq, _ := strconv.Atoi(matches[1])
		if seq != matchIndex {
			return fmt.Errorf("expected sequence %d, got %d", matchIndex, seq)
		}
		matchIndex++
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	minMatches := 5
	if matchIndex < minMatches {
		return fmt.Errorf("expected at least %d matches", minMatches)
	}
	return nil
}

func (t *PingAddressTest) Clean(w io.Writer, success bool) (err error) {
	if t.unikernel != "" {
		fmt.Fprintf(w, "[+] removing %s\n", t.unikernel)
		os.Remove(t.unikernel)
	}

	if t.domain != nil {
		cmd := t.domain.Destroy()
		cmd.Stderr = w
		cmd.Stdout = w
		err = cmd.Run()
	}

	if t.bridge != "" {
		fmt.Fprintf(w, "[+] removing bridge %s\n", t.bridge)
		err = ip.Down(t.bridge)
		if err == nil {
			err = brctl.Delete(t.bridge)
		}
	}

	return
}
