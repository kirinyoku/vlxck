# vlxck - Secure Command-Line Password Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

vlxck is a secure, lightweight command-line password manager that helps you store and manage sensitive information with strong encryption. It's designed to be simple, fast, and secure, with all data encrypted before being written to disk.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [From Source](#from-source)
  - [Using Go Install](#using-go-install)
- [Getting Started](#getting-started)
  - [Initialize a New Store](#initialize-a-new-store)
- [Usage](#usage)
  - [Add a New Secret](#add-a-new-secret)
  - [Update an Existing Secret](#update-an-existing-secret)
  - [Retrieve a Secret](#retrieve-a-secret)
  - [List All Secrets](#list-all-secrets)
  - [Generate a Strong Password](#generate-a-strong-password)
  - [Delete a Secret](#delete-a-secret)
  - [Change Master Password](#change-master-password)
  - [Export Your Secrets](#export-your-secrets)
  - [Import Secrets](#import-secrets)
  - [Backup and Restore](#backup-and-restore)
    - [Create a Backup](#create-a-backup)
    - [List Available Backups](#list-available-backups)
    - [Restore from Backup](#restore-from-backup)
- [Security](#security)
- [License](#license)
- [Contributing](#contributing)

## Features

- 🔒 **Secure Storage**: End-to-end encryption using AES-256-GCM
- 🔑 **Password Protection**: Secure master password with Argon2id key derivation
- 🔄 **Password Generation**: Create strong, customizable passwords
- 📂 **Organization**: Categorize and manage secrets efficiently
- 🔄 **Seamless Updates**: Modify existing secrets with ease
- 💾 **Export & Import**: Export and import your encrypted store
- 🔍 **Quick Access**: Retrieve secrets instantly when needed
- 🚫 **Offline-First**: No internet connection required
- 💻 **Cross-Platform**: Works on Windows, macOS, and Linux
- 💾 **Backups**: Create and manage backups
- 🔄 **Easy Restore**: Restore from any previous backup with a single command

## Installation

### Prerequisites

- Go 1.16 or later
- Git (for building from source)

### From Source

```bash
# Clone the repository
git clone https://github.com/kirinyoku/vlxck.git
cd vlxck

# Build the binary
go build -o vlxck

# Move the binary to your PATH
sudo mv vlxck /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/kirinyoku/vlxck@latest
```

## Getting Started

### Initialize a New Store

When you run your first command, vlxck will automatically create a new encrypted store:

```bash
vlxck add -n example.com -V yourpassword
```

You'll be prompted to set a master password. This password will be required to access your secrets.

## Usage

### Add a New Secret

```bash
# Add a secret with a specific value
vlxck add -n example.com -V yourpassword -c websites

# Generate a random password with default settings (16 chars, letters only)
vlxck add -n example.com -g -c websites

# Generate a custom password (24 chars with symbols and digits)
vlxck add -n example.com -gdsl 24 -c websites

# Interactive mode (guided prompts)
vlxck add -i
```

Options:
- `-n, --name`: Name/identifier for the secret (required in non-interactive mode)
- `-V, --value`: The secret value (either this or --generate is required in non-interactive mode)
- `-g, --generate`: Generate a random password (overrides -V if both specified)
- `-l, --length`: Length of the generated password (default: 16)
- `-s, --symbols`: Include special characters in generated password
- `-d, --digits`: Include digits in generated password
- `-c, --category`: Category for organization (optional)
- `-i, --interactive`: Use interactive mode (overrides other flags)

Password Generation Examples:
- `-g`: 16-character letters-only password
- `-gsd`: 16-character password with symbols and digits
- `-gdsl 24`: 24-character password with digits and symbols
- `-gsl 20`: 20-character password with symbols (no digits)

### Update an Existing Secret

```bash
# Update a secret with a new value
vlxck update -n example.com -V newpassword

# Update just the category
vlxck update -n example.com -c social-media

# Generate a new random password for the secret
vlxck update -n example.com -g

# Update both value and category
vlxck update -n example.com -V newpassword -c work

# Interactive mode (guided prompts)
vlxck update -i
```

Options:
- `-n, --name`: Name/identifier of the secret to update (required in non-interactive mode)
- `-V, --value`: New secret value (either this or --generate is required in non-interactive mode)
- `-g, --generate`: Generate a new random password for the secret
- `-l, --length`: Length of the generated password (default: 16)
- `-s, --symbols`: Include special characters in generated password
- `-d, --digits`: Include digits in generated password
- `-c, --category`: Update the category (optional)
- `-i, --interactive`: Use interactive mode (overrides other flags)

### Retrieve a Secret

Retrieve a secret and automatically copy it to your clipboard:

```bash
# Interactive mode - select from a list of secrets
vlxck get -i

# Non-interactive mode - specify the secret name
vlxck get -n example.com
```

Options:
- `-n, --name`: Name/identifier of the secret to retrieve (required in non-interactive mode)
- `-i, --interactive`: Use interactive mode to select from a list of secrets

The secret value will be copied to your clipboard automatically. This helps prevent accidentally displaying sensitive information in your terminal history or on screen.

### List All Secrets

```bash
# List all secrets
vlxck list

# Filter by category
vlxck list -c websites
```

### Generate a Strong Password

```bash
# Generate a 24-character password with symbols and digits
vlxck generate -sdl 24
```

Options:
- `-l, --length`: Length of the password (default: 16)
- `-s, --symbols`: Include special characters
- `-d, --digits`: Include digits

### Delete a Secret

Delete a secret from the store by its name or through interactive selection:

```bash
# Interactive mode - select secret to delete from a list
vlxck delete -i

# Non-interactive mode - specify the secret name
vlxck delete -n example.com
```

Options:
- `-n, --name`: Name of the secret to delete (required in non-interactive mode)
- `-i, --interactive`: Use interactive mode to select from a list of secrets

Interactive Mode:
When using interactive mode (`-i`), you'll be guided through the deletion process:
1. Select which secret to delete from a list
2. Confirm the deletion to prevent accidental data loss
3. The secret will be permanently removed if confirmed

### Change Master Password

```bash
vlxck change-master
```

### Export Your Secrets

Export your encrypted secrets to a backup location:

```bash
# Export to a specific directory
vlxck export -d /path/to/backup/directory
```

This will create a `store.dat` file in the specified directory containing your encrypted secrets.

Options:
- `-d, --dir`: Directory to export the store file to (required)

### Import Secrets

Import secrets from an encrypted backup file. You can choose to either replace your current store or merge with existing secrets:

```bash
# Basic import (replaces current store)
vlxck import -f /path/to/backup/store.dat

# Use the current store's password for import
vlxck import -f /path/to/backup/store.dat -p

# Merge with existing secrets (interactive conflict resolution)
vlxck import -f /path/to/backup/store.dat -m

# Merge using the current store's password
vlxck import -f /path/to/backup/store.dat -m -p
```

#### Merge Behavior
When using the merge option (`-m`), the import process will:
1. Keep all unique secrets from both the current store and import file
2. For secrets that exist in both:
   - Show a comparison of both versions
   - Prompt you to choose an action:
     - `[l]` Keep local version
     - `[i]` Use imported version
     - `[s]` Skip this secret (neither version will be included)

Options:
- `-f, --file`: Path to the import file (required)
- `-p, --use-store-password`: Use the current store's master password for import
- `-m, --merge`: Merge secrets from import file into existing store (interactive)

**Warning:** Without the `-m` flag, this will replace your current store with the imported one. Make sure you have a backup if needed.

## Backup and Restore

vlxck provides robust backup and restore functionality to keep your data safe.

### Create a Backup

Create a compressed backup of your entire password store:

```bash
# Create a backup in the default location (~/.vlxck/backups/)
vlxck backup

# Create a backup in a specific directory
vlxck backup /path/to/backup/directory
```

Backups are stored as timestamped zip archives (e.g., `backup_20250620-183238.zip`).

### List Available Backups

View all available backups with their sizes and creation times:

```bash
# List backups in the default location
vlxck list-backups

# List backups in a specific directory
vlxck list-backups /path/to/backup/directory
```

### Restore from Backup

Restore your password store from a previous backup:

```bash
# Interactive mode - choose from a list of available backups
vlxck restore -i

# Restore a specific backup file
vlxck restore /path/to/backup/backup_20250620-183238.zip

# Restore to a specific directory
vlxck restore -i /custom/restore/path
```

Options:
- `-i, --interactive`: Show an interactive menu to select from available backups
- `[backup-file]`: Path to a specific backup file to restore from
- `[target-dir]`: (Optional) Directory to restore the backup to (default: ~/.vlxck)

## Security

- All data is encrypted using AES-256-GCM
- Uses Argon2id for key derivation
- Encrypted data is stored in `~/.vlxck/store.dat`

### Password Caching

vlxck caches your master password in memory for 5 minutes after successful verification to improve usability. This means you won't need to re-enter your password for subsequent commands within this time window.

- The password is stored securely in an encrypted cache file
- The cache is automatically cleared after 5 minutes
- You can clear the cache immediately by pressing Ctrl+C or waiting for the timeout
- The cache is never written to disk in plaintext

For maximum security, the cache is also cleared when:
- You change your master password
- The program is interrupted
- The system is shut down or restarted

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
