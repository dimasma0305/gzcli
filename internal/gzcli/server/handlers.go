package server

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/dimasma0305/gzcli/internal/log"
)

//go:embed notification.mp3
var notificationSound []byte

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

    <!-- Tailwind CSS -->
    <script src="https://cdn.tailwindcss.com"></script>

    <!-- Google Fonts: Space Grotesk (Headings) & Inter (Body) & JetBrains Mono (Code) -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&family=Space+Grotesk:wght@500;700&display=swap" rel="stylesheet">

    <script>
        tailwind.config = {
            theme: {
                extend: {
                    fontFamily: {
                        sans: ['Inter', 'sans-serif'],
                        display: ['Space Grotesk', 'sans-serif'],
                        mono: ['JetBrains Mono', 'monospace'],
                    },
                    colors: {
                        glass: 'rgba(255, 255, 255, 0.03)',
                        'glass-hover': 'rgba(255, 255, 255, 0.08)',
                        'glass-border': 'rgba(255, 255, 255, 0.1)',
                        brand: '#6366f1', // Indigo
                        accent: '#a855f7', // Purple
                        success: '#22c55e',
                        danger: '#ef4444',
                        warning: '#eab308',
                    },
                    animation: {
                        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
                        'gradient-x': 'gradient-x 15s ease infinite',
                        'fadeIn': 'fadeIn 0.2s ease-out forwards',
                    },
                    keyframes: {
                        'gradient-x': {
                            '0%, 100%': {
                                'background-size': '200% 200%',
                                'background-position': 'left center'
                            },
                            '50%': {
                                'background-size': '200% 200%',
                                'background-position': 'right center'
                            },
                        },
                        fadeIn: {
                            '0%': { opacity: '0', transform: 'translateY(-5px)' },
                            '100%': { opacity: '1', transform: 'translateY(0)' },
                        }
                    }
                }
            }
        }
    </script>

    <style>
        body {
            background-color: #050505;
            background-image:
                radial-gradient(circle at 15% 50%, rgba(99, 102, 241, 0.15) 0%, transparent 25%),
                radial-gradient(circle at 85% 30%, rgba(168, 85, 247, 0.15) 0%, transparent 25%);
        }

        .bento-card {
            background: var(--tw-colors-glass);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid var(--tw-colors-glass-border);
            border-radius: 1.5rem;
            transition: all 0.3s ease;
        }

        .bento-card:hover {
            border-color: rgba(255, 255, 255, 0.2);
            background: var(--tw-colors-glass-hover);
        }

        /* Custom Scrollbar */
        .custom-scroll::-webkit-scrollbar { width: 6px; }
        .custom-scroll::-webkit-scrollbar-track { background: transparent; }
        .custom-scroll::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 10px; }
        .custom-scroll::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.2); }
    </style>
</head>
<body class="text-white min-h-screen p-4 md:p-8 flex flex-col items-center justify-center selection:bg-brand selection:text-white">

    <div class="max-w-6xl w-full mx-auto grid grid-cols-1 md:grid-cols-12 gap-6">

        <!-- Header / Nav -->
        <div class="col-span-1 md:col-span-12 flex justify-between items-center mb-4">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 flex items-center justify-center p-1.5">
                    <svg viewBox="0 0 4800 4800" class="w-full h-full">
                        <path fill="white" fill-rule="evenodd" class="icon-EMC348" d="M2994.48,4244.61L505.28,2807.47V1992.53l256.572-148.14L1287,2285l258-307,160.39,135.56L1209.27,2400l1786.1,1031.21V2427.79L3517,1806,2420.98,886.5l573.5-331.11,705.76,407.474V3837.14Z"></path>
                        <g id="Flag">
                            <path id="Flag_0" fill="#00bfa5" fill-rule="evenodd" d="M1280.55,582.029L2046.6,1224.82l-771.35,919.25L509.21,1501.28Z"></path>
                            <path id="Flag_1" fill="#007f6e" fill-rule="evenodd" d="M1225.95,1580.54l306.42,257.11-257.12,306.42Z"></path>
                            <path id="Flag_2" fill="#1de9b6" fill-rule="evenodd" d="M2636.97,2699.25l-32.14,38.31-264.98,315.78L1812.4,2748.51l332.8-396.63-919.25-771.34L1997.3,661.284,3376.18,1818.3ZM1880,3601.24l0.15-.04L1351.4,4231.34,891.769,3845.67l460.361-548.64Z"></path>
                        </g>
                    </svg>
                </div>
                <h1 class="font-display font-bold text-2xl tracking-tight">GZCLI Launcher</h1>
            </div>
            <div class="flex items-center gap-4">
                <div class="hidden md:flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/5 border border-white/10 text-xs font-medium text-gray-400">
                    <span id="connection-dot" class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
                    <span id="connection-text">Disconnected</span>
                </div>
            </div>
        </div>

        <!-- 1. Hero Card (Challenge Info) -->
        <div class="col-span-1 md:col-span-8 bento-card p-8 relative overflow-hidden group">
            <div class="absolute top-0 right-0 p-8 opacity-20 group-hover:opacity-40 transition-opacity duration-500">
                 <!-- Abstract Graphic -->
                 <svg width="120" height="120" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" class="text-white transform rotate-12">
                    <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
                </svg>
            </div>

            <div class="relative z-10 flex flex-col h-full justify-between">
                <div>
                    <div class="flex gap-2 mb-4">
                        <span class="px-3 py-1 rounded-full text-xs font-medium bg-brand/20 text-brand border border-brand/20">{{.Category}}</span>
                        <span class="px-3 py-1 rounded-full text-xs font-medium bg-white/5 text-gray-400 border border-white/10">{{.Event}}</span>
                    </div>
                    <h2 class="font-display text-5xl md:text-6xl font-bold mb-2 tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-white via-white to-gray-400">
                        {{.Name}}
                    </h2>
                    <p class="text-gray-400 max-w-lg mt-2 text-lg">
                        {{.Description}}
                    </p>
                </div>

                <div class="mt-8 flex items-center gap-4 text-sm text-gray-500">
                    <span id="user-count" class="flex items-center gap-2">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
                        0 users online
                    </span>
                </div>
            </div>
        </div>

        <!-- 2. Control Center Card -->
        <div class="col-span-1 md:col-span-4 bento-card p-6 flex flex-col justify-center items-center text-center relative overflow-hidden">
            <!-- Background Glow -->
            <div id="status-glow" class="absolute inset-0 bg-brand/10 blur-3xl opacity-0 transition-opacity duration-700"></div>

            <div class="relative z-10 w-full">
                <div class="mb-2 text-xs font-mono text-gray-500 uppercase tracking-widest">Instance Status</div>
                <div id="status-text" class="text-2xl font-display font-bold text-white mb-8">Unknown</div>

                <div class="flex gap-4 justify-center w-full">
                    <button id="btn-start" onclick="startChallenge()" disabled class="group relative flex items-center justify-center w-20 h-20 rounded-2xl bg-white text-black hover:scale-105 transition-all duration-300 shadow-xl shadow-white/10 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100">
                        <svg class="w-8 h-8 fill-current translate-x-0.5 group-hover:scale-110 transition-transform" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                    </button>

                    <button id="btn-restart" onclick="requestRestart()" disabled class="group flex items-center justify-center w-20 h-20 rounded-2xl bg-white/5 border border-white/10 text-white hover:bg-white/10 hover:border-white/20 transition-all duration-300 disabled:opacity-30 disabled:cursor-not-allowed">
                        <svg class="w-8 h-8 stroke-current group-hover:rotate-180 transition-transform duration-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                    </button>
                </div>
            </div>
        </div>

        <!-- 3. Voting Card -->
        <div id="voting-panel" class="hidden col-span-1 md:col-span-12 bento-card p-1 bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500">
            <div class="bg-black/90 h-full w-full rounded-[1.3rem] p-6 flex flex-col md:flex-row items-center justify-between gap-6 backdrop-blur-xl">
                <div class="flex items-center gap-4">
                    <div class="p-3 bg-white/10 rounded-xl animate-bounce">
                        <span class="text-2xl">🗳️</span>
                    </div>
                    <div>
                        <h3 class="font-bold text-lg">Restart Requested</h3>
                        <p id="vote-info" class="text-gray-400 text-sm">Consensus required to reboot instance.</p>
                    </div>
                </div>

                <div class="flex-1 w-full md:max-w-md">
                    <div class="flex justify-between text-xs font-mono mb-2 text-gray-400">
                        <span>YES</span>
                        <span>NO</span>
                    </div>
                    <div class="h-4 bg-white/10 rounded-full overflow-hidden flex relative">
                        <div id="yes-bar" class="h-full bg-success transition-all duration-500 flex items-center justify-center text-[10px] font-bold text-black" style="width: 0%"></div>
                        <div id="no-bar" class="h-full bg-danger transition-all duration-500 flex items-center justify-center text-[10px] font-bold text-white" style="width: 0%"></div>
                    </div>
                </div>

                <div class="flex items-center gap-3">
                    <button onclick="vote('yes')" class="px-6 py-2.5 rounded-xl bg-success/20 text-success border border-success/20 hover:bg-success hover:text-black font-semibold transition-all">Vote Yes</button>
                    <button onclick="vote('no')" class="px-6 py-2.5 rounded-xl bg-danger/20 text-danger border border-danger/20 hover:bg-danger hover:text-white font-semibold transition-all">Vote No</button>
                    <button onclick="stopAlarm()" class="w-10 h-10 flex items-center justify-center rounded-xl border border-white/10 hover:bg-white/10 text-gray-400 hover:text-white transition-all" title="Mute Sound">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" stroke-linejoin="round"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2"></path></svg>
                    </button>
                </div>
            </div>
        </div>

        <!-- 4. Ports Card -->
        <div class="col-span-1 md:col-span-4 bento-card p-6 flex flex-col">
            <h3 class="font-display font-bold text-lg mb-4 flex items-center gap-2">
                <svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
                Active Ports
            </h3>

            <div id="ports-list" class="space-y-3 flex-1 overflow-y-auto custom-scroll min-h-[140px]">
                {{if .Ports}}
                    {{range .Ports}}
                    <div class="group flex items-center justify-between p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 hover:border-white/20 transition-all">
                        <div class="flex flex-col">
                            <span class="text-xs text-gray-500 font-mono">TCP / Port {{.}}</span>
                            <span class="text-sm font-mono text-brand group-hover:text-white transition-colors">Port {{.}}</span>
                        </div>
                    </div>
                    {{end}}
                {{else}}
                <!-- Placeholder State -->
                <div class="h-full flex flex-col items-center justify-center text-gray-600 text-sm border border-dashed border-gray-800 rounded-xl">
                    <span>No active ports</span>
                </div>
                {{end}}
            </div>
        </div>

        <!-- 5. Terminal / Logs Card -->
        <div class="col-span-1 md:col-span-8 bento-card p-0 overflow-hidden flex flex-col h-80 md:h-auto">
            <div class="px-6 py-4 border-b border-white/5 bg-black/20 flex items-center justify-between">
                <h3 class="font-mono text-sm text-gray-400">system_logs.log</h3>
                <div class="flex gap-1.5">
                    <div class="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50"></div>
                    <div class="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50"></div>
                    <div class="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50"></div>
                </div>
            </div>
            <div id="messages" class="flex-1 p-6 font-mono text-xs md:text-sm overflow-y-auto custom-scroll space-y-2 bg-black/40 text-gray-300">
                <div class="text-gray-600 italic">// Waiting for connection...</div>
            </div>
        </div>

    </div>

    <script>
        // Configuration
        const slug = '{{.Slug}}';

        let ws = null;
        let reconnectAttempts = 0;
        const maxReconnectDelay = 30000;

        // --- Connection Logic ---
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
                case 'pong': break;
                case 'status': updateStatus(msg.data); break;
                case 'vote_started':
                    showVotingPanel();
                    playAlarm();
                    showMessage('info', 'Restart vote initiated by user');
                    break;
                case 'vote_update': updateVoteProgress(msg.data); break;
                case 'vote_ended':
                    hideVotingPanel();
                    stopAlarm();
                    showMessage('info', 'Vote ended: ' + msg.data.result);
                    break;
                case 'error': showMessage('error', msg.message); break;
                case 'info':
                    showMessage('info', msg.message);
                    if (msg.message.includes('started successfully') || msg.message.includes('ready')) {
                        showNotification('Challenge Ready', msg.message);
                    }
                    break;
            }
        }

        function updateConnectionStatus(status) {
            const dot = document.getElementById('connection-dot');
            const text = document.getElementById('connection-text');
            dot.className = 'w-2 h-2 rounded-full ' + (status === 'connected' ? 'bg-green-500' :
                status === 'connecting' ? 'bg-yellow-500 animate-pulse' : 'bg-red-500');

            const statusText = {
                'connecting': 'Connecting...',
                'connected': 'Connected',
                'disconnected': 'Disconnected'
            };
            text.textContent = statusText[status] || status;
        }

        function copyToClipboard(text, element) {
            navigator.clipboard.writeText(text).then(function() {
                let btn = element;
                if (element.tagName !== 'BUTTON') {
                    const parent = element.closest('.group');
                    btn = parent.querySelector('button');
                }

                if (btn) {
                    const copyIcon = btn.querySelector('.copy-icon');
                    const checkIcon = btn.querySelector('.check-icon');

                    if (copyIcon && checkIcon) {
                        copyIcon.classList.add('hidden');
                        checkIcon.classList.remove('hidden');

                        setTimeout(function() {
                            copyIcon.classList.remove('hidden');
                            checkIcon.classList.add('hidden');
                        }, 2000);
                    }
                }
                showMessage('success', 'Port copied to clipboard: ' + text);
            }).catch(function(err) {
                console.error('Could not copy text: ', err);
                showMessage('error', 'Failed to copy port');
            });
        }

        function updateStatus(data) {
            const statusEl = document.getElementById('status-text');
            if (statusEl) statusEl.textContent = data.status || 'Unknown';

            const countEl = document.getElementById('user-count');
            if (countEl) {
                countEl.innerHTML =
                    '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>' +
                    ' ' + data.connected_users + ' user' + (data.connected_users !== 1 ? 's' : '') + ' online';
            }

            // Update ports section
            const portsList = document.getElementById('ports-list');
            if (portsList) {
                if (data.status === 'running' && data.allocated_ports && data.allocated_ports.length > 0) {
                    let html = '';
                    data.allocated_ports.forEach(function(portMapping) {
                        const parts = portMapping.split(':');
                        const extPort = parts[0];
                        const intPort = parts[1];
                        const hostname = window.location.hostname;
                        const httpUrl = 'http://' + hostname + ':' + extPort;
                        const ncCmd = 'nc ' + hostname + ' ' + extPort;

                        html += '<div class="group flex flex-col p-3 rounded-lg bg-white/5 border border-white/5 hover:bg-white/10 hover:border-white/20 transition-all gap-2">' +
                            '<div class="flex items-center justify-between">' +
                                '<div class="flex flex-col">' +
                                    '<span class="text-xs text-gray-500 font-mono text-gray-400">TCP Port Mapping</span>' +
                                    '<div class="flex items-center gap-2">' +
                                        '<span class="text-lg font-mono font-bold text-white transition-colors">' + extPort + '</span>' +
                                        '<span class="text-sm text-gray-500 font-mono">→</span>' +
                                        '<span class="text-sm text-gray-400 font-mono">' + intPort + '</span>' +
                                    '</div>' +
                                '</div>' +
                            '</div>' +
                            '<div class="flex gap-2">' +
                                '<button onclick="copyToClipboard(\'' + httpUrl + '\', this)" class="flex-1 flex items-center justify-center gap-2 p-2 rounded-lg bg-brand/10 border border-brand/20 hover:bg-brand/20 text-xs font-mono text-brand transition-all" title="Copy HTTP URL">' +
                                    '<span class="copy-icon flex items-center gap-1"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path></svg> HTTP</span>' +
                                    '<span class="check-icon hidden text-green-500 font-bold">Copied!</span>' +
                                '</button>' +
                                '<button onclick="copyToClipboard(\'' + ncCmd + '\', this)" class="flex-1 flex items-center justify-center gap-2 p-2 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 text-xs font-mono text-gray-300 transition-all" title="Copy NC Command">' +
                                    '<span class="copy-icon flex items-center gap-1"><svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg> NC</span>' +
                                    '<span class="check-icon hidden text-green-500 font-bold">Copied!</span>' +
                                '</button>' +
                            '</div>' +
                        '</div>';
                    });
                    portsList.innerHTML = html;
                } else {
                     portsList.innerHTML =
                    '<div class="h-full flex flex-col items-center justify-center text-gray-600 text-sm border border-dashed border-gray-800 rounded-xl">' +
                        '<span>No active ports</span>' +
                    '</div>';
                }
            }

            const startBtn = document.getElementById('btn-start');
            const restartBtn = document.getElementById('btn-restart');

            if (startBtn) startBtn.disabled = ['starting', 'running', 'stopping'].includes(data.status);
            if (restartBtn) restartBtn.disabled = ['starting', 'stopping', 'restarting'].includes(data.status);
        }

        function showVotingPanel() {
            const panel = document.getElementById('voting-panel');
            if (panel) panel.style.display = 'block';
        }

        function hideVotingPanel() {
            const panel = document.getElementById('voting-panel');
            if (panel) panel.style.display = 'none';
        }

        function updateVoteProgress(data) {
            const yesBar = document.getElementById('yes-bar');
            const noBar = document.getElementById('no-bar');
            const info = document.getElementById('vote-info');

            if (yesBar) yesBar.style.width = data.yes_percent + '%';
            if (noBar) noBar.style.width = data.no_percent + '%';
            if (info) info.textContent = 'Total voters: ' + data.total_users + ' (waiting 15s handling...)';
        }

        function showMessage(type, text) {
            const messagesDiv = document.getElementById('messages');
            if (!messagesDiv) return;

            const msgDiv = document.createElement('div');
            // Mapping msg types to Tailwind classes
            let colorClass = 'text-brand';
            if (type === 'error') colorClass = 'text-red-500';
            else if (type === 'success') colorClass = 'text-green-500';

            msgDiv.className = 'p-2 rounded border border-white/5 bg-white/5 ' + colorClass;
            msgDiv.textContent = text;
            messagesDiv.insertBefore(msgDiv, messagesDiv.firstChild);

            // Remove old messages (keep last 5)
            while (messagesDiv.children.length > 5) {
                messagesDiv.removeChild(messagesDiv.lastChild);
            }
        }

        function startChallenge() { send('start'); }
        function requestRestart() { send('restart'); }
        function vote(value) { send('vote', { value }); }

        function requestNotificationPermission() {
            if ('Notification' in window && Notification.permission === 'default') {
                Notification.requestPermission();
            }
        }

        let alarmAudio = new Audio('/notification.mp3');
        alarmAudio.loop = true;

        function playAlarm() {
            alarmAudio.currentTime = 0;
            alarmAudio.play().catch(e => console.error("Error playing audio:", e));
            setTimeout(stopAlarm, 15000);
        }

        function stopAlarm() {
            alarmAudio.pause();
            alarmAudio.currentTime = 0;
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
		"Title":       challenge.Name,
		"Slug":        challenge.Slug,
		"Name":        challenge.Name,
		"Description": challenge.Description,
		"Event":       challenge.EventName,
		"Category":    challenge.Category,
		"Ports":       displayPorts,
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
	mux.HandleFunc("/notification.mp3", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(notificationSound)))
		if _, err := w.Write(notificationSound); err != nil {
			log.Error("Failed to write notification sound: %v", err)
		}
	})

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
