package ifconfig

import (
	"fmt"
	"net"
	"os/exec"
)

func SetIP(iface string, ip net.IP, mask net.IPMask) error {
	maskIP := net.IP(mask).To4()
	if maskIP == nil {
		return fmt.Errorf("invalid mask")
	}
	_, err := exec.Command(
		"ifconfig",
		iface,
		ip.String(),
		"netmask", maskIP.String(),
	).Output()
	return err
}
