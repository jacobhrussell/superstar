package cmd

import (
	"fmt"
	"strings"
)

var validAgents = []string{"claude", "codex", "cursor", "opencode"}

func validateAgent(agent string) error {
	for _, a := range validAgents {
		if agent == a {
			return nil
		}
	}
	return fmt.Errorf("agent must be one of: %s", strings.Join(validAgents, ", "))
}
