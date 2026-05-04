package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func resetSessionNewState() {
	sessionNewAgent = ""
	sessionNewDir = ""
	sessionNewProject = ""
	sessionNewName = ""
	sessionNewPrompt = ""
	viper.Reset()
}

func TestSessionNewAgentValidation(t *testing.T) {
	tests := []struct {
		name       string
		flag       string
		viperValue string
		wantErr    bool
		wantAgent  string
	}{
		{"empty flag and viper", "", "", true, ""},
		{"valid flag", "claude", "", false, "claude"},
		{"invalid flag", "gpt", "", true, "gpt"},
		{"flag overrides viper", "codex", "claude", false, "codex"},
		{"falls back to viper", "", "claude", false, "claude"},
		{"invalid viper value", "", "gpt", true, "gpt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetSessionNewState)

			sessionNewName = "my-session"
			sessionNewDir = "/tmp/test"
			sessionNewAgent = tt.flag
			if tt.viperValue != "" {
				viper.Set("default_agent", tt.viperValue)
			}

			err := sessionNewCmd.PreRunE(sessionNewCmd, nil)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sessionNewAgent != tt.wantAgent {
				t.Errorf("sessionNewAgent = %q, want %q", sessionNewAgent, tt.wantAgent)
			}
		})
	}
}

func TestSessionNewDirRequired(t *testing.T) {
	t.Cleanup(resetSessionNewState)

	sessionNewName = "my-session"
	sessionNewAgent = "claude"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err == nil {
		t.Fatal("expected error when --dir is empty, got nil")
	}
}

func TestSessionNewProjectResolvesDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/work/proj", Agent: "codex"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	sessionNewName = "my-session"
	sessionNewProject = "myproj"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionNewAgent != "codex" {
		t.Errorf("agent = %q, want %q", sessionNewAgent, "codex")
	}
	if sessionNewDir != "/work/proj" {
		t.Errorf("dir = %q, want %q", sessionNewDir, "/work/proj")
	}
}

func TestSessionNewFlagOverridesProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/work/proj", Agent: "codex"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	sessionNewName = "my-session"
	sessionNewProject = "myproj"
	sessionNewAgent = "claude"
	sessionNewDir = "/override"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessionNewAgent != "claude" {
		t.Errorf("agent = %q, want flag value %q", sessionNewAgent, "claude")
	}
	if sessionNewDir != "/override" {
		t.Errorf("dir = %q, want flag value %q", sessionNewDir, "/override")
	}
}

func TestSessionNewUnknownProjectErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetSessionNewState)

	sessionNewName = "my-session"
	sessionNewProject = "nope"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}
}
