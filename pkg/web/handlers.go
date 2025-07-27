package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"jwt-crack/internal/constants"
	"jwt-crack/pkg/validator"
)

// indexHandler serves the main web interface
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JWT-Crack Web Interface</title>
    <link href="/static/css/style.css" rel="stylesheet">
</head>
<body>
    <!-- Connection Status -->
    <div id="connection-status" class="connection-status disconnected">● Connecting...</div>
    
    <!-- Alerts Container -->
    <div id="alerts-container" style="position: fixed; top: 60px; right: 20px; z-index: 1000; width: 300px;"></div>
    
    <!-- Header -->
    <header class="header">
        <div class="container">
            <div class="header-content">
                <div class="logo">
                    <h1>🔐 JWT-Crack</h1>
                </div>
                <div class="system-info">
                    <span>OS: <strong id="system-os">Loading...</strong></span>
                    <span>CPU: <strong id="system-cpu">Loading...</strong></span>
                    <span>RAM: <strong id="system-ram">Loading...</strong></span>
                    <span>Arch: <strong id="system-arch">Loading...</strong></span>
                </div>
            </div>
        </div>
    </header>

    <!-- Main Content -->
    <main class="container">
        <!-- Navigation Tabs -->
        <div class="tabs">
            <div class="tab-nav">
                <button class="tab-button active" data-tab="analyze">JWT Analysis</button>
                <button class="tab-button" data-tab="attack">Attack</button>
                <button class="tab-button" data-tab="upload">Wordlists</button>
            </div>
        </div>

        <!-- JWT Analysis Tab -->
        <div id="analyze-tab" class="tab-content active">
            <div class="card">
                <div class="card-header">
                    <h2 class="card-title">JWT Token Analysis</h2>
                </div>
                <div class="card-body">
                    <div class="form-group">
                        <label class="form-label" for="jwt-token">JWT Token</label>
                        <textarea id="jwt-token" class="form-textarea" placeholder="Paste your JWT token here (e.g., eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ)"></textarea>
                    </div>
                    <button id="analyze-btn" class="btn btn-primary">Analyze Token</button>
                </div>
            </div>
            
            <div id="jwt-analysis-result"></div>
        </div>

        <!-- Attack Tab -->
        <div id="attack-tab" class="tab-content">
            <form id="attack-form">
                <div class="card">
                    <div class="card-header">
                        <h2 class="card-title">JWT Crack Attack</h2>
                    </div>
                    <div class="card-body">
                        <div class="form-group">
                            <label class="form-label" for="attack-token">JWT Token</label>
                            <textarea id="attack-token" name="token" class="form-textarea" placeholder="Paste your JWT token here" required></textarea>
                        </div>

                        <div class="grid grid-cols-2">
                            <div class="form-group">
                                <label class="form-label" for="attack-type">Attack Type</label>
                                <select id="attack-type" name="attack_type" class="form-select" required>
                                    <option value="smart">Smart Attack</option>
                                    <option value="wordlist">Wordlist Attack</option>
                                    <option value="charset">Brute Force</option>
                                </select>
                            </div>
                            
                            <div class="form-group">
                                <label class="form-label" for="performance">Performance</label>
                                <select name="performance" class="form-select">
                                    <option value="eco">Eco</option>
                                    <option value="balanced" selected>Balanced</option>
                                    <option value="performance">High Performance</option>
                                    <option value="maximum">Maximum</option>
                                </select>
                            </div>
                        </div>

                        <!-- Wordlist Options -->
                        <div id="wordlist-options" style="display: none;">
                            <div class="form-group">
                                <label class="form-label" for="wordlist-select">Wordlist</label>
                                <select id="wordlist-select" name="wordlist" class="form-select">
                                    <option value="">Select a wordlist...</option>
                                </select>
                            </div>
                        </div>

                        <!-- Charset Options -->
                        <div id="charset-options" style="display: none;">
                            <div class="charset-selection-panel">
                                <h4 class="charset-panel-title">🔤 Character Set Selection</h4>
                                
                                <!-- Mode Selector -->
                                <div class="charset-mode-selector">
                                    <button type="button" class="mode-btn active" data-mode="presets">📋 Quick Presets</button>
                                    <button type="button" class="mode-btn" data-mode="mixmatch">🔀 Mix & Match</button>
                                    <button type="button" class="mode-btn" data-mode="advanced">⚡ Advanced</button>
                                </div>

                                <!-- Quick Presets Mode -->
                                <div id="mode-presets" class="charset-mode active">
                                    <div class="charset-preset-grid">
                                        <button type="button" class="preset-btn" data-charset="lowercase" title="Lowercase letters: a-z">
                                            <span class="preset-icon">a-z</span>
                                            <span class="preset-label">Lowercase</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="uppercase" title="Uppercase letters: A-Z">
                                            <span class="preset-icon">A-Z</span>
                                            <span class="preset-label">Uppercase</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="digits" title="Numbers: 0-9">
                                            <span class="preset-icon">0-9</span>
                                            <span class="preset-label">Numbers</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="mixed" title="Lowercase + Uppercase: a-zA-Z">
                                            <span class="preset-icon">a-Z</span>
                                            <span class="preset-label">Mixed Case</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="alphanumeric" title="Alphanumeric: a-zA-Z0-9">
                                            <span class="preset-icon">aZ9</span>
                                            <span class="preset-label">Alphanumeric</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="special" title="Special characters: !@#$%^&*()...">
                                            <span class="preset-icon">!@#</span>
                                            <span class="preset-label">Special</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="printable" title="All printable ASCII">
                                            <span class="preset-icon">ALL</span>
                                            <span class="preset-label">Full ASCII</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="hex" title="Hexadecimal: 0-9a-f">
                                            <span class="preset-icon">HEX</span>
                                            <span class="preset-label">Hex</span>
                                        </button>
                                        <button type="button" class="preset-btn" data-charset="base64" title="Base64 charset: A-Za-z0-9+/">
                                            <span class="preset-icon">B64</span>
                                            <span class="preset-label">Base64</span>
                                        </button>
                                    </div>
                                </div>

                                <!-- Mix & Match Mode -->
                                <div id="mode-mixmatch" class="charset-mode">
                                    <div class="charset-toggles">
                                        <div class="toggle-item">
                                            <label class="toggle-switch">
                                                <input type="checkbox" id="toggle-lowercase" value="lowercase">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">Lowercase <code>a-z</code></span>
                                        </div>
                                        <div class="toggle-item">
                                            <label class="toggle-switch">
                                                <input type="checkbox" id="toggle-uppercase" value="uppercase">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">Uppercase <code>A-Z</code></span>
                                        </div>
                                        <div class="toggle-item">
                                            <label class="toggle-switch">
                                                <input type="checkbox" id="toggle-digits" value="digits">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">Numbers <code>0-9</code></span>
                                        </div>
                                        <div class="toggle-item">
                                            <label class="toggle-switch">
                                                <input type="checkbox" id="toggle-special" value="special">
                                                <span class="toggle-slider"></span>
                                            </label>
                                            <span class="toggle-label">Special <code>!@#$%^&*</code></span>
                                        </div>
                                    </div>
                                </div>

                                <!-- Advanced Mode -->
                                <div id="mode-advanced" class="charset-mode">
                                    <div class="advanced-input-group">
                                        <label class="advanced-label">Custom Charset Rule:</label>
                                        <input type="text" id="charset-custom" name="charset" class="advanced-input" 
                                               placeholder="Enter raw charset (abc123!@#) or hashcat rules (?l?u?d)">
                                        
                                        <div class="hashcat-help">
                                            <span class="help-text">Hashcat syntax:</span>
                                            <div class="hashcat-rules">
                                                <span class="rule-tag">?l</span> lowercase
                                                <span class="rule-tag">?u</span> uppercase  
                                                <span class="rule-tag">?d</span> digits
                                                <span class="rule-tag">?s</span> special
                                                <span class="rule-tag">?a</span> all printable
                                            </div>
                                        </div>
                                        
                                        <div class="rule-examples">
                                            <span class="examples-label">Quick examples:</span>
                                            <div class="example-item" data-rule="?l?u?d">?l?u?d</div>
                                            <div class="example-item" data-rule="?d?d?d?d">?d?d?d?d</div>
                                            <div class="example-item" data-rule="abc123">abc123</div>
                                        </div>
                                    </div>
                                </div>

                                <!-- Live Preview -->
                                <div class="charset-preview-container">
                                    <div class="preview-header">
                                        <span class="preview-label">Preview:</span>
                                        <span class="charset-stats">
                                            <strong id="charset-length">0</strong> chars | 
                                            Keyspace: <strong id="keyspace-estimate">0</strong>
                                        </span>
                                    </div>
                                    <div id="charset-preview-text" class="charset-preview-display">Select a character set...</div>
                                </div>
                            </div>
                            
                            <div class="grid grid-cols-2">
                                <div class="form-group">
                                    <label class="form-label" for="length-min">Min Length</label>
                                    <input type="number" name="length_min" class="form-input" min="1" value="1">
                                </div>
                                <div class="form-group">
                                    <label class="form-label" for="length-max">Max Length</label>
                                    <input type="number" name="length_max" class="form-input" min="1" value="6">
                                </div>
                            </div>
                        </div>

                        <div class="grid grid-cols-2">
                            <div class="form-group">
                                <label class="form-label" for="threads">Threads</label>
                                <input type="number" name="threads" class="form-input" min="1" placeholder="Auto">
                            </div>
                            
                            <div class="form-group">
                                <label class="form-label" for="timeout">Timeout (seconds)</label>
                                <input type="number" name="timeout" class="form-input" min="0" placeholder="No timeout">
                            </div>
                        </div>

                        <div class="form-group">
                            <button type="button" id="start-attack-btn" class="btn btn-primary">Start Attack</button>
                            <button type="button" id="stop-attack-btn" class="btn btn-danger" disabled>Stop Attack</button>
                        </div>
                    </div>
                </div>
            </form>

            <!-- Attack Progress -->
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">Attack Progress</h3>
                </div>
                <div class="card-body">
                    <div id="attack-progress">
                        <p class="text-center text-secondary">No active attack</p>
                    </div>
                    
                    <!-- Enhanced Speed Display -->
                    <div id="attack-progress-details" class="speed-display" style="display: none;">
                        <div class="speed-section">
                            <div class="speed-main">
                                <div class="speed-label">Current Speed</div>
                                <div id="progress-speed" class="speed-value">0 passwords/s</div>
                            </div>
                            <div class="speed-stats">
                                <div class="speed-stat">
                                    <span class="stat-label">Peak:</span>
                                    <span id="speed-peak" class="stat-value">0/s</span>
                                </div>
                                <div class="speed-stat">
                                    <span class="stat-label">Average:</span>
                                    <span id="speed-average" class="stat-value">0/s</span>
                                </div>
                            </div>
                        </div>
                        
                        <div class="progress-grid">
                            <div class="progress-metric">
                                <div id="progress-passwords" class="progress-metric-value">0</div>
                                <div class="progress-metric-label">Passwords Tried</div>
                            </div>
                            <div class="progress-metric">
                                <div id="progress-elapsed" class="progress-metric-value">0s</div>
                                <div class="progress-metric-label">Elapsed Time</div>
                            </div>
                            <div class="progress-metric">
                                <div id="progress-percent" class="progress-metric-value">0%</div>
                                <div class="progress-metric-label">Complete</div>
                            </div>
                        </div>
                    </div>
                    
                    <div id="attack-stats"></div>
                </div>
            </div>

            <!-- Attack Results -->
            <div id="attack-result"></div>
        </div>

        <!-- Upload Tab -->
        <div id="upload-tab" class="tab-content">
            <div class="card">
                <div class="card-header">
                    <h2 class="card-title">Upload Wordlist</h2>
                </div>
                <div class="card-body">
                    <div id="file-drop-zone" class="file-upload">
                        <p><strong>Drop wordlist file here</strong> or <a href="#" style="color: var(--primary-color);">click to browse</a></p>
                        <p class="text-sm text-secondary">Supported formats: .txt, .list, .dic, .dict (max 50MB)</p>
                        <input type="file" id="file-input" accept=".txt,.list,.dic,.dict" style="display: none;">
                    </div>
                </div>
            </div>
        </div>
    </main>

    <script src="/static/js/app.js"></script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// websocketHandler handles WebSocket connections
func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the connection supports hijacking
	if _, ok := w.(http.Hijacker); !ok {
		s.logger.Error("WebSocket upgrade failed: ResponseWriter does not implement http.Hijacker")
		http.Error(w, "WebSocket upgrade not supported", http.StatusInternalServerError)
		return
	}
	
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed: %v", err)
		return
	}
	
	// Generate client ID
	clientID := generateID()
	
	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		server: s,
		id:     clientID,
	}
	
	s.clientsMu.Lock()
	s.clients[conn] = client
	s.clientsMu.Unlock()
	
	s.logger.Debug("WebSocket client connected: %s", clientID)
	
	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// systemInfoHandler returns system information
func (s *Server) systemInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Refresh system info for real-time data
	sysInfo, err := s.getSystemInfo()
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get system information", err)
		return
	}
	
	s.writeJSONResponse(w, sysInfo)
}

// analyzeJWTHandler analyzes JWT tokens
func (s *Server) analyzeJWTHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}
	
	// Validate JWT token
	jwtValidator := validator.NewJWTValidator()
	if err := jwtValidator.ValidateToken(req.Token); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JWT token", err)
		return
	}
	
	// Analyze JWT token
	analysis, err := s.analyzeJWTToken(req.Token)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "JWT analysis failed", err)
		return
	}
	
	s.writeJSONResponse(w, analysis)
}

// startAttackHandler starts a new attack session
func (s *Server) startAttackHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		AttackType  string `json:"attack_type"`
		Wordlist    string `json:"wordlist,omitempty"`
		Charset     string `json:"charset,omitempty"`
		LengthMin   int    `json:"length_min,omitempty"`
		LengthMax   int    `json:"length_max,omitempty"`
		Threads     int    `json:"threads,omitempty"`
		Performance string `json:"performance,omitempty"`
		Timeout     int    `json:"timeout,omitempty"` // seconds
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}
	
	// Validate request
	if err := s.validateAttackRequest(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid attack request", err)
		return
	}
	
	// Create attack configuration
	attackConfig := s.createAttackConfig(&req)
	
	// Generate session ID
	sessionID := generateID()
	
	// Create attack session
	ctx, cancel := context.WithCancel(context.Background())
	if req.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
	}
	
	session := &AttackSession{
		ID:        sessionID,
		Status:    "starting",
		Config:    attackConfig,
		StartTime: time.Now(),
		Cancel:    cancel,
	}
	
	s.AddAttackSession(session)
	
	// Start attack in goroutine
	go s.executeAttack(ctx, sessionID, attackConfig)
	
	response := map[string]interface{}{
		"session_id": sessionID,
		"status":     "started",
		"message":    "Attack session started successfully",
	}
	
	s.writeJSONResponse(w, response)
}

// stopAttackHandler stops an active attack session
func (s *Server) stopAttackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	
	if sessionID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Session ID is required", nil)
		return
	}
	
	session, exists := s.GetAttackSession(sessionID)
	if !exists {
		s.writeErrorResponse(w, http.StatusNotFound, "Attack session not found", nil)
		return
	}
	
	// Cancel the attack
	if session.Cancel != nil {
		session.Cancel()
	}
	
	// Update session status
	s.attacksMu.Lock()
	if session, exists := s.attacks[sessionID]; exists {
		session.mu.Lock()
		session.Status = "stopped"
		session.mu.Unlock()
	}
	s.attacksMu.Unlock()
	
	// Broadcast stop message
	s.broadcastMessage("attack_stopped", map[string]string{
		"session_id": sessionID,
		"message":    "Attack stopped by user",
	}, sessionID)
	
	response := map[string]interface{}{
		"session_id": sessionID,
		"status":     "stopped",
		"message":    "Attack stopped successfully",
	}
	
	s.writeJSONResponse(w, response)
}

// attackStatusHandler returns attack session status
func (s *Server) attackStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]
	
	if sessionID == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "Session ID is required", nil)
		return
	}
	
	session, exists := s.GetAttackSession(sessionID)
	if !exists {
		s.writeErrorResponse(w, http.StatusNotFound, "Attack session not found", nil)
		return
	}
	
	session.mu.RLock()
	status := map[string]interface{}{
		"session_id": session.ID,
		"status":     session.Status,
		"start_time": session.StartTime,
		"progress":   session.Progress,
		"result":     session.Result,
	}
	session.mu.RUnlock()
	
	s.writeJSONResponse(w, status)
}

// listAttacksHandler returns all active attack sessions
func (s *Server) listAttacksHandler(w http.ResponseWriter, r *http.Request) {
	s.attacksMu.RLock()
	attacks := make([]map[string]interface{}, 0, len(s.attacks))
	
	for _, session := range s.attacks {
		session.mu.RLock()
		attacks = append(attacks, map[string]interface{}{
			"session_id": session.ID,
			"status":     session.Status,
			"start_time": session.StartTime,
			"progress":   session.Progress,
		})
		session.mu.RUnlock()
	}
	s.attacksMu.RUnlock()
	
	s.writeJSONResponse(w, map[string]interface{}{
		"attacks": attacks,
		"count":   len(attacks),
	})
}

// uploadHandler handles wordlist file uploads
func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Limit file size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxUploadSize)
	
	err := r.ParseMultipartForm(constants.MaxUploadSize)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "File too large or invalid form", err)
		return
	}
	
	file, header, err := r.FormFile("wordlist")
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "No file provided", err)
		return
	}
	defer file.Close()
	
	// Validate file
	if err := s.validateUploadedFile(header); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid file", err)
		return
	}
	
	// Create secure upload directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create upload directory", err)
		return
	}
	
	// Generate secure filename
	filename := s.generateSecureFilename(header.Filename)
	filePath := filepath.Join(uploadDir, filename)
	
	// Prevent path traversal attacks
	if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(uploadDir)) {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid file path", nil)
		return
	}
	
	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create file", err)
		return
	}
	defer dst.Close()
	
	// Copy file with size limit
	_, err = io.CopyN(dst, file, constants.MaxUploadSize)
	if err != nil && err != io.EOF {
		os.Remove(filePath) // Clean up on error
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save file", err)
		return
	}
	
	// Validate file content
	if err := s.validateWordlistContent(filePath); err != nil {
		os.Remove(filePath) // Clean up invalid file
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid wordlist content", err)
		return
	}
	
	s.logger.Info("File uploaded successfully: %s", filename)
	
	response := map[string]interface{}{
		"filename":    filename,
		"original":    header.Filename,
		"path":        filePath,
		"size":        header.Size,
		"upload_time": time.Now(),
		"status":      "uploaded",
	}
	
	s.writeJSONResponse(w, response)
}

// listWordlistsHandler returns available wordlists
func (s *Server) listWordlistsHandler(w http.ResponseWriter, r *http.Request) {
	wordlists := []map[string]interface{}{}
	
	// List built-in wordlists
	builtinDir := "examples"
	if entries, err := os.ReadDir(builtinDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
				info, _ := entry.Info()
				wordlists = append(wordlists, map[string]interface{}{
					"name":     entry.Name(),
					"path":     filepath.Join(builtinDir, entry.Name()),
					"size":     info.Size(),
					"type":     "builtin",
					"modified": info.ModTime(),
				})
			}
		}
	}
	
	// List uploaded wordlists
	uploadDir := "uploads"
	if entries, err := os.ReadDir(uploadDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				info, _ := entry.Info()
				wordlists = append(wordlists, map[string]interface{}{
					"name":     entry.Name(),
					"path":     filepath.Join(uploadDir, entry.Name()),
					"size":     info.Size(),
					"type":     "uploaded",
					"modified": info.ModTime(),
				})
			}
		}
	}
	
	s.writeJSONResponse(w, map[string]interface{}{
		"wordlists": wordlists,
		"count":     len(wordlists),
	})
}

// healthHandler returns server health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":           "healthy",
		"timestamp":        time.Now(),
		"version":          constants.AppVersion,
		"active_attacks":   len(s.attacks),
		"connected_clients": len(s.clients),
		"uptime":           time.Since(time.Now()), // This would be calculated from server start time
	}
	
	s.writeJSONResponse(w, health)
}

// Helper methods

// writeJSONResponse writes a JSON response
func (s *Server) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response: %v", err)
	}
}

// writeErrorResponse writes an error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, code int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	response := map[string]interface{}{
		"error":     true,
		"message":   message,
		"timestamp": time.Now(),
	}
	
	if err != nil {
		response["details"] = err.Error()
		s.logger.Error("HTTP %d: %s - %v", code, message, err)
	}
	
	json.NewEncoder(w).Encode(response)
}

// generateID generates a secure random ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// Client WebSocket methods

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.server.clientsMu.Lock()
		delete(c.server.clients, c.conn)
		c.server.clientsMu.Unlock()
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.server.logger.Error("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}