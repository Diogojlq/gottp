package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	count          int
	width          int
	height         int
	input          textinput.Model
	methods        []string
	selectedMethod int
	responseBody   string
	viewport       viewport.Model
	ready          bool
	focusedPanel   string
}

type httpResponseMsg string
type errorMsg error

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 35

	return model{
		input:          ti,
		methods:        []string{"GET", "POST", "PUT", "DELETE"},
		selectedMethod: 0,
		responseBody:   "Waiting request...",
		focusedPanel:   "methods",
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

		rightWidth := int(float64(m.width)*0.6) - 6 
		viewHeight := m.height - 10                 

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

		case "alt+left", "alt+h":
			if m.focusedPanel == "response" {
				m.focusedPanel = "methods"
				m.input.Blur()
			}
		case "alt+right", "alt+l":
			if m.focusedPanel == "methods" {
				m.focusedPanel = "response"
			}

		case "up":
			if m.focusedPanel == "methods" {
				if m.selectedMethod > 0 {
					m.selectedMethod--
				}
			} else {
				m.viewport.LineUp(1)
			}
		case "down":
			if m.focusedPanel == "methods" {
				if m.selectedMethod < len(m.methods)-1 {
					m.selectedMethod++
				}
			} else {
				m.viewport.LineDown(1)
			}
		case "pgup":
			if m.focusedPanel == "response" {
				m.viewport.HalfViewUp()
			}
		case "pgdown":
			if m.focusedPanel == "response" {
				m.viewport.HalfViewDown()
			}
		case "home":
			if m.focusedPanel == "response" {
				m.viewport.GotoTop()
			}
		case "end":
			if m.focusedPanel == "response" {
				m.viewport.GotoBottom()
			}
		case "tab":
			if m.input.Focused() {
				m.input.Blur()
			} else {
				m.input.Focus()
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
		m.viewport.SetContent(m.responseBody)
		return m, nil
	}

	if m.focusedPanel == "response" {
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

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

	leftWidth := int(float64(m.width)*0.4) - 2
	rightWidth := int(float64(m.width)*0.6) - 2
	mainHeight := m.height - 4

	var methodList strings.Builder
	methodListTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("178")).Render("Method")
	if m.focusedPanel == "methods" {
		methodListTitle += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("●")
	}
	methodList.WriteString(methodListTitle + "\n\n")

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
		"↑↓ select • Alt+Right go to response • Enter send • q quit",
	)

	leftBox := boxStyle.
		Width(leftWidth).
		Height(mainHeight).
		Render(leftContent)

	rightContent := lipgloss.JoinVertical(
		lipgloss.Left,
		func() string {
			title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("RESULT:")
			if m.focusedPanel == "response" {
				title += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("●")
			}
			return title
		}(),
		"\n",
		m.viewport.View(),
	)

	rightBox := boxStyle.
		Width(rightWidth).
		Height(mainHeight).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func formatResponse(body []byte) string {
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return string(body)
	}

	var prettyJSON bytes.Buffer
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		return string(body)
	}

	return prettyJSON.String()
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
		formatted := formatResponse(body)
		return httpResponseMsg(formatted)
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
