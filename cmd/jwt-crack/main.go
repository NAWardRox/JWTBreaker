package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	
	"jwt-crack/internal/constants"
	"jwt-crack/internal/errors"
	"jwt-crack/pkg/config"
	"jwt-crack/pkg/engine"
	"jwt-crack/pkg/logger"
	"jwt-crack/pkg/validator"
	"jwt-crack/pkg/web"
)

var (
	version = constants.AppVersion
	commit  = "dev"
	
	// Global configuration
	cfg    *config.Config
	log    *logger.Logger
	
	// CLI flags
	verbose    bool
	configFile string
)

func main() {
	// Initialize logger
	log = logger.Default()
	
	// Display author information
	displayAuthorInfo()
	
	// Create root command
	rootCmd := &cobra.Command{
		Use:   constants.AppName,
		Short: constants.AppDesc,
		Long: fmt.Sprintf(`%s

%s is a professional JWT secret bruteforce tool designed for 
security testing and penetration testing purposes.

⚠️  DISCLAIMER: This tool is for authorized security testing only.
Only use on systems you own or have explicit permission to test.

Repository: %s`, constants.AppDesc, constants.AppName, constants.AppRepo),
		Version: fmt.Sprintf("%s (commit: %s)", version, commit),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set log level based on verbose flag
			if verbose {
				log.SetLevel(logger.DEBUG)
			}
			
			// Load configuration
			cfg = config.DefaultConfig()
			if configFile != "" {
				log.Debug("Loading configuration from file: %s", configFile)
				// TODO: Implement config file loading
			}
		},
	}
	
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file path")
	
	// Add subcommands
	rootCmd.AddCommand(crackCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(validateCmd())
	
	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.Error("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func crackCmd() *cobra.Command {
	var (
		token       string
		wordlist    string
		charset     string
		lengthMin   int
		lengthMax   int
		threads     int
		output      string
		performance string
		smart       bool
		timeout     time.Duration
	)
	
	cmd := &cobra.Command{
		Use:   "crack",
		Short: "Crack JWT secret using various attack methods",
		Long: `Crack JWT secret using smart patterns, wordlist, or charset brute force.

Examples:
  # Smart attack (recommended first step)
  jwt-crack crack --token TOKEN --smart
  
  # Wordlist attack
  jwt-crack crack --token TOKEN --wordlist /path/to/wordlist.txt
  
  # Charset brute force
  jwt-crack crack --token TOKEN --charset password --length-min 1 --length-max 8`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCrack(token, wordlist, charset, lengthMin, lengthMax, 
				threads, output, performance, smart, timeout)
		},
	}
	
	// Required flags
	cmd.Flags().StringVarP(&token, "token", "t", "", "JWT token to crack (required)")
	cmd.MarkFlagRequired("token")
	
	// Attack method flags
	cmd.Flags().BoolVar(&smart, "smart", false, "Use smart attack with common patterns")
	cmd.Flags().StringVarP(&wordlist, "wordlist", "w", "", "Wordlist file path")
	cmd.Flags().StringVarP(&charset, "charset", "c", constants.DefaultCharset, 
		"Charset for brute force: digits, alpha, password, full")
	
	// Length flags
	cmd.Flags().IntVar(&lengthMin, "length-min", constants.DefaultLengthMin, "Minimum password length")
	cmd.Flags().IntVar(&lengthMax, "length-max", constants.DefaultLengthMax, "Maximum password length")
	
	// Performance flags
	cmd.Flags().IntVar(&threads, "threads", runtime.NumCPU(), "Number of concurrent threads")
	cmd.Flags().StringVar(&performance, "performance", constants.DefaultPerformance, 
		"Performance level: eco, balanced, performance, maximum")
	
	// Output flags
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (JSON/CSV/TXT)")
	cmd.Flags().DurationVar(&timeout, "timeout", 0, "Attack timeout (0 = no timeout)")
	
	return cmd
}

func serveCmd() *cobra.Command {
	var (
		port int
	)
	
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start web interface server",
		Long: `Start the web interface server for interactive JWT cracking.

The web interface provides:
- JWT token analysis
- Real-time attack progress
- Multiple attack methods
- Results visualization

Example:
  jwt-crack serve --port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(port)
		},
	}
	
	cmd.Flags().IntVar(&port, "port", constants.DefaultWebPort, "Web server port")
	
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", constants.AppName, version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("CPU cores: %d\n", runtime.NumCPU())
		},
	}
}

func validateCmd() *cobra.Command {
	var token string
	
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate JWT token format and structure",
		Long: `Validate JWT token format, decode headers and payload, and check algorithm support.

This command helps verify that a JWT token is properly formatted before attempting
to crack it.

Example:
  jwt-crack validate --token TOKEN`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(token)
		},
	}
	
	cmd.Flags().StringVarP(&token, "token", "t", "", "JWT token to validate (required)")
	cmd.MarkFlagRequired("token")
	
	return cmd
}

func runCrack(token, wordlist, charset string, lengthMin, lengthMax, threads int,
	output, performance string, smart bool, timeout time.Duration) error {
	
	// Build configuration
	cfg.Token = token
	cfg.Wordlist = wordlist
	cfg.Charset = charset
	cfg.LengthMin = lengthMin
	cfg.LengthMax = lengthMax
	cfg.Threads = threads
	cfg.Output = output
	cfg.Performance = performance
	cfg.Smart = smart
	
	// Validate configuration
	inputValidator := validator.NewInputValidator()
	jwtValidator := validator.NewJWTValidator()
	
	if err := jwtValidator.ValidateToken(token); err != nil {
		return errors.ErrValidation(errors.ErrInvalidToken, "invalid JWT token", err)
	}
	
	if wordlist != "" {
		fileValidator := validator.NewFileValidator(constants.MaxFileSize)
		if err := fileValidator.ValidateWordlistFile(wordlist); err != nil {
			return errors.ErrFile(errors.ErrInvalidFile, "invalid wordlist file", err)
		}
	}
	
	if err := inputValidator.ValidateCharset(charset); err != nil {
		return err
	}
	
	if err := inputValidator.ValidateLength(lengthMin, lengthMax); err != nil {
		return err
	}
	
	if err := inputValidator.ValidateThreads(threads); err != nil {
		return err
	}
	
	if err := inputValidator.ValidatePerformance(performance); err != nil {
		return err
	}
	
	if output != "" {
		if err := inputValidator.ValidateOutputPath(output); err != nil {
			return err
		}
	}
	
	// Adjust configuration for performance
	cfg.AdjustForPerformance()
	
	// Create engine
	attackEngine, err := engine.New(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create attack engine: %w", err)
	}
	
	// Set up progress callback
	progressCallback := func(attempts uint64, rate float64, eta time.Duration, status string) {
		log.ProgressUpdate(attempts, rate, eta)
		if verbose {
			fmt.Printf("\r%s - Attempts: %d, Rate: %.1f/s", status, attempts, rate)
		}
	}
	
	// Set progress callback on engine
	attackEngine.SetProgressCallback(progressCallback)
	
	// Set up context with timeout
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Info("Received interrupt signal, stopping attack...")
		cancel()
	}()
	
	// Run attack
	result, err := attackEngine.Attack(ctx)
	if err != nil {
		if err == context.Canceled {
			log.Info("Attack cancelled by user")
			return nil
		}
		if err == context.DeadlineExceeded {
			log.Info("Attack timed out")
			return nil
		}
		return fmt.Errorf("attack failed: %w", err)
	}
	
	// Display results
	displayResults(result)
	
	// Save results if output file specified
	if output != "" {
		if err := saveResults(result, output); err != nil {
			log.Error("Failed to save results: %v", err)
		} else {
			log.Info("Results saved to: %s", output)
		}
	}
	
	return nil
}

func runServe(port int) error {
	// Validate port
	inputValidator := validator.NewInputValidator()
	if err := inputValidator.ValidateWebPort(port); err != nil {
		return err
	}
	
	// Create web server configuration
	webConfig := config.DefaultConfig()
	webConfig.WebPort = port
	
	// Create web server
	server, err := web.New(webConfig, log)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}
	
	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Info("Received interrupt signal, shutting down web server...")
		cancel()
	}()
	
	// Start server
	log.Info("Starting web server on port %d", port)
	log.Info("Open http://localhost:%d in your browser", port)
	
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("web server error: %w", err)
	}
	
	return nil
}

func runValidate(token string) error {
	validator := validator.NewJWTValidator()
	
	if err := validator.ValidateToken(token); err != nil {
		log.Error("Token validation failed: %v", err)
		return err
	}
	
	log.Info("✅ JWT token is valid and supported")
	
	// TODO: Add detailed token analysis output
	
	return nil
}

func displayResults(result *engine.Result) {
	fmt.Println(strings.Repeat("─", 50))
	if result.Success {
		fmt.Printf("✅ SECRET FOUND!\n")
		fmt.Printf("Secret: %s\n", result.Secret)
		fmt.Printf("Algorithm: %s\n", result.Algorithm)
		fmt.Printf("Attack Mode: %s\n", result.AttackMode)
		fmt.Printf("Attempts: %s\n", formatNumber(result.Attempts))
		fmt.Printf("Duration: %s\n", result.Duration)
	} else {
		fmt.Printf("❌ Secret not found\n")
		fmt.Printf("Algorithm: %s\n", result.Algorithm)
		fmt.Printf("Attack Mode: %s\n", result.AttackMode)
		fmt.Printf("Attempts: %s\n", formatNumber(result.Attempts))
		fmt.Printf("Duration: %s\n", result.Duration)
	}
	fmt.Println(strings.Repeat("─", 50))
}

func saveResults(result *engine.Result, filename string) error {
	// TODO: Implement result saving in multiple formats
	return nil
}

func formatNumber(n uint64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	// For very high values, cap at millions for readability
	return fmt.Sprintf("%.0fM", float64(n)/1000000)
}

func displayAuthorInfo() {
	fmt.Println("╭─────────────────────────────────────────╮")
	fmt.Printf("│             JWT-Crack v%-8s        │\n", version)
	fmt.Println("│         Created by NAWardRox            │")
	fmt.Println("╰─────────────────────────────────────────╯")
	fmt.Println()
}