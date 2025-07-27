package errors

import (
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ValidationError ErrorType = "validation"
	ConfigError     ErrorType = "config"
	FileError       ErrorType = "file"
	NetworkError    ErrorType = "network"
	AlgorithmError  ErrorType = "algorithm"
	AttackError     ErrorType = "attack"
	SystemError     ErrorType = "system"
)

// JWTCrackError represents application-specific errors
type JWTCrackError struct {
	Type    ErrorType
	Code    string
	Message string
	Cause   error
}

// Error implements the error interface
func (e *JWTCrackError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s:%s] %s: %v", e.Type, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *JWTCrackError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *JWTCrackError) Is(target error) bool {
	if t, ok := target.(*JWTCrackError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// NewError creates a new JWTCrackError
func NewError(errType ErrorType, code, message string, cause error) *JWTCrackError {
	return &JWTCrackError{
		Type:    errType,
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Predefined error codes
const (
	// Validation errors
	ErrInvalidToken     = "INVALID_TOKEN"
	ErrInvalidCharset   = "INVALID_CHARSET"
	ErrInvalidLength    = "INVALID_LENGTH"
	ErrInvalidThreads   = "INVALID_THREADS"
	ErrInvalidFile      = "INVALID_FILE"
	ErrInvalidOutput    = "INVALID_OUTPUT"
	
	// Config errors
	ErrConfigLoad       = "CONFIG_LOAD"
	ErrConfigValidation = "CONFIG_VALIDATION"
	
	// File errors
	ErrFileNotFound     = "FILE_NOT_FOUND"
	ErrFileTooBig       = "FILE_TOO_BIG"
	ErrFileRead         = "FILE_READ"
	ErrFileWrite        = "FILE_WRITE"
	
	// Network errors
	ErrWebServerStart   = "WEB_SERVER_START"
	ErrWebSocketUpgrade = "WEBSOCKET_UPGRADE"
	
	// Algorithm errors
	ErrUnsupportedAlg   = "UNSUPPORTED_ALGORITHM"
	ErrSignatureInvalid = "SIGNATURE_INVALID"
	
	// Attack errors
	ErrAttackInit       = "ATTACK_INIT"
	ErrAttackExecution  = "ATTACK_EXECUTION"
	ErrAttackTimeout    = "ATTACK_TIMEOUT"
	
	// System errors
	ErrInsufficientMemory = "INSUFFICIENT_MEMORY"
	ErrSystemResource     = "SYSTEM_RESOURCE"
)

// Convenience functions for common errors

// ErrValidation creates a validation error
func ErrValidation(code, message string, cause error) *JWTCrackError {
	return NewError(ValidationError, code, message, cause)
}

// ErrConfig creates a configuration error  
func ErrConfig(code, message string, cause error) *JWTCrackError {
	return NewError(ConfigError, code, message, cause)
}

// ErrFile creates a file error
func ErrFile(code, message string, cause error) *JWTCrackError {
	return NewError(FileError, code, message, cause)
}

// ErrNetwork creates a network error
func ErrNetwork(code, message string, cause error) *JWTCrackError {
	return NewError(NetworkError, code, message, cause)
}

// ErrAlgorithm creates an algorithm error
func ErrAlgorithm(code, message string, cause error) *JWTCrackError {
	return NewError(AlgorithmError, code, message, cause)
}

// ErrAttack creates an attack error
func ErrAttack(code, message string, cause error) *JWTCrackError {
	return NewError(AttackError, code, message, cause)
}

// ErrSystem creates a system error
func ErrSystem(code, message string, cause error) *JWTCrackError {
	return NewError(SystemError, code, message, cause)
}

// Common error instances
var (
	ErrTokenRequired = ErrValidation(ErrInvalidToken, "JWT token is required", nil)
	ErrTokenMalformed = ErrValidation(ErrInvalidToken, "JWT token is malformed", nil)
	ErrNoWordlistFile = ErrValidation(ErrInvalidFile, "wordlist file is required for wordlist attack", nil)
	ErrThreadCountInvalid = ErrValidation(ErrInvalidThreads, "thread count must be between 1 and 64", nil)
)

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if jwtErr, ok := err.(*JWTCrackError); ok {
		return jwtErr.Type == ValidationError
	}
	return false
}

// IsFileError checks if error is a file error
func IsFileError(err error) bool {
	if jwtErr, ok := err.(*JWTCrackError); ok {
		return jwtErr.Type == FileError
	}
	return false
}

// IsNetworkError checks if error is a network error
func IsNetworkError(err error) bool {
	if jwtErr, ok := err.(*JWTCrackError); ok {
		return jwtErr.Type == NetworkError
	}
	return false
}