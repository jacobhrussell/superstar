package cmd

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var sessionDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a session (kills tmux + removes worktree)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Sessions) == 0 {
			return errors.New("no sessions to delete")
		}

		name, err := pickSessionName(args, cfg, "Delete which session?")
		if err != nil {
			return err
		}
		s := cfg.Sessions[name]

		fullName := tmuxName(name)
		if tmuxHasSession(fullName) {
			if err := tmuxKillSession(fullName); err != nil {
				return err
			}
		}
		if s.Worktree != "" {
			if err := gitRemoveWorktree(s.Dir, s.Worktree); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not remove worktree: %v\n", err)
			}
		}
		if s.Branch != "" {
			if err := gitDeleteBranch(s.Dir, s.Branch); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not delete branch: %v\n", err)
			}
		}

		delete(cfg.Sessions, name)
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "deleted session %q\n", name)
		return nil
	},
}

func pickSessionName(args []string, cfg *Config, title string) (string, error) {
	if len(args) == 1 {
		name := args[0]
		if _, ok := cfg.Sessions[name]; !ok {
			return "", fmt.Errorf("session %q not found", name)
		}
		return name, nil
	}
	var name string
	err := huh.NewSelect[string]().
		Title(title).
		Options(huh.NewOptions(sessionNames(cfg.Sessions)...)...).
		Filtering(true).
		Value(&name).
		Run()
	return name, err
}

func sessionNames(sessions map[string]SessionConfig) []string {
	out := make([]string, 0, len(sessions))
	for k := range sessions {
		out = append(out, k)
	}
	return out
}

func init() {
	sessionCmd.AddCommand(sessionDeleteCmd)
}
