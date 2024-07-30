package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func getAllBroadcast() []net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	var allBroadcast []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			p4 := ipnet.IP.To4()
			if p4 != nil && !p4.IsLoopback() {
				mask := ipnet.Mask
				broadcast := make(net.IP, 4)
				for i := 0; i < 4; i++ {
					broadcast[i] = p4[i] | ^mask[i]
				}
				allBroadcast = append(allBroadcast, broadcast)
			}
		}
	}
	return allBroadcast
}

func findOtherDevices(payload) {
	pc, err := net.ListenPacket("udp4", "")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	payload := PAYLOAD_PREFIX + strconv.Itoa(int(time.Now().UnixMicro()))

	go func() {
		for _, broadcast := range getAllBroadcast() {
			addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcast, PORT))
			if err != nil {
				panic(err)
			}

			_, err = pc.WriteTo([]byte(payload), addr)
			if err != nil {
				panic(err)
			}
		}
	}()

	buf := make([]byte, 1024)
	set := make(map[string]bool)
	for {
		pc.SetDeadline(time.Now().Add(time.Microsecond * time.Duration(TIMEOUT_MS)))
		n, raddr, err := pc.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				os.Exit(0)
			} else {
				panic(err)
			}
		}
		if string(buf[:n]) == payload {
			ip := strings.SplitN(raddr.String(), ":", 2)[0]
			if !set[ip] {
				set[ip] = true
				fmt.Println(ip)
			}
		}
	}
}
