// Package cache provides secure password caching for the master password.
// The cache is stored in an encrypted file in the user's config directory.
package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

type cacheData struct {
	Password  string    `json:"password"`
	ExpiresAt time.Time `json:"expires_at"`
}

var (
	cacheFile     string
	cacheKey      []byte
	cacheKeyMutex sync.RWMutex
)

func init() {
	// Initialize cache file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	cacheDir := filepath.Join(configDir, "vlxck")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		// Fallback to temp dir if we can't create config dir
		cacheDir = os.TempDir()
	}
	cacheFile = filepath.Join(cacheDir, "password.cache")

	// Generate a key from system information
	hostname, _ := os.Hostname()
	cacheKey = argon2.IDKey([]byte(hostname), []byte("vlxck-cache-key"), 3, 32*1024, 4, 32)
}

// SetMasterPassword caches the master password with the specified timeout
func SetMasterPassword(password string, timeout time.Duration) error {
	if timeout <= 0 {
		return ClearMasterPassword()
	}

	data := cacheData{
		Password:  password,
		ExpiresAt: time.Now().Add(timeout),
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	block, err := aes.NewCipher(cacheKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write to a temp file first, then rename to ensure atomicity
	tempFile := cacheFile + ".tmp"
	if err := os.WriteFile(tempFile, []byte(base64.StdEncoding.EncodeToString(ciphertext)), 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tempFile, cacheFile); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to update cache file: %w", err)
	}

	return nil
}

// GetMasterPassword retrieves the cached master password if it exists and is not expired
func GetMasterPassword() (string, error) {
	cacheKeyMutex.RLock()
	defer cacheKeyMutex.RUnlock()

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No cache file exists
		}
		return "", fmt.Errorf("failed to read cache file: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode cache data: %w", err)
	}

	block, err := aes.NewCipher(cacheKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cache: %w", err)
	}

	var cacheData cacheData
	if err := json.Unmarshal(plaintext, &cacheData); err != nil {
		return "", fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	if time.Now().After(cacheData.ExpiresAt) {
		ClearMasterPassword() // Clean up expired cache
		return "", nil
	}

	return cacheData.Password, nil
}

// ClearMasterPassword removes the cached master password
func ClearMasterPassword() error {
	if _, err := os.Stat(cacheFile); err == nil {
		return os.Remove(cacheFile)
	}
	return nil
}
