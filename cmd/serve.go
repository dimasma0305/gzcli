package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/server"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	serveHost string
	servePort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the challenge launcher web server",
	Long: `Start an HTTP/WebSocket server for managing challenge launchers.

The server provides a web interface to start, stop, and restart challenges
with dashboard configuration. Features include:

  • WebSocket-based real-time communication
  • IP-based user tracking
  • 50% threshold voting system for restarts
  • Automatic challenge stop when no users are connected
  • Restart cooldown protection
  • Rate limiting per IP
  • Health monitoring
  • Browser notifications

The server discovers all challenges with dashboard configuration across
all events and makes them accessible via secret URLs based on their slugs.`,
	Example: `  # Start server on default localhost:8080
  gzcli serve

  # Start server on custom host and port
  gzcli serve --host 0.0.0.0 --port 3000

  # Start server with short flags
  gzcli serve -H 0.0.0.0 -p 3000`,
	Run: func(_ *cobra.Command, _ []string) {
		log.Info("Starting GZCLI Challenge Launcher Server...")

		if err := server.RunServer(serveHost, servePort); err != nil {
			log.Error("Server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Flags
	serveCmd.Flags().StringVarP(&serveHost, "host", "H", "localhost", "Host to bind the server to")
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to bind the server to")
}
