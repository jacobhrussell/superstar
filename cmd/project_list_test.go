package cmd

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestProjectItemMethods(t *testing.T) {
	i := projectItem{name: "alpha", agent: "claude", dir: "/tmp/alpha"}
	if got := i.Title(); got != "alpha" {
		t.Errorf("Title() = %q, want %q", got, "alpha")
	}
	if got := i.FilterValue(); got != "alpha" {
		t.Errorf("FilterValue() = %q, want %q", got, "alpha")
	}
	if got := i.Description(); got != "agent: claude · dir: /tmp/alpha" {
		t.Errorf("Description() = %q", got)
	}
}

func TestBuildProjectItemsSorted(t *testing.T) {
	projects := map[string]ProjectConfig{
		"charlie": {Dir: "/c", Agent: "claude"},
		"alpha":   {Dir: "/a", Agent: "codex"},
		"bravo":   {Dir: "/b", Agent: "cursor"},
	}
	items := buildProjectItems(projects)

	want := []string{"alpha", "bravo", "charlie"}
	if len(items) != len(want) {
		t.Fatalf("got %d items, want %d", len(items), len(want))
	}
	for i, name := range want {
		got := items[i].(projectItem).name
		if got != name {
			t.Errorf("items[%d].name = %q, want %q", i, got, name)
		}
	}
}

func TestBuildProjectItemsEmpty(t *testing.T) {
	if items := buildProjectItems(nil); len(items) != 0 {
		t.Errorf("expected empty, got %d items", len(items))
	}
}

func TestProjectListEnterRecordsSelection(t *testing.T) {
	items := buildProjectItems(map[string]ProjectConfig{
		"alpha": {Dir: "/a", Agent: "claude"},
		"beta":  {Dir: "/b", Agent: "codex"},
	})
	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	m := projectListModel{list: l}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	got := updated.(projectListModel).selected
	if got != "alpha" {
		t.Errorf("selected = %q, want %q", got, "alpha")
	}
}
