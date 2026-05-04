package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	projectNewDir   string
	projectNewAgent string
)

var projectNewCmd = &cobra.Command{
	Use:     "new <name>",
	Aliases: []string{"n"},
	Short:   "Create a new project entry in the config",
	Args:    cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if projectNewDir == "" {
			return errors.New("--dir is required")
		}
		if projectNewAgent == "" {
			return errors.New("--agent is required")
		}
		return validateAgent(projectNewAgent)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" {
			return errors.New("project name cannot be empty")
		}

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

		key := "projects." + name
		if viper.IsSet(key) {
			return fmt.Errorf("project %q already exists", name)
		}

		viper.Set(key+".dir", projectNewDir)
		viper.Set(key+".agent", projectNewAgent)

		viper.SetConfigFile(path)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}

		fmt.Printf("created project %q in %s\n", name, path)
		return nil
	},
}

func init() {
	projectNewCmd.Flags().StringVarP(&projectNewDir, "dir", "d", "", "directory for the project")
	projectNewCmd.Flags().StringVarP(&projectNewAgent, "agent", "a", "", fmt.Sprintf("agent for the project (%s)", strings.Join(validAgents, ", ")))
	projectCmd.AddCommand(projectNewCmd)
}
