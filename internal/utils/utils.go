// Package utils provides a collection of general-purpose utility functions and helpers
// that are used across the application. These utilities are designed to be reusable,
// well-tested, and follow Go best practices.
package utils

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/manifoldco/promptui"
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

// PromptForInput prompts the user for input using the promptui library.
// It displays a label and allows the user to enter a value.
//
// Parameters:
//   - label: The label to display to the user
//   - defaultValue: The default value to display in the input field
//   - validate: A validation function to validate the input
//
// Returns:
//   - string: The input value entered by the user
//   - error: Any error that occurred during the input operation
func PromptForInput(label, defaultValue string, validate func(string) error) (string, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %v", err)
	}
	return result, nil
}

// PromptForSelect prompts the user for a selection using the promptui library.
// It displays a label and a list of items, and allows the user to select one.
//
// Parameters:
//   - label: The label to display to the user
//   - items: The list of items to display in the selection menu
//
// Returns:
//   - string: The selected item
//   - error: Any error that occurred during the selection operation
func PromptForSelect(label string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("select failed: %v", err)
	}
	return result, nil
}

// PromptForInt prompts the user for an integer input with validation.
//
// Parameters:
//   - label: The prompt label to display
//   - defaultValue: The default value to use if input is empty
//   - min: Minimum allowed value (inclusive)
//   - max: Maximum allowed value (inclusive)
//
// Returns:
//   - int: The validated integer input
//   - error: Any error that occurred during the input operation
func PromptForInt(label string, defaultValue, min, max int) (int, error) {
	validate := func(input string) error {
		if input == "" {
			return nil // Use default value
		}
		val, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("must be a valid number")
		}
		if val < min || val > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}

	input, err := PromptForInput(label, strconv.Itoa(defaultValue), validate)
	if err != nil {
		return 0, err
	}

	// If input is empty, use default value
	if input == "" {
		return defaultValue, nil
	}

	// We already validated the input, so we can ignore the error
	result, _ := strconv.Atoi(input)
	return result, nil
}

// PromptForSecretName prompts the user for a secret name using the promptui library.
// It validates that the name is not empty and not already in use.
//
// Parameters:
//   - existingSecrets: A slice of existing secrets to check for duplicates
//
// Returns:
//   - string: The secret name entered by the user
//   - error: Any error that occurred during the input operation
func PromptForSecretName(currentSecrets []store.Secret) (string, error) {
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("name cannot be empty")
		}
		for _, secret := range currentSecrets {
			if secret.Name == input {
				return fmt.Errorf("secret with name '%s' already exists", input)
			}
		}
		return nil
	}
	return PromptForInput("Enter secret name", "", validate)
}

// PromptForSecretValue prompts the user for a secret value using the promptui library.
// It displays a label and allows the user to enter a value.
//
// Returns:
//   - string: The secret value entered by the user
//   - error: Any error that occurred during the input operation
func PromptForSecretValue() (string, error) {
	options := []string{"Enter value manually", "Generate password"}
	choice, err := PromptForSelect("Choose value input method", options)
	if err != nil {
		return "", err
	}

	if choice == "Enter value manually" {
		validate := func(input string) error {
			if input == "" {
				return fmt.Errorf("value cannot be empty")
			}
			return nil
		}
		return PromptForInput("Enter secret value", "", validate)
	}

	lengthStr, err := PromptForInput("Enter password length", "16", func(input string) error {
		if n, err := strconv.Atoi(input); err != nil || n <= 0 {
			return fmt.Errorf("length must be a positive integer")
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	length, _ := strconv.Atoi(lengthStr)

	symbols, err := PromptForSelect("Include symbols?", []string{"Yes", "No"})
	if err != nil {
		return "", err
	}
	numbers, err := PromptForSelect("Include numbers?", []string{"Yes", "No"})
	if err != nil {
		return "", err
	}

	value, err := GeneratePassword(length, symbols == "Yes", numbers == "Yes")
	if err != nil {
		return "", err
	}

	CopyToClipboard(value)

	return value, nil
}

// PromptForCategory prompts the user for a category using the promptui library.
// It displays a label and allows the user to enter a value.
//
// Returns:
//   - string: The category entered by the user
//   - error: Any error that occurred during the input operation
func PromptForCategory() (string, error) {
	return PromptForInput("Enter category (optional)", "", nil)
}

// PromptForConfirm prompts the user for a yes/no confirmation.
//
// Parameters:
//   - prompt: The confirmation prompt to display
//
// Returns:
//   - bool: true if confirmed, false otherwise
//   - error: Any error that occurred during the prompt
func PromptForConfirm(prompt string) (bool, error) {
	p := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		if err == promptui.ErrAbort || err == promptui.ErrInterrupt {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// PromptForSecret prompts the user for a secret using the promptui library.
// It displays a label and allows the user to select a secret from a list.
//
// Parameters:
//   - secrets: The list of secrets to display in the selection menu
//
// Returns:
//   - store.Secret: The selected secret
//   - error: Any error that occurred during the selection operation
func PromptForSecret(secrets []store.Secret) (store.Secret, error) {
	if len(secrets) == 0 {
		return store.Secret{}, fmt.Errorf("no secrets available")
	}
	items := make([]string, len(secrets))
	for i, secret := range secrets {
		items[i] = fmt.Sprintf("%s (%s)", secret.Name, secret.Category)
	}
	choice, err := PromptForSelect("Select secret", items)
	if err != nil {
		return store.Secret{}, err
	}
	for _, secret := range secrets {
		if strings.HasPrefix(choice, secret.Name) {
			return secret, nil
		}
	}
	return store.Secret{}, fmt.Errorf("selected secret not found")
}

// PromptForCategoryFilter prompts the user for a category filter using the promptui library.
// It displays a label and allows the user to select a category from a list.
//
// Parameters:
//   - categories: The list of categories to display in the selection menu
//
// Returns:
//   - string: The selected category
//   - error: Any error that occurred during the selection operation
func PromptForCategoryFilter(categories []string) (string, error) {
	if len(categories) == 0 {
		return "", nil
	}
	items := append([]string{"All categories"}, categories...)
	return PromptForSelect("Select category to filter", items)
}

// EncryptFile encrypts a file using AES-GCM
//
// Parameters:
//   - data: The data to encrypt
//   - password: The password to use for encryption
//
// Returns:
//   - The encrypted data
//   - An error if encryption fails
func EncryptFile(data []byte, password string) ([]byte, error) {
	if password == "" {
		return data, nil
	}

	key := []byte(password)
	if len(key) < 32 {
		key = append(key, bytes.Repeat([]byte{0}, 32-len(key))...)
	} else {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// DecryptFile decrypts a file using AES-GCM
//
// Parameters:
//   - data: The data to decrypt
//   - password: The password to use for decryption
//
// Returns:
//   - The decrypted data
//   - An error if decryption fails
func DecryptFile(data []byte, password string) ([]byte, error) {
	if len(data) < 12 {
		return data, nil
	}

	key := []byte(password)
	if len(key) < 32 {
		key = append(key, bytes.Repeat([]byte{0}, 32-len(key))...)
	} else {
		key = key[:32]
	}

	nonce := data[:12]
	ciphertext := data[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}
