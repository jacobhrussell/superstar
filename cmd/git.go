package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

func isGitRepo(dir string) bool {
	return exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run() == nil
}

func gitCreateWorktree(repoDir, worktreePath, branch string) error {
	out, err := exec.Command("git", "-C", repoDir, "worktree", "add", "-b", branch, worktreePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func gitRemoveWorktree(repoDir, worktreePath string) error {
	out, err := exec.Command("git", "-C", repoDir, "worktree", "remove", "--force", worktreePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func gitDeleteBranch(repoDir, branch string) error {
	out, err := exec.Command("git", "-C", repoDir, "branch", "-D", branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -D: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func gitFetchPRBranch(repoDir, prNumber, localBranch string) error {
	refspec := fmt.Sprintf("pull/%s/head:%s", prNumber, localBranch)
	out, err := exec.Command("git", "-C", repoDir, "fetch", "origin", refspec).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		// If the local branch already exists, reuse it.
		if strings.Contains(msg, "already exists") {
			return nil
		}
		return fmt.Errorf("git fetch pr: %s: %w", msg, err)
	}
	return nil
}

func gitAddWorktree(repoDir, worktreePath, branch string) error {
	out, err := exec.Command("git", "-C", repoDir, "worktree", "add", worktreePath, branch).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
