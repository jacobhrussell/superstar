package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage superstar configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		agent := cfg.DefaultAgent

		opts := []huh.Option[string]{huh.NewOption("(none)", "")}
		for _, a := range validAgents {
			opts = append(opts, huh.NewOption(a, a))
		}

		if err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Default agent").
					Description("Used when --agent is omitted and no --project is selected").
					Options(opts...).
					Filtering(true).
					Value(&agent),
			),
		).Run(); err != nil {
			return err
		}

		cfg.DefaultAgent = agent
		if err := saveConfig(cfg); err != nil {
			return err
		}
		path, _ := configPath()
		fmt.Fprintf(cmd.OutOrStdout(), "saved %s\n", path)
		return nil
	},
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "superstar"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func init() {
	rootCmd.AddCommand(configCmd)
}
