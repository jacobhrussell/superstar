package cmd

import (
	"sort"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Short:   "Manage projects",
}

func projectNames(projects map[string]ProjectConfig) []string {
	names := make([]string, 0, len(projects))
	for n := range projects {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
