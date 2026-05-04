package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
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
		if sessionNewAgent == "" {
			sessionNewAgent = projectAgent
		}
		if sessionNewAgent == "" {
			sessionNewAgent = viper.GetString("default_agent")
		}

		if sessionNewAgent != "" {
			if err := validateAgent(sessionNewAgent); err != nil {
				return err
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if anySessionFieldMissing() {
			if isInteractiveTerm() {
				if err := promptMissingSessionFields(); err != nil {
					return err
				}
			} else {
				if sessionNewDir == "" {
					return errors.New("--dir is required (or use --project)")
				}
				if sessionNewAgent == "" {
					return errors.New("--agent is required (or use --project, or set default_agent in config)")
				}
			}
		}

		if err := validateAgent(sessionNewAgent); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "name:", sessionNewName)
		fmt.Fprintln(cmd.OutOrStdout(), "dir:", sessionNewDir)
		fmt.Fprintln(cmd.OutOrStdout(), "agent:", sessionNewAgent)
		fmt.Fprintln(cmd.OutOrStdout(), "prompt:", sessionNewPrompt)
		return nil
	},
}

func anySessionFieldMissing() bool {
	return sessionNewAgent == "" || sessionNewDir == "" || sessionNewPrompt == "" || sessionNewName == ""
}

func isInteractiveTerm() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func promptMissingSessionFields() error {
	var groups []*huh.Group

	if sessionNewAgent == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[string]().
				Title("Agent").
				Options(huh.NewOptions(validAgents...)...).
				Filtering(true).
				Value(&sessionNewAgent),
		))
	}
	if sessionNewDir == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Directory").
				Value(&sessionNewDir).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("directory cannot be empty")
					}
					return nil
				}),
		))
	}
	if sessionNewPrompt == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Prompt").
				Description("Initial prompt for the agent (optional, leave blank to skip)").
				Value(&sessionNewPrompt),
		))
	}
	if sessionNewName == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Description("Session name (optional, leave blank to skip)").
				Value(&sessionNewName),
		))
	}

	if len(groups) == 0 {
		return nil
	}
	return huh.NewForm(groups...).Run()
}

func init() {
	sessionNewCmd.Flags().StringVarP(&sessionNewName, "name", "n", "", "name for the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewDir, "dir", "d", "", "directory to open in the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewAgent, "agent", "a", "", fmt.Sprintf("agent to use (%s)", strings.Join(validAgents, ", ")))
	sessionNewCmd.Flags().StringVarP(&sessionNewProject, "project", "p", "", "use defaults from a named project in config")
	sessionNewCmd.Flags().StringVar(&sessionNewPrompt, "prompt", "", "initial prompt to send to the agent")
	sessionCmd.AddCommand(sessionNewCmd)
}
