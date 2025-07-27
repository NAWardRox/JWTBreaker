package engine

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"jwt-crack/internal/constants"
	"jwt-crack/internal/errors"
	"jwt-crack/pkg/config"
	"jwt-crack/pkg/logger"
)

// Result represents the result of a brute force attack
type Result struct {
	Success   bool          `json:"success"`
	Secret    string        `json:"secret,omitempty"`
	Algorithm string        `json:"algorithm"`
	Attempts  uint64        `json:"attempts"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	AttackMode string       `json:"attack_mode"`
}

// ProgressCallback is called during attack progress
type ProgressCallback func(attempts uint64, rate float64, eta time.Duration, status string)

// Engine represents the JWT brute force engine
type Engine struct {
	config        *config.Config
	jwtParts      []string
	algorithm     string
	hashFunc      func() hash.Hash
	attempts      uint64
	logger        *logger.Logger
	progressCallback ProgressCallback
	startTime     time.Time
}

// JWTHeader represents JWT header structure
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// New creates a new brute force engine
func New(cfg *config.Config, log *logger.Logger) (*Engine, error) {
	if cfg == nil {
		return nil, errors.ErrConfig(errors.ErrConfigValidation, "configuration cannot be nil", nil)
	}
	
	if log == nil {
		log = logger.Default()
	}
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, errors.ErrConfig(errors.ErrConfigValidation, "invalid configuration", err)
	}
	
	// Parse JWT token
	parts := strings.Split(cfg.Token, ".")
	if len(parts) != 3 {
		return nil, errors.ErrValidation(errors.ErrInvalidToken, 
			fmt.Sprintf("JWT token must have 3 parts, got %d", len(parts)), nil)
	}
	
	// Parse and validate header
	algorithm, hashFunc, err := parseJWTHeader(parts[0])
	if err != nil {
		return nil, errors.ErrValidation(errors.ErrInvalidToken, "failed to parse JWT header", err)
	}
	
	// Adjust thread count based on performance setting
	cfg.AdjustForPerformance()
	
	engine := &Engine{
		config:    cfg,
		jwtParts:  parts,
		algorithm: algorithm,
		hashFunc:  hashFunc,
		logger:    log,
	}
	
	log.TokenAnalyzed(algorithm, true)
	
	return engine, nil
}

// SetProgressCallback sets the progress callback function
func (e *Engine) SetProgressCallback(callback ProgressCallback) {
	e.progressCallback = callback
}

// Attack executes the brute force attack based on configuration
func (e *Engine) Attack(ctx context.Context) (*Result, error) {
	startTime := time.Now()
	e.startTime = startTime
	
	e.logger.AttackStarted(e.algorithm, e.getAttackMode(), e.config.Threads)
	
	var result *Result
	var err error
	
	// Execute attack based on mode
	switch {
	case e.config.Smart:
		result, err = e.smartAttack(ctx)
	case e.config.Wordlist != "":
		result, err = e.wordlistAttack(ctx)
	default:
		result, err = e.charsetAttack(ctx)
	}
	
	if err != nil {
		return nil, errors.ErrAttack(errors.ErrAttackExecution, "attack failed", err)
	}
	
	// Complete result
	duration := time.Since(startTime)
	result.Duration = duration
	result.Timestamp = time.Now()
	result.Algorithm = e.algorithm
	result.AttackMode = e.getAttackMode()
	
	e.logger.AttackCompleted(result.Success, result.Attempts, duration, result.Secret)
	
	return result, nil
}

// smartAttack performs smart pattern-based attack
func (e *Engine) smartAttack(ctx context.Context) (*Result, error) {
	result := &Result{}
	patterns := constants.SmartPatterns
	
	for _, pattern := range patterns {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		
		attempts := atomic.AddUint64(&e.attempts, 1)
		
		if e.verifySecret(pattern) {
			result.Success = true
			result.Secret = pattern
			result.Attempts = attempts
			return result, nil
		}
		
		// Report progress
		if e.progressCallback != nil {
			elapsed := time.Since(e.startTime).Seconds()
			rate := 0.0
			// Only calculate rate if we have meaningful elapsed time (at least 100ms)
			if elapsed > 0.1 {
				rate = float64(attempts) / elapsed
				// Cap unrealistic rates (max 100M/s which is already very high)
				if rate > 100000000 {
					rate = 100000000
				}
			}
			e.progressCallback(attempts, rate, 0, fmt.Sprintf("Testing pattern: %s", pattern))
		}
		
		// Small delay to prevent overwhelming
		time.Sleep(constants.SmartAttackDelay)
	}
	
	result.Attempts = atomic.LoadUint64(&e.attempts)
	return result, nil
}

// wordlistAttack performs wordlist-based attack
func (e *Engine) wordlistAttack(ctx context.Context) (*Result, error) {
	result := &Result{}
	
	file, err := os.Open(e.config.Wordlist)
	if err != nil {
		return nil, errors.ErrFile(errors.ErrFileRead, 
			fmt.Sprintf("failed to open wordlist file: %s", e.config.Wordlist), err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			result.Attempts = atomic.LoadUint64(&e.attempts)
			return result, ctx.Err()
		default:
		}
		
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue // Skip empty lines
		}
		
		attempts := atomic.AddUint64(&e.attempts, 1)
		
		if e.verifySecret(word) {
			result.Success = true
			result.Secret = word
			result.Attempts = attempts
			return result, nil
		}
		
		// Report progress periodically
		if e.progressCallback != nil && attempts%constants.WordlistProgressFreq == 0 {
			elapsed := time.Since(e.startTime).Seconds()
			rate := 0.0
			// Only calculate rate if we have meaningful elapsed time (at least 100ms)
			if elapsed > 0.1 {
				rate = float64(attempts) / elapsed
				// Cap unrealistic rates (max 100M/s which is already very high)
				if rate > 100000000 {
					rate = 100000000
				}
			}
			e.progressCallback(attempts, rate, 0, fmt.Sprintf("Tested %d passwords", attempts))
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, errors.ErrFile(errors.ErrFileRead, "error reading wordlist file", err)
	}
	
	result.Attempts = atomic.LoadUint64(&e.attempts)
	return result, nil
}

// charsetAttack performs charset-based brute force attack
func (e *Engine) charsetAttack(ctx context.Context) (*Result, error) {
	result := &Result{}
	
	var charset string
	
	// Check if it's a predefined charset name
	if predefinedCharset, exists := constants.Charsets[e.config.Charset]; exists {
		charset = predefinedCharset
	} else {
		// Use the raw charset string (for custom charsets and processed hashcat rules)
		charset = e.config.Charset
	}
	
	// Validate charset is not empty
	if charset == "" {
		return nil, errors.ErrValidation(errors.ErrInvalidCharset, 
			"charset cannot be empty", nil)
	}
	
	for length := e.config.LengthMin; length <= e.config.LengthMax; length++ {
		if found, err := e.bruteforceLength(ctx, charset, length, result); err != nil {
			return nil, err
		} else if found {
			return result, nil
		}
	}
	
	result.Attempts = atomic.LoadUint64(&e.attempts)
	return result, nil
}

// bruteforceLength performs brute force for a specific length
func (e *Engine) bruteforceLength(ctx context.Context, charset string, length int, result *Result) (bool, error) {
	indices := make([]int, length)
	charsetLen := len(charset)
	
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		
		// Generate candidate
		candidate := make([]byte, length)
		for i, idx := range indices {
			candidate[i] = charset[idx]
		}
		
		attempts := atomic.AddUint64(&e.attempts, 1)
		
		if e.verifySecret(string(candidate)) {
			result.Success = true
			result.Secret = string(candidate)
			result.Attempts = attempts
			return true, nil
		}
		
		// Report progress periodically
		if e.progressCallback != nil && attempts%constants.CharsetProgressFreq == 0 {
			elapsed := time.Since(e.startTime).Seconds()
			rate := 0.0
			// Only calculate rate if we have meaningful elapsed time (at least 100ms)
			if elapsed > 0.1 {
				rate = float64(attempts) / elapsed
				// Cap unrealistic rates (max 100M/s which is already very high)
				if rate > 100000000 {
					rate = 100000000
				}
			}
			status := fmt.Sprintf("Testing length %d: %s", length, string(candidate))
			e.progressCallback(attempts, rate, 0, status)
		}
		
		// Increment indices (like an odometer)
		carry := 1
		for i := length - 1; i >= 0 && carry > 0; i-- {
			indices[i] += carry
			if indices[i] >= charsetLen {
				indices[i] = 0
				carry = 1
			} else {
				carry = 0
			}
		}
		
		// If carry is still 1, we've exhausted all combinations for this length
		if carry > 0 {
			break
		}
	}
	
	return false, nil
}

// verifySecret verifies if a secret produces the correct JWT signature
func (e *Engine) verifySecret(secret string) bool {
	h := hmac.New(e.hashFunc, []byte(secret))
	h.Write([]byte(e.jwtParts[0] + "." + e.jwtParts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	
	return expectedSignature == e.jwtParts[2]
}

// getAttackMode returns the current attack mode as string
func (e *Engine) getAttackMode() string {
	switch {
	case e.config.Smart:
		return constants.AttackModeSmart
	case e.config.Wordlist != "":
		return constants.AttackModeWordlist
	default:
		return constants.AttackModeCharset
	}
}

// parseJWTHeader parses JWT header and returns algorithm and hash function
func parseJWTHeader(headerPart string) (string, func() hash.Hash, error) {
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerPart)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode header: %w", err)
	}
	
	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", nil, fmt.Errorf("failed to parse header JSON: %w", err)
	}
	
	var hashFunc func() hash.Hash
	switch header.Alg {
	case constants.AlgorithmHS256:
		hashFunc = sha256.New
	case constants.AlgorithmHS384:
		hashFunc = sha512.New384
	case constants.AlgorithmHS512:
		hashFunc = sha512.New
	default:
		return "", nil, errors.ErrAlgorithm(errors.ErrUnsupportedAlg, 
			fmt.Sprintf("unsupported algorithm: %s", header.Alg), nil)
	}
	
	return header.Alg, hashFunc, nil
}

// GetStats returns current attack statistics
func (e *Engine) GetStats() (uint64, string) {
	attempts := atomic.LoadUint64(&e.attempts)
	mode := e.getAttackMode()
	return attempts, mode
}