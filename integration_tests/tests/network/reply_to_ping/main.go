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
var ipNetmask string
var ipGateway string

//export Main
func Main(unused int) {
	if ipAddress == "" {
		ipAddress = "10.0.100.2"
		fmt.Printf("[*] warning: using default IP address (%v)\n", ipAddress)
	} else {
		fmt.Printf("[+] using IP address %v\n", ipAddress)
	}

	if ipNetmask == "" {
		ipNetmask = "255.255.255.0"
		fmt.Printf("[*] warning: using default IP netmask (%v)\n", ipNetmask)
	} else {
		fmt.Printf("[+] using IP netmask %v\n", ipNetmask)
	}

	if ipGateway == "" {
		fmt.Printf("[*] warning: not using an IP gateway\n")
	} else {
		fmt.Printf("[+] using IP gateway %v\n", ipGateway)
	}

	parseIP := net.ParseIP(ipAddress)
	if parseIP != nil {
		parseIP = parseIP.To4()
	}
	if parseIP == nil {
		panic("invalid IPv4 address")
	}

	parseNetmask := net.ParseIP(ipNetmask)
	if parseNetmask != nil {
		parseNetmask = parseNetmask.To4()
	}
	if parseNetmask == nil {
		panic("invalid IPv4 netmask")
	}

	parseGateway := net.ParseIP(ipGateway)
	if parseGateway != nil {
		parseGateway = parseGateway.To4()
	}
	if ipGateway == "" {
		parseGateway = nil
	} else if parseGateway == nil {
		panic("invalid IPv4 gateway")
	}

	var sourceIP ipv4.Address
	for i := 0; i < 4; i++ {
		sourceIP[i] = parseIP[i]
	}

	var sourceNetmask ipv4.Address
	for i := 0; i < 4; i++ {
		sourceNetmask[i] = parseNetmask[i]
	}

	sourceGateway := new(ipv4.Address)
	if parseGateway != nil {
		for i := 0; i < 4; i++ {
			sourceGateway[i] = parseGateway[i]
		}
	} else {
		sourceGateway = nil
	}

	nic := ethernet.NewNIC()
	eth := ethernet.NewLayer(nic)
	arp := ipv4.NewARP(nic.GetMAC(), sourceIP, eth)
	router := ipv4.NewRouter(arp, sourceIP, sourceNetmask, sourceGateway)
	ip := ipv4.NewLayer(sourceIP, router, eth)
	icmp.NewLayer(ip)

	nic.Start()

	fmt.Println("[+] network is ready")

	m := make(chan int)
	m <- 0
}
