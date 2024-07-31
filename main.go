package main

import (
	"flag"
	"time"
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
	flag.StringVar(&secretKey, "k", "", "secret key (no crypter if leave empty)")
	flag.BoolVar(&serverMode, "s", false, "server mode")
	flag.StringVar(&serverCmd, "exec", "", "server response command output when received broadcast")
	flag.Parse()

	var crypter Crypter = noneCrypter{}
	if secretKey != "" {
		crypter = NewAESCrypter(secretKey)
	}
	prefix := []byte("github.com/urie96/ip-discovery")

	if serverMode {
		Serve(port, prefix, crypter)
	} else {
		Broadcast(prefix, port, crypter, time.Millisecond*time.Duration(timeoutMs), serverCmd)
	}
}
