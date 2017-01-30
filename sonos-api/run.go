package main

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const (
	// UDP protocol for ResolveUDPAddr
	UDP = "udp"
	// SsdpPort default port for Simple Service Discovery Protocol
	SsdpPort = ":1900"
	// Multicast default address for UDP Multicast (SSDP)
	Multicast = "239.255.255.250" + SsdpPort
)

// SsdpData ssdp response message
type SsdpData struct {
	IP   net.IP
	Data string
}

var ssdpData chan SsdpData
var timeout chan bool

func main() {
	timeout := make(chan bool, 1)
	ssdpData = make(chan SsdpData)

	go ssdpMulticast()

	for {
		select {
		case msg := <-ssdpData:
			fmt.Println(msg.Data)
		case <-timeout:
			break
		}
	}
}

func ssdpMulticast() {
	addr, _ := net.ResolveUDPAddr(UDP, SsdpPort)

	conn, _ := net.ListenUDP(UDP, addr)
	conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	defer conn.Close()
	defer func() {
		if err := recover(); err != nil {
			timeout <- true
		}
	}()

	mcastaddr, _ := net.ResolveUDPAddr(UDP, Multicast)
	msg := buildMulticastDiscoveryPackage()
	conn.WriteTo(msg, mcastaddr)

	for {
		buf := make([]byte, 4096)
		_, addr, _ := conn.ReadFromUDP(buf)
		ssdpData <- SsdpData{addr.IP, string(buf)}
	}
}

func buildMulticastDiscoveryPackage() []byte {
	hpackage := new(bytes.Buffer)
	hpackage.WriteString("M-SEARCH * HTTP/1.1\r\n")
	hpackage.WriteString("HOST: 239.255.255.250:1900\r\n")
	hpackage.WriteString("MAN: \"ssdp:discover\"\r\n")
	hpackage.WriteString("ST: urn:schemas-upnp-org:device:ZonePlayer:1\r\n")
	hpackage.WriteString("\r\n")
	return hpackage.Bytes()
}
