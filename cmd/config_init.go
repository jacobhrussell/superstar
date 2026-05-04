package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a config file with all keys commented out",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := configDir()
		if err != nil {
			return fmt.Errorf("could not resolve home directory: %w", err)
		}
		path, err := configPath()
		if err != nil {
			return fmt.Errorf("could not resolve home directory: %w", err)
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			if os.IsExist(err) {
				return fmt.Errorf("config already exists at %s", path)
			}
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(configTemplate()); err != nil {
			return err
		}

		fmt.Println("created", path)
		return nil
	},
}

func configTemplate() string {
	return strings.Join([]string{
		"# Superstar configuration",
		"",
		"# Fallback agent for `session new --agent` when the flag is omitted.",
		"# Valid values: " + strings.Join(validAgents, ", "),
		"# default_agent:",
		"",
	}, "\n")
}

func init() {
	configCmd.AddCommand(configInitCmd)
}
