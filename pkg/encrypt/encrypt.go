// Package encrypt provides functions for encrypting and decrypting data using AES-GCM.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/argon2"
	"io"
)

// DeriveKey generates a 32-byte encryption key using Argon2id
func DeriveKey(masterPassword string, userSalt []byte) []byte {
	return argon2.IDKey([]byte(masterPassword), userSalt, 1, 64*1024, 4, 32)
}

// Encrypt takes a master password, user salt, and plaintext data and returns:
// - ciphertext: the encrypted data as a byte slice
// - nonce: the initialization vector (IV) used for encryption as a byte slice
// - error: any error encountered during encryption
// The function uses AES-GCM for encryption with a key derived from the master password
// and user salt using Argon2id. A random nonce is generated for each encryption operation.
// Example:
//
//	masterPassword := "user-secure-password"
//	userSalt := []byte{...} // Salt retrieved from user's record
//	plaintext := []byte("sensitive data to encrypt")
//
//	ciphertext, nonce, err := encrypt(masterPassword, userSalt, plaintext)
//	if err != nil {
//	    // Handle encryption error
//	}
//
//	// Store ciphertext and nonce in the database
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

// ComputeChecksum calculates a SHA-256 hash of the encrypted data
func ComputeChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
