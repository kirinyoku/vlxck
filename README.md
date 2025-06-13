# vlxck - Secure Command-Line Password Manager

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

vlxck is a secure, lightweight command-line password manager that helps you store and manage sensitive information with strong encryption. It's designed to be simple, fast, and secure, with all data encrypted before being written to disk.

## Features

- 🔒 End-to-end encryption using AES-256-GCM
- 🔑 Secure master password protection
- 🔄 Easy password generation with customizable complexity
- 📂 Organized storage with categories
- 💻 Cross-platform support
- 🌐 No internet connection required

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
vlxck add -n example.com -v yourpassword -c websites
```

Options:
- `-n, --name`: Name/identifier for the secret (required)
- `-v, --value`: The secret value (required)
- `-c, --category`: Category for organization (optional)

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
