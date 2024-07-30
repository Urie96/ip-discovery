package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

func server(port int64, prefix []byte, cipher Crypter) {
	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	log.Printf("Listening UDP at %s\n", pc.LocalAddr())

	buflen := 1024
	buf := make([]byte, buflen)
	prefixBuf := make([]byte, len(prefix))
	idBuf := [8]byte{}
	dataBuf := make([]byte, buflen-len(prefixBuf)-len(idBuf))
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			panic(err)
		}
		log.Printf("received udp from %s\n", addr)

		dataLen := n - len(idBuf) - len(prefix)
		if dataLen < 0 {
			continue
		}
		splitBytes(buf, prefixBuf, idBuf[:], dataBuf)
		if !bytes.Equal(prefixBuf, prefix) {
			continue
		}
		id := bytesToUint64(idBuf)
		data := string(dataBuf)
		log.Printf("id: %d, data: %s", id, dataBuf)

		cmd := exec.Command("echo", "a")

		if strings.HasPrefix(string(buf[:n])) {
			_, err = pc.WriteTo(buf[:n], addr)
			if err != nil {
				log.Println("error response to client", err)
			} else {
				log.Printf("response successfully\n")
			}
		}
	}
}
