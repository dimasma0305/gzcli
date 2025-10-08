package socket

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/watchertypes"
)

// Handler provides implementations for socket command handling
// This is a placeholder interface - the actual watcher will implement this
type Handler interface {
	HandleStatusCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleListChallengesCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleGetMetricsCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleGetLogsCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleStopScriptCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleRestartChallengeCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleGetScriptExecutionsCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
	HandleStopEventCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse
}

// DefaultCommandHandler implements CommandHandler by routing to Handler methods
type DefaultCommandHandler struct {
	handler Handler
}

// NewDefaultCommandHandler creates a new default command handler
func NewDefaultCommandHandler(handler Handler) *DefaultCommandHandler {
	return &DefaultCommandHandler{handler: handler}
}

// HandleCommand processes a socket command
func (h *DefaultCommandHandler) HandleCommand(cmd watchertypes.WatcherCommand) watchertypes.WatcherResponse {
	switch cmd.Action {
	case "status":
		return h.handler.HandleStatusCommand(cmd)
	case "list_challenges":
		return h.handler.HandleListChallengesCommand(cmd)
	case "get_metrics":
		return h.handler.HandleGetMetricsCommand(cmd)
	case "get_logs":
		return h.handler.HandleGetLogsCommand(cmd)
	case "stop_script":
		return h.handler.HandleStopScriptCommand(cmd)
	case "restart_challenge":
		return h.handler.HandleRestartChallengeCommand(cmd)
	case "get_script_executions":
		return h.handler.HandleGetScriptExecutionsCommand(cmd)
	case "stop_event":
		return h.handler.HandleStopEventCommand(cmd)
	default:
		return watchertypes.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown command: %s", cmd.Action),
		}
	}
}
