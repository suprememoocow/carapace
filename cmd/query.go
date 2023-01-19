package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/suprememoocow/carapace/internal/list"
)

var timeout time.Duration

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query all shelly devices on your network using jq",
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var query string
		if len(args) == 0 {
			query = "."
		} else {
			query = args[0]
		}

		err := list.QueryShellies(query, timeout)
		if err != nil {
			return fmt.Errorf("failed to query shellies: %w", err)
		}

		return nil
	},
}


func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().DurationVar(&timeout, "timeout", time.Second * 10, "Timeout for mDNS responses")
}
