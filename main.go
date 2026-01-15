package main

import (
	"fmt"
	"os"
	"strings"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
)

type model struct {
	count 	int
	width 	int
	height 	int
  input   textinput.Model
	methods []string
	selectedMethod int
	responseBody   string
	viewport viewport.Model
  ready    bool
}

type httpResponseMsg string
type errorMsg error

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 30

	return model{
		input: ti,
		methods:        []string{"GET", "POST", "PUT", "DELETE"},
		selectedMethod: 0,
		responseBody:   "Waiting request...",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
  
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	
	  rightWidth := int(float64(m.width)*0.6) - 6 // Ajuste de margem
    viewHeight := m.height - 10 // Ajuste de margem vertical

    if !m.ready {
        m.viewport = viewport.New(rightWidth, viewHeight)         
				m.ready = true 
		} else {
        m.viewport.Width = rightWidth
        m.viewport.Height = viewHeight
    }

	case tea.KeyMsg:

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up":
			if m.selectedMethod > 0 {
				m.selectedMethod--
			}
		case "down":
			if m.selectedMethod < len(m.methods)-1 {
				m.selectedMethod++
			}
		case "enter":
			m.responseBody = "Making request..."
			m.viewport.SetContent(m.responseBody)
			return m, m.makeRequest()
		}

	case httpResponseMsg:
		m.responseBody = string(msg)
		m.viewport.SetContent(m.responseBody)
		return m, nil

	case errorMsg:
		m.responseBody = "Erro: " + msg.Error()
		return m, nil
	}

		var vpCmd tea.Cmd
    m.viewport, vpCmd = m.viewport.Update(msg)
    cmds = append(cmds, vpCmd)

    var inputCmd tea.Cmd
    m.input, inputCmd = m.input.Update(msg)
    cmds = append(cmds, inputCmd)

    return m, tea.Batch(cmds...)
}

var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("33")).
	Padding(1, 2)

func (m model) View() string {

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftWidth := int(float64(m.width) * 0.4) - 2
	rightWidth := int(float64(m.width) * 0.6) - 2
	mainHeight := m.height - 4

	var methodList strings.Builder
	methodList.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("178")).Render("Method") + "\n\n")

	for i, method := range m.methods {
		if i == m.selectedMethod {
			methodList.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("21")).Render("> "+method) + "\n")
		} else {
			methodList.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render("  "+method) + "\n")
		}
	}

	leftContent := lipgloss.JoinVertical(
		lipgloss.Left,
		methodList.String(),
		"\n",
		m.input.View(),
		"\n",
		"↑ / ↓ to change • q to quit",
	)
	
	leftBox := boxStyle.
		Width(leftWidth).
		Height(mainHeight).
		Render(leftContent)

	rightContent := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("RESULT:"),
		"\n",
		m.viewport.View(),
	)

	rightBox := boxStyle.
		Width(rightWidth).
		Height(mainHeight).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (m model) makeRequest() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 10 * time.Second}
		
		method := m.methods[m.selectedMethod]
		url := m.input.Value()

		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return errorMsg(err)
		}

		res, err := client.Do(req)
		if err != nil {
			return errorMsg(err)
		}
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		return httpResponseMsg(string(body))
	}
}


func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
