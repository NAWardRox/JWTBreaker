// JWT-Crack Web Interface
class JWTCrackApp {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.currentAttack = null;
        this.systemInfo = null;
        this.speedHistory = [];
        this.peakSpeed = 0;
        this.currentMode = 'presets';
        this.selectedPreset = null;
        this.charsets = {
            lowercase: 'abcdefghijklmnopqrstuvwxyz',
            uppercase: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
            digits: '0123456789',
            mixed: 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ',
            alphanumeric: 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789',
            special: '!@#$%^&*()-_=+[]{}|;:\'",.<>?/\\`',
            printable: ' !"#$%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~',
            hex: '0123456789abcdef',
            base64: 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/',
            // Backend compatibility mappings
            alpha: 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ',
            password: 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*',
            full: 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:\'",.<>?/\\`'
        };
        
        this.init();
    }

    async init() {
        await this.loadSystemInfo();
        this.setupEventListeners();
        this.connectWebSocket();
        this.initializeTabs();
        await this.loadWordlists();
    }

    // WebSocket Connection
    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.updateConnectionStatus(true);
        };
        
        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };
        
        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.updateConnectionStatus(false);
            // Don't auto-reconnect if WebSocket is not supported
            if (this.reconnectAttempts < 3) {
                this.reconnectWebSocket();
            } else {
                console.warn('WebSocket reconnection failed multiple times, falling back to HTTP polling');
                this.fallbackToPolling();
            }
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus(false);
        };
    }

    fallbackToPolling() {
        console.log('Using HTTP polling instead of WebSocket');
        this.updateConnectionStatus(true); // Show as connected since we can use HTTP
        
        // Poll for attack status if there's an active attack
        if (this.currentAttack) {
            this.pollAttackStatus();
        }
    }

    async pollAttackStatus() {
        if (!this.currentAttack) return;
        
        try {
            const response = await fetch(`/api/attack/status/${this.currentAttack}`);
            if (response.ok) {
                const status = await response.json();
                if (status.progress) {
                    this.updateAttackProgress(status.progress, this.currentAttack);
                }
                if (status.result) {
                    this.updateAttackResult(status.result, this.currentAttack);
                    return; // Stop polling when attack is complete
                }
            }
        } catch (error) {
            console.error('Failed to poll attack status:', error);
        }
        
        // Continue polling every 2 seconds
        setTimeout(() => this.pollAttackStatus(), 2000);
    }

    reconnectWebSocket() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.pow(2, this.reconnectAttempts) * 1000; // Exponential backoff
            console.log(`Reconnecting WebSocket in ${delay}ms (attempt ${this.reconnectAttempts})`);
            setTimeout(() => this.connectWebSocket(), delay);
        }
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'progress':
                this.updateAttackProgress(message.data, message.session);
                break;
            case 'result':
                this.updateAttackResult(message.data, message.session);
                break;
            case 'attack_started':
                this.onAttackStarted(message.data, message.session);
                break;
            case 'attack_stopped':
                this.onAttackStopped(message.data, message.session);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    updateConnectionStatus(connected) {
        const statusEl = document.getElementById('connection-status');
        if (statusEl) {
            statusEl.className = `connection-status ${connected ? 'connected' : 'disconnected'}`;
            statusEl.textContent = connected ? '● Connected' : '● Disconnected';
        }
    }

    // System Information
    async loadSystemInfo() {
        try {
            const response = await fetch('/api/system');
            if (response.ok) {
                this.systemInfo = await response.json();
                this.displaySystemInfo();
            }
        } catch (error) {
            console.error('Failed to load system info:', error);
        }
    }

    displaySystemInfo() {
        if (!this.systemInfo) return;
        
        const elements = {
            'system-os': this.systemInfo.platform || this.systemInfo.os,
            'system-cpu': `${this.systemInfo.cpu_cores} cores`,
            'system-ram': this.systemInfo.total_ram,
            'system-arch': this.systemInfo.architecture
        };
        
        Object.entries(elements).forEach(([id, value]) => {
            const el = document.getElementById(id);
            if (el) el.textContent = value;
        });
    }

    // Tab Management
    initializeTabs() {
        document.querySelectorAll('.tab-button').forEach(button => {
            button.addEventListener('click', () => this.switchTab(button.dataset.tab));
        });
        
        // Activate first tab
        this.switchTab('analyze');
    }

    switchTab(tabId) {
        // Preserve JWT token across tabs
        this.preserveJWTToken();
        
        // Update buttons
        document.querySelectorAll('.tab-button').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabId);
        });
        
        // Update content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabId}-tab`);
        });
        
        // Restore JWT token in new tab
        this.restoreJWTToken();
    }
    
    preserveJWTToken() {
        // Get token from analyze tab
        const analyzeToken = document.getElementById('jwt-token');
        const attackToken = document.getElementById('attack-token');
        
        if (analyzeToken && analyzeToken.value && attackToken) {
            this.savedToken = analyzeToken.value;
        } else if (attackToken && attackToken.value && analyzeToken) {
            this.savedToken = attackToken.value;
        }
    }
    
    restoreJWTToken() {
        if (!this.savedToken) return;
        
        const analyzeToken = document.getElementById('jwt-token');
        const attackToken = document.getElementById('attack-token');
        
        if (analyzeToken && !analyzeToken.value) {
            analyzeToken.value = this.savedToken;
        }
        if (attackToken && !attackToken.value) {
            attackToken.value = this.savedToken;
        }
    }

    // JWT Analysis
    async analyzeJWT() {
        const token = document.getElementById('jwt-token').value.trim();
        const resultEl = document.getElementById('jwt-analysis-result');
        
        if (!token) {
            this.showAlert('error', 'Please enter a JWT token');
            return;
        }
        
        resultEl.innerHTML = '<div class="spinner"></div> Analyzing token...';
        
        try {
            const response = await fetch('/api/analyze', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ token })
            });
            
            if (response.ok) {
                const analysis = await response.json();
                this.displayJWTAnalysis(analysis);
            } else {
                const error = await response.json();
                resultEl.innerHTML = `<div class="alert alert-error">${error.message}</div>`;
            }
        } catch (error) {
            resultEl.innerHTML = `<div class="alert alert-error">Analysis failed: ${error.message}</div>`;
        }
    }

    displayJWTAnalysis(analysis) {
        const resultEl = document.getElementById('jwt-analysis-result');
        
        let html = `
            <div class="card">
                <div class="card-header">
                    <h3 class="card-title">JWT Analysis Results</h3>
                </div>
                <div class="card-body">
                    <div class="stats-grid">
                        <div class="stat-card">
                            <div class="stat-value">${analysis.algorithm || 'Unknown'}</div>
                            <div class="stat-label">Algorithm</div>
                        </div>
                        <div class="stat-card">
                            <div class="stat-value ${analysis.supported ? 'text-success' : 'text-error'}">
                                ${analysis.supported ? 'Yes' : 'No'}
                            </div>
                            <div class="stat-label">Supported</div>
                        </div>
                        <div class="stat-card">
                            <div class="stat-value ${analysis.expired ? 'text-error' : 'text-success'}">
                                ${analysis.expired ? 'Yes' : 'No'}
                            </div>
                            <div class="stat-label">Expired</div>
                        </div>
                    </div>
                    
                    <div class="grid grid-cols-2">
                        <div>
                            <h4 class="font-bold mb-2">Header</h4>
                            <pre class="form-textarea font-mono text-sm">${JSON.stringify(analysis.header, null, 2)}</pre>
                        </div>
                        <div>
                            <h4 class="font-bold mb-2">Payload</h4>
                            <pre class="form-textarea font-mono text-sm">${JSON.stringify(analysis.payload, null, 2)}</pre>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        resultEl.innerHTML = html;
    }

    // Attack Management
    async startAttack() {
        const form = document.getElementById('attack-form');
        const formData = new FormData(form);
        
        const attackData = {
            token: formData.get('token'),
            attack_type: formData.get('attack_type'),
            wordlist: formData.get('wordlist'),
            charset: this.getCurrentCharset(),
            length_min: parseInt(formData.get('length_min')) || 0,
            length_max: parseInt(formData.get('length_max')) || 0,
            threads: parseInt(formData.get('threads')) || 0,
            performance: formData.get('performance'),
            timeout: parseInt(formData.get('timeout')) || 0
        };
        
        if (!attackData.token) {
            this.showAlert('error', 'JWT token is required');
            return;
        }
        
        try {
            const response = await fetch('/api/attack/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(attackData)
            });
            
            if (response.ok) {
                const result = await response.json();
                this.currentAttack = result.session_id;
                this.showAlert('success', 'Attack started successfully');
                this.updateAttackUI('running');
                
                // If WebSocket is not connected, start polling
                if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
                    this.pollAttackStatus();
                }
            } else {
                const error = await response.json();
                this.showAlert('error', error.message);
            }
        } catch (error) {
            this.showAlert('error', `Failed to start attack: ${error.message}`);
        }
    }

    async stopAttack() {
        if (!this.currentAttack) {
            this.showAlert('warning', 'No active attack to stop');
            return;
        }
        
        try {
            const response = await fetch(`/api/attack/stop/${this.currentAttack}`, {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showAlert('info', 'Attack stopped');
                this.updateAttackUI('stopped');
            } else {
                const error = await response.json();
                this.showAlert('error', error.message);
            }
        } catch (error) {
            this.showAlert('error', `Failed to stop attack: ${error.message}`);
        }
    }

    updateAttackUI(status) {
        const startBtn = document.getElementById('start-attack-btn');
        const stopBtn = document.getElementById('stop-attack-btn');
        
        if (status === 'running') {
            startBtn.disabled = true;
            stopBtn.disabled = false;
        } else {
            startBtn.disabled = false;
            stopBtn.disabled = true;
        }
    }

    updateAttackProgress(progress, sessionId) {
        if (sessionId !== this.currentAttack) return;
        
        const progressEl = document.getElementById('attack-progress');
        const detailsEl = document.getElementById('attack-progress-details');
        
        // Show enhanced progress details
        detailsEl.style.display = 'block';
        
        if (progressEl) {
            const percentage = progress.percent || 0;
            progressEl.innerHTML = `
                <div class="progress">
                    <div class="progress-bar" style="width: ${percentage}%"></div>
                </div>
                <div class="text-sm text-center mt-2">
                    ${percentage.toFixed(1)}% complete - ${progress.status}
                </div>
            `;
        }
        
        // Track speed history for average calculation
        const currentSpeed = progress.rate || 0;
        
        // Only add valid speeds to history
        if (currentSpeed > 0 && !isNaN(currentSpeed) && isFinite(currentSpeed)) {
            this.speedHistory.push(currentSpeed);
            
            // Keep only last 10 measurements for rolling average
            if (this.speedHistory.length > 10) {
                this.speedHistory.shift();
            }
            
            // Update peak speed
            if (currentSpeed > this.peakSpeed) {
                this.peakSpeed = currentSpeed;
            }
        }
        
        // Calculate average speed from valid history
        const averageSpeed = this.speedHistory.length > 0 
            ? this.speedHistory.reduce((sum, speed) => sum + speed, 0) / this.speedHistory.length 
            : 0;
        
        // Format speed values with proper handling
        const formatSpeed = (speed) => {
            if (!speed || isNaN(speed) || !isFinite(speed)) return '0';
            return this.formatSpeedNumber(speed);
        };
        
        // Update enhanced speed display - use backend's pre-formatted speed when available
        const speedDisplay = progress.speed || `${formatSpeed(currentSpeed)} passwords/s`;
        this.updateProgressMetric('progress-speed', speedDisplay);
        this.updateProgressMetric('speed-peak', `${formatSpeed(this.peakSpeed)}/s`);
        this.updateProgressMetric('speed-average', `${formatSpeed(averageSpeed)}/s`);
        
        // Update other metrics
        this.updateProgressMetric('progress-passwords', (progress.current || 0).toLocaleString());
        this.updateProgressMetric('progress-elapsed', progress.elapsed_time || '0s');
        this.updateProgressMetric('progress-percent', `${(progress.percent || 0).toFixed(1)}%`);
        
    }

    updateProgressMetric(id, value) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = value;
        }
    }

    updateAttackResult(result, sessionId) {
        if (sessionId !== this.currentAttack) return;
        
        const resultEl = document.getElementById('attack-result');
        
        if (result.success) {
            resultEl.innerHTML = `
                <div class="result-success">
                    <h3 class="font-bold text-success mb-2">🎉 Secret Found!</h3>
                    <div class="mb-2">
                        <strong>Secret:</strong>
                        <div class="result-secret">${result.secret}</div>
                    </div>
                    <div class="grid grid-cols-2 text-sm">
                        <div><strong>Algorithm:</strong> ${result.algorithm}</div>
                        <div><strong>Mode:</strong> ${result.mode}</div>
                        <div><strong>Attempts:</strong> ${result.attempts.toLocaleString()}</div>
                        <div><strong>Duration:</strong> ${result.duration}</div>
                    </div>
                </div>
            `;
        } else {
            resultEl.innerHTML = `
                <div class="alert alert-error">
                    <h3 class="font-bold mb-2">Attack Failed</h3>
                    <p>No secret found after ${result.attempts.toLocaleString()} attempts in ${result.duration}</p>
                    ${result.error ? `<p class="text-sm mt-2">${result.error}</p>` : ''}
                </div>
            `;
        }
        
        this.currentAttack = null;
        this.updateAttackUI('completed');
    }

    onAttackStarted(data, sessionId) {
        console.log('Attack started:', sessionId);
        this.currentAttack = sessionId;
        
        // Reset speed tracking for new attack
        this.speedHistory = [];
        this.peakSpeed = 0;
        
        this.updateAttackUI('running');
    }

    onAttackStopped(data, sessionId) {
        console.log('Attack stopped:', sessionId);
        if (sessionId === this.currentAttack) {
            this.currentAttack = null;
            this.updateAttackUI('stopped');
        }
    }

    // File Upload
    async uploadWordlist(file) {
        const formData = new FormData();
        formData.append('wordlist', file);
        
        try {
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });
            
            if (response.ok) {
                const result = await response.json();
                this.showAlert('success', `File uploaded: ${result.filename}`);
                await this.loadWordlists();
            } else {
                const error = await response.json();
                this.showAlert('error', error.message);
            }
        } catch (error) {
            this.showAlert('error', `Upload failed: ${error.message}`);
        }
    }

    async loadWordlists() {
        try {
            const response = await fetch('/api/wordlists');
            if (response.ok) {
                const data = await response.json();
                this.displayWordlists(data.wordlists);
            }
        } catch (error) {
            console.error('Failed to load wordlists:', error);
        }
    }

    displayWordlists(wordlists) {
        const select = document.getElementById('wordlist-select');
        select.innerHTML = '<option value="">Select a wordlist...</option>';
        
        wordlists.forEach(wordlist => {
            const option = document.createElement('option');
            option.value = wordlist.path;
            option.textContent = `${wordlist.name} (${this.formatFileSize(wordlist.size)}) - ${wordlist.type}`;
            select.appendChild(option);
        });
    }

    // Utility Functions
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    formatNumber(n) {
        if (!n || isNaN(n) || !isFinite(n) || n === 0) return '0';
        if (n < 1000) return Math.round(n).toString();
        const k = 1000;
        const sizes = ['', 'K', 'M', 'B', 'T'];
        const i = Math.floor(Math.log(n) / Math.log(k));
        const value = n / Math.pow(k, i);
        return parseFloat(value.toFixed(1)) + sizes[i];
    }
    
    formatSpeedNumber(n) {
        if (!n || isNaN(n) || !isFinite(n) || n === 0) return '0';
        
        // Ensure we're dealing with a positive number
        const num = Math.abs(n);
        
        // Handle small numbers
        if (num < 1000) {
            return Math.round(num).toString();
        }
        
        // Use proper SI units for speed formatting
        const units = ['', 'K', 'M', 'G', 'T'];
        let unitIndex = 0;
        let value = num;
        
        while (value >= 1000 && unitIndex < units.length - 1) {
            value /= 1000;
            unitIndex++;
        }
        
        // Format with appropriate precision
        if (value >= 100) {
            return Math.round(value) + units[unitIndex];
        } else if (value >= 10) {
            return value.toFixed(1) + units[unitIndex];
        } else {
            return value.toFixed(2) + units[unitIndex];
        }
    }

    showAlert(type, message) {
        const alertsContainer = document.getElementById('alerts-container');
        const alert = document.createElement('div');
        alert.className = `alert alert-${type}`;
        alert.textContent = message;
        
        alertsContainer.appendChild(alert);
        
        setTimeout(() => {
            alert.remove();
        }, 5000);
    }

    // Event Listeners
    setupEventListeners() {
        // JWT Analysis
        document.getElementById('analyze-btn').addEventListener('click', () => this.analyzeJWT());
        
        // Attack Controls
        document.getElementById('start-attack-btn').addEventListener('click', () => this.startAttack());
        document.getElementById('stop-attack-btn').addEventListener('click', () => this.stopAttack());
        
        // Attack Type Change
        document.getElementById('attack-type').addEventListener('change', (e) => {
            this.toggleAttackOptions(e.target.value);
        });
        
        // Character Set Management
        this.setupCharsetListeners();
        
        // File Upload
        const fileInput = document.getElementById('file-input');
        const dropZone = document.getElementById('file-drop-zone');
        
        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.uploadWordlist(e.target.files[0]);
            }
        });
        
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('dragover');
        });
        
        dropZone.addEventListener('dragleave', () => {
            dropZone.classList.remove('dragover');
        });
        
        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('dragover');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                this.uploadWordlist(files[0]);
            }
        });
        
        dropZone.addEventListener('click', () => {
            fileInput.click();
        });
    }

    setupCharsetListeners() {
        // Mode selector buttons
        document.querySelectorAll('.mode-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                this.switchCharsetMode(btn.dataset.mode);
            });
        });
        
        // Preset buttons
        document.querySelectorAll('.preset-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                this.selectCharsetPreset(btn.dataset.charset);
            });
        });
        
        // Toggle switches
        document.querySelectorAll('.toggle-switch input').forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                // Switch to mix & match mode if not already there
                if (this.currentMode !== 'mixmatch') {
                    this.switchCharsetMode('mixmatch');
                } else {
                    this.updateCharsetFromCheckboxes();
                }
            });
        });
        
        // Custom charset input
        const customInput = document.getElementById('charset-custom');
        if (customInput) {
            customInput.addEventListener('input', () => {
                // Clear other selections when typing in advanced input
                this.selectedPreset = null;
                document.querySelectorAll('.preset-btn').forEach(btn => {
                    btn.classList.remove('active');
                });
                document.querySelectorAll('.toggle-switch input').forEach(cb => {
                    cb.checked = false;
                });
                this.updateCharsetPreview();
            });
            customInput.addEventListener('blur', () => this.processHashcatRules());
        }
        
        // Example rule clicks
        document.querySelectorAll('.example-item').forEach(item => {
            item.addEventListener('click', () => {
                const rule = item.dataset.rule;
                if (rule && customInput) {
                    // Switch to advanced mode first
                    if (this.currentMode !== 'advanced') {
                        this.switchCharsetMode('advanced');
                    }
                    customInput.value = rule;
                    this.processHashcatRules();
                }
            });
        });
        
        // Length input changes for keyspace calculation
        document.querySelectorAll('input[name="length_min"], input[name="length_max"]').forEach(input => {
            input.addEventListener('input', () => this.updateCharsetPreview());
        });
        
        // JWT token synchronization
        this.setupTokenSync();
        
        // Initialize with default preset
        this.selectCharsetPreset('lowercase');
        
        // Initialize preview
        this.updateCharsetPreview();
    }
    
    setupTokenSync() {
        const analyzeToken = document.getElementById('jwt-token');
        const attackToken = document.getElementById('attack-token');
        
        if (analyzeToken) {
            analyzeToken.addEventListener('input', (e) => {
                if (attackToken) {
                    attackToken.value = e.target.value;
                }
            });
            
            analyzeToken.addEventListener('paste', (e) => {
                setTimeout(() => {
                    if (attackToken) {
                        attackToken.value = analyzeToken.value;
                    }
                }, 10);
            });
        }
        
        if (attackToken) {
            attackToken.addEventListener('input', (e) => {
                if (analyzeToken) {
                    analyzeToken.value = e.target.value;
                }
            });
            
            attackToken.addEventListener('paste', (e) => {
                setTimeout(() => {
                    if (analyzeToken) {
                        analyzeToken.value = attackToken.value;
                    }
                }, 10);
            });
        }
    }
    
    switchCharsetMode(mode) {
        // Update mode selector buttons
        document.querySelectorAll('.mode-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.mode === mode);
        });
        
        // Update mode content visibility
        document.querySelectorAll('.charset-mode').forEach(modeEl => {
            modeEl.classList.toggle('active', modeEl.id === `mode-${mode}`);
        });
        
        this.currentMode = mode;
        
        // Clear selections when switching modes
        if (mode === 'presets') {
            // Clear checkboxes and advanced input
            document.querySelectorAll('.toggle-switch input').forEach(cb => {
                cb.checked = false;
            });
            const customInput = document.getElementById('charset-custom');
            if (customInput) {
                customInput.value = '';
            }
        } else if (mode === 'mixmatch') {
            // Clear preset selection and advanced input
            this.selectedPreset = null;
            document.querySelectorAll('.preset-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            const customInput = document.getElementById('charset-custom');
            if (customInput) {
                customInput.value = '';
            }
        } else if (mode === 'advanced') {
            // Clear presets and checkboxes
            this.selectedPreset = null;
            document.querySelectorAll('.preset-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            document.querySelectorAll('.toggle-switch input').forEach(cb => {
                cb.checked = false;
            });
        }
        
        // Always update preview after mode change
        this.updateCharsetPreview();
    }

    selectCharsetPreset(preset) {
        // Store the selected preset
        this.selectedPreset = preset;
        
        // Clear other selections
        document.querySelectorAll('.toggle-switch input').forEach(cb => {
            cb.checked = false;
        });
        
        // Clear advanced input
        const customInput = document.getElementById('charset-custom');
        if (customInput) {
            customInput.value = '';
        }
        
        // Highlight active preset button
        document.querySelectorAll('.preset-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.charset === preset);
        });
        
        // Stay in presets mode (don't switch modes)
        if (this.currentMode !== 'presets') {
            this.switchCharsetMode('presets');
        }
        
        // Update preview immediately
        this.updateCharsetPreview();
    }

    updateCharsetFromCheckboxes() {
        // Clear preset selection when using checkboxes
        this.selectedPreset = null;
        document.querySelectorAll('.preset-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // Clear advanced input when using checkboxes
        const customInput = document.getElementById('charset-custom');
        if (customInput) {
            customInput.value = '';
        }
        
        // Always update the preview
        this.updateCharsetPreview();
    }
    
    getCurrentCharset() {
        // Advanced mode: prioritize custom input
        if (this.currentMode === 'advanced') {
            const customInput = document.getElementById('charset-custom');
            if (customInput && customInput.value.trim()) {
                return customInput.value.trim();
            }
            return '';
        }
        
        // Mix & Match mode: get charset from checked toggles
        if (this.currentMode === 'mixmatch') {
            const checkboxes = document.querySelectorAll('.toggle-switch input:checked');
            let charset = '';
            checkboxes.forEach(cb => {
                const charsetName = cb.value;
                if (this.charsets[charsetName]) {
                    charset += this.charsets[charsetName];
                }
            });
            // Remove duplicates and return
            return [...new Set(charset)].join('');
        }
        
        // Presets mode: get charset from selected preset
        if (this.currentMode === 'presets') {
            if (this.selectedPreset && this.charsets[this.selectedPreset]) {
                return this.charsets[this.selectedPreset];
            }
            return '';
        }
        
        // Fallback
        return '';
    }

    processHashcatRules() {
        const input = document.getElementById('charset-custom');
        let value = input.value;
        
        // Process hashcat-style rules
        const hashcatRules = {
            '?l': this.charsets.lowercase,
            '?u': this.charsets.uppercase,
            '?d': this.charsets.digits,
            '?s': this.charsets.special,
            '?a': this.charsets.printable
        };
        
        // Replace hashcat rules with actual characters
        Object.entries(hashcatRules).forEach(([rule, chars]) => {
            value = value.replace(new RegExp('\\?' + rule.substring(1), 'g'), chars);
        });
        
        // Remove duplicates
        value = [...new Set(value)].join('');
        
        input.value = value;
        this.updateCharsetPreview();
    }

    updateCharsetPreview() {
        const previewEl = document.getElementById('charset-preview-text');
        const lengthEl = document.getElementById('charset-length');
        const keyspaceEl = document.getElementById('keyspace-estimate');
        
        if (!previewEl || !lengthEl) return;
        
        // Use the centralized charset getter for consistency
        let charset = this.getCurrentCharset();
        
        // Process hashcat rules if present in the charset
        if (charset) {
            const hashcatRules = {
                '?l': this.charsets.lowercase,
                '?u': this.charsets.uppercase,
                '?d': this.charsets.digits,
                '?s': this.charsets.special,
                '?a': this.charsets.printable
            };
            
            Object.entries(hashcatRules).forEach(([rule, chars]) => {
                // Fix regex pattern - escape the ? properly
                charset = charset.replace(new RegExp('\\?' + rule.substring(1), 'g'), chars);
            });
            
            // Remove duplicates
            charset = [...new Set(charset)].join('');
        }
        
        if (charset && charset.length > 0) {
            // Show first 80 characters with ellipsis if longer
            const display = charset.length > 80 ? charset.substring(0, 80) + '...' : charset;
            previewEl.textContent = display;
            lengthEl.textContent = charset.length;
            
            // Calculate keyspace estimate
            if (keyspaceEl) {
                const minLen = parseInt(document.querySelector('input[name="length_min"]')?.value) || 1;
                const maxLen = parseInt(document.querySelector('input[name="length_max"]')?.value) || 6;
                
                // Ensure valid length range
                if (minLen > 0 && maxLen > 0 && minLen <= maxLen) {
                    const keyspace = this.calculateKeyspace(charset.length, minLen, maxLen);
                    
                    if (keyspace === Infinity) {
                        keyspaceEl.textContent = '∞ (Very Large)';
                    } else if (keyspace > 1e15) {
                        keyspaceEl.textContent = '> 1 Quadrillion';
                    } else if (keyspace === 0) {
                        keyspaceEl.textContent = '0';
                    } else {
                        keyspaceEl.textContent = this.formatNumber(keyspace);
                    }
                } else {
                    keyspaceEl.textContent = 'Invalid length range';
                }
            }
        } else {
            previewEl.textContent = 'Select a character set...';
            lengthEl.textContent = '0';
            if (keyspaceEl) {
                keyspaceEl.textContent = '0';
            }
        }
    }
    
    calculateKeyspace(charsetSize, minLen, maxLen) {
        if (!charsetSize || charsetSize === 0 || !minLen || !maxLen || minLen < 0 || maxLen < 0) {
            return 0;
        }
        if (minLen > maxLen) return 0;
        
        // For very large keyspaces, use approximation to prevent overflow
        const maxSafeExponent = 15; // Roughly 10^15 combinations
        
        let total = 0;
        for (let i = minLen; i <= maxLen; i++) {
            // Check if the calculation would overflow
            const logCombinations = i * Math.log10(charsetSize);
            if (logCombinations > maxSafeExponent) {
                return Infinity;
            }
            
            const combinations = Math.pow(charsetSize, i);
            if (combinations === Infinity || !isFinite(combinations)) {
                return Infinity;
            }
            
            total += combinations;
            if (total === Infinity || !isFinite(total)) {
                return Infinity;
            }
        }
        
        return total;
    }

    toggleAttackOptions(attackType) {
        const wordlistOptions = document.getElementById('wordlist-options');
        const charsetOptions = document.getElementById('charset-options');
        
        wordlistOptions.style.display = attackType === 'wordlist' ? 'block' : 'none';
        charsetOptions.style.display = attackType === 'charset' ? 'block' : 'none';
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.jwtCrackApp = new JWTCrackApp();
});