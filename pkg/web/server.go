package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"jwt-crack/pkg/config"
	"jwt-crack/pkg/engine"
	"jwt-crack/pkg/logger"
	"jwt-crack/pkg/system"
)

// Server represents the web server
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	router     *mux.Router
	server     *http.Server
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]*Client
	clientsMu  sync.RWMutex
	broadcast  chan []byte
	attacks    map[string]*AttackSession
	attacksMu  sync.RWMutex
	systemInfo *system.Info
}

// Client represents a WebSocket client
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	server *Server
	id     string
}

// AttackSession represents an active attack session
type AttackSession struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Config    *config.Config         `json:"config"`
	StartTime time.Time              `json:"start_time"`
	Cancel    context.CancelFunc     `json:"-"`
	Progress  *AttackProgress        `json:"progress"`
	Result    *AttackResult          `json:"result"`
	mu        sync.RWMutex          `json:"-"`
}

// AttackProgress represents real-time attack progress
type AttackProgress struct {
	Type        string    `json:"type"`
	Current     uint64    `json:"current"`
	Total       uint64    `json:"total,omitempty"`
	Percent     float64   `json:"percent,omitempty"`
	Rate        float64   `json:"rate"`
	Speed       string    `json:"speed"`
	ETA         string    `json:"eta,omitempty"`
	Status      string    `json:"status"`
	ElapsedTime string    `json:"elapsed_time"`
	Timestamp   time.Time `json:"timestamp"`
}

// AttackResult represents attack completion result
type AttackResult struct {
	Type      string    `json:"type"`
	Success   bool      `json:"success"`
	Secret    string    `json:"secret,omitempty"`
	Algorithm string    `json:"algorithm"`
	Mode      string    `json:"mode"`
	Attempts  uint64    `json:"attempts"`
	Duration  string    `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// WebMessage represents WebSocket message structure
type WebMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Session string      `json:"session,omitempty"`
}

// New creates a new web server instance
func New(cfg *config.Config, log *logger.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	if log == nil {
		log = logger.Default()
	}
	
	// Get system information
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		log.Warn("Failed to get system information: %v", err)
		// Create basic system info as fallback
		sysInfo = &system.Info{
			OS:           "unknown",
			Architecture: "unknown",
			CPUCores:     1,
			TotalRAM:     "unknown",
			Platform:     "unknown",
		}
	}
	
	server := &Server{
		config:     cfg,
		logger:     log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			EnableCompression: false,
		},
		clients:    make(map[*websocket.Conn]*Client),
		broadcast:  make(chan []byte, 256),
		attacks:    make(map[string]*AttackSession),
		systemInfo: sysInfo,
	}
	
	server.setupRoutes()
	return server, nil
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.WebPort),
		Handler: s.router,
		// Remove timeouts that might interfere with WebSocket connections
	}
	
	// Start WebSocket hub
	go s.runHub()
	
	s.logger.WebServerStarted(s.config.WebPort)
	
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()
	
	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down web server...")
		
		// Shutdown gracefully
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		return s.server.Shutdown(shutdownCtx)
		
	case err := <-errChan:
		return fmt.Errorf("web server error: %w", err)
	}
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()
	
	// WebSocket endpoint (no middleware to avoid hijacker issues)
	s.router.HandleFunc("/ws", s.websocketHandler)
	
	// Create a subrouter for routes that need middleware
	web := s.router.NewRoute().Subrouter()
	web.Use(s.securityMiddleware)
	web.Use(s.loggingMiddleware)
	
	// Static files
	web.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("pkg/web/static/"))),
	)
	
	// Main page
	web.HandleFunc("/", s.indexHandler).Methods("GET")
	
	// API endpoints
	api := web.PathPrefix("/api").Subrouter()
	api.Use(s.jsonMiddleware)
	
	// System information
	api.HandleFunc("/system", s.systemInfoHandler).Methods("GET")
	
	// JWT analysis
	api.HandleFunc("/analyze", s.analyzeJWTHandler).Methods("POST")
	
	// Attack management
	api.HandleFunc("/attack/start", s.startAttackHandler).Methods("POST")
	api.HandleFunc("/attack/stop/{id}", s.stopAttackHandler).Methods("POST")
	api.HandleFunc("/attack/status/{id}", s.attackStatusHandler).Methods("GET")
	api.HandleFunc("/attack/list", s.listAttacksHandler).Methods("GET")
	
	// File upload
	api.HandleFunc("/upload", s.uploadHandler).Methods("POST")
	api.HandleFunc("/wordlists", s.listWordlistsHandler).Methods("GET")
	
	// Health check
	api.HandleFunc("/health", s.healthHandler).Methods("GET")
}

// runHub manages WebSocket connections and broadcasts
func (s *Server) runHub() {
	for {
		select {
		case message := <-s.broadcast:
			s.clientsMu.RLock()
			for conn, client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, conn)
				}
			}
			s.clientsMu.RUnlock()
		}
	}
}

// broadcastMessage sends a message to all connected clients
func (s *Server) broadcastMessage(msgType string, data interface{}, sessionID ...string) {
	message := WebMessage{
		Type: msgType,
		Data: data,
	}
	
	if len(sessionID) > 0 {
		message.Session = sessionID[0]
	}
	
	if jsonData, err := s.marshalJSON(message); err == nil {
		select {
		case s.broadcast <- jsonData:
		default:
			s.logger.Warn("Broadcast channel full, dropping message")
		}
	}
}

// GetAttackSession retrieves an attack session by ID
func (s *Server) GetAttackSession(id string) (*AttackSession, bool) {
	s.attacksMu.RLock()
	defer s.attacksMu.RUnlock()
	session, exists := s.attacks[id]
	return session, exists
}

// AddAttackSession adds a new attack session
func (s *Server) AddAttackSession(session *AttackSession) {
	s.attacksMu.Lock()
	defer s.attacksMu.Unlock()
	s.attacks[session.ID] = session
}

// RemoveAttackSession removes an attack session
func (s *Server) RemoveAttackSession(id string) {
	s.attacksMu.Lock()
	defer s.attacksMu.Unlock()
	if session, exists := s.attacks[id]; exists {
		if session.Cancel != nil {
			session.Cancel()
		}
		delete(s.attacks, id)
	}
}

// UpdateAttackProgress updates attack progress and broadcasts to clients
func (s *Server) UpdateAttackProgress(sessionID string, progress *AttackProgress) {
	s.attacksMu.Lock()
	if session, exists := s.attacks[sessionID]; exists {
		session.mu.Lock()
		session.Progress = progress
		session.mu.Unlock()
	}
	s.attacksMu.Unlock()
	
	// Broadcast progress update
	s.broadcastMessage("progress", progress, sessionID)
}

// UpdateAttackResult updates attack result and broadcasts to clients
func (s *Server) UpdateAttackResult(sessionID string, result *AttackResult) {
	s.attacksMu.Lock()
	if session, exists := s.attacks[sessionID]; exists {
		session.mu.Lock()
		session.Result = result
		session.Status = "completed"
		session.mu.Unlock()
	}
	s.attacksMu.Unlock()
	
	// Broadcast result
	s.broadcastMessage("result", result, sessionID)
}

// marshalJSON safely marshalls JSON data
func (s *Server) marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// executeAttack runs the actual attack in a separate goroutine
func (s *Server) executeAttack(ctx context.Context, sessionID string, cfg *config.Config) {
	session, exists := s.GetAttackSession(sessionID)
	if !exists {
		s.logger.Error("Attack session not found: %s", sessionID)
		return
	}

	// Update session status
	s.attacksMu.Lock()
	session.mu.Lock()
	session.Status = "running"
	session.mu.Unlock()
	s.attacksMu.Unlock()

	// Broadcast start message
	s.broadcastMessage("attack_started", map[string]interface{}{
		"session_id": sessionID,
		"status":     "running",
		"message":    "Attack started",
	}, sessionID)

	// Create progress callback
	progressCallback := func(attempts uint64, rate float64, eta time.Duration, status string) {
		progress := &AttackProgress{
			Type:        "progress",
			Current:     attempts,
			Total:       0, // Total is not available for all attack types
			Rate:        rate,
			Speed:       formatFloatNumber(rate) + "/s",
			Status:      status,
			ElapsedTime: formatDuration(time.Since(session.StartTime)),
			Timestamp:   time.Now(),
		}

		if eta > 0 {
			progress.ETA = formatDuration(eta)
		}

		s.UpdateAttackProgress(sessionID, progress)
	}

	// Run the attack
	engineInstance, err := engine.New(cfg, s.logger)
	if err != nil {
		s.logger.Error("Failed to create attack engine: %v", err)
		return
	}
	
	// Set up progress callback
	engineInstance.SetProgressCallback(progressCallback)
	
	// Execute the attack
	result, err := engineInstance.Attack(ctx)
	if err != nil {
		s.logger.Error("Attack failed: %v", err)
		// Create error result
		result = &engine.Result{
			Success:   false,
			Algorithm: "HS256", // Default algorithm, should be extracted from token
			Attempts:  0,
			Duration:  time.Since(session.StartTime),
			Timestamp: time.Now(),
			AttackMode: "error",
		}
	}

	// Update session with result
	attackResult := &AttackResult{
		Type:      "result",
		Success:   result.Success,
		Secret:    result.Secret,
		Algorithm: result.Algorithm,
		Mode:      result.AttackMode,
		Attempts:  result.Attempts,
		Duration:  formatDuration(result.Duration),
		Timestamp: time.Now(),
	}

	if err != nil {
		attackResult.Error = err.Error()
	}

	s.UpdateAttackResult(sessionID, attackResult)

	// Clean up session after completion
	go func() {
		time.Sleep(30 * time.Minute) // Keep results for 30 minutes
		s.RemoveAttackSession(sessionID)
	}()
}