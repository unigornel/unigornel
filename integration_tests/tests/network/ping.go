package network

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/unigornel/unigornel/integration_tests/brctl"
	"github.com/unigornel/unigornel/integration_tests/ip"
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
}

func (t *PingTest) GetName() string {
	return "ping"
}

func (t *PingTest) GetCategory() string {
	return "network"
}

func (t *PingTest) Build(w io.Writer) error {
	if err := t.setupNetwork(w); err != nil {
		return err
	}

	file, err := tests.Build(
		w, "ping", tests.SimpleTestPackage("network", "ping"),
		"--ldflags", "-X main.ipAddress="+t.network.unikernelIP.String(),
	)
	t.unikernel = file
	return err
}

func (t *PingTest) setupNetwork(w io.Writer) error {
	var ipnet *net.IPNet
	var err error
	if v := os.Getenv("TEST_PING_NETMASK"); v != "" {
		fmt.Fprintln(w, "[+] using network from TEST_PING_NETMASK:", v)
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
	if err := ip.Ifconfig(bridge, t.network.xenIP, t.network.netmask); err != nil {
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
	return nil
}

func (t *PingTest) Check(w io.Writer) error {
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
