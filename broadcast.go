package main

import (
	"fmt"
	"net"
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

func Broadcast(prefix []byte, serverPort int, crypter Crypter, timeout time.Duration, serverCmd string) {
	pc, err := net.ListenPacket("udp4", "")
	if err != nil {
		panic(err)
	}
	defer pc.Close()
	id := uint64(time.Now().UnixMilli())

	go func() {
		for _, broadcast := range getAllBroadcast() {
			addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", broadcast, serverPort))
			if err != nil {
				panic(err)
			}

			req := Request{}
			if serverCmd != "" {
				req.Method = "shell"
				req.Body = serverCmd
			}

			_, err = pc.WriteTo(buildPayload(prefix, id, req, crypter), addr)
			if err != nil {
				panic(err)
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		pc.SetDeadline(time.Now().Add(timeout))
		n, raddr, err := pc.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return
			} else {
				panic(err)
			}
		}
		var resp Response
		ok, respID := parsePayload(buf[:n], prefix, &resp, crypter)
		if !ok || respID != id {
			continue
		}
		if resp.Code != 0 {
			fmt.Printf("%s error: %s", raddr, resp.Body)
		} else {
			fmt.Printf("%s %s", raddr, resp.Body)
		}
	}
}
