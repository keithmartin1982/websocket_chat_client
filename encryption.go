package websocket_chat_client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	
	"golang.org/x/crypto/pbkdf2"
)

func encrypt(password, plaintext []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("error generating salt: %v", err)
	}
	key := pbkdf2.Key(password, salt, 1000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("NewGCM: %s", err)
	}
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate iv: %s", err)
	}
	ciphertext := gcm.Seal(nil, iv, plaintext, nil)
	hexSalt := make([]byte, hex.EncodedLen(len(salt)))
	hex.Encode(hexSalt, salt)
	hexIv := make([]byte, hex.EncodedLen(len(iv)))
	hex.Encode(hexIv, iv)
	hexCiphertext := make([]byte, hex.EncodedLen(len(ciphertext)))
	hex.Encode(hexCiphertext, ciphertext)
	return append(append(append(append(hexSalt, []byte("-")...), hexIv...), []byte("-")...), hexCiphertext...), nil // lol
}

func decrypt(password, cipherText []byte) ([]byte, error) {
	if !bytes.ContainsAny(cipherText, "-") {
		return nil, fmt.Errorf("invalid data")
	}
	data := bytes.Split(cipherText, []byte("-"))
	salt := make([]byte, hex.DecodedLen(len(data[0])))
	hex.Decode(salt, data[0])
	iv := make([]byte, hex.DecodedLen(len(data[1])))
	hex.Decode(iv, data[1])
	ciphertext := make([]byte, hex.DecodedLen(len(data[2])))
	hex.Decode(ciphertext, data[2])
	key := pbkdf2.Key(password, salt, 1000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %v", err)
	}
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %v", err)
	}
	return plaintext, nil
}
