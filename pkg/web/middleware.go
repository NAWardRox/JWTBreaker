package web

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"jwt-crack/internal/constants"
	"jwt-crack/internal/errors"
	"jwt-crack/pkg/config"
	"jwt-crack/pkg/system"
	"jwt-crack/pkg/validator"
)

// securityMiddleware adds security headers
func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:")
		
		// CORS headers for API endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(lrw, r)
		
		duration := time.Since(start)
		
		s.logger.Debug("HTTP %s %s %d %v %s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

// jsonMiddleware sets JSON content type for API endpoints
func (s *Server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status codes
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Validation and utility functions

// validateAttackRequest validates attack request parameters
func (s *Server) validateAttackRequest(req *struct {
	Token       string `json:"token"`
	AttackType  string `json:"attack_type"`
	Wordlist    string `json:"wordlist,omitempty"`
	Charset     string `json:"charset,omitempty"`
	LengthMin   int    `json:"length_min,omitempty"`
	LengthMax   int    `json:"length_max,omitempty"`
	Threads     int    `json:"threads,omitempty"`
	Performance string `json:"performance,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
}) error {
	// Validate JWT token
	jwtValidator := validator.NewJWTValidator()
	if err := jwtValidator.ValidateToken(req.Token); err != nil {
		return errors.ErrValidation(errors.ErrInvalidToken, "invalid JWT token", err)
	}
	
	// Validate attack type
	validTypes := map[string]bool{
		constants.AttackModeSmart:    true,
		constants.AttackModeWordlist: true,
		constants.AttackModeCharset:  true,
	}
	if !validTypes[req.AttackType] {
		return errors.ErrValidation(errors.ErrInvalidFile, 
			fmt.Sprintf("invalid attack type: %s", req.AttackType), nil)
	}
	
	// Validate wordlist for wordlist attacks
	if req.AttackType == constants.AttackModeWordlist {
		if req.Wordlist == "" {
			return errors.ErrValidation(errors.ErrInvalidFile, "wordlist is required for wordlist attacks", nil)
		}
		
		// Validate wordlist file exists and is accessible
		if _, err := os.Stat(req.Wordlist); os.IsNotExist(err) {
			return errors.ErrFile(errors.ErrFileNotFound, 
				fmt.Sprintf("wordlist file not found: %s", req.Wordlist), err)
		}
	}
	
	// Validate input parameters using existing validators
	inputValidator := validator.NewInputValidator()
	
	if req.Charset != "" {
		// Process and validate charset (handles hashcat rules and custom charsets)
		processedCharset := inputValidator.ProcessHashcatCharset(req.Charset)
		if len(processedCharset) == 0 {
			return errors.ErrValidation(errors.ErrInvalidFile, "processed charset cannot be empty", nil)
		}
		if len(processedCharset) > 1000 {
			return errors.ErrValidation(errors.ErrInvalidFile, "charset too long after processing", nil)
		}
	}
	
	if req.LengthMin > 0 || req.LengthMax > 0 {
		lengthMin := req.LengthMin
		lengthMax := req.LengthMax
		if lengthMin == 0 {
			lengthMin = constants.DefaultLengthMin
		}
		if lengthMax == 0 {
			lengthMax = constants.DefaultLengthMax
		}
		
		if err := inputValidator.ValidateLength(lengthMin, lengthMax); err != nil {
			return err
		}
	}
	
	if req.Threads > 0 {
		if err := inputValidator.ValidateThreads(req.Threads); err != nil {
			return err
		}
	}
	
	if req.Performance != "" {
		if err := inputValidator.ValidatePerformance(req.Performance); err != nil {
			return err
		}
	}
	
	// Validate timeout
	if req.Timeout < 0 || req.Timeout > 3600 { // Max 1 hour
		return errors.ErrValidation("INVALID_TIMEOUT", "timeout must be between 0 and 3600 seconds", nil)
	}
	
	return nil
}

// createAttackConfig creates attack configuration from request
func (s *Server) createAttackConfig(req *struct {
	Token       string `json:"token"`
	AttackType  string `json:"attack_type"`
	Wordlist    string `json:"wordlist,omitempty"`
	Charset     string `json:"charset,omitempty"`
	LengthMin   int    `json:"length_min,omitempty"`
	LengthMax   int    `json:"length_max,omitempty"`
	Threads     int    `json:"threads,omitempty"`
	Performance string `json:"performance,omitempty"`
	Timeout     int    `json:"timeout,omitempty"`
}) *config.Config {
	cfg := config.DefaultConfig()
	
	cfg.Token = req.Token
	cfg.Smart = req.AttackType == constants.AttackModeSmart
	cfg.Wordlist = req.Wordlist
	
	if req.Charset != "" {
		// Process hashcat rules and custom charsets
		inputValidator := validator.NewInputValidator()
		cfg.Charset = inputValidator.ProcessHashcatCharset(req.Charset)
	}
	
	if req.LengthMin > 0 {
		cfg.LengthMin = req.LengthMin
	}
	if req.LengthMax > 0 {
		cfg.LengthMax = req.LengthMax
	}
	
	if req.Threads > 0 {
		cfg.Threads = req.Threads
	} else {
		cfg.Threads = runtime.NumCPU()
	}
	
	if req.Performance != "" {
		cfg.Performance = req.Performance
	}
	
	// Adjust for performance
	cfg.AdjustForPerformance()
	
	return cfg
}

// validateUploadedFile validates uploaded file
func (s *Server) validateUploadedFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > constants.MaxUploadSize {
		return errors.ErrFile(errors.ErrFileTooBig, 
			fmt.Sprintf("file too large: %d bytes (max: %d)", header.Size, constants.MaxUploadSize), nil)
	}
	
	// Check filename
	filename := header.Filename
	if filename == "" {
		return errors.ErrValidation(errors.ErrInvalidFile, "filename cannot be empty", nil)
	}
	
	// Check for dangerous characters in filename
	if strings.ContainsAny(filename, "/<>:|\"?*\\") {
		return errors.ErrValidation(errors.ErrInvalidFile, "filename contains invalid characters", nil)
	}
	
	// Check file extension
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := map[string]bool{
		".txt":  true,
		".list": true,
		".dic":  true,
		".dict": true,
		"":      true, // Allow files without extension
	}
	
	if !validExts[ext] {
		return errors.ErrValidation(errors.ErrInvalidFile, 
			fmt.Sprintf("invalid file extension: %s", ext), nil)
	}
	
	return nil
}

// generateSecureFilename generates a secure filename
func (s *Server) generateSecureFilename(original string) string {
	// Extract extension
	ext := filepath.Ext(original)
	
	// Create hash of original filename + timestamp
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%d", original, time.Now().UnixNano())))
	
	// Use first 16 characters of hash as filename
	filename := fmt.Sprintf("%x%s", hash[:8], ext)
	
	return filename
}

// validateWordlistContent validates wordlist file content
func (s *Server) validateWordlistContent(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Read first 1KB to validate content
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	content := string(buffer[:n])
	
	// Check if content is valid UTF-8
	if !utf8.ValidString(content) {
		return errors.ErrValidation(errors.ErrInvalidFile, "file contains invalid UTF-8 characters", nil)
	}
	
	// Check for extremely long lines (potential DoS)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len(line) > constants.MaxLineLength {
			return errors.ErrValidation(errors.ErrInvalidFile, 
				fmt.Sprintf("line %d is too long (%d characters, max: %d)", 
					i+1, len(line), constants.MaxLineLength), nil)
		}
	}
	
	return nil
}

// getSystemInfo returns current system information
func (s *Server) getSystemInfo() (*system.Info, error) {
	return system.GetSystemInfo()
}

// analyzeJWTToken analyzes a JWT token and returns detailed information
func (s *Server) analyzeJWTToken(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	
	analysis := map[string]interface{}{
		"valid": true,
		"parts": map[string]interface{}{
			"header":    parts[0],
			"payload":   parts[1],
			"signature": parts[2],
		},
	}
	
	// Decode and parse header
	headerBytes, err := decodeBase64URLSafe(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	
	var header map[string]interface{}
	if err := s.unmarshalJSON(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header JSON: %w", err)
	}
	analysis["header"] = header
	
	// Decode and parse payload
	payloadBytes, err := decodeBase64URLSafe(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	
	var payload map[string]interface{}
	if err := s.unmarshalJSON(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload JSON: %w", err)
	}
	analysis["payload"] = payload
	
	// Extract algorithm
	if alg, ok := header["alg"].(string); ok {
		analysis["algorithm"] = alg
		
		// Check if algorithm is supported
		supportedAlgs := []string{constants.AlgorithmHS256, constants.AlgorithmHS384, constants.AlgorithmHS512}
		supported := false
		for _, supportedAlg := range supportedAlgs {
			if alg == supportedAlg {
				supported = true
				break
			}
		}
		analysis["supported"] = supported
	}
	
	// Add timing information if present in payload
	if iat, ok := payload["iat"].(float64); ok {
		analysis["issued_at"] = time.Unix(int64(iat), 0)
	}
	if exp, ok := payload["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		analysis["expires_at"] = expTime
		analysis["expired"] = time.Now().After(expTime)
	}
	if nbf, ok := payload["nbf"].(float64); ok {
		analysis["not_before"] = time.Unix(int64(nbf), 0)
	}
	
	return analysis, nil
}

// Helper functions

// decodeBase64URLSafe safely decodes base64 URL encoded strings
func decodeBase64URLSafe(data string) ([]byte, error) {
	// Add padding if needed
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}
	
	// Try different base64 encodings
	decoders := []func(string) ([]byte, error){
		func(s string) ([]byte, error) {
			// Remove padding first for RawURLEncoding
			s = strings.TrimRight(s, "=")
			return base64.RawURLEncoding.DecodeString(s)
		},
		base64.URLEncoding.DecodeString,
		base64.StdEncoding.DecodeString,
	}
	
	for _, decoder := range decoders {
		if decoded, err := decoder(data); err == nil {
			return decoded, nil
		}
	}
	
	return nil, fmt.Errorf("failed to decode base64 data")
}

// unmarshalJSON safely unmarshals JSON data
func (s *Server) unmarshalJSON(data []byte, v interface{}) error {
	// In production, you might want to use a more secure JSON parser
	// For now, we'll use the standard library
	return json.Unmarshal(data, v)
}

// formatDuration formats duration into human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// formatNumber formats large numbers with suffixes
func formatNumber(n uint64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else {
		// For very high values, cap at millions for readability
		return fmt.Sprintf("%.0fM", float64(n)/1000000)
	}
}

// formatFloatNumber formats float64 numbers with appropriate precision and suffixes
func formatFloatNumber(n float64) string {
	if n < 0 {
		return "0"
	}
	if n < 1000 {
		if n < 1 {
			return fmt.Sprintf("%.2f", n)
		} else if n < 10 {
			return fmt.Sprintf("%.1f", n)
		} else {
			return fmt.Sprintf("%.0f", n)
		}
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", n/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", n/1000000)
	} else {
		return fmt.Sprintf("%.1fG", n/1000000000)
	}
}