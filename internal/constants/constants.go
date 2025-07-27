package constants

import "time"

// Application constants
const (
	AppName     = "jwt-crack"
	AppVersion  = "2.0.0"
	AppDesc     = "A high-performance JWT secret bruteforcer for security testing"
	AppAuthor   = "Security Research Team"
	AppRepo     = "https://github.com/security-tools/jwt-crack"
)

// Algorithm constants
const (
	AlgorithmHS256 = "HS256"
	AlgorithmHS384 = "HS384"
	AlgorithmHS512 = "HS512"
)

// Attack mode constants
const (
	AttackModeSmart   = "smart"
	AttackModeWordlist = "wordlist"
	AttackModeCharset = "charset"
)

// Charset constants
const (
	CharsetDigits   = "digits"
	CharsetAlpha    = "alpha" 
	CharsetPassword = "password"
	CharsetFull     = "full"
)

// Performance level constants
const (
	PerformanceEco    = "eco"
	PerformanceBalanced = "balanced"
	PerformanceHigh   = "performance"
	PerformanceMaximum = "maximum"
)

// Default values
const (
	DefaultThreads          = 0 // Will be set to runtime.NumCPU()
	DefaultLengthMin        = 1
	DefaultLengthMax        = 8
	DefaultWebPort          = 8080
	DefaultPerformance      = PerformanceBalanced
	DefaultCharset          = CharsetPassword
	DefaultProgressInterval = 100 * time.Millisecond
	DefaultRequestTimeout   = 30 * time.Second
)

// Limits and constraints
const (
	MaxThreads           = 64
	MaxPasswordLength    = 20
	MinPasswordLength    = 1
	MaxFileSize          = 100 * 1024 * 1024 // 100MB
	MaxUploadSize        = 10 * 1024 * 1024  // 10MB
	MaxLineLength        = 1000              // Max characters per wordlist line
	MinWebPort           = 1024              // Avoid privileged ports
	MaxWebPort           = 65535
	SmartAttackDelay     = 10 * time.Millisecond
	WordlistProgressFreq = 100               // Report progress every N attempts
	CharsetProgressFreq  = 1000              // Report progress every N attempts
)

// Individual character sets
var (
	CharsetLowercase      = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercase      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetDigitsOnly     = "0123456789"
	CharsetSpecial        = "!@#$%^&*()-_=+[]{}|;:'\",.<>?/\\`"
	CharsetPrintableASCII = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	CharsetHex            = "0123456789abcdef"
	CharsetBase64         = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	CharsetAlphanumeric   = CharsetLowercase + CharsetUppercase + CharsetDigitsOnly
	CharsetMixed          = CharsetLowercase + CharsetUppercase
)

// Character sets
var Charsets = map[string]string{
	CharsetDigits:     CharsetDigitsOnly,
	CharsetAlpha:      CharsetMixed,
	CharsetPassword:   CharsetLowercase + CharsetUppercase + CharsetDigitsOnly + "!@#$%^&*",
	CharsetFull:       CharsetLowercase + CharsetUppercase + CharsetDigitsOnly + CharsetSpecial,
	"lowercase":       CharsetLowercase,
	"uppercase":       CharsetUppercase,
	"mixed":           CharsetMixed,
	"alphanumeric":    CharsetAlphanumeric,
	"special":         CharsetSpecial,
	"printable":       CharsetPrintableASCII,
	"hex":             CharsetHex,
	"base64":          CharsetBase64,
}

// Hashcat-style charset rules
var HashcatCharsets = map[string]string{
	"?l": CharsetLowercase,
	"?u": CharsetUppercase,
	"?d": CharsetDigitsOnly,
	"?s": CharsetSpecial,
	"?a": CharsetPrintableASCII,
}

// Smart attack patterns - common JWT secrets
var SmartPatterns = []string{
	"secret", "password", "123456", "admin", "test", "key", "jwt", "token",
	"secretkey", "jwtkey", "mysecret", "supersecret", "qwerty", "password123",
	"your-256-bit-secret", "your-secret", "secret-key", "jwt-secret",
	"", "null", "undefined", "your_secret_here", "change_me", "default",
	"demo", "example", "sample", "temp", "temporary", "dev", "development",
	"prod", "production", "staging", "testing", "debug", "localhost",
}

// Supported file extensions
var SupportedWordlistExtensions = []string{
	".txt", ".list", ".dic", ".dict",
}

var SupportedOutputExtensions = []string{
	".json", ".csv", ".txt",
}

// Web interface constants
const (
	WebStaticPath    = "/static/"
	WebAPIPath       = "/api/"
	WebSocketPath    = "/ws"
	MaxWebSocketConns = 100
	WebSocketTimeout  = 60 * time.Second
)

// Logging constants
const (
	LogTimeFormat = "2006-01-02 15:04:05"
	LogMaxSize    = 100 * 1024 * 1024 // 100MB
)

// Output format constants
const (
	OutputFormatJSON = "json"
	OutputFormatCSV  = "csv"
	OutputFormatTXT  = "txt"
)