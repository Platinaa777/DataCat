package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Cipher struct{ aead cipher.AEAD }

func New(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("a 32-byte encryption key is required")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead}, nil
}
func (c *Cipher) Seal(scope, value string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(c.aead.Seal(nonce, nonce, []byte(value), []byte(scope))), nil
}
func (c *Cipher) Open(scope, value string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(b) < c.aead.NonceSize() {
		return "", fmt.Errorf("invalid encrypted value")
	}
	n := c.aead.NonceSize()
	plain, err := c.aead.Open(nil, b[:n], b[n:], []byte(scope))
	if err != nil {
		return "", fmt.Errorf("decrypt value: %w", err)
	}
	return string(plain), nil
}
