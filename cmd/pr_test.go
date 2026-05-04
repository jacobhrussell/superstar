package cmd

import (
	"strings"
	"testing"
)

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantNum   string
		wantErr   bool
	}{
		{"https url", "https://github.com/jacobhrussell/superstar/pull/42", "jacobhrussell", "superstar", "42", false},
		{"http url", "http://github.com/foo/bar/pull/1", "foo", "bar", "1", false},
		{"no scheme", "github.com/foo/bar/pull/1", "foo", "bar", "1", false},
		{"trailing path", "https://github.com/foo/bar/pull/123/files", "foo", "bar", "123", false},
		{"with whitespace", "  https://github.com/foo/bar/pull/7  ", "foo", "bar", "7", false},
		{"issue url", "https://github.com/foo/bar/issues/9", "", "", "", true},
		{"non-github", "https://gitlab.com/foo/bar/pull/1", "", "", "", true},
		{"empty", "", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Owner != tt.wantOwner || got.Repo != tt.wantRepo || got.Number != tt.wantNum {
				t.Errorf("got %+v, want owner=%q repo=%q num=%q", got, tt.wantOwner, tt.wantRepo, tt.wantNum)
			}
		})
	}
}

func TestFindProjectByGithub(t *testing.T) {
	projects := map[string]ProjectConfig{
		"alpha": {Dir: "/a", Agent: "claude", Github: "jacobhrussell/superstar"},
		"beta":  {Dir: "/b", Agent: "codex"},
		"gamma": {Dir: "/g", Agent: "claude", Github: "Other/Repo"},
	}

	t.Run("matches case-insensitively", func(t *testing.T) {
		name, proj, ok := findProjectByGithub(projects, "JacobHRussell/Superstar")
		if !ok {
			t.Fatal("expected match")
		}
		if name != "alpha" || proj.Dir != "/a" {
			t.Errorf("got name=%q dir=%q, want alpha /a", name, proj.Dir)
		}
	})

	t.Run("no match", func(t *testing.T) {
		_, _, ok := findProjectByGithub(projects, "missing/repo")
		if ok {
			t.Error("expected no match")
		}
	})

	t.Run("ignores empty github fields", func(t *testing.T) {
		_, _, ok := findProjectByGithub(projects, "")
		if ok {
			t.Error("empty target should not match the project with empty github")
		}
	})
}

func TestSessionNewPullRequestResolvesProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"superstar": {Dir: "/work/superstar", Agent: "codex", Github: "jacobhrussell/superstar"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	sessionNewPullRequest = "https://github.com/jacobhrussell/superstar/pull/7"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionNewProject != "superstar" {
		t.Errorf("project = %q, want %q", sessionNewProject, "superstar")
	}
	if sessionNewDir != "/work/superstar" {
		t.Errorf("dir = %q, want %q", sessionNewDir, "/work/superstar")
	}
	if sessionNewAgent != "codex" {
		t.Errorf("agent = %q, want %q", sessionNewAgent, "codex")
	}
}

func TestSessionNewPullRequestNoMatchingProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"other": {Dir: "/x", Agent: "claude", Github: "someone/else"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	sessionNewPullRequest = "https://github.com/jacobhrussell/superstar/pull/7"

	err := sessionNewCmd.PreRunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "jacobhrussell/superstar") {
		t.Errorf("error %q should mention the missing repo", err.Error())
	}
	if !strings.Contains(err.Error(), "project new") {
		t.Errorf("error %q should point user to project new", err.Error())
	}
}

func TestSessionNewPullRequestRejectsCombinedWithProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	sessionNewPullRequest = "https://github.com/foo/bar/pull/1"
	sessionNewProject = "anything"

	err := sessionNewCmd.PreRunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error when both flags are set")
	}
	if !strings.Contains(err.Error(), "--project") {
		t.Errorf("error %q should mention --project", err.Error())
	}
}

func TestSessionNewPullRequestInvalidURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	sessionNewPullRequest = "not a url"

	err := sessionNewCmd.PreRunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
