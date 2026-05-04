package cmd

import (
	"bytes"
	"strings"
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

func TestSessionNewPreRunResolvesAgent(t *testing.T) {
	tests := []struct {
		name       string
		flag       string
		viperValue string
		wantErr    bool
		wantAgent  string
	}{
		{"empty flag and viper", "", "", false, ""},
		{"valid flag", "claude", "", false, "claude"},
		{"invalid flag", "gpt", "", true, "gpt"},
		{"flag overrides viper", "codex", "claude", false, "codex"},
		{"falls back to viper", "", "claude", false, "claude"},
		{"invalid viper value", "", "gpt", true, "gpt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetSessionNewState)

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

func TestSessionNewRunRequiresDirNonInteractive(t *testing.T) {
	t.Cleanup(resetSessionNewState)

	sessionNewAgent = "claude"
	var out bytes.Buffer
	sessionNewCmd.SetOut(&out)
	t.Cleanup(func() { sessionNewCmd.SetOut(nil) })

	err := sessionNewCmd.RunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error when --dir is empty in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "--dir is required") {
		t.Errorf("error = %q, want to contain '--dir is required'", err.Error())
	}
}

func TestSessionNewRunRequiresAgentNonInteractive(t *testing.T) {
	t.Cleanup(resetSessionNewState)

	sessionNewDir = "/tmp/test"
	var out bytes.Buffer
	sessionNewCmd.SetOut(&out)
	t.Cleanup(func() { sessionNewCmd.SetOut(nil) })

	err := sessionNewCmd.RunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error when --agent is empty in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "--agent is required") {
		t.Errorf("error = %q, want to contain '--agent is required'", err.Error())
	}
}

func TestSessionNewRunSucceedsWithRequiredOnly(t *testing.T) {
	t.Cleanup(resetSessionNewState)

	sessionNewDir = "/tmp/test"
	sessionNewAgent = "claude"
	var out bytes.Buffer
	sessionNewCmd.SetOut(&out)
	t.Cleanup(func() { sessionNewCmd.SetOut(nil) })

	if err := sessionNewCmd.RunE(sessionNewCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"dir: /tmp/test", "agent: claude"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q, got: %q", want, out.String())
		}
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

	sessionNewProject = "nope"

	if err := sessionNewCmd.PreRunE(sessionNewCmd, nil); err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}
}

func TestAnySessionFieldMissing(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		expected bool
	}{
		{"all set", func() {
			sessionNewAgent = "claude"
			sessionNewDir = "/x"
			sessionNewPrompt = "do thing"
			sessionNewName = "n"
		}, false},
		{"agent missing", func() {
			sessionNewDir = "/x"
			sessionNewPrompt = "p"
			sessionNewName = "n"
		}, true},
		{"prompt missing", func() {
			sessionNewAgent = "claude"
			sessionNewDir = "/x"
			sessionNewName = "n"
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(resetSessionNewState)
			tt.setup()
			if got := anySessionFieldMissing(); got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
