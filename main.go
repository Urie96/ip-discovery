package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	PORT           = 0
	PAYLOAD_KEY    = ""
	PAYLOAD_PREFIX = ""
	TIMEOUT_MS     = 0
	SERVER_MODE    = false
)

func init() {
	flag.IntVar(&PORT, "p", 25615, "server port")
	flag.IntVar(&TIMEOUT_MS, "t", 1000, "timeout(ms)")
	flag.StringVar(&PAYLOAD_KEY, "k", "", "payload key")
	flag.BoolVar(&SERVER_MODE, "s", false, "server mode")
	flag.Parse()
	PAYLOAD_PREFIX = "(github.com/urie96/ip-discovery)" + PAYLOAD_KEY
}

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

func findOtherDevices() {
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

func server() {
	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", PORT))
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	fmt.Printf("Listening UDP at %s\n", pc.LocalAddr())

	buf := make([]byte, 1024)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("received udp from %s\n", addr)
		if strings.HasPrefix(string(buf[:n]), PAYLOAD_PREFIX) {
			_, err = pc.WriteTo(buf[:n], addr)
			if err != nil {
				fmt.Println("error response to client", err)
			} else {
				fmt.Printf("response successfully\n")
			}
		}
	}
}

func main() {
	if SERVER_MODE {
		server()
	} else {
		findOtherDevices()
	}
}
