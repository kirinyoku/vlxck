// Package store provides functionality for securely storing and managing secrets.
// It handles encryption, decryption, and persistence of sensitive data.
package store

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kirinyoku/vlxck/internal/crypto"
)

// Store represents the main data structure for storing secrets.
// It includes version information for backward compatibility
// and a collection of secrets.
type Store struct {
	// Version indicates the data schema version
	Version int `json:"version"`
	// Secrets is a collection of stored secret items
	Secrets []Secret `json:"secrets"`
}

// Secret represents a single secret item with its metadata.
type Secret struct {
	// Name is the unique identifier for the secret
	Name string `json:"name"`
	// Value is the actual secret value (encrypted at rest)
	Value string `json:"value"`
	// Category helps in organizing secrets into groups
	Category string `json:"category"`
	// CreatedAt records when the secret was created
	CreatedAt time.Time `json:"created_at"`
}

// LoadStore reads and decrypts the store from the specified file.
// It handles the decryption of the stored data using the provided password.
//
// Parameters:
//   - filePath: Path to the encrypted store file
//   - password: Password used for decryption
//
// Returns:
//   - *Store: Pointer to the loaded and decrypted store or an empty store if the file does not exist
//   - error: Any error that occurred during file operations, decryption, or JSON unmarshaling
//
// Note: The file format is expected to be [16-byte salt][12-byte nonce][encrypted data].
// The function uses AES-256-GCM for decryption with the provided password and stored salt.
func LoadStore(filePath, password string) (*Store, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Secrets: []Secret{}}, fmt.Errorf("store file %s does not exist", filePath)
		}
		return nil, fmt.Errorf("failed to read store: %v", err)
	}

	if len(data) == 0 {
		return &Store{Secrets: []Secret{}}, fmt.Errorf("store file %s is empty", filePath)
	}

	salt := data[:16]
	nonce := data[16:28]
	encrypted := data[28:]
	key := crypto.DeriveKey(password, salt)

	plaintext, err := crypto.Decrypt(encrypted, key, nonce)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := json.Unmarshal(plaintext, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

// SaveStore encrypts and writes the store to the specified file.
// If the file exists, it reuses the existing salt; otherwise, it generates a new one.
//
// Parameters:
//   - filePath: Path where the store should be saved
//   - password: Password used for encryption
//   - store: Pointer to the Store struct to be saved
//
// Returns:
//   - error: Any error that occurred during file operations, encryption, or JSON marshaling
//
// Note: The file format is [16-byte salt][12-byte nonce][encrypted data].
// The function creates any necessary parent directories with 0700 permissions.
// The file is saved with 0600 permissions for security.
func SaveStore(filePath, password string, store *Store) error {
	var salt []byte
	if _, err := os.Stat(filePath); err == nil {
		data, _ := os.ReadFile(filePath)
		salt = data[:16]
	} else {
		salt = make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return err
		}
	}
	key := crypto.DeriveKey(password, salt)
	plaintext, _ := json.Marshal(store)
	encrypted, nonce, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filePath)
	os.MkdirAll(dir, 0700)
	data := append(salt, nonce...)
	data = append(data, encrypted...)
	return os.WriteFile(filePath, data, 0600)
}

// InitializeStore creates a new, empty store file with default settings.
// It generates a new random salt and initializes the store with version 1.
//
// Parameters:
//   - filePath: Path where the new store should be created
//   - password: Password to be used for encrypting the store
//
// Returns:
//   - error: Any error that occurred during store creation or initialization
//
// Note: This function creates a new store with an empty secrets slice and version 1.
// The store is immediately saved to disk using SaveStore with the provided password.
// The salt is randomly generated using crypto/rand for secure key derivation.
func InitializeStore(filePath, password string) error {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	store := &Store{Version: 1, Secrets: []Secret{}}
	return SaveStore(filePath, password, store)
}
