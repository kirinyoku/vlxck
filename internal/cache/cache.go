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

// cacheData represents the data stored in the cache file.
// This struct defines the format of data stored in the cache.
type cacheData struct {
	Password  string    `json:"password"`
	ExpiresAt time.Time `json:"expires_at"`
}

var (
	cacheFile     string       // Path to the cache file (~/.config/vlxck/password.cache)
	cacheKey      []byte       // Encryption key used for encrypting/decrypting the cache (generated from hostname)
	cacheKeyMutex sync.RWMutex // Mutex for synchronizing access to cacheKey
)

func init() {
	// Initialize cache file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to temp directory (/tmp on Linux) if unavailable
		configDir = os.TempDir()
	}
	cacheDir := filepath.Join(configDir, "vlxck")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		// If creating the directory fails, fall back to the temp directory
		cacheDir = os.TempDir()
	}
	cacheFile = filepath.Join(cacheDir, "password.cache")

	// Generate an encryption key based on system information
	hostname, _ := os.Hostname()
	cacheKey = argon2.IDKey([]byte(hostname), []byte("vlxck-cache-key"), 3, 32*1024, 4, 32)
}

// SetMasterPassword caches the master password with the specified timeout.
// If timeout is less than or equal to 0, the cache is cleared.
//
// Parameters:
//   - password: The master password to cache
//   - timeout: The duration for which the password should be cached
//
// Returns:
//   - error: Any error that occurred during caching
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

	// Create an AES cipher based on cacheKey.
	// AES (Advanced Encryption Standard) is a symmetric encryption algorithm that uses one key
	// for both encryption and decryption
	block, err := aes.NewCipher(cacheKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM (Galois/Counter Mode) for encryption.
	// GCM provides both confidentiality and data integrity (verifies data hasn't been tampered with)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a nonce (number used once) — a unique random number for each encryption.
	// Nonce ensures that identical data encrypts differently each time, enhancing security
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data. Seal prepends the nonce to the ciphertext.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write encrypted data to a temporary file, then rename it for atomicity.
	// This ensures the main file isn't corrupted if the write operation is interrupted
	tempFile := cacheFile + ".tmp"
	if err := os.WriteFile(tempFile, []byte(base64.StdEncoding.EncodeToString(ciphertext)), 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Rename the temporary file to the main cache file
	if err := os.Rename(tempFile, cacheFile); err != nil {
		os.Remove(tempFile) // Clean up the temp file on error
		return fmt.Errorf("failed to update cache file: %w", err)
	}

	return nil
}

// GetMasterPassword retrieves the cached master password if it exists and is not expired.
//
// Returns:
//   - string: The cached master password
//   - error: Any error that occurred during retrieval
func GetMasterPassword() (string, error) {
	cacheKeyMutex.RLock()
	defer cacheKeyMutex.RUnlock()

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // If the file doesn't exist, return an empty password (no cache)
		}
		return "", fmt.Errorf("failed to read cache file: %w", err)
	}

	// Decode the data from Base64 back to binary format
	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode cache data: %w", err)
	}

	// Recreate the AES cipher and GCM for decryption.
	// This is needed to decrypt data previously encrypted
	block, err := aes.NewCipher(cacheKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and encrypted data from ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Split nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data.
	// Open verifies integrity and decrypts the data.
	// If the data was tampered with or the key doesn't match, it returns an error
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt cache: %w", err)
	}
	var cacheData cacheData
	if err := json.Unmarshal(plaintext, &cacheData); err != nil {
		return "", fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	// Check if the cache has expired
	if time.Now().After(cacheData.ExpiresAt) {
		ClearMasterPassword() // Clean up expired cache
		return "", nil
	}

	return cacheData.Password, nil
}

// ClearMasterPassword removes the cached master password.
//
// Returns:
//   - error: Any error that occurred during removal
func ClearMasterPassword() error {
	if _, err := os.Stat(cacheFile); err == nil {
		return os.Remove(cacheFile)
	}
	return nil
}
