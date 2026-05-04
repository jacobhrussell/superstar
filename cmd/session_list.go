package cmd

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type sessionItem struct {
	name, project, agent, dir, worktree string
}

func (i sessionItem) Title() string {
	if i.project != "" {
		return fmt.Sprintf("[%s] %s", i.project, i.name)
	}
	return i.name
}
func (i sessionItem) Description() string {
	if i.worktree != "" {
		return fmt.Sprintf("agent: %s · worktree: %s", i.agent, i.worktree)
	}
	return fmt.Sprintf("agent: %s · dir: %s", i.agent, i.dir)
}
func (i sessionItem) FilterValue() string {
	if i.project != "" {
		return i.project + " " + i.name
	}
	return i.name
}

type sessionListModel struct {
	list     list.Model
	selected string
}

func (m sessionListModel) Init() tea.Cmd { return nil }

func (m sessionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(sessionItem); ok {
					m.selected = item.name
					return m, tea.Quit
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m sessionListModel) View() string { return m.list.View() }

func buildSessionItems(sessions map[string]SessionConfig, alive map[string]bool) []list.Item {
	type pair struct {
		name string
		s    SessionConfig
	}
	pairs := make([]pair, 0, len(sessions))
	for name, s := range sessions {
		if alive != nil && !alive[name] {
			continue
		}
		pairs = append(pairs, pair{name, s})
	}
	// Sort: project first (empty last), then name.
	sort.Slice(pairs, func(i, j int) bool {
		pi, pj := pairs[i].s.Project, pairs[j].s.Project
		if pi == "" && pj != "" {
			return false
		}
		if pi != "" && pj == "" {
			return true
		}
		if pi != pj {
			return pi < pj
		}
		return pairs[i].name < pairs[j].name
	})

	items := make([]list.Item, 0, len(pairs))
	for _, p := range pairs {
		items = append(items, sessionItem{
			name:     p.name,
			project:  p.s.Project,
			agent:    p.s.Agent,
			dir:      p.s.Dir,
			worktree: p.s.Worktree,
		})
	}
	return items
}

var sessionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List superstar sessions (enter to attach)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Sessions) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no sessions")
			return nil
		}

		var alive map[string]bool
		if err := tmuxAvailable(); err == nil {
			alive, _ = tmuxListSuperstarSessions()
		}

		items := buildSessionItems(cfg.Sessions, alive)
		if len(items) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no live sessions (config has stale entries; clean with `session delete`)")
			return nil
		}

		l := list.New(items, list.NewDefaultDelegate(), 0, 0)
		l.Title = "Sessions"
		l.SetFilteringEnabled(true)

		final, err := tea.NewProgram(sessionListModel{list: l}, tea.WithAltScreen()).Run()
		if err != nil {
			return err
		}
		m, ok := final.(sessionListModel)
		if !ok || m.selected == "" {
			return nil
		}
		return tmuxAttach(tmuxName(m.selected))
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
}
