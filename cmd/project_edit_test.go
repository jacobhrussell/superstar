package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetProjectEditState() {
	projectEditDir = ""
	projectEditAgent = ""
	projectEditAfterScript = ""
	viper.Reset()
	projectEditCmd.ResetFlags()
	projectEditCmd.Flags().StringVarP(&projectEditDir, "dir", "d", "", "")
	projectEditCmd.Flags().StringVarP(&projectEditAgent, "agent", "a", "", "")
	projectEditCmd.Flags().StringVar(&projectEditAfterScript, "session-after-script", "", "")
}

func TestProjectEditUpdatesFlaggedFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/old", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectEditCmd.Flags().Set("dir", "/new"); err != nil {
		t.Fatalf("set dir: %v", err)
	}
	if err := projectEditCmd.Flags().Set("agent", "codex"); err != nil {
		t.Fatalf("set agent: %v", err)
	}

	if err := projectEditCmd.RunE(projectEditCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	got := cfg.Projects["myproj"]
	if got.Dir != "/new" || got.Agent != "codex" {
		t.Errorf("got %+v, want {Dir:/new Agent:codex}", got)
	}
}

func TestProjectEditPartialUpdateKeepsOtherField(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/old", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectEditCmd.Flags().Set("agent", "codex"); err != nil {
		t.Fatalf("set agent: %v", err)
	}

	if err := projectEditCmd.RunE(projectEditCmd, []string{"myproj"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	got := cfg.Projects["myproj"]
	if got.Dir != "/old" {
		t.Errorf("dir changed unexpectedly: %q, want %q", got.Dir, "/old")
	}
	if got.Agent != "codex" {
		t.Errorf("agent = %q, want %q", got.Agent, "codex")
	}
}

func TestProjectEditUpdatesAfterScript(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/old", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectEditCmd.Flags().Set("session-after-script", "/scripts/after.sh"); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	if err := projectEditCmd.RunE(projectEditCmd, []string{"myproj"}); err != nil {
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
	if got.Dir != "/old" || got.Agent != "claude" {
		t.Errorf("other fields changed: %+v", got)
	}
}

func TestProjectEditUnknownProjectErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"only": {Dir: "/x", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectEditCmd.Flags().Set("dir", "/new"); err != nil {
		t.Fatalf("set dir: %v", err)
	}

	err := projectEditCmd.RunE(projectEditCmd, []string{"nope"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want to contain 'not found'", err.Error())
	}
}

func TestProjectEditNoProjectsErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	err := projectEditCmd.RunE(projectEditCmd, []string{"anything"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectEditRejectsInvalidAgent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(resetProjectEditState)

	if err := saveConfig(&Config{Projects: map[string]ProjectConfig{
		"myproj": {Dir: "/old", Agent: "claude"},
	}}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := projectEditCmd.Flags().Set("agent", "gpt"); err != nil {
		t.Fatalf("set agent: %v", err)
	}

	err := projectEditCmd.RunE(projectEditCmd, []string{"myproj"})
	if err == nil {
		t.Fatal("expected error for invalid agent, got nil")
	}
}
