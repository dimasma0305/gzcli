package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli/uploadserver"
	"github.com/dimasma0305/gzcli/internal/log"
)

var (
	uploadServerHost string
	uploadServerPort int
)

var uploadServerCmd = &cobra.Command{
	Use:   "upload-server",
	Short: "Start the challenge upload web server",
	Long: `Start an HTTP server dedicated to uploading challenge packages.

The upload server lets contributors download the challenge template ZIP and
submit completed challenge archives that comply with the gzcli structure.`,
	Example: `  # Start server on default localhost:8090
  gzcli upload-server

  # Start server on custom host and port
  gzcli upload-server --host 0.0.0.0 --port 4000`,
	Run: func(_ *cobra.Command, _ []string) {
		opts := uploadserver.Options{
			Host: uploadServerHost,
			Port: uploadServerPort,
		}

		log.Info("Starting GZCLI Challenge Upload Server...")
		if err := uploadserver.Run(opts); err != nil {
			log.Error("Upload server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadServerCmd)

	uploadServerCmd.Flags().StringVarP(&uploadServerHost, "host", "H", "localhost", "Host to bind the upload server")
	uploadServerCmd.Flags().IntVarP(&uploadServerPort, "port", "p", 8090, "Port to bind the upload server")
}
