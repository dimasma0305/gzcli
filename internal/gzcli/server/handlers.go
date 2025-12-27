package server

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
)

// HTML Templates
const homeTemplate = `{{define "home"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #0d1117;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            color: #c9d1d9;
        }
        .container {
            text-align: center;
            padding: 60px 40px;
            background: #161b22;
            border-radius: 8px;
            border: 1px solid #30363d;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
        }
        h1 { font-size: 2.5em; margin-bottom: 10px; color: #58a6ff; font-weight: 600; }
        p { font-size: 1.1em; color: #8b949e; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 {{.Title}}</h1>
        <p>{{.Message}}</p>
    </div>
</body>
</html>
{{end}}`

const challengeTemplate = `{{define "challenge"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - GZCLI Launcher</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #0d1117;
            min-height: 100vh;
            padding: 12px;
            color: #c9d1d9;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: #161b22;
            border-radius: 6px;
            padding: 20px;
            border: 1px solid #30363d;
        }
        h1 {
            color: #58a6ff;
            margin-bottom: 6px;
            font-size: 1.5em;
            font-weight: 600;
        }
        .meta {
            color: #8b949e;
            margin-bottom: 16px;
            font-size: 0.8em;
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
        .meta span {
            padding: 3px 6px;
            background: #21262d;
            border-radius: 3px;
            border: 1px solid #30363d;
            white-space: nowrap;
        }
        .ports {
            margin-bottom: 12px;
            padding: 10px;
            background: #0d1117;
            border-radius: 4px;
            border: 1px solid #30363d;
            font-size: 0.8em;
        }
        .ports strong {
            color: #58a6ff;
            margin-right: 6px;
            display: block;
            margin-bottom: 6px;
        }
        .ports .port {
            display: inline-block;
            margin: 3px 3px 3px 0;
            padding: 3px 8px;
            background: #21262d;
            border: 1px solid #3fb950;
            border-radius: 3px;
            color: #3fb950;
            font-family: 'Courier New', monospace;
            font-size: 0.85em;
        }
        .status {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 10px;
            background: #0d1117;
            border-radius: 4px;
            margin-bottom: 12px;
            font-size: 0.8em;
            border: 1px solid #30363d;
        }
        .status-dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            flex-shrink: 0;
        }
        .status-dot.connected { background: #3fb950; }
        .status-dot.disconnected { background: #f85149; }
        .status-dot.connecting { background: #d29922; }
        .controls {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 8px;
            margin-bottom: 12px;
        }
        button {
            padding: 10px;
            font-size: 0.85em;
            border: 1px solid #30363d;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.15s;
            font-weight: 500;
            background: #21262d;
            color: #c9d1d9;
            white-space: nowrap;
            touch-action: manipulation;
        }
        button:hover:not(:disabled) {
            background: #30363d;
            border-color: #58a6ff;
        }
        button:active:not(:disabled) {
            transform: scale(0.98);
        }
        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        .btn-start:hover:not(:disabled) { border-color: #3fb950; }
        .btn-restart:hover:not(:disabled) { border-color: #d29922; }
        .voting-panel {
            display: none;
            padding: 12px;
            background: #21262d;
            border-radius: 4px;
            margin-bottom: 12px;
            border: 1px solid #d29922;
        }
        .voting-panel h3 {
            margin-bottom: 10px;
            color: #d29922;
            font-size: 0.9em;
            font-weight: 600;
        }
        .vote-progress {
            display: flex;
            gap: 6px;
            margin-bottom: 10px;
        }
        .vote-bar {
            flex: 1;
            height: 20px;
            background: #0d1117;
            border-radius: 3px;
            overflow: hidden;
            border: 1px solid #30363d;
        }
        .vote-bar-fill {
            height: 100%;
            transition: width 0.3s;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 0.75em;
            font-weight: 600;
        }
        .vote-yes { background: #2ea043; }
        .vote-no { background: #da3633; }
        .vote-buttons {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 6px;
        }
        .vote-info {
            font-size: 0.75em;
            color: #8b949e;
            margin-bottom: 8px;
        }
        .info-panel {
            padding: 8px 10px;
            border-radius: 4px;
            margin-bottom: 6px;
            font-size: 0.8em;
            border: 1px solid;
        }
        .info-panel.success { background: #0d1117; color: #3fb950; border-color: #2ea043; }
        .info-panel.error { background: #0d1117; color: #f85149; border-color: #da3633; }
        .info-panel.info { background: #0d1117; color: #58a6ff; border-color: #1f6feb; }
        .messages {
            max-height: 150px;
            overflow-y: auto;
        }
        .messages::-webkit-scrollbar { width: 4px; }
        .messages::-webkit-scrollbar-track { background: #0d1117; }
        .messages::-webkit-scrollbar-thumb { background: #30363d; border-radius: 2px; }

        /* Mobile optimizations */
        @media (max-width: 480px) {
            body { padding: 8px; }
            .container { padding: 16px; }
            h1 { font-size: 1.3em; }
            .meta { font-size: 0.75em; gap: 6px; }
            button { padding: 8px; font-size: 0.8em; }
            .controls { gap: 6px; }
        }

        /* Tablet */
        @media (min-width: 481px) and (max-width: 768px) {
            .container { margin: 20px auto; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Name}}</h1>
        <div class="meta">
            <span>📁 {{.Category}}</span>
            <span>🎯 {{.Event}}</span>
            <span id="user-count">👥 0 users</span>
        </div>

        {{if .Ports}}
        <div class="ports">
            <strong>Ports:</strong>
            {{range .Ports}}
            <span class="port">{{.}}</span>
            {{end}}
        </div>
        {{end}}

        <div class="status">
            <div class="status-dot connecting" id="connection-status"></div>
            <div>
                <strong>Connection:</strong> <span id="connection-text">Connecting...</span><br>
                <strong>Status:</strong> <span id="challenge-status">Unknown</span>
            </div>
        </div>

        <div id="messages" class="messages"></div>

        <div id="voting-panel" class="voting-panel">
            <h3>🗳️ Restart Vote in Progress</h3>
            <div class="vote-progress">
                <div class="vote-bar">
                    <div class="vote-bar-fill vote-yes" id="yes-bar" style="width: 0%">
                        <span id="yes-percent">0%</span>
                    </div>
                </div>
                <div class="vote-bar">
                    <div class="vote-bar-fill vote-no" id="no-bar" style="width: 0%">
                        <span id="no-percent">0%</span>
                    </div>
                </div>
            </div>
            <p class="vote-info" id="vote-info">Waiting for votes...</p>
            <div class="vote-buttons">
                <button class="btn-start" onclick="vote('yes')">✅ Vote Yes</button>
                <button class="btn-stop" onclick="vote('no')">❌ Vote No</button>
            </div>
            <div style="margin-top: 8px; text-align: center;">
                 <button onclick="stopAlarm()" style="font-size: 0.75em; padding: 4px 8px;">🔕 Stop Sound</button>
            </div>
        </div>

        <div class="controls">
            <button class="btn-start" id="btn-start" onclick="startChallenge()">▶️ Start</button>
            <button class="btn-restart" id="btn-restart" onclick="requestRestart()">🔄 Restart</button>
        </div>
    </div>

    <script>
        const slug = '{{.Slug}}';
        let ws = null;
        let reconnectAttempts = 0;
        const maxReconnectDelay = 30000;

        function connect() {
            updateConnectionStatus('connecting');

            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/' + slug + '/ws';

            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                console.log('WebSocket connected');
                updateConnectionStatus('connected');
                reconnectAttempts = 0;
                requestNotificationPermission();
            };

            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    handleMessage(msg);
                } catch (e) {
                    console.error('Failed to parse message:', e);
                }
            };

            ws.onclose = (event) => {
                console.log('WebSocket disconnected', event.code, event.reason);
                if (event.code !== 1000) { // Not normal closure
                    updateConnectionStatus('disconnected');
                    reconnect();
                }
            };

            ws.onerror = (error) => {
                // Don't log error during reconnection attempts
                if (reconnectAttempts === 0) {
                    console.error('WebSocket error:', error);
                }
            };
        }

        function reconnect() {
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), maxReconnectDelay);
            reconnectAttempts++;

            if (reconnectAttempts <= 3) {
                console.log('Reconnecting in ' + delay + 'ms... (attempt ' + reconnectAttempts + ')');
            }

            setTimeout(connect, delay);
        }

        function send(type, data = {}) {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type, data }));
            }
        }

        function handleMessage(msg) {
            console.log('Received:', msg);

            switch (msg.type) {
                case 'pong':
                    break;
                case 'status':
                    updateStatus(msg.data);
                    break;
                case 'vote_started':
                    showVotingPanel();
                    playAlarm();
                    showMessage('info', 'Restart vote initiated by user');
                    break;
                case 'vote_update':
                    updateVoteProgress(msg.data);
                    break;
                case 'vote_ended':
                    hideVotingPanel();
                    stopAlarm();
                    showMessage('info', 'Vote ended: ' + msg.data.result);
                    break;
                case 'error':
                    showMessage('error', msg.message);
                    break;
                case 'info':
                    showMessage('info', msg.message);
                    if (msg.message.includes('started successfully') || msg.message.includes('ready')) {
                        showNotification('Challenge Ready', msg.message);
                    }
                    break;
            }
        }

        function updateConnectionStatus(status) {
            const dot = document.getElementById('connection-status');
            const text = document.getElementById('connection-text');
            dot.className = 'status-dot ' + status;

            const statusText = {
                'connecting': 'Connecting...',
                'connected': 'Connected',
                'disconnected': 'Disconnected'
            };
            text.textContent = statusText[status] || status;
        }

        function updateStatus(data) {
            document.getElementById('challenge-status').textContent = data.status;
            document.getElementById('user-count').textContent = '👥 ' + data.connected_users + ' user' + (data.connected_users !== 1 ? 's' : '');

            // Update ports section
            const portsContainer = document.querySelector('.ports');
            if (data.status === 'running' && data.allocated_ports && data.allocated_ports.length > 0) {
                let html = '<strong>Ports:</strong> ';
                data.allocated_ports.forEach(port => {
                    html += '<span class="port">' + port + '</span>';
                });

                if (portsContainer) {
                    portsContainer.innerHTML = html;
                    portsContainer.style.display = 'block';
                } else {
                    // Create if not exists
                    const newContainer = document.createElement('div');
                    newContainer.className = 'ports';
                    newContainer.innerHTML = html;
                    // Insert after meta
                    const meta = document.querySelector('.meta');
                    meta.parentNode.insertBefore(newContainer, meta.nextSibling);
                }
            } else {
                if (portsContainer) {
                    portsContainer.style.display = 'none';
                }
            }

            const startBtn = document.getElementById('btn-start');
            const restartBtn = document.getElementById('btn-restart');

            startBtn.disabled = ['starting', 'running', 'stopping'].includes(data.status);
            restartBtn.disabled = ['starting', 'stopping', 'restarting'].includes(data.status);
        }

        function showVotingPanel() {
            document.getElementById('voting-panel').style.display = 'block';
        }

        function hideVotingPanel() {
            document.getElementById('voting-panel').style.display = 'none';
        }

        function updateVoteProgress(data) {
            document.getElementById('yes-bar').style.width = data.yes_percent + '%';
            document.getElementById('yes-percent').textContent = Math.round(data.yes_percent) + '%';
            document.getElementById('no-bar').style.width = data.no_percent + '%';
            document.getElementById('no-percent').textContent = Math.round(data.no_percent) + '%';
            document.getElementById('vote-info').textContent =
                'Total voters: ' + data.total_users + ' (waiting 15s handling...)';
        }

        function showMessage(type, text) {
            const messagesDiv = document.getElementById('messages');
            const msgDiv = document.createElement('div');
            msgDiv.className = 'info-panel ' + type;
            msgDiv.textContent = text;
            messagesDiv.insertBefore(msgDiv, messagesDiv.firstChild);

            // Remove old messages (keep last 5)
            while (messagesDiv.children.length > 5) {
                messagesDiv.removeChild(messagesDiv.lastChild);
            }
        }

        function startChallenge() {
            send('start');
        }

        function requestRestart() {
            send('restart');
        }

        function vote(value) {
            send('vote', { value });
        }

        function requestNotificationPermission() {
            if ('Notification' in window && Notification.permission === 'default') {
                Notification.requestPermission();
            }
        }

        let audioCtx;
        let oscillator;
        let gainNode;

        function playAlarm() {
            if (!audioCtx) {
                audioCtx = new (window.AudioContext || window.webkitAudioContext)();
            }

            // Create oscillator
            oscillator = audioCtx.createOscillator();
            gainNode = audioCtx.createGain();

            oscillator.connect(gainNode);
            gainNode.connect(audioCtx.destination);

            oscillator.type = 'sawtooth';
            oscillator.frequency.setValueAtTime(440, audioCtx.currentTime); // Start at 440Hz

            // Siren effect: pitch modulation
            const now = audioCtx.currentTime;
            oscillator.frequency.linearRampToValueAtTime(880, now + 0.5);
            oscillator.frequency.linearRampToValueAtTime(440, now + 1.0);

            // Loop the frequency sweep
            // Using setInterval for simplicity in this context, or precise scheduling
            // For a simple alarm, we can just let it run or restart it.
            // A better way for siren is LFO, but let's stick to a simple repeating ramp manually or just a simple beep-beep if easier.
            // Let's do a simple LFO-like effect using oscillator parameters

            // Re-creating oscillator for a proper LFO modulation is better:
            // Carrier
            const carrier = audioCtx.createOscillator();
            carrier.type = 'sawtooth';
            carrier.frequency.value = 600;

            // LFO
            const lfo = audioCtx.createOscillator();
            lfo.type = 'sine';
            lfo.frequency.value = 2; // 2Hz siren speed

            const lfoGain = audioCtx.createGain();
            lfoGain.gain.value = 200; // Modulation depth

            lfo.connect(lfoGain);
            lfoGain.connect(carrier.frequency);

            carrier.connect(gainNode);
            gainNode.connect(audioCtx.destination);

            carrier.start();
            lfo.start();

            oscillator = { stop: () => { carrier.stop(); lfo.stop(); } }; // Mock object for stop

            // Auto stop after 15 seconds
            setTimeout(stopAlarm, 15000);
        }

        function stopAlarm() {
            if (oscillator) {
                try {
                    oscillator.stop();
                } catch (e) {}
                oscillator = null;
            }
        }

        function showNotification(title, body) {
            if ('Notification' in window && Notification.permission === 'granted') {
                new Notification(title, { body, icon: '/favicon.ico' });
            }
        }

        // Start connection
        connect();

        // Send ping every 30 seconds
        setInterval(() => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                send('ping');
            }
        }, 30000);
    </script>
</body>
</html>
{{end}}`

// Server handles HTTP requests
type Server struct {
	challenges *ChallengeManager
	wsManager  *WSManager
	templates  *template.Template
}

// NewServer creates a new HTTP server handler
func NewServer(challenges *ChallengeManager, wsManager *WSManager) *Server {
	return &Server{
		challenges: challenges,
		wsManager:  wsManager,
	}
}

// LoadTemplates loads HTML templates
func (s *Server) LoadTemplates() error {
	// Parse embedded or file-based templates
	tmpl, err := template.New("").Parse(homeTemplate)
	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(challengeTemplate)
	if err != nil {
		return err
	}

	s.templates = tmpl
	return nil
}

// HandleHome handles the homepage
func (s *Server) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":   "GZCLI Challenge Launcher",
		"Message": "Welcome to GZCLI Challenge Launcher",
	}

	if err := s.templates.ExecuteTemplate(w, "home", data); err != nil {
		log.Error("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// HandleChallenge handles the challenge launcher page
func (s *Server) HandleChallenge(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path
	path := strings.TrimPrefix(r.URL.Path, "/")
	slug := strings.TrimSuffix(path, "/ws")

	// Handle WebSocket upgrade
	if strings.HasSuffix(r.URL.Path, "/ws") {
		s.wsManager.HandleWebSocket(w, r, slug)
		return
	}

	// Get challenge info
	challenge, exists := s.challenges.GetChallenge(slug)
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Determine initial ports to display
	var displayPorts []string
	if challenge.GetStatus() == StatusRunning {
		displayPorts = challenge.GetAllocatedPorts()
	}

	// Render challenge page
	data := map[string]interface{}{
		"Title":    challenge.Name,
		"Slug":     challenge.Slug,
		"Name":     challenge.Name,
		"Event":    challenge.EventName,
		"Category": challenge.Category,
		"Ports":    displayPorts,
	}

	if err := s.templates.ExecuteTemplate(w, "challenge", data); err != nil {
		log.Error("Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SetupRoutes sets up HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Homepage
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.HandleHome(w, r)
		} else {
			// All other paths are treated as challenge slugs
			s.HandleChallenge(w, r)
		}
	})

	return mux
}
