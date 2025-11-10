package goencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// GoEncrypt is an interface that defines the methods for encrypting and decrypting data.
// It is used to encrypt and decrypt data in a secure way.
type GoEncrypt interface {
	Encrypt(data any) (encrypted string, err error)
	Decrypt(data string) (result []byte, err error)
	DecryptFromBytes(data []byte) (result []byte, err error)
}

// cryp is a struct that implements the GoEncrypt interface.
// It is used to encrypt and decrypt data in a secure way.
type cryp struct {
	secret string
	size   []byte
}

// New is a function that creates a new GoEncrypt instance.
// It takes a secret string and returns a GoEncrypt instance.
func New(secret string) GoEncrypt {
	bb := make([]byte, 16)
	rand.Read(bb)

	if len(trimSpace(secret)) < 24 {
		panic("secret for invt_cryptography must be at least 16 characters long")
	}

	return &cryp{
		secret: secret,
		size:   bb,
	}
}

// toBytes is a function that converts the data to bytes.
// It takes a data any and returns a byte slice and an error.
func (c *cryp) toBytes(data any) ([]byte, error) {
	return json.Marshal(data)
}

// encode is a function that encodes the data to a base64 string.
// It takes a byte slice and returns a string.
func (c *cryp) encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// decode is a function that decodes the data from a base64 string.
// It takes a string and returns a byte slice and an error.
func (c *cryp) decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// getBlock is a function that returns a new AES cipher block.
// It takes a secret string and returns a cipher.Block and an error.
func (c *cryp) getBlock() (cipher.Block, error) {
	return aes.NewCipher([]byte(c.secret))
}

// Encrypt is a function that encrypts the data.
// It takes a data any and returns a string and an error.
func (c *cryp) Encrypt(data any) (string, error) {
	var (
		enc string
		err error
	)

	b, err := c.toBytes(data)
	if err != nil {
		return enc, err
	}

	blk, err := c.getBlock()
	if err != nil {
		return enc, err
	}

	cfb := cipher.NewCFBEncrypter(blk, c.size)
	cText := make([]byte, len(b))
	cfb.XORKeyStream(cText, b)

	enc = c.encode(cText)

	return enc, nil
}

// Decrypt is a function that decrypts the data.
// It takes a string and returns a byte slice and an error.
func (c *cryp) Decrypt(data string) ([]byte, error) {
	blk, err := c.getBlock()
	if err != nil {
		return nil, err
	}

	cText, err := c.decode(data)
	if err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBDecrypter(blk, c.size)
	plain := make([]byte, len(cText))
	cfb.XORKeyStream(plain, cText)

	return plain, nil
}

// DecryptFromBytes is a function that decrypts the data from a byte slice.
// It takes a byte slice and returns a byte slice and an error.
func (c *cryp) DecryptFromBytes(data []byte) ([]byte, error) {
	dec, err := c.Decrypt(string(data))
	if err != nil {
		return nil, err
	}

	return dec, nil
}

// trimSpace is a function that trims the space from the data.
// It takes a string and returns a string.
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
