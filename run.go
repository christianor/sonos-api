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

func main() {
	go ssdpMulticast()

	for {
		fmt.Println("waiting for ssdp messages")
		msg := <-ssdpData
		fmt.Println(msg.Data)
	}
}

func ssdpMulticast() {
	ssdpData = make(chan SsdpData)
	addr, _ := net.ResolveUDPAddr(UDP, SsdpPort)

	conn, _ := net.ListenUDP(UDP, addr)
	conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	defer conn.Close()

	mcastaddr, _ := net.ResolveUDPAddr(UDP, Multicast)
	msg := buildMulticastDiscoveryPackage()
	conn.WriteTo(msg, mcastaddr)

	buf := make([]byte, 4096)

	for {
		fmt.Println("reading data")
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
