package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	projectEditDir         string
	projectEditAgent       string
	projectEditGithub      string
	projectEditAfterScript string
)

var projectEditCmd = &cobra.Command{
	Use:     "edit [name]",
	Aliases: []string{"e"},
	Short:   "Edit an existing project",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Projects) == 0 {
			return errors.New("no projects to edit; create one with `superstar project new`")
		}

		name, err := resolveProjectName(args, cfg, "Edit which project?")
		if err != nil {
			return err
		}

		current := cfg.Projects[name]
		newDir, newAgent, newGithub, newAfterScript := current.Dir, current.Agent, current.Github, current.SessionAfterScript

		dirChanged := cmd.Flags().Changed("dir")
		agentChanged := cmd.Flags().Changed("agent")
		githubChanged := cmd.Flags().Changed("github")
		afterScriptChanged := cmd.Flags().Changed("session-after-script")
		anyFlag := dirChanged || agentChanged || githubChanged || afterScriptChanged

		if dirChanged {
			newDir = projectEditDir
		} else if !anyFlag {
			if err := huh.NewInput().
				Title("Directory").
				Value(&newDir).
				Run(); err != nil {
				return err
			}
		}

		if agentChanged {
			newAgent = projectEditAgent
		} else if !anyFlag {
			if err := huh.NewSelect[string]().
				Title("Agent").
				Options(huh.NewOptions(validAgents...)...).
				Filtering(true).
				Value(&newAgent).
				Run(); err != nil {
				return err
			}
		}

		if githubChanged {
			newGithub = projectEditGithub
		} else if !anyFlag {
			if err := huh.NewInput().
				Title("GitHub repo (optional)").
				Description("owner/repo, e.g. jacobhrussell/superstar").
				Value(&newGithub).
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					return validateGithubRepo(s)
				}).
				Run(); err != nil {
				return err
			}
		}

		if afterScriptChanged {
			newAfterScript = projectEditAfterScript
		} else if !anyFlag {
			if err := huh.NewInput().
				Title("Session-after script").
				Description("Path to a script run after each new session (optional)").
				Value(&newAfterScript).
				Run(); err != nil {
				return err
			}
		}

		if err := validateAgent(newAgent); err != nil {
			return err
		}
		if newDir == "" {
			return errors.New("directory cannot be empty")
		}
		if newGithub != "" {
			if err := validateGithubRepo(newGithub); err != nil {
				return err
			}
		}

		cfg.Projects[name] = ProjectConfig{Dir: newDir, Agent: newAgent, Github: newGithub, SessionAfterScript: newAfterScript}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "updated project %q\n", name)
		return nil
	},
}

// resolveProjectName picks a name from args[0] (validating it exists) or
// prompts the user with a fuzzy-find select if no arg was given.
func resolveProjectName(args []string, cfg *Config, title string) (string, error) {
	if len(args) == 1 {
		name := args[0]
		if _, ok := cfg.Projects[name]; !ok {
			return "", fmt.Errorf("project %q not found", name)
		}
		return name, nil
	}
	var name string
	err := huh.NewSelect[string]().
		Title(title).
		Options(huh.NewOptions(projectNames(cfg.Projects)...)...).
		Filtering(true).
		Value(&name).
		Run()
	return name, err
}

func init() {
	projectEditCmd.Flags().StringVarP(&projectEditDir, "dir", "d", "", "new directory for the project")
	projectEditCmd.Flags().StringVarP(&projectEditAgent, "agent", "a", "", fmt.Sprintf("new agent for the project (%s)", strings.Join(validAgents, ", ")))
	projectEditCmd.Flags().StringVarP(&projectEditGithub, "github", "g", "", "new GitHub repo for the project (owner/repo, empty to clear)")
	projectEditCmd.Flags().StringVar(&projectEditAfterScript, "session-after-script", "", "path to a script run after a session is created (empty string to clear)")
	projectCmd.AddCommand(projectEditCmd)
}
