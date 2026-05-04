package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetProjectNewState() {
	projectNewDir = ""
	projectNewAgent = ""
	projectNewGithub = ""
	projectNewAfterScript = ""
	viper.Reset()
}

func TestProjectNewCreatesEntry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectNewState)

	projectNewDir = "/work/proj"
	projectNewAgent = "claude"

	if err := projectNewCmd.PreRunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("PreRunE: %v", err)
	}
	if err := projectNewCmd.RunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	got, ok := cfg.Projects["myproj"]
	if !ok {
		t.Fatalf("project not saved")
	}
	if got.Dir != "/work/proj" || got.Agent != "claude" {
		t.Errorf("got %+v", got)
	}
}

func TestProjectNewSavesAfterScript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectNewState)

	projectNewDir = "/work/proj"
	projectNewAgent = "claude"
	projectNewAfterScript = "/scripts/after.sh"

	if err := projectNewCmd.PreRunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("PreRunE: %v", err)
	}
	if err := projectNewCmd.RunE(projectNewCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	got := cfg.Projects["myproj"]
	if got.SessionAfterScript != "/scripts/after.sh" {
		t.Errorf("SessionAfterScript = %q, want %q", got.SessionAfterScript, "/scripts/after.sh")
	}
}

func TestProjectNewRefusesDuplicate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectNewState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/old", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

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

func TestValidateNewProjectName(t *testing.T) {
	existing := map[string]ProjectConfig{
		"taken": {Dir: "/x", Agent: "claude"},
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"duplicate", "taken", true},
		{"valid new name", "fresh", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNewProjectName(tt.input, existing)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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
