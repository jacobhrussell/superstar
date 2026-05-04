package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sessionNewDir     string
	sessionNewAgent   string
	sessionNewProject string
	sessionNewName    string
	sessionNewPrompt  string
)

var sessionNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new session",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var projectAgent, projectDir string
		if sessionNewProject != "" {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("could not load config: %w", err)
			}
			proj, ok := cfg.Projects[sessionNewProject]
			if !ok {
				return fmt.Errorf("project %q not found in config", sessionNewProject)
			}
			projectAgent = proj.Agent
			projectDir = proj.Dir
		}

		if sessionNewDir == "" {
			sessionNewDir = projectDir
		}
		if sessionNewDir == "" {
			return errors.New("--dir is required (or use --project)")
		}

		if sessionNewAgent == "" {
			sessionNewAgent = projectAgent
		}
		if sessionNewAgent == "" {
			sessionNewAgent = viper.GetString("default_agent")
		}
		if sessionNewAgent == "" {
			return errors.New("--agent is required (or use --project, or set default_agent in config)")
		}

		return validateAgent(sessionNewAgent)
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("name:", sessionNewName)
		fmt.Println("dir:", sessionNewDir)
		fmt.Println("agent:", sessionNewAgent)
		fmt.Println("prompt:", sessionNewPrompt)
	},
}

func init() {
	sessionNewCmd.Flags().StringVarP(&sessionNewName, "name", "n", "", "name for the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewDir, "dir", "d", "", "directory to open in the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewAgent, "agent", "a", "", fmt.Sprintf("agent to use (%s)", strings.Join(validAgents, ", ")))
	sessionNewCmd.Flags().StringVarP(&sessionNewProject, "project", "p", "", "use defaults from a named project in config")
	sessionNewCmd.Flags().StringVar(&sessionNewPrompt, "prompt", "", "initial prompt to send to the agent")
	sessionCmd.AddCommand(sessionNewCmd)
}
