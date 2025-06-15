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
- [Security](#security)
- [License](#license)
- [Contributing](#contributing)

## Features

- 🔒 **Secure Storage**: End-to-end encryption using AES-256-GCM
- 🔑 **Password Protection**: Secure master password with Argon2id key derivation
- 🔄 **Password Generation**: Create strong, customizable passwords
- 📂 **Organization**: Categorize and manage secrets efficiently
- 🔄 **Seamless Updates**: Modify existing secrets with ease
- 💾 **Backup & Restore**: Export and import your encrypted store
- 🔍 **Quick Access**: Retrieve secrets instantly when needed
- 🚫 **Offline-First**: No internet connection required
- 💻 **Cross-Platform**: Works on Windows, macOS, and Linux

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
vlxck add -n example.com -v yourpassword
```

You'll be prompted to set a master password. This password will be required to access your secrets.

## Usage

### Add a New Secret

```bash
# Add a secret with a specific value
vlxck add -n example.com -v yourpassword -c websites

# Or generate a random password
vlxck add -n example.com -g -c websites
```

Options:
- `-n, --name`: Name/identifier for the secret (required)
- `-v, --value`: The secret value (either this or --generate is required)
- `-g, --generate`: Generate a random password for the secret
- `-c, --category`: Category for organization (optional)

Note: If both `--value` and `--generate` are specified, `--value` takes precedence.

### Update an Existing Secret

```bash
# Update a secret with a new value
vlxck update -n example.com -v newpassword

# Update just the category
vlxck update -n example.com -c social-media

# Generate a new random password for the secret
vlxck update -n example.com -g

# Update both value and category
vlxck update -n example.com -v newpassword -c work
```

Options:
- `-n, --name`: Name/identifier of the secret to update (required)
- `-v, --value`: New secret value (either this or --generate is required)
- `-g, --generate`: Generate a new random password for the secret
- `-c, --category`: Update the category (optional)

Note: If both `--value` and `--generate` are specified, `--value` takes precedence.

### Retrieve a Secret

```bash
vlxck get -n example.com
```

### List All Secrets

```bash
# List all secrets
vlxck list

# Filter by category
vlxck list -c websites
```

### Generate a Strong Password

```bash
# Generate a 16-character password with symbols and numbers
vlxck generate -l 16 -s -n
```

Options:
- `-l, --length`: Length of the password (default: 12)
- `-s, --symbols`: Include special characters
- `-n, --numbers`: Include digits

### Delete a Secret

```bash
vlxck delete -n example.com
```

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

## Security

- All data is encrypted using AES-256-GCM
- Master password is never stored
- Uses Argon2id for key derivation
- Encrypted data is stored in `~/.vlxck/store.dat`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request
