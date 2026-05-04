package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var projectDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a project",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Projects) == 0 {
			return errors.New("no projects to delete")
		}

		name, err := resolveProjectName(args, cfg, "Delete which project?")
		if err != nil {
			return err
		}

		delete(cfg.Projects, name)
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "deleted project %q\n", name)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectDeleteCmd)
}
