package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const tmuxSuffix = "-superstar"

func tmuxName(name string) string {
	return name + tmuxSuffix
}

func sessionNameFromTmux(tmuxName string) (string, bool) {
	if !strings.HasSuffix(tmuxName, tmuxSuffix) {
		return "", false
	}
	return strings.TrimSuffix(tmuxName, tmuxSuffix), true
}

func tmuxAvailable() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return errors.New("tmux is not installed (try: brew install tmux)")
	}
	return nil
}

func tmuxHasSession(fullName string) bool {
	return exec.Command("tmux", "has-session", "-t", fullName).Run() == nil
}

func tmuxNewSession(fullName, dir string) error {
	out, err := exec.Command("tmux", "new-session", "-d", "-s", fullName, "-c", dir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-session: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func tmuxSendLine(fullName, text string) error {
	return tmuxSendKeys(fullName, text, "Enter")
}

func tmuxSendText(fullName, text string) error {
	return tmuxSendKeys(fullName, text)
}

func tmuxPressEnter(fullName string) error {
	return tmuxSendKeys(fullName, "Enter")
}

func tmuxSendKeys(fullName string, keys ...string) error {
	args := append([]string{"send-keys", "-t", fullName}, keys...)
	out, err := exec.Command("tmux", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux send-keys: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func tmuxKillSession(fullName string) error {
	out, err := exec.Command("tmux", "kill-session", "-t", fullName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux kill-session: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func tmuxRenameSession(oldFullName, newFullName string) error {
	out, err := exec.Command("tmux", "rename-session", "-t", oldFullName, newFullName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux rename-session: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func tmuxAttach(fullName string) error {
	cmd := exec.Command("tmux", "attach-session", "-t", fullName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// tmuxListSuperstarSessions returns the bare names (without suffix) of running tmux sessions
// that match the superstar suffix.
func tmuxListSuperstarSessions() (map[string]bool, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		// tmux exits non-zero with "no server running" when no sessions exist
		if exitErr, ok := err.(*exec.ExitError); ok {
			if strings.Contains(string(exitErr.Stderr), "no server") {
				return map[string]bool{}, nil
			}
		}
		return nil, fmt.Errorf("tmux list-sessions: %w", err)
	}
	result := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if name, ok := sessionNameFromTmux(line); ok {
			result[name] = true
		}
	}
	return result, nil
}
