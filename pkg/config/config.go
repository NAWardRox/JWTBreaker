package config

import (
	"fmt"
	"runtime"
	"time"
)

// Config holds all configuration for the JWT cracker
type Config struct {
	// Core settings
	Token       string `json:"token" yaml:"token"`
	Wordlist    string `json:"wordlist" yaml:"wordlist"`
	Charset     string `json:"charset" yaml:"charset"`
	LengthMin   int    `json:"length_min" yaml:"length_min"`
	LengthMax   int    `json:"length_max" yaml:"length_max"`
	
	// Performance settings
	Threads     int    `json:"threads" yaml:"threads"`
	Performance string `json:"performance" yaml:"performance"`
	
	// Attack settings
	Smart       bool   `json:"smart" yaml:"smart"`
	Output      string `json:"output" yaml:"output"`
	Verbose     bool   `json:"verbose" yaml:"verbose"`
	
	// Web settings
	WebEnabled  bool   `json:"web_enabled" yaml:"web_enabled"`
	WebPort     int    `json:"web_port" yaml:"web_port"`
	
	// Timeouts and limits
	RequestTimeout    time.Duration `json:"request_timeout" yaml:"request_timeout"`
	MaxFileSize       int64         `json:"max_file_size" yaml:"max_file_size"`
	MaxUploadSize     int64         `json:"max_upload_size" yaml:"max_upload_size"`
	ProgressInterval  time.Duration `json:"progress_interval" yaml:"progress_interval"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Charset:           "password",
		LengthMin:         1,
		LengthMax:         8,
		Threads:           runtime.NumCPU(),
		Performance:       "balanced",
		WebPort:           8080,
		RequestTimeout:    30 * time.Second,
		MaxFileSize:       100 * 1024 * 1024, // 100MB
		MaxUploadSize:     10 * 1024 * 1024,  // 10MB
		ProgressInterval:  100 * time.Millisecond,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("JWT token is required")
	}
	
	if c.LengthMin < 1 {
		return fmt.Errorf("minimum length must be at least 1")
	}
	
	if c.LengthMax < c.LengthMin {
		return fmt.Errorf("maximum length must be greater than or equal to minimum length")
	}
	
	if c.LengthMax > 20 {
		return fmt.Errorf("maximum length cannot exceed 20 characters")
	}
	
	if c.Threads < 1 {
		return fmt.Errorf("thread count must be at least 1")
	}
	
	if c.Threads > 64 {
		return fmt.Errorf("thread count cannot exceed 64")
	}
	
	validPerformance := map[string]bool{
		"eco": true, "balanced": true, "performance": true, "maximum": true,
	}
	if !validPerformance[c.Performance] {
		return fmt.Errorf("invalid performance level: %s", c.Performance)
	}
	
	// Validate charset - can be predefined name or raw character string
	if c.Charset == "" {
		return fmt.Errorf("charset cannot be empty")
	}
	
	// Check if it's a predefined charset or allow raw character string
	validCharsets := map[string]bool{
		"digits": true, "alpha": true, "password": true, "full": true,
	}
	
	// If it's not a predefined charset, validate as raw charset string
	if !validCharsets[c.Charset] {
		// Allow raw character strings (processed from frontend)
		if len(c.Charset) == 0 {
			return fmt.Errorf("charset cannot be empty")
		}
		if len(c.Charset) > 1000 {
			return fmt.Errorf("charset too long (%d characters, max: 1000)", len(c.Charset))
		}
	}
	
	if c.WebPort < 1 || c.WebPort > 65535 {
		return fmt.Errorf("web port must be between 1 and 65535")
	}
	
	return nil
}

// AdjustForPerformance modifies thread count based on performance setting
func (c *Config) AdjustForPerformance() {
	baseCPU := runtime.NumCPU()
	
	switch c.Performance {
	case "eco":
		c.Threads = max(1, baseCPU/4)
	case "balanced":
		c.Threads = max(1, baseCPU/2)
	case "performance":
		c.Threads = baseCPU
	case "maximum":
		c.Threads = baseCPU * 2
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}