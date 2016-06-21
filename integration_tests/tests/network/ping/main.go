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
	ip := [4]byte{10, 0, 100, 2}

	nic := ethernet.NewNIC()
	ethDemux := ethernet.NewDemux(nic.Receive(), ethernet.DemuxLog)

	arp := ipv4.NewARP(nic.GetMAC(), ip, nic.Send())
	arp.Bind(ethDemux)

	ipl := ipv4.NewLayer([4]byte{10, 0, 100, 2}, arp, nic.Send())
	ipDemux := ipv4.NewDemux(ipl.Receive(), ipv4.DemuxLog)
	ipl.Bind(ethDemux)

	icmpl := icmp.NewLayer(ipl.Send())
	icmpl.Bind(ipDemux)

	m := make(chan int)
	m <- 0
}
