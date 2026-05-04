package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var githubRepoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

func validateGithubRepo(s string) error {
	if !githubRepoPattern.MatchString(s) {
		return errors.New("github repo must be in owner/repo form")
	}
	return nil
}

var (
	projectNewDir         string
	projectNewAgent       string
	projectNewGithub      string
	projectNewAfterScript string
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

		var name, dir, agent, github, afterScript string
		if isInteractive(args) {
			name, dir, agent, github, afterScript, err = promptProject(cfg.Projects)
			if err != nil {
				return err
			}
		} else {
			name = args[0]
			dir = projectNewDir
			agent = projectNewAgent
			github = projectNewGithub
			afterScript = projectNewAfterScript
		}

		if name == "" {
			return errors.New("project name cannot be empty")
		}
		if _, exists := cfg.Projects[name]; exists {
			return fmt.Errorf("project %q already exists", name)
		}
		if github != "" {
			if err := validateGithubRepo(github); err != nil {
				return err
			}
		}

		cfg.Projects[name] = ProjectConfig{Dir: dir, Agent: agent, Github: github, SessionAfterScript: afterScript}
		if err := saveConfig(cfg); err != nil {
			return err
		}

		path, _ := configPath()
		fmt.Fprintf(cmd.OutOrStdout(), "created project %q in %s\n", name, path)
		return nil
	},
}

func isInteractive(args []string) bool {
	return len(args) == 0 && projectNewDir == "" && projectNewAgent == "" && projectNewGithub == "" && projectNewAfterScript == ""
}

func promptProject(existing map[string]ProjectConfig) (name, dir, agent, github, afterScript string, err error) {
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Value(&name).
				Validate(func(s string) error {
					return validateNewProjectName(s, existing)
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Directory").
				Value(&dir).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("directory cannot be empty")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Agent").
				Options(huh.NewOptions(validAgents...)...).
				Filtering(true).
				Value(&agent),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub repo (optional)").
				Description("owner/repo, e.g. jacobhrussell/superstar").
				Value(&github).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					return validateGithubRepo(s)
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Session-after script").
				Description("Path to a script run after each new session (optional)").
				Value(&afterScript),
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
	projectNewCmd.Flags().StringVarP(&projectNewGithub, "github", "g", "", "GitHub repo for the project (owner/repo)")
	projectNewCmd.Flags().StringVar(&projectNewAfterScript, "session-after-script", "", "path to a script run after a session is created")
	projectCmd.AddCommand(projectNewCmd)
}
