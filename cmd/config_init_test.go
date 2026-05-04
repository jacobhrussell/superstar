package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigInitCreatesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := configInitCmd.RunE(configInitCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	path := filepath.Join(home, ".config", "superstar", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "default_agent") {
		t.Errorf("template missing 'default_agent', got:\n%s", content)
	}
	for _, agent := range validAgents {
		if !strings.Contains(content, agent) {
			t.Errorf("template missing agent %q, got:\n%s", agent, content)
		}
	}
}

func TestConfigInitRefusesOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", "superstar")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := configInitCmd.RunE(configInitCmd, nil)
	if err == nil {
		t.Fatal("expected error for existing file, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want to contain 'already exists'", err.Error())
	}

	data, _ := os.ReadFile(path)
	if string(data) != "existing" {
		t.Errorf("file was overwritten: %q", string(data))
	}
}
