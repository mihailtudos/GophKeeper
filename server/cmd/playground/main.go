package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// DeriveKey generates a 32-byte encryption key using Argon2id
func DeriveKey(masterPassword string, userSalt []byte) []byte {
	return argon2.IDKey([]byte(masterPassword), userSalt, 1, 64*1024, 4, 32)
}

// Encrypt encrypts plaintext or binary data using AES-GCM
func Encrypt(masterPassword string, userSalt, plaintext []byte) ([]byte, []byte, error) {
	key := DeriveKey(masterPassword, userSalt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// Generate a new nonce (IV)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	// Encrypt data
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(masterPassword string, userSalt, ciphertext, nonce []byte) ([]byte, error) {
	key := DeriveKey(masterPassword, userSalt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Test the encryption and decryption
func main() {
	masterPassword := "strongpassword123"
	userSalt := []byte("random-usersalt123") // Normally, this comes from the DB

	secret := []byte("My super secret data")

	// Encrypt
	ciphertext, nonce, err := Encrypt(masterPassword, userSalt, secret)
	if err != nil {
		fmt.Println("Encryption Error:", err)
		return
	}

	fmt.Printf("Encrypted Data: %x\n", ciphertext)
	fmt.Printf("Nonce: %x\n", nonce)

	// Decrypt
	decryptedText, err := Decrypt(masterPassword, userSalt, ciphertext, nonce)
	if err != nil {
		fmt.Println("Decryption Error:", err)
		return
	}

	fmt.Println("Decrypted Text:", string(decryptedText))
}
