package cmd

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type projectItem struct {
	name, agent, dir, github, afterScript string
}

func (i projectItem) Title() string { return i.name }
func (i projectItem) Description() string {
	desc := fmt.Sprintf("agent: %s · dir: %s", i.agent, i.dir)
	if i.github != "" {
		desc += " · github: " + i.github
	}
	if i.afterScript != "" {
		desc += " · session-after-script: " + i.afterScript
	}
	return desc
}
func (i projectItem) FilterValue() string { return i.name }

type projectListModel struct {
	list     list.Model
	selected string
}

func (m projectListModel) Init() tea.Cmd { return nil }

func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(projectItem); ok {
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

func (m projectListModel) View() string {
	return m.list.View()
}

func buildProjectItems(projects map[string]ProjectConfig) []list.Item {
	items := make([]list.Item, 0, len(projects))
	for _, name := range projectNames(projects) {
		p := projects[name]
		items = append(items, projectItem{name: name, agent: p.Agent, dir: p.Dir, github: p.Github, afterScript: p.SessionAfterScript})
	}
	return items
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List projects with details (enter copies name to clipboard)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if len(cfg.Projects) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "no projects configured")
			return nil
		}

		l := list.New(buildProjectItems(cfg.Projects), list.NewDefaultDelegate(), 0, 0)
		l.Title = "Projects"
		l.SetFilteringEnabled(true)

		final, err := tea.NewProgram(projectListModel{list: l}, tea.WithAltScreen()).Run()
		if err != nil {
			return err
		}

		m, ok := final.(projectListModel)
		if !ok || m.selected == "" {
			return nil
		}
		if err := clipboard.WriteAll(m.selected); err != nil {
			return fmt.Errorf("could not copy to clipboard: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "copied %q to clipboard\n", m.selected)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}
