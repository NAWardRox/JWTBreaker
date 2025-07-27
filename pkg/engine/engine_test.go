package engine

import (
	"context"
	"os"
	"testing"
	"time"

	"jwt-crack/internal/constants"
	"jwt-crack/pkg/config"
	"jwt-crack/pkg/logger"
)

// Test JWT tokens with known secrets
const (
	validTokenSecret = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	// Token signed with "abc" (3 char secret for faster charset testing)
	validTokenABC = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.TJiJR0JGn5zLLe6OmfZS26TjZhQBi7-KJIFUZoQ1FHI"
	invalidToken = "invalid.jwt.token"
	malformedToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Token:       validTokenSecret,
				Threads:     4,
				LengthMin:   1,
				LengthMax:   5,
				Charset:     constants.CharsetPassword,
				Performance: constants.PerformanceBalanced,
				WebPort:     8080,
			},
			wantErr: false,
		},
		{
			name:    "nil configuration",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty token",
			config: &config.Config{
				Token:   "",
				Threads: 4,
			},
			wantErr: true,
		},
		{
			name: "invalid token format",
			config: &config.Config{
				Token:   invalidToken,
				Threads: 4,
			},
			wantErr: true,
		},
		{
			name: "malformed token",
			config: &config.Config{
				Token:   malformedToken,
				Threads: 4,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logger.New(logger.ERROR, os.Stderr) // Suppress logs during tests
			engine, err := New(tt.config, logger)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}
			
			if engine == nil {
				t.Errorf("New() returned nil engine")
			}
		})
	}
}

func TestSmartAttack(t *testing.T) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Smart:       true,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	result, err := engine.Attack(ctx)
	if err != nil {
		t.Fatalf("Smart attack failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected smart attack to succeed")
	}
	
	if result.Secret != "your-256-bit-secret" {
		t.Errorf("Expected secret 'your-256-bit-secret', got '%s'", result.Secret)
	}
	
	if result.AttackMode != constants.AttackModeSmart {
		t.Errorf("Expected attack mode '%s', got '%s'", constants.AttackModeSmart, result.AttackMode)
	}
}

func TestWordlistAttack(t *testing.T) {
	// Create temporary wordlist file
	tmpfile, err := os.CreateTemp("", "wordlist_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	
	// Write test wordlist
	wordlist := []string{
		"wrong1",
		"wrong2", 
		"secret",  // This should match
		"wrong3",
	}
	
	for _, word := range wordlist {
		if _, err := tmpfile.WriteString(word + "\n"); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	tmpfile.Close()
	
	cfg := &config.Config{
		Token:       validTokenSecret,
		Wordlist:    tmpfile.Name(),
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	result, err := engine.Attack(ctx)
	if err != nil {
		t.Fatalf("Wordlist attack failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected wordlist attack to succeed")
	}
	
	if result.Secret != "secret" {
		t.Errorf("Expected secret 'secret', got '%s'", result.Secret)
	}
	
	if result.AttackMode != constants.AttackModeWordlist {
		t.Errorf("Expected attack mode '%s', got '%s'", constants.AttackModeWordlist, result.AttackMode)
	}
}

func TestCharsetAttack(t *testing.T) {
	// Use a token with a short secret for charset attack
	cfg := &config.Config{
		Token:       validTokenABC, // This token uses "abc" as secret
		Charset:     constants.CharsetAlpha,
		LengthMin:   3,
		LengthMax:   3, // Exactly 3 characters
		Threads:     2,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	result, err := engine.Attack(ctx)
	if err != nil {
		t.Fatalf("Charset attack failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected charset attack to succeed")
	}
	
	if result.Secret != "abc" {
		t.Errorf("Expected secret 'abc', got '%s'", result.Secret)
	}
	
	if result.AttackMode != constants.AttackModeCharset {
		t.Errorf("Expected attack mode '%s', got '%s'", constants.AttackModeCharset, result.AttackMode)
	}
}

func TestAttackWithCancellation(t *testing.T) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Charset:     constants.CharsetFull,
		LengthMin:   10,
		LengthMax:   15, // This will take a while
		Threads:     1,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	start := time.Now()
	result, err := engine.Attack(ctx)
	duration := time.Since(start)
	
	// Should be cancelled quickly
	if duration > time.Second {
		t.Errorf("Attack took too long before cancellation: %v", duration)
	}
	
	if err == nil {
		t.Errorf("Expected an error due to cancellation, got none")
	}
	
	if result != nil && result.Success {
		t.Errorf("Attack should not have succeeded after cancellation")
	}
}

func TestProgressCallback(t *testing.T) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Smart:       true,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	var progressCalls int
	var lastAttempts uint64
	
	engine.SetProgressCallback(func(attempts uint64, rate float64, eta time.Duration, status string) {
		progressCalls++
		lastAttempts = attempts
		
		if attempts == 0 {
			t.Errorf("Progress callback called with 0 attempts")
		}
		
		if status == "" {
			t.Errorf("Progress callback called with empty status")
		}
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	result, err := engine.Attack(ctx)
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected attack to succeed")
	}
	
	// Should have made some progress calls
	if progressCalls == 0 {
		t.Errorf("Expected progress callbacks to be called")
	}
	
	if lastAttempts == 0 {
		t.Errorf("Expected last attempts to be > 0")
	}
}

func TestVerifySecret(t *testing.T) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	tests := []struct {
		secret   string
		expected bool
	}{
		{"your-256-bit-secret", true},
		{"wrong-secret", false},
		{"", false},
		{"another-wrong-secret", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.secret, func(t *testing.T) {
			result := engine.verifySecret(tt.secret)
			if result != tt.expected {
				t.Errorf("verifySecret(%s) = %v, expected %v", tt.secret, result, tt.expected)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Smart:       true,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Initial stats
	attempts, mode := engine.GetStats()
	if attempts != 0 {
		t.Errorf("Expected initial attempts to be 0, got %d", attempts)
	}
	
	if mode != constants.AttackModeSmart {
		t.Errorf("Expected mode to be '%s', got '%s'", constants.AttackModeSmart, mode)
	}
	
	// Run attack and check stats again
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err = engine.Attack(ctx)
	if err != nil {
		t.Fatalf("Attack failed: %v", err)
	}
	
	attempts, mode = engine.GetStats()
	if attempts == 0 {
		t.Errorf("Expected attempts > 0 after attack, got %d", attempts)
	}
}

// Benchmark tests
func BenchmarkSmartAttack(b *testing.B) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Smart:       true,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset attempts counter for each run
		engine.attempts = 0
		
		_, err := engine.smartAttack(ctx)
		if err != nil {
			b.Fatalf("Smart attack failed: %v", err)
		}
	}
}

func BenchmarkVerifySecret(b *testing.B) {
	cfg := &config.Config{
		Token:       validTokenSecret,
		Threads:     1,
		LengthMin:   1,
		LengthMax:   8,
		Charset:     constants.CharsetPassword,
		Performance: constants.PerformanceBalanced,
		WebPort:     8080,
	}
	
	logger := logger.New(logger.ERROR, os.Stderr)
	engine, err := New(cfg, logger)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	
	secret := "test-secret"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.verifySecret(secret)
	}
}