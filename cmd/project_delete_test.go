package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestProjectDeleteRemovesEntry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/x", Agent: "claude"},
		"other":  {Dir: "/y", Agent: "codex"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectDeleteCmd.RunE(projectDeleteCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if _, exists := cfg.Projects["myproj"]; exists {
		t.Errorf("project still present after delete")
	}
	if _, exists := cfg.Projects["other"]; !exists {
		t.Errorf("unrelated project removed")
	}
}

func TestProjectDeleteUnknownProjectErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"only": {Dir: "/x", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := projectDeleteCmd.RunE(projectDeleteCmd, []string{"nope"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

func TestProjectDeleteNoProjectsErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { viper.Reset() })

	err := projectDeleteCmd.RunE(projectDeleteCmd, []string{"anything"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
