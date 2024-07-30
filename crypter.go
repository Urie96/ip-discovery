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

type crypter struct {
	cipher.AEAD
}

func NewCrypter(key string) Crypter {
	key256 := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(key256[:])
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	return crypter{gcm}
}

func (c crypter) Encrypt(plain []byte) []byte {
	nonce := make([]byte, c.NonceSize())
	rand.Read(nonce)
	t := c.Seal(nil, nonce, plain, nil)
	return append(nonce, t...)
}

func (c crypter) Decrypt(crypted []byte) ([]byte, error) {
	nonceSize := c.NonceSize()
	if len(crypted) < nonceSize {
		return nil, fmt.Errorf("Ciphertext too short.")
	}
	nonce := crypted[0:nonceSize]
	msg := crypted[nonceSize:]
	return c.Open(nil, nonce, msg, nil)
}
