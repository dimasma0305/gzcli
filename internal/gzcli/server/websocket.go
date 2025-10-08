package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/dimasma0305/gzcli/internal/log"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true // Allow all origins (adjust for production)
	},
}

// WSManager manages WebSocket connections
type WSManager struct {
	clients        map[string]map[*Client]bool // challenge slug -> set of clients
	challenges     *ChallengeManager
	executor       *Executor
	voting         *VotingManager
	rateLimiter    *RateLimiter
	mu             sync.RWMutex
	autoStopTimers map[string]*time.Timer // challenge slug -> auto-stop timer
	autoStopMu     sync.Mutex
}

// NewWSManager creates a new WebSocket manager
func NewWSManager(challenges *ChallengeManager, executor *Executor, voting *VotingManager, rateLimiter *RateLimiter) *WSManager {
	return &WSManager{
		clients:        make(map[string]map[*Client]bool),
		challenges:     challenges,
		executor:       executor,
		voting:         voting,
		rateLimiter:    rateLimiter,
		autoStopTimers: make(map[string]*time.Timer),
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (wm *WSManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, slug string) {
	// Get client IP
	ip := getClientIP(r)

	// Check rate limit
	if allowed, waitTime := wm.rateLimiter.AllowAction(ip, "websocket"); !allowed {
		http.Error(w, fmt.Sprintf("Rate limit exceeded. Try again in %v", waitTime), http.StatusTooManyRequests)
		return
	}

	// Verify challenge exists
	challenge, exists := wm.challenges.GetChallenge(slug)
	if !exists {
		http.Error(w, "Challenge not found", http.StatusNotFound)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Failed to upgrade connection: %v", err)
		return
	}

	// Create client
	client := &Client{
		Conn:      conn,
		IP:        ip,
		Challenge: slug,
		Send:      make(chan []byte, 256),
	}

	// Register client
	wm.register(client)

	// Add IP to challenge's connected users
	challenge.AddConnectedIP(ip)

	// Cancel auto-stop if any
	wm.cancelAutoStop(slug)

	// Broadcast updated status
	wm.broadcastStatus(slug)

	log.InfoH3("WebSocket connected: %s (IP: %s)", slug, maskIP(ip))

	// Start client goroutines
	go wm.writePump(client)
	go wm.readPump(client)
}

// register registers a client to a challenge
func (wm *WSManager) register(client *Client) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.clients[client.Challenge] == nil {
		wm.clients[client.Challenge] = make(map[*Client]bool)
	}
	wm.clients[client.Challenge][client] = true
}

// unregister unregisters a client
func (wm *WSManager) unregister(client *Client) {
	wm.mu.Lock()

	if clients, ok := wm.clients[client.Challenge]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)

			// Clean up empty map first
			if len(clients) == 0 {
				delete(wm.clients, client.Challenge)
			}
		}
	}

	wm.mu.Unlock()

	// Close send channel after unlocking to avoid deadlock
	select {
	case <-client.Send:
		// Channel already closed
	default:
		close(client.Send)
	}

	// Remove IP from challenge's connected users
	if challenge, exists := wm.challenges.GetChallenge(client.Challenge); exists {
		challenge.RemoveConnectedIP(client.IP)

		// Check if this was the last user - schedule auto-stop
		if challenge.GetConnectedUsers() == 0 && challenge.GetStatus() == StatusRunning {
			go wm.scheduleAutoStop(client.Challenge)
		}
	}
}

// broadcast sends a message to all clients of a challenge
func (wm *WSManager) broadcast(slug string, message []byte) {
	wm.mu.RLock()
	clients, exists := wm.clients[slug]
	wm.mu.RUnlock()

	if !exists || len(clients) == 0 {
		return // No clients to broadcast to
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			// Channel full, skip this client
			log.Debug("Skipping broadcast to client %s (channel full)", maskIP(client.IP))
		}
	}
}

// readPump reads messages from the WebSocket connection
func (wm *WSManager) readPump(client *Client) {
	defer func() {
		wm.unregister(client)
		_ = client.Conn.Close()
		log.InfoH3("WebSocket disconnected: %s (IP: %s)", client.Challenge, maskIP(client.IP))
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	_ = client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("WebSocket error: %v", err)
			}
			break
		}

		// Handle message
		wm.handleMessage(client, message)
	}
}

// writePump writes messages to the WebSocket connection
func (wm *WSManager) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (wm *WSManager) handleMessage(client *Client, message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		wm.sendError(client, "Invalid message format")
		return
	}

	switch msg.Type {
	case "ping":
		wm.handlePing(client)
	case "start":
		wm.handleStart(client)
	case "restart":
		wm.handleRestartRequest(client)
	case "vote":
		wm.handleVote(client, msg)
	default:
		wm.sendError(client, fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handlePing responds to ping messages
func (wm *WSManager) handlePing(client *Client) {
	response := WSMessage{
		Type: "pong",
	}
	wm.sendToClient(client, response)
}

// handleStart handles challenge start requests
func (wm *WSManager) handleStart(client *Client) {
	// Check rate limit
	if allowed, waitTime := wm.rateLimiter.AllowAction(client.IP, "start"); !allowed {
		wm.sendError(client, fmt.Sprintf("Rate limit exceeded. Try again in %v", waitTime))
		return
	}

	challenge, exists := wm.challenges.GetChallenge(client.Challenge)
	if !exists {
		wm.sendError(client, "Challenge not found")
		return
	}

	// Check current status
	if challenge.GetStatus() == StatusRunning {
		wm.sendError(client, "Challenge is already running")
		return
	}

	// Set status to starting
	challenge.SetStatus(StatusStarting)
	wm.broadcastStatus(client.Challenge)

	// Start in background
	go func() {
		if err := wm.executor.Start(challenge); err != nil {
			log.Error("Failed to start challenge %s: %v", challenge.Name, err)
			challenge.SetStatus(StatusStopped)
			wm.broadcastError(client.Challenge, "Failed to start challenge. Please check server logs.")
		} else {
			challenge.SetStatus(StatusRunning)
			wm.broadcastInfo(client.Challenge, "Challenge started successfully")
		}
		wm.broadcastStatus(client.Challenge)
	}()
}

// handleRestartRequest handles restart vote initiation
func (wm *WSManager) handleRestartRequest(client *Client) {
	// Check rate limit
	if allowed, waitTime := wm.rateLimiter.AllowAction(client.IP, "restart"); !allowed {
		wm.sendError(client, fmt.Sprintf("Rate limit exceeded. Try again in %v", waitTime))
		return
	}

	challenge, exists := wm.challenges.GetChallenge(client.Challenge)
	if !exists {
		wm.sendError(client, "Challenge not found")
		return
	}

	// Check cooldown
	if inCooldown, remaining := challenge.IsInCooldown(); inCooldown {
		wm.sendError(client, fmt.Sprintf("Restart on cooldown. Wait %v", remaining.Round(time.Second)))
		return
	}

	// Check if vote already exists
	if wm.voting.HasActiveVote(client.Challenge) {
		wm.sendError(client, "Restart vote already in progress")
		return
	}

	// Start vote
	if err := wm.voting.StartVote(client.Challenge); err != nil {
		log.Error("Failed to start vote for %s: %v", challenge.Name, err)
		wm.sendError(client, "Failed to start vote")
		return
	}

	// Broadcast vote started
	voteMsg := VoteMessage{
		InitiatorIP: maskIP(client.IP),
	}
	wm.broadcastVoteStarted(client.Challenge, voteMsg)

	// Automatically vote yes for the initiator
	_ = wm.voting.CastVote(client.Challenge, client.IP, true)
	wm.checkAndBroadcastVoteUpdate(client.Challenge)
}

// handleVote handles vote submissions
func (wm *WSManager) handleVote(client *Client, msg WSMessage) {
	// Check rate limit
	if allowed, waitTime := wm.rateLimiter.AllowAction(client.IP, "vote"); !allowed {
		wm.sendError(client, fmt.Sprintf("Rate limit exceeded. Try again in %v", waitTime))
		return
	}

	// Parse vote value
	voteValue, ok := msg.Data.(map[string]interface{})["value"].(string)
	if !ok {
		wm.sendError(client, "Invalid vote value")
		return
	}

	voteYes := voteValue == "yes"

	// Cast vote
	if err := wm.voting.CastVote(client.Challenge, client.IP, voteYes); err != nil {
		log.Error("Failed to cast vote for %s from %s: %v", client.Challenge, maskIP(client.IP), err)
		wm.sendError(client, "Failed to cast vote")
		return
	}

	// Check threshold and broadcast update
	wm.checkAndBroadcastVoteUpdate(client.Challenge)
}

// checkAndBroadcastVoteUpdate checks vote threshold and broadcasts updates
func (wm *WSManager) checkAndBroadcastVoteUpdate(slug string) {
	challenge, exists := wm.challenges.GetChallenge(slug)
	if !exists {
		return
	}

	// Get vote status
	yesPercent, noPercent, totalVoters, _ := wm.voting.GetVoteStatus(slug, challenge.ConnectedIPs)

	// Broadcast vote update
	voteMsg := VoteMessage{
		YesPercent: yesPercent,
		NoPercent:  noPercent,
		TotalUsers: totalVoters,
	}
	wm.broadcastVoteUpdate(slug, voteMsg)

	// Check threshold
	approved, rejected, inProgress := wm.voting.CheckThreshold(slug, challenge.ConnectedIPs)

	switch {
	case approved:
		// Execute restart
		wm.voting.EndVote(slug, "approved")
		wm.broadcastVoteEnded(slug, VoteMessage{Result: "approved"})
		wm.executeRestart(challenge)
	case rejected:
		// Cancel vote
		wm.voting.EndVote(slug, "rejected")
		wm.broadcastVoteEnded(slug, VoteMessage{Result: "rejected"})
	case !inProgress:
		// Shouldn't happen, but just in case
		wm.voting.EndVote(slug, "unknown")
	}
}

// executeRestart executes a challenge restart
func (wm *WSManager) executeRestart(challenge *ChallengeInfo) {
	challenge.SetStatus(StatusRestarting)
	wm.broadcastStatus(challenge.Slug)

	go func() {
		if err := wm.executor.Restart(challenge); err != nil {
			log.Error("Failed to restart challenge %s: %v", challenge.Name, err)
			challenge.SetStatus(StatusStopped)
			wm.broadcastError(challenge.Slug, "Failed to restart challenge. Please check server logs.")
		} else {
			challenge.SetStatus(StatusRunning)
			challenge.SetLastRestart(time.Now())
			wm.broadcastInfo(challenge.Slug, "Challenge restarted successfully")
		}
		wm.broadcastStatus(challenge.Slug)
	}()
}

// scheduleAutoStop schedules an auto-stop for a challenge
func (wm *WSManager) scheduleAutoStop(slug string) {
	challenge, exists := wm.challenges.GetChallenge(slug)
	if !exists {
		return
	}

	// Only auto-stop if challenge is running
	if challenge.GetStatus() != StatusRunning {
		return
	}

	gracePeriod := challenge.CalculateGracePeriod()

	log.InfoH3("Scheduling auto-stop for %s in %v", challenge.Name, gracePeriod)

	wm.autoStopMu.Lock()
	// Cancel existing timer if any
	if timer, exists := wm.autoStopTimers[slug]; exists {
		timer.Stop()
		delete(wm.autoStopTimers, slug)
	}

	// Create new timer
	timer := time.AfterFunc(gracePeriod, func() {
		wm.performAutoStop(slug)
	})

	wm.autoStopTimers[slug] = timer
	wm.autoStopMu.Unlock()

	// Broadcast auto-stop scheduled (only if there are no connected users)
	if challenge.GetConnectedUsers() == 0 {
		log.Info("Auto-stop scheduled for %s (no broadcast - no users connected)", slug)
	}
}

// cancelAutoStop cancels a scheduled auto-stop
func (wm *WSManager) cancelAutoStop(slug string) {
	wm.autoStopMu.Lock()

	if timer, exists := wm.autoStopTimers[slug]; exists {
		timer.Stop()
		delete(wm.autoStopTimers, slug)
		wm.autoStopMu.Unlock()

		log.InfoH3("Cancelled auto-stop for %s", slug)
		wm.broadcastInfo(slug, "Auto-stop cancelled (user reconnected)")
	} else {
		wm.autoStopMu.Unlock()
	}
}

// performAutoStop performs the auto-stop action
func (wm *WSManager) performAutoStop(slug string) {
	challenge, exists := wm.challenges.GetChallenge(slug)
	if !exists {
		return
	}

	// Double-check no users are connected
	if challenge.GetConnectedUsers() > 0 {
		log.InfoH3("Auto-stop cancelled for %s (users reconnected)", challenge.Name)
		return
	}

	log.InfoH2("Auto-stopping challenge: %s", challenge.Name)

	challenge.SetStatus(StatusStopping)
	wm.broadcastStatus(slug)

	if err := wm.executor.Stop(challenge); err != nil {
		log.Error("Auto-stop failed for %s: %v", challenge.Name, err)
		challenge.SetStatus(StatusRunning)
		wm.broadcastError(slug, "Auto-stop failed. Challenge still running.")
	} else {
		challenge.SetStatus(StatusStopped)
		wm.broadcastInfo(slug, "Challenge auto-stopped (no users connected)")
	}
	wm.broadcastStatus(slug)
}

// Broadcast helper methods

func (wm *WSManager) broadcastStatus(slug string) {
	challenge, exists := wm.challenges.GetChallenge(slug)
	if !exists {
		return
	}

	statusMsg := StatusMessage{
		Status:         string(challenge.GetStatus()),
		ConnectedUsers: challenge.GetConnectedUsers(),
	}

	msg := WSMessage{
		Type: "status",
		Data: statusMsg,
	}

	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) broadcastError(slug, message string) {
	msg := WSMessage{
		Type:    "error",
		Message: message,
	}
	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) broadcastInfo(slug, message string) {
	msg := WSMessage{
		Type:    "info",
		Message: message,
	}
	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) broadcastVoteStarted(slug string, voteMsg VoteMessage) {
	msg := WSMessage{
		Type: "vote_started",
		Data: voteMsg,
	}
	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) broadcastVoteUpdate(slug string, voteMsg VoteMessage) {
	msg := WSMessage{
		Type: "vote_update",
		Data: voteMsg,
	}
	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) broadcastVoteEnded(slug string, voteMsg VoteMessage) {
	msg := WSMessage{
		Type: "vote_ended",
		Data: voteMsg,
	}
	data, _ := json.Marshal(msg)
	wm.broadcast(slug, data)
}

func (wm *WSManager) sendToClient(client *Client, msg WSMessage) {
	data, _ := json.Marshal(msg)
	select {
	case client.Send <- data:
	default:
		// Client's send buffer is full, skip
	}
}

func (wm *WSManager) sendError(client *Client, message string) {
	msg := WSMessage{
		Type:    "error",
		Message: message,
	}
	wm.sendToClient(client, msg)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Get first IP in the list
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	ip := r.RemoteAddr
	// Strip port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
