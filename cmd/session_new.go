package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	sessionNewDir         string
	sessionNewAgent       string
	sessionNewProject     string
	sessionNewName        string
	sessionNewPrompt      string
	sessionNewAfterScript string
	sessionNewPullRequest string
)

var sessionNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "Create a new session (tmux + optional git worktree + agent)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if sessionNewPullRequest != "" && sessionNewProject != "" {
			return errors.New("--pull-request cannot be combined with --project")
		}

		var projectAgent, projectDir, projectAfterScript string
		if sessionNewPullRequest != "" {
			pr, err := parsePRURL(sessionNewPullRequest)
			if err != nil {
				return err
			}
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("could not load config: %w", err)
			}
			name, proj, found := findProjectByGithub(cfg.Projects, pr.ownerRepo())
			if !found {
				return fmt.Errorf(
					"--pull-request is only supported for projects with a configured GitHub repo. "+
						"No project matches %s.\n"+
						"Run `superstar project new` (or `superstar project edit <name>`) and set --github to %s.",
					pr.ownerRepo(), pr.ownerRepo(),
				)
			}
			sessionNewProject = name
			projectAgent = proj.Agent
			projectDir = proj.Dir
			projectAfterScript = proj.SessionAfterScript
		} else if sessionNewProject != "" {
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
			projectAfterScript = proj.SessionAfterScript
		}

		if sessionNewDir == "" {
			sessionNewDir = projectDir
		}
		if sessionNewAgent == "" {
			sessionNewAgent = projectAgent
		}
		sessionNewAfterScript = projectAfterScript
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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

		var pr prRef
		if sessionNewPullRequest != "" {
			pr, _ = parsePRURL(sessionNewPullRequest)
		}

		if sessionNewName == "" {
			if sessionNewPullRequest != "" {
				sessionNewName = "pr-" + pr.Number
			} else {
				sessionNewName = fmt.Sprintf("s-%d", time.Now().Unix())
			}
		}

		absDir, err := filepath.Abs(sessionNewDir)
		if err != nil {
			return fmt.Errorf("resolve dir: %w", err)
		}
		sessionNewDir = absDir

		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, exists := cfg.Sessions[sessionNewName]; exists {
			return fmt.Errorf("session %q already exists", sessionNewName)
		}
		if err := tmuxAvailable(); err != nil {
			return err
		}
		fullName := tmuxName(sessionNewName)
		if tmuxHasSession(fullName) {
			return fmt.Errorf("tmux session %q already exists", fullName)
		}

		// Optionally create a git worktree.
		worktreePath := ""
		branchName := ""
		tmuxDir := sessionNewDir
		if sessionNewPullRequest != "" {
			if !isGitRepo(sessionNewDir) {
				return fmt.Errorf("project dir %s is not a git repo; cannot use --pull-request", sessionNewDir)
			}
			prBranch, err := ghPRHeadBranch(pr)
			if err != nil {
				return err
			}
			if err := gitFetchPRBranch(sessionNewDir, pr.Number, prBranch); err != nil {
				return err
			}
			worktreePath = sessionNewDir + "-" + sessionNewName
			branchName = prBranch
			if err := gitAddWorktree(sessionNewDir, worktreePath, prBranch); err != nil {
				return err
			}
			tmuxDir = worktreePath
			defer func() {
				if err != nil {
					_ = gitRemoveWorktree(sessionNewDir, worktreePath)
				}
			}()
		} else if isGitRepo(sessionNewDir) {
			worktreePath = sessionNewDir + "-" + sessionNewName
			branchName = sessionNewName
			if err := gitCreateWorktree(sessionNewDir, worktreePath, branchName); err != nil {
				return err
			}
			tmuxDir = worktreePath
			defer func() {
				if err != nil {
					_ = gitRemoveWorktree(sessionNewDir, worktreePath)
					_ = gitDeleteBranch(sessionNewDir, branchName)
				}
			}()
		}

		// Create the tmux session.
		if err := tmuxNewSession(fullName, tmuxDir); err != nil {
			return err
		}
		defer func() {
			if err != nil {
				_ = tmuxKillSession(fullName)
			}
		}()

		// Start the agent and optionally send the prompt.
		if err := tmuxSendLine(fullName, sessionNewAgent); err != nil {
			return err
		}
		if sessionNewPrompt != "" {
			// Wait for the agent's TUI to render before typing the prompt,
			// then pause briefly so the input is registered before submitting.
			time.Sleep(2 * time.Second)
			if err := tmuxSendText(fullName, sessionNewPrompt); err != nil {
				return err
			}
			time.Sleep(300 * time.Millisecond)
			if err := tmuxPressEnter(fullName); err != nil {
				return err
			}
		}

		// Persist.
		cfg.Sessions[sessionNewName] = SessionConfig{
			Project:  sessionNewProject,
			Dir:      sessionNewDir,
			Agent:    sessionNewAgent,
			Worktree: worktreePath,
			Branch:   branchName,
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}

		if sessionNewAfterScript != "" {
			if scriptErr := runSessionAfterScript(sessionNewAfterScript, sessionAfterScriptEnv{
				Name:     sessionNewName,
				TmuxName: fullName,
				Dir:      sessionNewDir,
				TmuxDir:  tmuxDir,
				Agent:    sessionNewAgent,
				Project:  sessionNewProject,
				Worktree: worktreePath,
				Branch:   branchName,
			}); scriptErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: session-after-script failed: %v\n", scriptErr)
			}
		}

		attachCmd := fmt.Sprintf("tmux attach -t %s", fullName)
		fmt.Fprintf(cmd.OutOrStdout(), "created session %q (tmux: %s)\n", sessionNewName, fullName)
		if worktreePath != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "worktree: %s (branch: %s)\n", worktreePath, branchName)
		}
		if clipErr := clipboard.WriteAll(attachCmd); clipErr == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "copied to clipboard: %s\n", attachCmd)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "attach with: %s\n", attachCmd)
		}
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
				Description("Session name (optional, auto-generated if blank)").
				Value(&sessionNewName),
		))
	}

	if len(groups) == 0 {
		return nil
	}
	return huh.NewForm(groups...).Run()
}

type sessionAfterScriptEnv struct {
	Name     string
	TmuxName string
	Dir      string
	TmuxDir  string
	Agent    string
	Project  string
	Worktree string
	Branch   string
}

func runSessionAfterScript(scriptPath string, env sessionAfterScriptEnv) error {
	resolved, err := expandHome(scriptPath)
	if err != nil {
		return err
	}
	c := exec.Command(resolved)
	c.Env = append(os.Environ(),
		"SUPERSTAR_SESSION_NAME="+env.Name,
		"SUPERSTAR_TMUX_SESSION="+env.TmuxName,
		"SUPERSTAR_SESSION_DIR="+env.Dir,
		"SUPERSTAR_SESSION_TMUX_DIR="+env.TmuxDir,
		"SUPERSTAR_SESSION_AGENT="+env.Agent,
		"SUPERSTAR_SESSION_PROJECT="+env.Project,
		"SUPERSTAR_SESSION_WORKTREE="+env.Worktree,
		"SUPERSTAR_SESSION_BRANCH="+env.Branch,
	)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func expandHome(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

func init() {
	sessionNewCmd.Flags().StringVarP(&sessionNewName, "name", "n", "", "name for the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewDir, "dir", "d", "", "directory to open in the session")
	sessionNewCmd.Flags().StringVarP(&sessionNewAgent, "agent", "a", "", fmt.Sprintf("agent to use (%s)", strings.Join(validAgents, ", ")))
	sessionNewCmd.Flags().StringVarP(&sessionNewProject, "project", "p", "", "use defaults from a named project in config")
	sessionNewCmd.Flags().StringVar(&sessionNewPrompt, "prompt", "", "initial prompt to send to the agent")
	sessionNewCmd.Flags().StringVar(&sessionNewPullRequest, "pull-request", "", "GitHub PR URL — match a configured project and check out the PR's branch in a new worktree (alias: --pr)")
	sessionNewCmd.Flags().SetNormalizeFunc(func(_ *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "pr" {
			name = "pull-request"
		}
		return pflag.NormalizedName(name)
	})
	sessionCmd.AddCommand(sessionNewCmd)
}
