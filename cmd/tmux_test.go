package cmd

import "testing"

func TestTmuxNameRoundtrip(t *testing.T) {
	if got := tmuxName("alpha"); got != "alpha-superstar" {
		t.Errorf("tmuxName = %q, want alpha-superstar", got)
	}
	if name, ok := sessionNameFromTmux("alpha-superstar"); !ok || name != "alpha" {
		t.Errorf("sessionNameFromTmux(alpha-superstar) = (%q, %v)", name, ok)
	}
	if _, ok := sessionNameFromTmux("alpha"); ok {
		t.Errorf("sessionNameFromTmux(alpha) should not match")
	}
	if _, ok := sessionNameFromTmux("plain-tmux-session"); ok {
		t.Errorf("non-superstar session should not match")
	}
}
