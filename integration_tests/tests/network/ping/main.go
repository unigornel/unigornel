package main

import "C"
import (
	"github.com/unigornel/go-tcpip/ethernet"
	"github.com/unigornel/go-tcpip/icmp"
	"github.com/unigornel/go-tcpip/ipv4"
)

func main() {}

//export Main
func Main(unused int) {
	sourceIP := [4]byte{10, 0, 100, 2}

	nic := ethernet.NewNIC()
	eth := ethernet.NewLayer(nic)
	arp := ipv4.NewARP(nic.GetMAC(), sourceIP, eth)
	ip := ipv4.NewLayer(sourceIP, arp, eth)
	icmp.NewLayer(ip)

	nic.Start()

	m := make(chan int)
	m <- 0
}
