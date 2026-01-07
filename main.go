package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	count int
	width int
	height int
  input  textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	return model{
		count: 0,
		input: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
  
	case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	
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
  m.input, cmd = m.input.Update(msg)
	return m, cmd
}

var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("67")).
	Padding(1, 2)

func (m model) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"API testing",
		fmt.Sprintf("Count: %d", m.count),
		"",
		"",
		m.input.View(),
		"↑ / ↓ to change • q to quit",
	)
	
	box := boxStyle.
		Width(m.width - 4).
		Height(m.height - 2)

	return box.Render(content)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
