package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetProjectNewState() {
	projectNewDir = ""
	projectNewAgent = ""
	viper.Reset()
}

func TestProjectNewCreatesEntry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Cleanup(resetProjectNewState)

	projectNewDir = "/work/proj"
	projectNewAgent = "claude"

	if err := projectNewCmd.PreRunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("PreRunE: %v", err)
	}
	if err := projectNewCmd.RunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	path := filepath.Join(home, ".config", "superstar", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	content := string(data)
	for _, want := range []string{"myproj", "/work/proj", "claude"} {
		if !strings.Contains(content, want) {
			t.Errorf("config missing %q, got:\n%s", want, content)
		}
	}
}

func TestProjectNewRefusesDuplicate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Cleanup(resetProjectNewState)

	viper.Set("projects.myproj.agent", "claude")
	viper.Set("projects.myproj.dir", "/old")

	projectNewDir = "/new"
	projectNewAgent = "codex"

	err := projectNewCmd.RunE(projectNewCmd, []string{"myproj"})
	if err == nil {
		t.Fatal("expected error for duplicate project, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want to contain 'already exists'", err.Error())
	}
}

func TestPromptProjectHappyPath(t *testing.T) {
	t.Cleanup(resetProjectNewState)

	in := strings.NewReader("my-proj\n/work/proj\nclaude\n")
	var out bytes.Buffer

	name, dir, agent, err := promptProject(in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "my-proj" {
		t.Errorf("name = %q, want %q", name, "my-proj")
	}
	if dir != "/work/proj" {
		t.Errorf("dir = %q, want %q", dir, "/work/proj")
	}
	if agent != "claude" {
		t.Errorf("agent = %q, want %q", agent, "claude")
	}
}

func TestPromptProjectRepromptsOnInvalidAgent(t *testing.T) {
	t.Cleanup(resetProjectNewState)

	in := strings.NewReader("my-proj\n/work/proj\ngpt\nclaude\n")
	var out bytes.Buffer

	_, _, agent, err := promptProject(in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent != "claude" {
		t.Errorf("agent = %q, want %q after re-prompt", agent, "claude")
	}
	if !strings.Contains(out.String(), "invalid agent") {
		t.Errorf("expected 'invalid agent' message, got: %q", out.String())
	}
}

func TestPromptProjectRepromptsOnDuplicateName(t *testing.T) {
	t.Cleanup(resetProjectNewState)

	viper.Set("projects.taken.agent", "claude")
	viper.Set("projects.taken.dir", "/old")

	in := strings.NewReader("taken\nfresh\n/work/proj\nclaude\n")
	var out bytes.Buffer

	name, _, _, err := promptProject(in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "fresh" {
		t.Errorf("name = %q, want %q after re-prompt", name, "fresh")
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("expected duplicate-name message, got: %q", out.String())
	}
}

func TestProjectNewInteractiveMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Cleanup(resetProjectNewState)

	in := strings.NewReader("interactive-proj\n/tmp/interactive\ncodex\n")
	var out bytes.Buffer
	projectNewCmd.SetIn(in)
	projectNewCmd.SetOut(&out)
	t.Cleanup(func() {
		projectNewCmd.SetIn(nil)
		projectNewCmd.SetOut(nil)
	})

	if err := projectNewCmd.RunE(projectNewCmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	path := filepath.Join(home, ".config", "superstar", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	content := string(data)
	for _, want := range []string{"interactive-proj", "/tmp/interactive", "codex"} {
		if !strings.Contains(content, want) {
			t.Errorf("config missing %q, got:\n%s", want, content)
		}
	}
}

func TestProjectNewRequiresFlags(t *testing.T) {
	tests := []struct {
		name  string
		dir   string
		agent string
	}{
		{"missing both", "", ""},
		{"missing dir", "", "claude"},
		{"missing agent", "/tmp", ""},
		{"invalid agent", "/tmp", "gpt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetProjectNewState)
			projectNewDir = tt.dir
			projectNewAgent = tt.agent

			if err := projectNewCmd.PreRunE(projectNewCmd, []string{"myproj"}); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
