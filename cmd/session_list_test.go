package cmd

import "testing"

func TestSessionItemTitleAndDescription(t *testing.T) {
	tests := []struct {
		name    string
		item    sessionItem
		title   string
		descIn  string
	}{
		{
			"with project, no worktree",
			sessionItem{name: "alpha", project: "backend", agent: "claude", dir: "/tmp/a"},
			"[backend] alpha",
			"agent: claude · dir: /tmp/a",
		},
		{
			"no project, with worktree",
			sessionItem{name: "alpha", agent: "codex", dir: "/repo", worktree: "/repo-alpha"},
			"alpha",
			"agent: codex · worktree: /repo-alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.Title(); got != tt.title {
				t.Errorf("Title = %q, want %q", got, tt.title)
			}
			if got := tt.item.Description(); got != tt.descIn {
				t.Errorf("Description = %q, want %q", got, tt.descIn)
			}
		})
	}
}

func TestBuildSessionItemsSorting(t *testing.T) {
	sessions := map[string]SessionConfig{
		"loose":     {Agent: "claude", Dir: "/x"},
		"backend-2": {Project: "backend", Agent: "claude", Dir: "/x"},
		"backend-1": {Project: "backend", Agent: "claude", Dir: "/x"},
		"frontend":  {Project: "frontend", Agent: "claude", Dir: "/x"},
	}
	items := buildSessionItems(sessions, nil)
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = it.(sessionItem).name
	}
	want := []string{"backend-1", "backend-2", "frontend", "loose"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("items[%d] = %q, want %q (full: %v)", i, got[i], w, got)
		}
	}
}

func TestBuildSessionItemsAliveFilter(t *testing.T) {
	sessions := map[string]SessionConfig{
		"alive": {Agent: "claude", Dir: "/x"},
		"dead":  {Agent: "claude", Dir: "/x"},
	}
	alive := map[string]bool{"alive": true}
	items := buildSessionItems(sessions, alive)
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if name := items[0].(sessionItem).name; name != "alive" {
		t.Errorf("got %q, want %q", name, "alive")
	}
}
