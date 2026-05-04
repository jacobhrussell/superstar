package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sessionNewDir   string
	sessionNewAgent string
)

var validAgents = []string{"claude", "codex", "cursor", "opencode"}

var sessionNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new session",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if sessionNewDir == "" {
			return errors.New("--dir is required")
		}
		if sessionNewAgent == "" {
			sessionNewAgent = viper.GetString("default_agent")
		}
		if sessionNewAgent == "" {
			return errors.New("--agent is required (or set default_agent in config)")
		}
		for _, a := range validAgents {
			if sessionNewAgent == a {
				return nil
			}
		}
		return fmt.Errorf("--agent must be one of: %s", strings.Join(validAgents, ", "))
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello, superstar")
		fmt.Println("dir:", sessionNewDir)
		fmt.Println("agent:", sessionNewAgent)
	},
}

func init() {
	sessionNewCmd.Flags().StringVarP(&sessionNewDir, "dir", "d", "", "directory to open in the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewAgent, "agent", "a", "", fmt.Sprintf("agent to use (%s)", strings.Join(validAgents, ", ")))
	sessionCmd.AddCommand(sessionNewCmd)
}
