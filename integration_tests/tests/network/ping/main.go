package main

import "C"
import (
	"fmt"
	"net"

	"github.com/unigornel/go-tcpip/ethernet"
	"github.com/unigornel/go-tcpip/icmp"
	"github.com/unigornel/go-tcpip/ipv4"
)

func main() {}

var ipAddress string

//export Main
func Main(unused int) {
	if ipAddress == "" {
		ipAddress = "10.0.100.2"
		fmt.Printf("[*] warning: using default IP address (%v)\n", ipAddress)
	} else {
		fmt.Printf("[+] using IP address %v\n", ipAddress)
	}

	parseIP := net.ParseIP(ipAddress)
	if parseIP != nil {
		parseIP = parseIP.To4()
	}
	if parseIP == nil {
		panic("invalid IPv4 address")
	}

	var sourceIP ipv4.Address
	for i := 0; i < 4; i++ {
		sourceIP[i] = parseIP[i]
	}

	nic := ethernet.NewNIC()
	eth := ethernet.NewLayer(nic)
	arp := ipv4.NewARP(nic.GetMAC(), sourceIP, eth)
	ip := ipv4.NewLayer(sourceIP, arp, eth)
	icmp.NewLayer(ip)

	nic.Start()

	fmt.Println("[+] network is ready")

	m := make(chan int)
	m <- 0
}
