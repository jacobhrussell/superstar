package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	projectNewDir   string
	projectNewAgent string
)

var projectNewCmd = &cobra.Command{
	Use:     "new [name]",
	Aliases: []string{"n"},
	Short:   "Create a new project entry in the config",
	Args:    cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if isInteractive(args) {
			return nil
		}
		if projectNewDir == "" {
			return errors.New("--dir is required")
		}
		if projectNewAgent == "" {
			return errors.New("--agent is required")
		}
		return validateAgent(projectNewAgent)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		var name, dir, agent string
		if isInteractive(args) {
			name, dir, agent, err = promptProject(cfg.Projects)
			if err != nil {
				return err
			}
		} else {
			name = args[0]
			dir = projectNewDir
			agent = projectNewAgent
		}

		if name == "" {
			return errors.New("project name cannot be empty")
		}
		if _, exists := cfg.Projects[name]; exists {
			return fmt.Errorf("project %q already exists", name)
		}

		cfg.Projects[name] = ProjectConfig{Dir: dir, Agent: agent}
		if err := saveConfig(cfg); err != nil {
			return err
		}

		path, _ := configPath()
		fmt.Fprintf(cmd.OutOrStdout(), "created project %q in %s\n", name, path)
		return nil
	},
}

func isInteractive(args []string) bool {
	return len(args) == 0 && projectNewDir == "" && projectNewAgent == ""
}

func promptProject(existing map[string]ProjectConfig) (name, dir, agent string, err error) {
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Value(&name).
				Validate(func(s string) error {
					return validateNewProjectName(s, existing)
				}),
			huh.NewInput().
				Title("Directory").
				Value(&dir).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("directory cannot be empty")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Agent").
				Options(huh.NewOptions(validAgents...)...).
				Filtering(true).
				Value(&agent),
		),
	).Run()
	return
}

func validateNewProjectName(name string, existing map[string]ProjectConfig) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name cannot be empty")
	}
	if _, exists := existing[name]; exists {
		return fmt.Errorf("project %q already exists", name)
	}
	return nil
}

func init() {
	projectNewCmd.Flags().StringVarP(&projectNewDir, "dir", "d", "", "directory for the project")
	projectNewCmd.Flags().StringVarP(&projectNewAgent, "agent", "a", "", fmt.Sprintf("agent for the project (%s)", strings.Join(validAgents, ", ")))
	projectCmd.AddCommand(projectNewCmd)
}
