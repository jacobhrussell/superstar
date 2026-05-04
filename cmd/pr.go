package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var prURLPattern = regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/pull/(\d+)`)

type prRef struct {
	Owner  string
	Repo   string
	Number string
}

func (p prRef) ownerRepo() string {
	return p.Owner + "/" + p.Repo
}

func parsePRURL(s string) (prRef, error) {
	m := prURLPattern.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return prRef{}, fmt.Errorf("invalid GitHub PR URL %q (expected https://github.com/owner/repo/pull/N)", s)
	}
	return prRef{Owner: m[1], Repo: m[2], Number: m[3]}, nil
}

func findProjectByGithub(projects map[string]ProjectConfig, ownerRepo string) (string, ProjectConfig, bool) {
	target := strings.ToLower(ownerRepo)
	if target == "" {
		return "", ProjectConfig{}, false
	}
	for name, p := range projects {
		if p.Github != "" && strings.ToLower(p.Github) == target {
			return name, p, true
		}
	}
	return "", ProjectConfig{}, false
}

func ghAvailable() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New("gh is not installed (try: brew install gh)")
	}
	return nil
}

func ghPRHeadBranch(p prRef) (string, error) {
	if err := ghAvailable(); err != nil {
		return "", err
	}
	cmd := exec.Command("gh", "pr", "view", p.Number,
		"--repo", p.ownerRepo(),
		"--json", "headRefName")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("gh pr view: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("gh pr view: %w", err)
	}
	var data struct {
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return "", fmt.Errorf("parse gh pr view output: %w", err)
	}
	if data.HeadRefName == "" {
		return "", errors.New("PR head branch name is empty")
	}
	return data.HeadRefName, nil
}
