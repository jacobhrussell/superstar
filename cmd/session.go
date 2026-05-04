package cmd

import "github.com/spf13/cobra"

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}
