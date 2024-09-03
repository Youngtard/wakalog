package interact

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type choiceModel[T fmt.Stringer] struct {
	title   string
	choices []T
	choice  T
	cursor  int
}

func (m choiceModel[T]) Init() tea.Cmd {
	return nil
}

func (m choiceModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			m.choice = m.choices[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

func (m choiceModel[T]) View() string {

	title := fmt.Sprintf("%s\n", m.title)
	selectedText := "> "
	unselectedText := "  "

	s := strings.Builder{}
	s.WriteString(title)

	for i, choice := range m.choices {
		if m.cursor == i {
			s.WriteString(selectedText)
		} else {
			s.WriteString(unselectedText)
		}
		s.WriteString(choice.String())
		s.WriteString("\n")
	}

	return s.String()
}

func Choice[T fmt.Stringer](title string, choices []T, dest *T) error {

	p := tea.NewProgram(choiceModel[T]{title: title, choices: choices})

	m, err := p.Run()

	if err != nil {
		return fmt.Errorf("error running prompt: %v", err)
	}

	if m, ok := m.(choiceModel[T]); ok {
		*dest = m.choice
		return nil
	}

	return fmt.Errorf("failed to get a choice")

}
