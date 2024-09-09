package interact

// import (
// 	"fmt"
// 	"strings"

// 	tea "github.com/charmbracelet/bubbletea"
// )

// type multiOptionModel struct {
// 	title   string
// 	options []Option
// 	cursor  int
// }

// type Option struct {
// 	Option  string
// 	Checked bool
// }

// func (m multiOptionModel) Init() tea.Cmd {
// 	return nil
// }

// func (m multiOptionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c", "q", "esc":
// 			return m, tea.Quit

// 		case "enter":
// 			m.options[m.cursor].Checked = !m.options[m.cursor].Checked

// 		case "down", "j":
// 			m.cursor++
// 			if m.cursor >= len(m.options) {
// 				m.cursor = 0
// 			}

// 		case "up", "k":
// 			m.cursor--
// 			if m.cursor < 0 {
// 				m.cursor = len(m.options) - 1
// 			}
// 		}
// 	}
// 	return m, nil
// }

// func (m multiOptionModel) View() string {

// 	title := fmt.Sprintf("%s\n", m.title)
// 	selectedText := "> "
// 	unselectedText := "  "

// 	checkedText := "[✓] "
// 	unCheckedText := "[ ] "

// 	s := strings.Builder{}
// 	s.WriteString(title)

// 	for i, option := range m.options {
// 		if i == m.cursor {
// 			s.WriteString(selectedText)
// 		} else {
// 			s.WriteString(unselectedText)
// 		}

// 		if option.Checked {
// 			s.WriteString(checkedText)
// 		} else {
// 			s.WriteString(unCheckedText)
// 		}

// 		s.WriteString(option.Option)
// 		s.WriteString("\n")
// 	}

// 	return s.String()

// }

// func MultiOption(title string, options []Option, cursor int) error {

// 	p := tea.NewProgram(multiOptionModel{title: title, options: options, cursor: cursor})

// 	_, err := p.Run()

// 	if err != nil {
// 		return fmt.Errorf("error running prompt: %v", err)
// 	}

// 	return nil

// }
