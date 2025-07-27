package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJWTValidator_ValidateToken(t *testing.T) {
	validator := NewJWTValidator()
	
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid HS256 token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token with whitespace",
			token:   "  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c  ",
			wantErr: false,
		},
		{
			name:    "token with only 2 parts",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			wantErr: true,
		},
		{
			name:    "token with 4 parts",
			token:   "part1.part2.part3.part4",
			wantErr: true,
		},
		{
			name:    "token with invalid base64",
			token:   "invalid!.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
		{
			name:    "token with empty header",
			token:   ".eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
		{
			name:    "token with invalid JSON header",
			token:   "aW52YWxpZA.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
		{
			name:    "token with unsupported algorithm",
			token:   "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileValidator_ValidateWordlistFile(t *testing.T) {
	// Create temporary test files
	validFile := createTempFile(t, "valid.txt", "word1\nword2\nword3\n")
	defer os.Remove(validFile)
	
	largeFile := createTempFile(t, "large.txt", generateLargeContent(2*1024*1024)) // 2MB
	defer os.Remove(largeFile)
	
	invalidExtFile := createTempFile(t, "invalid.exe", "word1\nword2\n")
	defer os.Remove(invalidExtFile)
	
	longLineFile := createTempFile(t, "longline.txt", generateLongLine(2000)+"\nword2\n")
	defer os.Remove(longLineFile)
	
	validator := NewFileValidator(1024 * 1024) // 1MB limit
	
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{
			name:     "valid wordlist file",
			filepath: validFile,
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filepath: "/path/to/non/existent/file.txt",
			wantErr:  true,
		},
		{
			name:     "file too large",
			filepath: largeFile,
			wantErr:  true,
		},
		{
			name:     "invalid file extension",
			filepath: invalidExtFile,
			wantErr:  true,
		},
		{
			name:     "file with long lines",
			filepath: longLineFile,
			wantErr:  true,
		},
		{
			name:     "empty file path",
			filepath: "",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateWordlistFile(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWordlistFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateCharset(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name    string
		charset string
		wantErr bool
	}{
		{"valid digits", "digits", false},
		{"valid alpha", "alpha", false},
		{"valid password", "password", false},
		{"valid full", "full", false},
		{"invalid charset", "invalid", true},
		{"empty charset", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCharset(tt.charset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCharset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateLength(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name    string
		min     int
		max     int
		wantErr bool
	}{
		{"valid range", 1, 8, false},
		{"min equals max", 5, 5, false},
		{"min zero", 0, 8, true},
		{"min negative", -1, 8, true},
		{"max less than min", 8, 1, true},
		{"max too large", 1, 25, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLength(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateThreads(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name    string
		threads int
		wantErr bool
	}{
		{"valid thread count", 4, false},
		{"single thread", 1, false},
		{"max threads", 64, false},
		{"zero threads", 0, true},
		{"negative threads", -1, true},
		{"too many threads", 65, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateThreads(tt.threads)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThreads() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidatePerformance(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		performance string
		wantErr     bool
	}{
		{"valid eco", "eco", false},
		{"valid balanced", "balanced", false},
		{"valid performance", "performance", false},
		{"valid maximum", "maximum", false},
		{"invalid performance", "invalid", true},
		{"empty performance", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePerformance(tt.performance)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePerformance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateOutputPath(t *testing.T) {
	validator := NewInputValidator()
	
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_output_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	tests := []struct {
		name string
		path string
		wantErr bool
	}{
		{"valid JSON output", filepath.Join(tmpDir, "output.json"), false},
		{"valid CSV output", filepath.Join(tmpDir, "output.csv"), false},
		{"valid TXT output", filepath.Join(tmpDir, "output.txt"), false},
		{"empty path", "", false}, // Optional parameter
		{"invalid extension", filepath.Join(tmpDir, "output.xml"), true},
		{"non-existent directory", "/path/to/non/existent/dir/output.json", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInputValidator_ValidateWebPort(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port", 8080, false},
		{"high port", 65535, false},
		{"privileged port", 80, true},
		{"port zero", 0, true},
		{"negative port", -1, true},
		{"port too high", 65536, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateWebPort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebPort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions

func createTempFile(t *testing.T, name, content string) string {
	tmpDir, err := os.MkdirTemp("", "validator_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	filePath := filepath.Join(tmpDir, name)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	
	return filePath
}

func generateLargeContent(size int) string {
	content := make([]byte, size)
	for i := 0; i < size; i++ {
		content[i] = 'a'
	}
	return string(content)
}

func generateLongLine(length int) string {
	line := make([]byte, length)
	for i := 0; i < length; i++ {
		line[i] = 'a'
	}
	return string(line)
}