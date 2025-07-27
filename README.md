# JWT-Crack

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tests](https://github.com/security-tools/jwt-crack/workflows/Tests/badge.svg)](https://github.com/security-tools/jwt-crack/actions)

A high-performance JWT secret bruteforcer with web interface, designed for security testing and penetration testing.

## ⚠️ Legal Disclaimer

**This tool is intended for authorized security testing and educational purposes only.** Only use on systems you own or have explicit written permission to test. Unauthorized use is illegal and unethical. The developers are not responsible for any misuse of this tool.

## ✨ Features

### Core Capabilities
- **Multiple Attack Methods**: Smart patterns, wordlist attacks, and charset bruteforce
- **Web Interface**: Modern, interactive web UI with real-time progress tracking
- **High Performance**: Multi-threaded with configurable performance levels and optimized speed calculations
- **Advanced Character Sets**: Mix & match character sets, hashcat-style rules support
- **Comprehensive Validation**: Robust input validation and error handling
- **Professional Logging**: Structured logging with multiple levels

### Web Interface Features
- **JWT Token Analysis**: Decode and analyze JWT structure before attacking
- **Real-time Progress**: Live speed metrics, progress tracking, and ETA calculations
- **Multiple Attack Modes**: Smart, wordlist, and brute force attacks through web UI
- **Character Set Selection**: Interactive preset selection, mix & match, and custom rules
- **System Information**: CPU, memory, and performance monitoring
- **WebSocket Support**: Real-time updates with HTTP polling fallback

### Algorithm Support
- **HS256** (HMAC-SHA256)
- **HS384** (HMAC-SHA384) 
- **HS512** (HMAC-SHA512)

## 🚀 Installation

### Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/security-tools/jwt-crack/releases).

### Build from Source

```bash
git clone https://github.com/security-tools/jwt-crack.git
cd jwt-crack
make build
```

### Alternative Build Methods

```bash
# Direct Go build
go build -o jwt-crack ./cmd/jwt-crack

# Using Go install
go install github.com/security-tools/jwt-crack/cmd/jwt-crack@latest

# Build with version info
make install  # Install dependencies
make build    # Build with version tags
```

## 📋 Requirements

- **Go 1.22+** (for building from source)
- **Supported platforms**: Linux, macOS, Windows
- **Memory**: Minimum 512MB RAM
- **CPU**: Multi-core recommended for optimal performance

## 🎯 Quick Start

### Web Interface (Recommended)

```bash
# Start the web server
jwt-crack serve --port 8080

# Open in browser
open http://localhost:8080
```

### Command Line Interface

#### 1. Validate JWT Token
```bash
jwt-crack validate --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 2. Smart Attack (Recommended First Step)
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --smart
```

#### 3. Wordlist Attack
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --wordlist /path/to/wordlist.txt
```

#### 4. Charset Bruteforce
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --charset password --length-min 1 --length-max 6
```

## 📖 Detailed Usage

### Command Overview

```
jwt-crack [command] [flags]

Available Commands:
  crack       Crack JWT secret using various attack methods
  serve       Start web interface server
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

### Serve Command

```bash
jwt-crack serve [flags]

Flags:
      --port int   Web server port (default 8080)
```

### Performance Levels

- **eco**: 25% CPU usage, environmentally friendly
- **balanced**: 50% CPU usage, good performance/resource balance  
- **performance**: 100% CPU usage, maximum single-machine performance
- **maximum**: 200% CPU usage, overclocking mode (use with caution)

### Charset Options

- **digits**: `0123456789`
- **alpha**: `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
- **password**: `a-zA-Z0-9!@#$%^&*`
- **full**: All printable ASCII characters

### Hashcat-Style Rules (Web Interface)

The web interface supports hashcat-style character set rules:
- `?l` = lowercase letters (a-z)
- `?u` = uppercase letters (A-Z)  
- `?d` = digits (0-9)
- `?s` = special characters (!@#$...)
- `?a` = all printable characters

Example: `?l?l?d?d` = two lowercase letters followed by two digits

## 📊 Examples

### Basic Smart Attack

```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c" --smart
```

Output:
```
[2025-07-27 20:31:21] INFO  🔍 Token analyzed: valid JWT with HS256 algorithm
[2025-07-27 20:31:21] INFO  🎯 Attack started: HS256 algorithm, smart mode, 10 threads
[2025-07-27 20:31:21] INFO  ✅ Attack successful: found secret 'your-256-bit-secret' after 15 attempts in 151ms
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
  --wordlist ./examples/common-secrets.txt \
  --threads 16 \
  --performance maximum \
  --output results.json \
  --verbose
```

### Custom Charset Bruteforce with Timeout

```bash
jwt-crack crack \
  --token "eyJhbGciOiJIUzI1NiJ9..." \
  --charset digits \
  --length-min 4 \
  --length-max 6 \
  --threads 8 \
  --timeout 5m
```

### Web Interface Usage

```bash
# Start web server
jwt-crack serve --port 8080

# Advanced configuration
jwt-crack serve --port 3000 --verbose
```

Then navigate to `http://localhost:8080` in your browser for the interactive interface.

## 📁 Project Structure

```
jwt-crack/
├── cmd/jwt-crack/           # Main application entry point
├── pkg/                     # Public libraries
│   ├── config/             # Configuration management
│   ├── engine/             # Core attack engine
│   ├── logger/             # Structured logging
│   ├── system/             # System information
│   ├── validator/          # Input validation
│   └── web/                # Web interface server
├── internal/               # Private application code
│   ├── constants/          # Application constants
│   └── errors/             # Custom error types
├── examples/               # Example wordlists and configs
├── Makefile               # Build automation
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── LICENSE               # MIT license
```

## 🧪 Testing

Run the complete test suite:

```bash
# Run all tests
go test ./...
make test

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
- Make (optional but recommended)

### Building

```bash
# Build for current platform
make build

# Clean build artifacts
make clean

# Install dependencies
make install

# Run demo
make demo

# Start development server
make server
```

### Cross-Platform Building

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o jwt-crack-linux ./cmd/jwt-crack

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o jwt-crack.exe ./cmd/jwt-crack

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o jwt-crack-macos ./cmd/jwt-crack
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Run security checks (requires gosec)
gosec ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Reporting Issues

- Use the [GitHub issue tracker](https://github.com/security-tools/jwt-crack/issues)
- Include detailed steps to reproduce
- Provide sample tokens (non-sensitive only)
- Include system information and logs

### Feature Requests

- Open a GitHub issue with the `enhancement` label
- Describe the use case and expected behavior
- Consider implementing the feature yourself via pull request

## 🔒 Security Considerations

### Best Practices

- **Always use on authorized systems only**
- **Never use production JWT tokens for testing**
- **Rotate secrets immediately after successful attacks**
- **Use appropriate performance levels to avoid system overload**

### Rate Limiting

The tool includes built-in rate limiting and performance controls to prevent system abuse. Always respect system resources and network policies.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for excellent CLI framework
- [Gorilla](https://github.com/gorilla) for WebSocket and HTTP router libraries
- The security research community for feedback and testing

## 🔗 Related Projects

- [jwt.io](https://jwt.io/) - JWT debugger and token information
- [hashcat](https://hashcat.net/) - Advanced password recovery tool
- [john](https://www.openwall.com/john/) - John the Ripper password cracker
- [SecLists](https://github.com/danielmiessler/SecLists) - Security tester's wordlists

## 📞 Support

- 📖 [Documentation](https://github.com/security-tools/jwt-crack/wiki)
- 🐛 [Issues](https://github.com/security-tools/jwt-crack/issues)
- 💬 [Discussions](https://github.com/security-tools/jwt-crack/discussions)

## 🚨 Performance Notes

- **Typical speeds**: 1K-1M passwords/second depending on hardware
- **Memory usage**: ~50-200MB depending on wordlist size
- **CPU usage**: Configurable via performance levels
- **Network**: Web interface uses minimal bandwidth with WebSocket updates

---

**Remember: Use responsibly and only on systems you are authorized to test.**