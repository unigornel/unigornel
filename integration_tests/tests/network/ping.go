package network

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/unigornel/unigornel/integration_tests/brctl"
	"github.com/unigornel/unigornel/integration_tests/ifconfig"
	"github.com/unigornel/unigornel/integration_tests/ip"
	"github.com/unigornel/unigornel/integration_tests/ping"
	"github.com/unigornel/unigornel/integration_tests/tests"
	"github.com/unigornel/unigornel/integration_tests/xen"
)

type PingTest struct {
	unikernel string
	domain    *xen.Domain
	bridge    string
	network   struct {
		xenIP       net.IP
		unikernelIP net.IP
		netmask     net.IPMask
	}
	responses []ping.Response
}

func (t *PingTest) GetName() string {
	return "ping"
}

func (t *PingTest) GetCategory() string {
	return "network"
}

func (t *PingTest) GetInfo() string {
	return `ENVIRONMENT:
    TEST_PING_NETWORK
	    The network in CIDR style to use (default: 10.123.123.0/30).
		The network must have room for at least 2 hosts (<=/30).
`
}

func (t *PingTest) Build(w io.Writer) error {
	if err := t.setupNetwork(w); err != nil {
		return err
	}

	p := tests.SimpleTestPackage("network", "ping")

	if err := tests.GoGet(w, p); err != nil {
		return err
	}

	if err := tests.UpdateLibs(w); err != nil {
		return err
	}

	file, err := tests.Build(
		w, "ping", p,
		"--ldflags", "-X main.ipAddress="+t.network.unikernelIP.String(),
	)
	t.unikernel = file
	return err
}

func (t *PingTest) setupNetwork(w io.Writer) error {
	var ipnet *net.IPNet
	var err error
	if v := os.Getenv("TEST_PING_NETWORK"); v != "" {
		fmt.Fprintln(w, "[+] using network from TEST_PING_NETWORK:", v)
		_, ipnet, err = net.ParseCIDR(v)
	} else {
		v = "10.123.123.0/30"
		fmt.Fprintln(w, "[*] warning: using default network (collisions may fail the test):", v)
		_, ipnet, err = net.ParseCIDR(v)
	}

	if err != nil {
		return err
	}

	ones, _ := ipnet.Mask.Size()
	if ones > 30 {
		return fmt.Errorf("invalid mask size: %v", ones)
	}

	local := ipnet.IP.To4()
	mask := net.IP(ipnet.Mask).To4()
	if local == nil || mask == nil {
		return fmt.Errorf("network is not an IPv4 network")
	}
	local[3] += 1
	unikernel := net.IP([]byte{local[0], local[1], local[2], local[3] + 1})

	t.network.xenIP = local
	t.network.unikernelIP = unikernel
	t.network.netmask = ipnet.Mask
	return nil
}

func (t *PingTest) Setup(w io.Writer) error {
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

func (t *PingTest) Run(w io.Writer) error {
	consoleBuffer := bytes.NewBuffer(nil)
	pingBuffer := bytes.NewBuffer(nil)

	console := t.domain.Console()
	console.Stdout = io.MultiWriter(w, consoleBuffer)
	console.Stderr = io.MultiWriter(w, consoleBuffer)
	_, err := console.StdinPipe()
	if err != nil {
		return err
	}

	pingCmd := ping.Ping(t.network.unikernelIP.String(), "-c", "10", "-i", "0.5", "-W", "1")
	pingCmd.Stdout = io.MultiWriter(w, pingBuffer)
	pingCmd.Stderr = w

	fmt.Println(w, "[+] attaching to the console")
	if err := console.Start(); err != nil {
		return err
	}

	done := make(chan struct{})
	exited := make(chan error)
	timeout := make(chan struct{})
	networkReady := make(chan error)
	pingReady := make(chan error)

	waitForConsole := func() {
		console.Wait()
		close(done)
	}
	waitForTimeout := func() {
		time.Sleep(10 * time.Second)
		close(timeout)
	}
	checkDomainState := func() {
		for {
			select {
			case <-done:
				return
			case <-timeout:
				return
			case <-time.After(1 * time.Second):
				c, err := t.domain.Update()
				if err != nil {
					exited <- err
					close(exited)
					return
				}
				if c.State.Check(xen.DomainStateShutdown) || c.State.Check(xen.DomainStateCrashed) {
					close(exited)
					return
				}
			}
		}
	}
	waitForReady := func() {
		defer close(networkReady)
		fmt.Fprintln(w, "[+] waiting for unikernel network setup")
		scanner := bufio.NewScanner(consoleBuffer)
		r := regexp.MustCompile("network.*ready")
		for scanner.Scan() {
			line := scanner.Text()
			if r.MatchString(line) {
				fmt.Fprintln(w, "[+] unikernel network is ready")
				return
			}
		}

		networkReady <- scanner.Err()
	}
	waitForPing := func() {
		defer close(pingReady)

		err := <-networkReady
		if err != nil {
			pingReady <- err
			return
		}

		fmt.Fprintln(w, "[+] starting ping:", pingCmd)
		if err := pingCmd.Start(); err != nil {
			pingReady <- err
			return
		}

		if err := pingCmd.Wait(); err != nil {
			pingReady <- err
			return
		}

		responses, err := ping.Parse(pingBuffer)
		if err != nil {
			pingReady <- err
			return
		}
		t.responses = responses
	}

	go waitForConsole()
	go waitForTimeout()
	go checkDomainState()
	go waitForReady()
	go waitForPing()

	fmt.Fprintln(w, "[+] unpausing unikernel domain")
	if err := t.domain.Unpause().Run(); err != nil {
		console.Process.Kill()
		return err
	}

	select {
	case <-done:
		return errors.New("console unexpectedly exited")
	case err := <-exited:
		// Make sure the console can catch up
		time.Sleep(1 * time.Second)
		console.Process.Kill()
		if err != nil {
			return err
		}
		<-done
	case <-timeout:
		console.Process.Kill()
		<-done
		return errors.New("test timeout")
	case err := <-pingReady:
		console.Process.Kill()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *PingTest) Check(w io.Writer) error {
	if n := len(t.responses); n < 10 {
		return fmt.Errorf("expected %d responses, got %d", 10, n)
	}

	for _, r := range t.responses {
		if r.Error != "" {
			return fmt.Errorf("ping error for echo request %d: %v", r.ICMPSeq, r.Error)
		}
	}

	return nil
}

func (t *PingTest) Clean(w io.Writer, success bool) (err error) {
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
