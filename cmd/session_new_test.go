package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

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
			t.Cleanup(func() {
				sessionNewAgent = ""
				sessionNewDir = ""
				viper.Reset()
			})

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
	t.Cleanup(func() {
		sessionNewAgent = ""
		sessionNewDir = ""
		viper.Reset()
	})

	sessionNewDir = ""
	sessionNewAgent = "claude"

	err := sessionNewCmd.PreRunE(sessionNewCmd, nil)
	if err == nil {
		t.Fatal("expected error when --dir is empty, got nil")
	}
}
