# JWT-Crack

A high-performance JWT secret bruteforcer with web interface for security testing.

## ⚠️ Legal Disclaimer

**This tool is for authorized security testing only.** Only use on systems you own or have explicit written permission to test. Unauthorized use is illegal.

## Installation

```bash
git clone https://github.com/security-tools/jwt-crack.git
cd jwt-crack
make build
```

## Usage

### Web Interface (Recommended)

Start the web server:
```bash
jwt-crack serve --port 8080
```

Open `http://localhost:8080` in your browser for the interactive interface.

### Command Line Interface

#### Validate JWT Token
```bash
jwt-crack validate --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Smart Attack (Try First)
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --smart
```

#### Wordlist Attack
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --wordlist /path/to/wordlist.txt
```

#### Brute Force Attack
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." --charset password --length-min 1 --length-max 6
```

### Command Reference

```
jwt-crack [command] [flags]

Commands:
  crack       Crack JWT secret
  serve       Start web interface
  validate    Validate JWT token
  version     Show version

Crack Flags:
  -t, --token string       JWT token to crack (required)
      --smart              Use smart attack patterns
  -w, --wordlist string    Wordlist file path
  -c, --charset string     Charset: digits, alpha, password, full (default "password")
      --length-min int     Minimum length (default 1)
      --length-max int     Maximum length (default 8)
      --threads int        Number of threads (default: CPU cores)
      --performance string Performance: eco, balanced, performance, maximum (default "balanced")
      --timeout duration   Attack timeout (0 = no timeout)
  -o, --output string      Output file

Serve Flags:
      --port int           Web server port (default 8080)

Global Flags:
  -v, --verbose            Enable verbose logging
```

### Examples

**Basic smart attack:**
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiJ9..." --smart
```

**High-performance wordlist attack:**
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiJ9..." --wordlist common-passwords.txt --threads 16 --performance maximum
```

**Brute force with timeout:**
```bash
jwt-crack crack --token "eyJhbGciOiJIUzI1NiJ9..." --charset digits --length-min 4 --length-max 6 --timeout 5m
```

### Character Sets

- **digits**: `0123456789`
- **alpha**: `a-zA-Z`
- **password**: `a-zA-Z0-9!@#$%^&*`
- **full**: All printable ASCII characters

### Performance Levels

- **eco**: 25% CPU usage
- **balanced**: 50% CPU usage
- **performance**: 100% CPU usage
- **maximum**: 200% CPU usage

---

**Use responsibly and only on authorized systems.**