package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage superstar configuration",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
