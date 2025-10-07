package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dimasma0305/gzcli/internal/gzcli"
	"github.com/dimasma0305/gzcli/internal/log"
)

var scoreboardCmd = &cobra.Command{
	Use:   "scoreboard",
	Short: "Generate CTFTime scoreboard feed",
	Long: `Generate a CTFTime-compatible scoreboard feed in JSON format.

The output can be used to submit your CTF scoreboard to CTFTime.org.`,
	Example: `  # Generate scoreboard
  gzcli scoreboard

  # Save to file
  gzcli scoreboard > scoreboard.json`,
	Run: func(_ *cobra.Command, _ []string) {
		gz := gzcli.MustInit()
		feed := gz.MustScoreboard2CTFTimeFeed()

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(feed); err != nil {
			log.Fatal(fmt.Errorf("JSON encoding failed: %w", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(scoreboardCmd)
}
