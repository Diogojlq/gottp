package func main()
	
import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	count int
}

func initialModel() model {
	return model{count: 0}
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

func (m model) View() string {
	return fmt.Sprintf(
		"Bubble Tea started 🍵\n\nCount: %d\n\n↑ / ↓ to change\nq to quit\n",
		m.count,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
