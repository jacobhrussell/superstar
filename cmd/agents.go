package cmd

import (
	"fmt"
	"strings"
)

var validAgents = []string{"claude", "codex", "cursor", "opencode"}

// agentCommands maps an agent name to the command used to launch it when the
// CLI binary differs from the agent name.
var agentCommands = map[string]string{
	"cursor": "agent",
}

func validateAgent(agent string) error {
	for _, a := range validAgents {
		if agent == a {
			return nil
		}
	}
	return fmt.Errorf("agent must be one of: %s", strings.Join(validAgents, ", "))
}

// agentCommand returns the shell command that launches the given agent.
func agentCommand(agent string) string {
	if cmd, ok := agentCommands[agent]; ok {
		return cmd
	}
	return agent
}
