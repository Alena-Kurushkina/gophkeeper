package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type model struct {
	inputs      []textinput.Model
	focused     int
	err         error
	successMsg  string
	showSuccess bool

	client pb.GophkeeperClient
}

type user struct {
	username string
	password string
}

func InitialModel(client pb.GophkeeperClient) model {
	var inputs = make([]textinput.Model, 2)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Логин"
	inputs[0].Focus()
	inputs[0].CharLimit = 20
	inputs[0].PromptStyle = focusedStyle
	inputs[0].TextStyle = focusedStyle

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Пароль"
	inputs[1].CharLimit = 30
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'

	return model{
		inputs:  inputs,
		focused: 0,
		err:     nil,
		client: client,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				// Регистрация при нажатии Enter на последнем поле
				u := user{
					username: m.inputs[0].Value(),
					password: m.inputs[1].Value(),
				}
				
				if err := validateUser(u); err != nil {
					m.err = err
					return m, nil
				}

				_, err:=m.client.Register(context.Background(),&pb.Credentials{
					Login: u.username,
					Password: u.password,
				})
				if err != nil {
					m.err = err
					return m, nil
				}

				// Сброс формы и показ сообщения об успехе
				m = InitialModel(m.client)
				m.showSuccess = true
				m.successMsg = "Пользователь успешно зарегистрирован!"
				return m, nil
			}
			m.nextInput()
			return m, tea.Batch(textinput.Blink)

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.prevInput()
			} else {
				m.nextInput()
			}
			return m, tea.Batch(textinput.Blink)

		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}

	// for i := range m.inputs {
	// 	m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	// }

	// Обновляем только текущее поле ввода, чтобы избежать мерцания
    m.inputs[m.focused], cmds[m.focused] = m.inputs[m.focused].Update(msg)

	return m, tea.Batch(cmds...)
}

func (m *model) nextInput() {
	m.inputs[m.focused].Blur()
    m.focused = (m.focused + 1) % len(m.inputs)
    m.inputs[m.focused].Focus()
}

func (m *model) prevInput() {
	m.inputs[m.focused].Blur()
    m.focused--
    if m.focused < 0 {
        m.focused = len(m.inputs) - 1
    }
    m.inputs[m.focused].Focus()
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("\n  Регистрация пользователя\n\n")

	for i := range m.inputs {
		b.WriteString("  ")
		if i == m.focused {
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
		} else {
			m.inputs[i].PromptStyle = blurredStyle
			m.inputs[i].TextStyle = blurredStyle
		}

		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("  Ошибка: " + m.err.Error()))
	}

	if m.showSuccess {
		b.WriteString("\n\n")
		b.WriteString(successStyle.Render("  " + m.successMsg))
	}

	b.WriteString("\n\n  ")
	instructions := blurredStyle.Render("Tab/↑/↓ - перемещение • Enter - подтвердить • Ctrl+C - выход")
	b.WriteString(instructions)
	b.WriteString("\n")

	return b.String()
}

func validateUser(u user) error {
	if len(u.username) < 3 {
		return fmt.Errorf("логин должен быть не менее 3 символов")
	}
	if len(u.password) < 6 {
		return fmt.Errorf("пароль должен быть не менее 6 символов")
	}
	return nil
}