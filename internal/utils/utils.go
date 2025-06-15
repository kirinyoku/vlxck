// Package utils provides a collection of general-purpose utility functions and helpers
// that are used across the application. These utilities are designed to be reusable,
// well-tested, and follow Go best practices.
package utils

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/kirinyoku/vlxck/internal/store"
	"golang.org/x/term"
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

// PromptForPassword prompts the user for a password and returns it as a string.
// It reads the password from the standard input without echoing it to the screen.
//
// Parameters:
//   - prompt: The prompt message to display to the user
//
// Returns:
//   - string: The password entered by the user
func PromptForPassword(prompt string) string {
	fmt.Print(prompt)
	password, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return strings.TrimSpace(string(password))
}

// PromptForConflictChoice prompts the user for a choice when a conflict is detected between a local secret and an imported secret.
// It displays the details of both secrets and allows the user to choose how to resolve the conflict.
//
// Parameters:
//   - localSecret: The local secret with the conflict
//   - importedSecret: The imported secret with the conflict
//
// Returns:
//   - string: The user's choice ('l' for local, 'i' for imported, 's' for skip)
func PromptForConflictChoice(localSecret, importedSecret store.Secret) string {
	fmt.Printf("Conflict detected for secret name '%s':\n", localSecret.Name)
	fmt.Printf("Local secret: Value=%s, Category=%s\n", localSecret.Value, localSecret.Category)
	fmt.Printf("Imported secret: Value=%s, Category=%s\n", importedSecret.Value, importedSecret.Category)
	fmt.Printf("Choose action: [l] keep local, [i] use imported, [s] skip: ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		choice := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if choice == "l" || choice == "i" || choice == "s" {
			return choice
		}
		fmt.Print("Invalid choice. Please enter [l], [i], or [s]: ")
	}
	fmt.Println("Error reading choice. Exiting.")
	os.Exit(1)
	return ""
}

// CopyToClipboard copies the specified text to the clipboard.
//
// Parameters:
//   - text: The text to copy to the clipboard
//
// Returns:
//   - error: Any error that occurred during the clipboard operation
func CopyToClipboard(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %v", err)
	}
	return nil
}
