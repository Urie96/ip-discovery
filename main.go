package main

import (
	"flag"
)

func main() {
	var (
		port       = 0
		secretKey  = ""
		serverCmd  = ""
		timeoutMs  = 0
		serverMode = false
	)
	flag.IntVar(&port, "p", 25615, "server port")
	flag.IntVar(&timeoutMs, "t", 1000, "timeout(ms)")
	flag.StringVar(&secretKey, "k", "", "payload key")
	flag.BoolVar(&serverMode, "s", false, "server mode")
	flag.StringVar(&serverCmd, "exec", "", "server response command output when received broadcast")
	flag.Parse()

	crypter := NewCrypter("hello")
	prefix := []byte("github.com/urie96/ip-discovery")

	if serverMode {
		server()
	} else {
		findOtherDevices()
	}
}
