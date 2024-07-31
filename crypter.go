package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

type Crypter interface {
	Encrypt(plain []byte) []byte
	Decrypt(crypted []byte) ([]byte, error)
}

type aesCrypter struct {
	cipher.AEAD
}

func NewAESCrypter(key string) Crypter {
	key256 := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(key256[:])
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	return aesCrypter{gcm}
}

func (c aesCrypter) Encrypt(plain []byte) []byte {
	nonce := make([]byte, c.NonceSize())
	rand.Read(nonce)
	t := c.Seal(nil, nonce, plain, nil)
	return append(nonce, t...)
}

func (c aesCrypter) Decrypt(crypted []byte) ([]byte, error) {
	nonceSize := c.NonceSize()
	if len(crypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := crypted[0:nonceSize]
	msg := crypted[nonceSize:]
	return c.Open(nil, nonce, msg, nil)
}

type noneCrypter struct{}

func (c noneCrypter) Encrypt(plain []byte) []byte {
	return plain
}

func (c noneCrypter) Decrypt(crypted []byte) ([]byte, error) {
	return crypted, nil
}
