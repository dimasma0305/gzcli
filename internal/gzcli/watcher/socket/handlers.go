package socket

import (
	"fmt"

	"github.com/dimasma0305/gzcli/internal/gzcli/watcher/types"
)

// Handler provides implementations for socket command handling
// This is a placeholder interface - the actual watcher will implement this
type Handler interface {
	HandleStatusCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleListChallengesCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleGetMetricsCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleGetLogsCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleStopScriptCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleRestartChallengeCommand(cmd types.WatcherCommand) types.WatcherResponse
	HandleGetScriptExecutionsCommand(cmd types.WatcherCommand) types.WatcherResponse
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
func (h *DefaultCommandHandler) HandleCommand(cmd types.WatcherCommand) types.WatcherResponse {
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
	default:
		return types.WatcherResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown command: %s", cmd.Action),
		}
	}
}
