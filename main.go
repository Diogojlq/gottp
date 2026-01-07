package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	count int
}

func initialModel() model {
	return model{
		count: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up":
			m.count++

		case "down":
			m.count--
		}
	}

	return m, nil
}

var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(1, 2)

func (m model) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"🍵 Bubble Tea App",
		fmt.Sprintf("Count: %d", m.count),
		"",
		"↑ / ↓ to change • q to quit",
	)

	return boxStyle.Render(content)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
