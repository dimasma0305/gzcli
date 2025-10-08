<!-- 7ea5ef00-4893-4c80-b504-5499f8f1093c 3b46ba5b-32c0-4532-9c55-3f3e1cae4367 -->
# Challenge Launcher Web Dashboard

## Architecture Overview

Create `gzcli serve` command that starts an HTTP/WebSocket server for managing challenge launchers. The system will:

- Use WebSocket connections with ping/pong for authentication
- Track unique users by IP address
- Implement a 50% threshold voting system for restarts
- Support all challenges with `dashboard` configuration across all events

## Implementation Steps

### 1. Server Package Structure

Create `internal/gzcli/server/` with:

- **`server.go`**: Main HTTP server setup
- **`websocket.go`**: WebSocket connection manager with IP-based tracking
- **`handlers.go`**: HTTP request handlers for homepage and challenge pages
- **`challenge.go`**: Challenge discovery and management
- **`executor.go`**: Docker Compose script execution (start/stop/restart)
- **`voting.go`**: Voting system for restart requests (50% threshold)
- **`types.go`**: Shared types and structures

### 2. Key Components

#### WebSocket Manager

- Track connections by IP address (unique user identification)
- Ping/pong for real person verification
- Broadcast challenge status updates to all connected clients
- Manage per-challenge connection rooms

#### Voting System

- Store active votes per challenge with voter IP tracking
- Calculate percentage (yes votes / total unique IPs)
- Trigger restart when ≥50% vote yes OR ≥50% vote no (cancel)
- Prevent duplicate votes from same IP

#### Challenge Manager

- Discover all challenges across all events with `dashboard` config
- Generate slugs using existing logic: `{category}_{name}`
- Load challenge configurations and Docker Compose paths
- Track challenge status: `stopped`, `starting`, `running`, `stopping`, `restarting`

#### Launcher Executor (Multi-Platform)

Support three deployment types based on `dashboard.launcher` configuration:

**1. Docker Compose** (most common):

```yaml
dashboard:
  launcher:
    type: "compose"
    config: "./docker-compose.yml"
```

- Execute: `docker compose -f {config} -p {slug} up -d`
- Stop: `docker compose -f {config} -p {slug} down --volumes`
- Health: `docker compose -f {config} -p {slug} ps --format json`

**2. Dockerfile** (single container):

```yaml
dashboard:
  launcher:
    type: "dockerfile"
    config: "./Dockerfile"
    ports: ["8080:80"]  # port mappings
```

- Build: `docker build -t {slug}:latest -f {config} .`
- Start: `docker run -d --name {slug} -p {ports} {slug}:latest`
- Stop: `docker stop {slug} && docker rm {slug}`
- Health: `docker ps --filter name={slug} --format json`
- Container name is always the slug (no custom naming)

**3. Kubernetes** (advanced):

```yaml
dashboard:
  launcher:
    type: "kubernetes"
    config: "./k8s-manifest.yaml"
```

- Start: `kubectl apply -f {config}`
- Stop: `kubectl delete -f {config}`
- Health: `kubectl get pods -l app={slug} -o json`
- Namespace should be defined within the k8s manifest file itself

### 3. Slug Generation Changes

**Update slug format globally** to: `<event>_<category>_<challenge_name>`

- Modify `internal/gzcli/config/challenges.go::generateSlug()` to:
  - Accept event name parameter
  - Return format: `{event}_{category}_{name}` (normalized and cleaned)
  - Apply same normalization as before (lowercase, replace spaces with underscores, remove special chars)
- Update all callers of `generateSlug()` to pass event name
- Update `processChallengeTemplate()` to use new slug format
- Update any existing tests that depend on slug format

### 4. HTTP Routes

```
GET  /                          - Homepage (simple welcome message)
GET  /{slug}                    - Challenge launcher page (secret URL)
WS   /{slug}/ws                 - WebSocket connection for ALL interactions
```

Routes use only the slug (no event prefix) to keep URLs secret. All challenge operations (start/stop/restart/vote) are handled through WebSocket messages.

### 5. Frontend (Go Templates)

Create `internal/gzcli/server/templates/`:

- **`home.html`**: Simple welcome page
- **`challenge.html`**: Challenge launcher interface with:
  - WebSocket connection status
  - Challenge info (name, status, connected users)
  - Start/Stop/Restart buttons
  - Voting UI (shows when vote is active)
  - Real-time updates via WebSocket messages

Use basic HTML/CSS/JavaScript (no external dependencies).

### 5. Command Implementation

Create `cmd/serve.go`:

```go
gzcli serve [--port PORT] [--host HOST]
```

Default: `localhost:8080`

### 6. Integration Points

- Use `internal/gzcli/config` for discovering events and challenges
- Use `internal/gzcli/challenge/script.go` for executing scripts
- Reuse slug generation from `internal/gzcli/config/challenges.go`
- Filter challenges to only those with `Dashboard` configuration

### 7. Critical Implementation Details

**Challenge Discovery**:

```go
// For each event in events/
//   For each category in event
//     For each challenge.yml
//       Parse and check if Dashboard != nil
//       Generate slug: generateSlug(challenge)
//       Store mapping: slug -> {event, category, name, cwd, scripts, dashboard}
```

**WebSocket Protocol**:

```json
// Client -> Server
{"type": "ping"}
{"type": "start"}                        // Start the challenge
{"type": "stop"}                         // Stop the challenge (manual)
{"type": "restart"}                      // Initiate restart vote
{"type": "vote", "value": "yes|no"}     // Vote on active restart proposal

// Server -> Client
{"type": "pong"}
{"type": "status", "status": "running|stopped|starting|stopping|restarting", "connected_users": 5}
{"type": "vote_started", "initiator_ip": "hidden"}
{"type": "vote_update", "yes_percent": 40, "no_percent": 10, "total_users": 10}
{"type": "vote_ended", "result": "approved|rejected"}
{"type": "error", "message": "..."}
{"type": "info", "message": "..."}      // General information messages
```

**Auto-Stop Feature**:

- Track active connections per challenge
- When last user disconnects, start a grace period timer (e.g., 2 minutes)
- If no one reconnects within grace period, automatically stop the challenge
- If someone reconnects during grace period, cancel auto-stop
- Broadcast `auto_stop_scheduled` and `auto_stop_cancelled` events

**Voting Logic**:

- Store votes as `map[challenge_slug]map[ip]vote_value`
- On each vote, recalculate percentages
- If yes% >= 50% → execute restart, clear votes
- If no% >= 50% → cancel vote, clear votes
- Timeout after 5 minutes if neither threshold reached

### 8. Additional Features

**Challenge Health Check**:

- Periodic health check every 30 seconds using `docker compose ps --format json`
- Compare expected state vs actual Docker state
- Auto-update status to "stopped" or "unhealthy" if containers crashed
- Broadcast health status changes to connected clients
- Create `internal/gzcli/server/health.go` for health monitoring

**Browser Notifications**:

- Use Web Notifications API in frontend JavaScript
- Request permission on page load
- Send notifications for:
  - "Challenge is ready!" when startup completes
  - "Restart vote started" when someone initiates restart
  - "Auto-stop in X minutes" as warning before shutdown
- Include sound for important notifications

**Automatic WebSocket Reconnection**:

- Implement exponential backoff: 1s, 2s, 4s, 8s, max 30s
- Show connection status indicator in UI (green/yellow/red dot)
- Auto-retry on disconnect, connection errors
- Reset backoff timer on successful reconnection

**Challenge Port Display**:

- Parse Docker Compose YAML to extract exposed ports
- Display port mapping in UI (e.g., "Web: http://localhost:8080")
- Support both `ports:` and `expose:` directives
- Handle port ranges and multiple services

**Rate Limiting**:

- Per-IP rate limits:
  - Start/stop/restart: max 5 actions per minute
  - Vote submissions: max 10 per minute
  - WebSocket connections: max 3 per minute
- Use token bucket algorithm
- Return error via WebSocket: "Rate limit exceeded, try again in X seconds"
- Create `internal/gzcli/server/ratelimit.go` for rate limiting logic

### 9. Files to Create

1. `cmd/serve.go` - New command
2. `internal/gzcli/server/server.go` - Main server
3. `internal/gzcli/server/websocket.go` - WebSocket manager
4. `internal/gzcli/server/handlers.go` - HTTP handlers
5. `internal/gzcli/server/challenge.go` - Challenge discovery
6. `internal/gzcli/server/executor.go` - Script execution
7. `internal/gzcli/server/voting.go` - Voting system
8. `internal/gzcli/server/types.go` - Types
9. `internal/gzcli/server/templates/home.html` - Home template
10. `internal/gzcli/server/templates/challenge.html` - Challenge template

### 9. Testing Strategy

- Unit tests for voting logic
- Integration tests for challenge discovery
- Mock WebSocket connections for testing
- Test script execution with dummy challenges

## Key Technical Decisions

- **Go templates** for simple, embedded HTML (no build step)
- **Gorilla WebSocket** library for WebSocket support (already used in watcher)
- **IP-based tracking** using `r.RemoteAddr` with proper proxy header support
- **In-memory state** (no persistence - restart clears votes)
- **Mutex protection** for concurrent challenge operations

### To-dos

- [ ] Create server package structure with types.go defining core data structures
- [ ] Implement challenge discovery to find all challenges with dashboard config across all events
- [ ] Implement script executor for start/stop/restart using existing challenge script infrastructure
- [ ] Implement WebSocket manager with IP-based client tracking and ping/pong authentication
- [ ] Implement voting system with 50% threshold logic and IP-based vote tracking
- [ ] Create HTTP handlers for homepage, challenge pages, and API endpoints
- [ ] Create Go templates for home page and challenge launcher interface
- [ ] Create gzcli serve command with port and host flags
- [ ] Add unit tests for voting logic, challenge discovery, and executor
- [ ] Update README and documentation with serve command usage