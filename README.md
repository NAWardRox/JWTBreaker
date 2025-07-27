# JWT-Crack

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tests](https://github.com/security-tools/jwt-crack/workflows/Tests/badge.svg)](https://github.com/security-tools/jwt-crack/actions)

A high-performance JWT secret bruteforcer designed for security testing and penetration testing.

## ⚠️ Legal Disclaimer

This tool is intended for authorized security testing and educational purposes only. Only use on systems you own or have explicit written permission to test. Unauthorized use is illegal and unethical.

## ✨ Features

- **Multiple Attack Methods**: Smart patterns, wordlist attacks, and charset bruteforce
- **High Performance**: Multi-threaded with configurable performance levels
- **Comprehensive Validation**: Robust input validation and error handling  
- **Professional Logging**: Structured logging with multiple levels
- **Format Support**: JSON, CSV, and TXT output formats
- **Modern Architecture**: Clean, maintainable Go codebase with full test coverage

## 🚀 Installation

### Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/security-tools/jwt-crack/releases).

### Build from Source

```bash
git clone https://github.com/security-tools/jwt-crack.git
cd jwt-crack
go build -o jwt-crack ./cmd/jwt-crack
```

### Using Go Install

```bash
go install github.com/security-tools/jwt-crack/cmd/jwt-crack@latest
```

## 📋 Requirements

- Go 1.22 or higher (for building from source)
- Supported platforms: Linux, macOS, Windows

## 🎯 Quick Start

### 1. Validate a JWT Token

```bash
jwt-crack validate --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. Smart Attack (Recommended First Step)

```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --smart
```

### 3. Wordlist Attack

```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --wordlist /path/to/wordlist.txt
```

### 4. Charset Bruteforce

```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --charset password --length-min 1 --length-max 6
```

## 📖 Usage

### Command Overview

```
jwt-crack [command] [flags]

Available Commands:
  crack       Crack JWT secret using various attack methods
  serve       Start web interface server (coming soon)
  validate    Validate JWT token format and structure
  version     Show version information

Global Flags:
  -v, --verbose   Enable verbose logging
      --config    Configuration file path
```

### Crack Command

```bash
jwt-crack crack [flags]

Required Flags:
  -t, --token string    JWT token to crack

Attack Method Flags:
      --smart           Use smart attack with common patterns
  -w, --wordlist string Wordlist file path
  -c, --charset string  Charset: digits, alpha, password, full (default "password")

Configuration Flags:
      --length-min int     Minimum password length (default 1)
      --length-max int     Maximum password length (default 8)
      --threads int        Number of concurrent threads (default: CPU cores)
      --performance string Performance level: eco, balanced, performance, maximum (default "balanced")
      --timeout duration   Attack timeout (0 = no timeout)

Output Flags:
  -o, --output string   Output file (JSON/CSV/TXT)
```

### Performance Levels

- **eco**: 25% CPU usage, environmentally friendly
- **balanced**: 50% CPU usage, good performance/resource balance  
- **performance**: 100% CPU usage, maximum single-machine performance
- **maximum**: 200% CPU usage, overclocking mode

### Charset Options

- **digits**: `0123456789`
- **alpha**: `a-zA-Z`
- **password**: `a-zA-Z0-9!@#$%^&*`
- **full**: All printable ASCII characters

## 📊 Examples

### Basic Smart Attack

```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c" --smart
```

Output:
```
[2025-07-26 20:31:21] INFO  🔍 Token analyzed: valid JWT with HS256 algorithm
[2025-07-26 20:31:21] INFO  🎯 Attack started: HS256 algorithm, smart mode, 10 threads
[2025-07-26 20:31:21] INFO  ✅ Attack successful: found secret 'your-256-bit-secret' after 15 attempts in 151ms
──────────────────────────────────────────────────
✅ SECRET FOUND!
Secret: your-256-bit-secret
Algorithm: HS256
Attack Mode: smart
Attempts: 15
Duration: 151.265792ms
──────────────────────────────────────────────────
```

### High-Performance Wordlist Attack

```bash
jwt-crack crack \
  --token "eyJhbGciOiJIUzI1NiJ9..." \
  --wordlist ./wordlists/common-passwords.txt \
  --threads 16 \
  --performance maximum \
  --output results.json \
  --verbose
```

### Custom Charset Bruteforce

```bash
jwt-crack crack \
  --token "eyJhbGciOiJIUzI1NiJ9..." \
  --charset digits \
  --length-min 4 \
  --length-max 6 \
  --threads 8 \
  --timeout 5m
```

## 📁 Project Structure

```
jwt-crack/
├── cmd/jwt-crack/           # Main application entry point
├── pkg/                     # Public libraries
│   ├── config/             # Configuration management
│   ├── engine/             # Core attack engine
│   ├── logger/             # Structured logging
│   └── validator/          # Input validation
├── internal/               # Private application code
│   ├── constants/          # Application constants
│   └── errors/             # Custom error types
├── examples/               # Example wordlists and configs
├── wordlists/              # Built-in wordlists
└── .github/workflows/      # CI/CD pipelines
```

## 🧪 Testing

Run the complete test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/engine -v

# Run benchmarks
go test -bench=. ./pkg/engine
```

## 🔧 Development

### Prerequisites

- Go 1.22+
- Git

### Building

```bash
# Build for current platform
go build -o jwt-crack ./cmd/jwt-crack

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o jwt-crack-linux ./cmd/jwt-crack
GOOS=windows GOARCH=amd64 go build -o jwt-crack.exe ./cmd/jwt-crack
GOOS=darwin GOARCH=arm64 go build -o jwt-crack-macos ./cmd/jwt-crack
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run tests
go test ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Reporting Issues

- Use the [GitHub issue tracker](https://github.com/security-tools/jwt-crack/issues)
- Include detailed steps to reproduce
- Provide sample tokens (non-sensitive)
- Include system information and logs

### Feature Requests

- Open a GitHub issue with the `enhancement` label
- Describe the use case and expected behavior
- Consider implementing the feature yourself

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [ProjectDiscovery](https://github.com/projectdiscovery) for inspiration on CLI design
- [Cobra](https://github.com/spf13/cobra) for excellent CLI framework
- The security research community for feedback and testing

## 🔗 Related Projects

- [jwt.io](https://jwt.io/) - JWT debugger and token information
- [hashcat](https://hashcat.net/) - Advanced password recovery tool
- [john](https://www.openwall.com/john/) - John the Ripper password cracker

## 📞 Support

- 📖 [Documentation](https://github.com/security-tools/jwt-crack/wiki)
- 🐛 [Issues](https://github.com/security-tools/jwt-crack/issues)
- 💬 [Discussions](https://github.com/security-tools/jwt-crack/discussions)

---

**Remember: Use responsibly and only on systems you are authorized to test.**