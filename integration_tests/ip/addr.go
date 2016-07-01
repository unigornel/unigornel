package ip

import (
	"fmt"
	"os/exec"
	"strings"
)

func Down(iface string) error {
	out, err := exec.Command("ip", "link", "set", "dev", iface, "down").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: ip link set dev %s down: %v", iface, strings.TrimSpace(string(out)))
	}
	return nil
}
