// Package utils provides a collection of general-purpose utility functions and helpers
// that are used across the application. These utilities are designed to be reusable,
// well-tested, and follow Go best practices.
package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// GeneratePassword generates a random password of the specified length.
// It includes lowercase and uppercase letters by default.
// Additional characters can be included based on the parameters.
//
// Parameters:
//   - length: The desired length of the password
//   - useSymbols: Whether to include special characters (!@#$%^&*()-_=+)
//   - useNumbers: Whether to include digits (0123456789)
//
// Returns:
//   - string: The generated password
//   - error: Any error that occurred during password generation
func GeneratePassword(length int, useSymbols, useNumbers bool) (string, error) {
	if length < 1 {
		return "", errors.New("length must be positive")
	}
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if useNumbers {
		chars += "0123456789"
	}
	if useSymbols {
		chars += "!@#$%^&*()-_=+"
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[idx.Int64()]
	}
	return string(result), nil
}
