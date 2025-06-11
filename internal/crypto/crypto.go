// Package crypto provides cryptographic functions for secure data encryption and decryption.
// It uses AES-256-GCM for authenticated encryption and Argon2id for key derivation,
// following best practices for secure cryptographic operations.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

// DeriveKey generates a cryptographic key from a password and salt using Argon2id.
// It's designed to be computationally intensive to prevent brute force attacks.
//
// Parameters:
//   - password: The plaintext password to derive the key from
//   - salt: A cryptographically secure random salt (recommended 16-32 bytes)
//
// Returns:
//   - A 32-byte key suitable for use with AES-256
//
// Note: The function uses the following Argon2id parameters:
//   - Time: 1 iteration (trade-off between security and performance)
//   - Memory: 64MB (64 * 1024 KB)
//   - Threads: 4
//   - Key length: 32 bytes (256 bits)
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
}

// Encrypt encrypts the given plaintext using AES-256-GCM (Galois/Counter Mode).
// It generates a random nonce for each encryption operation.
//
// Parameters:
//   - plaintext: The data to be encrypted
//   - key: A 32-byte key (typically derived using DeriveKey)
//
// Returns:
//   - encrypted: The encrypted data
//   - nonce: The randomly generated nonce used for encryption
//   - err: Any error that occurred during encryption
//
// The function returns an error if:
//   - The key is not a valid AES key (16, 24, or 32 bytes)
//   - There's an error generating random bytes for the nonce
//
// Note: The nonce must be stored along with the encrypted data and provided
// to Decrypt for successful decryption.
func Encrypt(plaintext []byte, key []byte) (encrypted []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	encrypted = aesgcm.Seal(nil, nonce, plaintext, nil)
	return encrypted, nonce, nil
}

// Decrypt decrypts data that was encrypted using the Encrypt function.
//
// Parameters:
//   - encrypted: The encrypted data
//   - key: The same key used for encryption
//   - nonce: The nonce that was used during encryption
//
// Returns:
//   - The decrypted plaintext data
//   - An error if decryption fails (e.g., if the key is incorrect or the data is corrupted)
//
// Note: This function verifies the authenticity of the encrypted data
// before decrypting it, protecting against tampering.
func Decrypt(encrypted []byte, key []byte, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesgcm.Open(nil, nonce, encrypted, nil)
}
