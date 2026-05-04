package cmd

import (
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
