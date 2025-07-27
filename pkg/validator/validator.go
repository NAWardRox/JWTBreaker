package validator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"jwt-crack/internal/constants"
)

// JWTValidator provides JWT token validation
type JWTValidator struct{}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator() *JWTValidator {
	return &JWTValidator{}
}

// ValidateToken validates a JWT token format and structure
func (v *JWTValidator) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("JWT token cannot be empty")
	}
	
	// Remove any surrounding whitespace
	token = strings.TrimSpace(token)
	
	// Check basic format
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("JWT token must have exactly 3 parts separated by dots, got %d parts", len(parts))
	}
	
	// Validate each part is valid base64
	for i, part := range parts {
		if err := v.validateBase64URLPart(part, i); err != nil {
			return err
		}
	}
	
	// Validate header
	if err := v.validateHeader(parts[0]); err != nil {
		return fmt.Errorf("invalid JWT header: %w", err)
	}
	
	// Validate payload
	if err := v.validatePayload(parts[1]); err != nil {
		return fmt.Errorf("invalid JWT payload: %w", err)
	}
	
	return nil
}

// validateBase64URLPart validates a base64 URL encoded part
func (v *JWTValidator) validateBase64URLPart(part string, partIndex int) error {
	if part == "" {
		partNames := []string{"header", "payload", "signature"}
		return fmt.Errorf("JWT %s cannot be empty", partNames[partIndex])
	}
	
	// Check for invalid characters
	validChars := regexp.MustCompile(`^[A-Za-z0-9_-]*$`)
	if !validChars.MatchString(part) {
		return fmt.Errorf("JWT part %d contains invalid base64url characters", partIndex+1)
	}
	
	// Try to decode
	_, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return fmt.Errorf("JWT part %d is not valid base64url: %w", partIndex+1, err)
	}
	
	return nil
}

// validateHeader validates JWT header structure
func (v *JWTValidator) validateHeader(headerPart string) error {
	data, err := base64.RawURLEncoding.DecodeString(headerPart)
	if err != nil {
		return fmt.Errorf("failed to decode header: %w", err)
	}
	
	var header map[string]interface{}
	if err := json.Unmarshal(data, &header); err != nil {
		return fmt.Errorf("header is not valid JSON: %w", err)
	}
	
	// Check required fields
	alg, ok := header["alg"]
	if !ok {
		return fmt.Errorf("header missing required 'alg' field")
	}
	
	algStr, ok := alg.(string)
	if !ok {
		return fmt.Errorf("'alg' field must be a string")
	}
	
	// Validate supported algorithms
	supportedAlgs := map[string]bool{
		"HS256": true,
		"HS384": true,
		"HS512": true,
	}
	
	if !supportedAlgs[algStr] {
		return fmt.Errorf("unsupported algorithm: %s (supported: HS256, HS384, HS512)", algStr)
	}
	
	return nil
}

// validatePayload validates JWT payload structure
func (v *JWTValidator) validatePayload(payloadPart string) error {
	data, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return fmt.Errorf("failed to decode payload: %w", err)
	}
	
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("payload is not valid JSON: %w", err)
	}
	
	// Payload can be empty or contain any valid JSON
	return nil
}

// FileValidator provides file validation
type FileValidator struct {
	maxSize int64
}

// NewFileValidator creates a new file validator
func NewFileValidator(maxSize int64) *FileValidator {
	return &FileValidator{maxSize: maxSize}
}

// ValidateWordlistFile validates a wordlist file
func (v *FileValidator) ValidateWordlistFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("wordlist file path cannot be empty")
	}
	
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wordlist file does not exist: %s", filePath)
		}
		return fmt.Errorf("cannot access wordlist file: %w", err)
	}
	
	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("wordlist path is not a regular file: %s", filePath)
	}
	
	// Check file size
	if info.Size() > v.maxSize {
		return fmt.Errorf("wordlist file too large: %d bytes (max: %d bytes)", info.Size(), v.maxSize)
	}
	
	// Check file extension  
	ext := strings.ToLower(filepath.Ext(filePath))
	validExts := map[string]bool{
		".txt": true, ".list": true, ".dic": true, ".dict": true, "": true,
	}
	if !validExts[ext] {
		return fmt.Errorf("invalid wordlist file extension: %s (allowed: .txt, .list, .dic, .dict)", ext)
	}
	
	// Try to read a few lines to validate content
	if err := v.validateWordlistContent(filePath); err != nil {
		return fmt.Errorf("invalid wordlist content: %w", err)
	}
	
	return nil
}

// validateWordlistContent validates the content of a wordlist file
func (v *FileValidator) validateWordlistContent(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Read first 1KB to validate
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	content := string(buffer[:n])
	
	// Check if content is valid UTF-8
	if !utf8.ValidString(content) {
		return fmt.Errorf("file contains invalid UTF-8 characters")
	}
	
	// Check for extremely long lines (potential DoS)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len(line) > 1000 {
			return fmt.Errorf("line %d is too long (%d characters, max: 1000)", i+1, len(line))
		}
	}
	
	return nil
}

// InputValidator provides general input validation
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateCharset validates charset parameter
func (v *InputValidator) ValidateCharset(charset string) error {
	if charset == "" {
		return fmt.Errorf("charset cannot be empty")
	}
	
	// Check if it's a predefined charset name
	if _, exists := constants.Charsets[charset]; exists {
		return nil
	}
	
	// If it's not a predefined charset, treat it as a raw charset or hashcat rules
	processedCharset := charset
	
	// Check if it contains hashcat-style rules and process them
	for rule := range constants.HashcatCharsets {
		if strings.Contains(charset, rule) {
			processedCharset = v.ProcessHashcatCharset(charset)
			break
		}
	}
	
	// Validate processed charset length
	if len(processedCharset) == 0 {
		return fmt.Errorf("charset cannot be empty after processing")
	}
	
	if len(processedCharset) > 1000 {
		return fmt.Errorf("charset too long (%d characters, max: 1000)", len(processedCharset))
	}
	
	// Basic validation for raw charset characters
	if !utf8.ValidString(processedCharset) {
		return fmt.Errorf("charset contains invalid UTF-8 characters")
	}
	
	return nil
}

// ProcessHashcatCharset converts hashcat-style charset rules to actual characters
func (v *InputValidator) ProcessHashcatCharset(charset string) string {
	// Replace hashcat rules with actual character sets
	for rule, chars := range constants.HashcatCharsets {
		charset = strings.ReplaceAll(charset, rule, chars)
	}
	
	// Remove duplicate characters
	seen := make(map[rune]bool)
	var result []rune
	
	for _, char := range charset {
		if !seen[char] {
			seen[char] = true
			result = append(result, char)
		}
	}
	
	return string(result)
}

// ValidateLength validates length parameters
func (v *InputValidator) ValidateLength(min, max int) error {
	if min < 1 {
		return fmt.Errorf("minimum length must be at least 1, got %d", min)
	}
	
	if max < min {
		return fmt.Errorf("maximum length (%d) must be >= minimum length (%d)", max, min)
	}
	
	if max > 20 {
		return fmt.Errorf("maximum length cannot exceed 20, got %d", max)
	}
	
	return nil
}

// ValidateThreads validates thread count
func (v *InputValidator) ValidateThreads(threads int) error {
	if threads < 1 {
		return fmt.Errorf("thread count must be at least 1, got %d", threads)
	}
	
	if threads > 64 {
		return fmt.Errorf("thread count cannot exceed 64, got %d", threads)
	}
	
	return nil
}

// ValidatePerformance validates performance setting
func (v *InputValidator) ValidatePerformance(performance string) error {
	validSettings := map[string]bool{
		"eco":         true,
		"balanced":    true,
		"performance": true,
		"maximum":     true,
	}
	
	if !validSettings[performance] {
		return fmt.Errorf("invalid performance setting '%s' (valid: eco, balanced, performance, maximum)", performance)
	}
	
	return nil
}

// ValidateOutputPath validates output file path
func (v *InputValidator) ValidateOutputPath(path string) error {
	if path == "" {
		return nil // Optional parameter
	}
	
	dir := filepath.Dir(path)
	if dir != "." {
		// Check if directory exists or can be created
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", dir)
		}
	}
	
	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	validExts := map[string]bool{
		".json": true,
		".csv":  true,
		".txt":  true,
	}
	
	if !validExts[ext] {
		return fmt.Errorf("invalid output file extension: %s (allowed: .json, .csv, .txt)", ext)
	}
	
	return nil
}

// ValidateWebPort validates web server port
func (v *InputValidator) ValidateWebPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	
	// Check for privileged ports
	if port < 1024 {
		return fmt.Errorf("port %d requires root privileges (use port >= 1024)", port)
	}
	
	return nil
}