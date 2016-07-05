package main

import "C"
import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/unigornel/go-tcpip/ethernet"
	"github.com/unigornel/go-tcpip/icmp"
	"github.com/unigornel/go-tcpip/ipv4"
)

func main() {}

var ipAddress string
var ipNetmask string
var ipGateway string
var ipDestination string

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

	if ipDestination == "" {
		ipDestination = "10.0.100.1"
		fmt.Printf("[*] warning: using default destination IP (%v)\n", ipDestination)
	} else {
		fmt.Printf("[+] using destination IP %v\n", ipDestination)
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

	parseDestination := net.ParseIP(ipDestination)
	if parseDestination != nil {
		parseDestination = parseDestination.To4()
	}
	if parseDestination == nil {
		panic("invalid IPv4 destination")
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

	var destination ipv4.Address
	for i := 0; i < 4; i++ {
		destination[i] = parseDestination[i]
	}

	nic := ethernet.NewNIC()
	eth := ethernet.NewLayer(nic)
	arp := ipv4.NewARP(nic.GetMAC(), sourceIP, eth)
	router := ipv4.NewRouter(arp, sourceIP, sourceNetmask, sourceGateway)
	ip := ipv4.NewLayer(sourceIP, router, eth)
	icmpLayer := icmp.NewLayer(ip)

	pinger := &pinger{
		requests: make(map[uint16]chan bool),
	}

	go pinger.handleReplies(icmpLayer)
	nic.Start()
	pinger.Start(icmpLayer, destination)
}

type pinger struct {
	sync.RWMutex
	requests map[uint16]chan bool
}

func (pinger *pinger) Start(layer icmp.Layer, destination ipv4.Address) {
	ident := uint16(rand.Int())
	seq := uint16(0)
	payload := []byte("abcdefghijklmnopqrstuvwxyz")
	for {
		seq += 1
		p := icmp.NewEchoRequest(ident, seq, payload)
		p.Address = destination

		pinger.Lock()
		pinger.requests[seq] = make(chan bool, 1)
		pinger.Unlock()

		if err := layer.Send(p); err != nil {
			fmt.Printf("[-] error: ping %v (seq %v): %v\n", destination, seq, err)
			pinger.RLock()
			close(pinger.requests[seq])
			pinger.RUnlock()
		}

		go pinger.waitForReply(seq)
		fmt.Printf("[+] sent ping to %v (seq %v)\n", destination, seq)
		time.Sleep(500 * time.Millisecond)
	}
}

func (pinger *pinger) waitForReply(seq uint16) {
	pinger.RLock()
	c := pinger.requests[seq]
	pinger.RUnlock()

	select {
	case <-time.After(1 * time.Second):
		fmt.Printf("[-] error: ping %v: timeout\n", seq)
	case flag := <-c:
		if flag {
			fmt.Printf("[+] got ping reply seq %v\n", seq)
		} else {
			fmt.Printf("[-] error for ping %v\n", seq)
		}
	}

}

func (pinger *pinger) handleReplies(icmpLayer icmp.Layer) {
	for p := range icmpLayer.Packets(icmp.EchoReplyType) {
		data := p.Data.(icmp.Echo)

		pinger.RLock()
		c := pinger.requests[data.Header.SequenceNumber]
		pinger.RUnlock()

		if c != nil {
			c <- true
			close(c)
		}
	}
}
