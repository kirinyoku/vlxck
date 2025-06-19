// Package cmd implements the command-line interface for the secure secret manager.
// This file contains the implementation of the 'list' command which is used to
// list all secrets in a formatted table with pagination.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kirinyoku/vlxck/internal/store"
	"github.com/spf13/cobra"
)

const (
	SecretsPerPage = 10           // Number of secrets to display per page
	MaxNameLen     = 20           // Maximum length for Name column
	MaxCategoryLen = 20           // Maximum length for Category column
	IndexWidth     = 5            // Width for Index column (e.g., "1")
	HeaderColor    = "\033[1;34m" // ANSI color for headers (blue)
	ResetColor     = "\033[0m"    // Reset ANSI color
	BorderColor    = "\033[1;30m" // ANSI color for borders (gray)
)

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// listCmd represents the 'list' command that allows users to list all secrets in the store.
// It prompts the user for the master password and displays the list of secrets in a formatted table.
// If the store is not found, it displays an error message.
// The list is paginated, showing SecretsPerPage secrets at a time, with user input to navigate pages.
//
// The command supports the following flags:
//   - category (-c): Optional category for filtering secrets
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets",
	Long: `List all secrets in the store.

Examples:
  # List all secrets
  vlxck list

  # List secrets by category
  vlxck list -c games`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := getStorePath()

		password, err := getPassword(false)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		s, err := store.LoadStore(filePath, password)
		if err == nil {
			cacheVerifiedPassword(password)
		}
		if err != nil {
			fmt.Println("Error loading store:", err)
			return
		}

		category, _ := cmd.Flags().GetString("category")

		var filteredSecrets []store.Secret
		for _, secret := range s.Secrets {
			if category == "" || secret.Category == category {
				filteredSecrets = append(filteredSecrets, secret)
			}
		}

		if len(filteredSecrets) == 0 {
			fmt.Println("No secrets found" + func() string {
				if category != "" {
					return " for category " + category
				}
				return ""
			}() + ".")
			return
		}

		// Initialize pagination
		currentPage := 0
		totalPages := (len(filteredSecrets) + SecretsPerPage - 1) / SecretsPerPage
		scanner := bufio.NewScanner(os.Stdin)

		for {
			// Display current page
			displaySecretsPage(filteredSecrets, currentPage, totalPages)

			// Prompt for navigation if multiple pages
			if totalPages > 1 {
				fmt.Printf("\nPage %d of %d. Enter (n)ext, (p)revious, or (q)uit: ", currentPage+1, totalPages)
				if scanner.Scan() {
					input := strings.ToLower(strings.TrimSpace(scanner.Text()))
					switch input {
					case "n":
						if currentPage < totalPages-1 {
							currentPage++
						}
					case "p":
						if currentPage > 0 {
							currentPage--
						}
					case "q":
						return
					default:
						fmt.Println("Invalid input. Use 'n' for next, 'p' for previous, or 'q' to quit.")
					}
				}
			} else {
				break
			}
		}
	},
}

// displaySecretsPage displays a single page of secrets in a formatted table.
func displaySecretsPage(secrets []store.Secret, currentPage, totalPages int) {
	// Initialize tabwriter with left alignment and consistent padding
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Define column headers
	headerIndex := "Index"
	headerName := "Name"
	headerCategory := "Category"

	// Calculate border lengths based on maximum content width
	borderIndex := strings.Repeat("─", IndexWidth)
	borderName := strings.Repeat("─", MaxNameLen)
	borderCategory := strings.Repeat("─", MaxCategoryLen)

	// Print top border
	fmt.Fprintf(w, "%s┌─%s─┬─%s─┬─%s─┐%s\n", BorderColor, borderIndex, borderName, borderCategory, ResetColor)

	// Print header row
	fmt.Fprintf(w, "%s│ %s%-*s%s │ %s%-*s%s │ %s%-*s%s │%s\n",
		BorderColor,
		HeaderColor, IndexWidth, headerIndex, ResetColor,
		HeaderColor, MaxNameLen, headerName, ResetColor,
		HeaderColor, MaxCategoryLen, headerCategory, ResetColor,
		ResetColor)

	// Print header separator
	fmt.Fprintf(w, "%s├─%s─┼─%s─┼─%s─┤%s\n", BorderColor, borderIndex, borderName, borderCategory, ResetColor)

	// Calculate page range
	start := currentPage * SecretsPerPage
	end := start + SecretsPerPage
	if end > len(secrets) {
		end = len(secrets)
	}

	// Print secrets
	for i, secret := range secrets[start:end] {
		name := truncateString(secret.Name, MaxNameLen)
		category := truncateString(secret.Category, MaxCategoryLen)
		fmt.Fprintf(w, "│ %-*d │ %-*s │ %-*s │\n", IndexWidth, start+i+1, MaxNameLen, name, MaxCategoryLen, category)
	}

	// Print bottom border
	fmt.Fprintf(w, "%s└─%s─┴─%s─┴─%s─┘%s\n", BorderColor, borderIndex, borderName, borderCategory, ResetColor)

	// Flush output to ensure proper rendering
	w.Flush()
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Define command flags with shorthand and descriptions
	listCmd.Flags().StringP("category", "c", "", "Filter by category")
}
