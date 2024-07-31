package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
			resp := dispatcher(ctx, req)
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
	Code int
	Body string
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
	dataBuf := make([]byte, len(buf)-len(idBuf))
	SplitBytes(decypted, idBuf[:], dataBuf)
	err = json.Unmarshal(dataBuf, dst)
	if err != nil {
		return
	}
	return true, BytesToUint64(idBuf)
}

func dispatcher(ctx context.Context, req Request) *Response {
	switch req.Method {
	case "echo":
		return &Response{Body: req.Body}
	default:
		return nil
	}
}

// func handle(ctx context.Context, body string) (string, error) {
// 	return []byte("hello"), nil
// }
