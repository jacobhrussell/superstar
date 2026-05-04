package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var sessionEditNewName string

var sessionEditCmd = &cobra.Command{
	Use:     "edit [name]",
	Aliases: []string{"e"},
	Short:   "Rename a session (also renames the tmux session)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Sessions) == 0 {
			return errors.New("no sessions to edit")
		}

		oldName, err := pickSessionName(args, cfg, "Edit which session?")
		if err != nil {
			return err
		}

		newName := sessionEditNewName
		if newName == "" {
			if !isInteractiveTerm() {
				return errors.New("--name is required in non-interactive mode")
			}
			newName = oldName
			if err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("New session name").
						Value(&newName).
						Validate(func(s string) error {
							s = strings.TrimSpace(s)
							if s == "" {
								return errors.New("name cannot be empty")
							}
							if s == oldName {
								return nil
							}
							if _, exists := cfg.Sessions[s]; exists {
								return fmt.Errorf("session %q already exists", s)
							}
							return nil
						}),
				),
			).Run(); err != nil {
				return err
			}
		}

		newName = strings.TrimSpace(newName)
		if newName == oldName {
			fmt.Fprintln(cmd.OutOrStdout(), "no change")
			return nil
		}
		if _, exists := cfg.Sessions[newName]; exists {
			return fmt.Errorf("session %q already exists", newName)
		}

		oldFull := tmuxName(oldName)
		newFull := tmuxName(newName)
		if tmuxHasSession(oldFull) {
			if err := tmuxRenameSession(oldFull, newFull); err != nil {
				return err
			}
		}

		cfg.Sessions[newName] = cfg.Sessions[oldName]
		delete(cfg.Sessions, oldName)
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "renamed %q to %q\n", oldName, newName)
		return nil
	},
}

func init() {
	sessionEditCmd.Flags().StringVarP(&sessionEditNewName, "name", "n", "", "new name for the session")
	sessionCmd.AddCommand(sessionEditCmd)
}
