package cmd

import "github.com/spf13/cobra"

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Short:   "Manage projects",
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
