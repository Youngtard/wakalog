package interact

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type textInputModel struct {
	title     string
	textInput textinput.Model
	err       error
}

type (
	errMsg error
)

func initialModel(title string) textInputModel {
	ti := textinput.New()
	ti.Placeholder = "Enter here..."
	ti.Focus()

	return textInputModel{
		title:     title,
		textInput: ti,
		err:       nil,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m textInputModel) View() string {
	return fmt.Sprintf(
		"%s\n%s\n\n",
		m.title,
		m.textInput.View(),
	)
}

func TextInput(title string, dest *string) error {
	p := tea.NewProgram(initialModel(title))

	m, err := p.Run()

	if err != nil {
		return fmt.Errorf("error running prompt: %v", err)
	}

	if m, ok := m.(textInputModel); ok {

		*dest = m.textInput.Value()

		return nil
	}

	return fmt.Errorf("failed to get input")
}
