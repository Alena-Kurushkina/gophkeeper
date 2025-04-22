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

type authState int

const (
	stateLogin authState = iota
	stateRegister
)

type model struct {
	state        authState
	inputs      []textinput.Model
	focused     int
	err         error
	successMsg  string
	showSuccess bool

	client *GophKeeperClient
}

type user struct {
	username string
	password string
}

func InitialModel(client *GophKeeperClient) model {
	return model{
		state:   stateLogin,
		inputs:  createInputs(stateLogin),
		focused: 0,
		err:     nil,
		client: client,
	}
}

func createInputs(state authState) []textinput.Model {
	var inputs []textinput.Model
	
	switch state {
	case stateLogin:
		inputs = make([]textinput.Model, 2)
		inputs[0] = newInput("Логин")
		inputs[1] = newPasswordInput("Пароль")
	 
	case stateRegister:
		inputs = make([]textinput.Model, 2)
		inputs[0] = newInput("Логин")
		inputs[1] = newPasswordInput("Пароль")
	}
	
	inputs[0].Focus()
	return inputs
}

func newInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 30
	ti.PromptStyle = focusedStyle
	ti.TextStyle = focusedStyle
	return ti
   }
   
   func newPasswordInput(placeholder string) textinput.Model {
	ti := newInput(placeholder)
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	return ti
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
				switch m.state {
				case stateLogin:
					m.authenticate()
					return m, tea.Quit
				 
				case stateRegister:
					m.register()
					m.state = stateLogin
					m.inputs = createInputs(stateLogin)
					m.successMsg = "Регистрация успешна! Войдите в систему"
					return m, nil
				}
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
		case tea.KeyCtrlS:
			if m.state == stateLogin {
				m.state = stateRegister
			} else {
				m.state = stateLogin
			}
			m.inputs = createInputs(m.state)
			m.focused = 0
			return m, nil
		}
	case error:
		m.err = msg
		return m, nil
	}

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

	title := "\n Вход в систему\n\n"
 	if m.state == stateRegister {
  		title = "\n Регистрация\n\n"
 	}

	b.WriteString(title)

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
	instructions := blurredStyle.Render("Tab/↑/↓ - перемещение • Enter - подтвердить • Ctrl+S - переключить режим • Ctrl+C - выход")
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

func (m *model) authenticate() {
	u := user{
		username: m.inputs[0].Value(),
		password: m.inputs[1].Value(),
	}
	
	if err := validateUser(u); err != nil {
		m.err = err
	}

	tokenResp, err:=m.client.Client.Login(context.Background(),&pb.Credentials{
		Login: u.username,
		Password: u.password,
	})
	if err != nil {
		m.err = err
	}

	m.client.Interceptor.Token=tokenResp.Token

	_, err=m.client.Client.SaveCredentials(context.Background(), &pb.SaveCredsRequest{Marking: "marking1", Creds:&pb.Credentials{Login: "some_login", Password: "34fdg65"}, Encrpassword: "12345", Metainfo: "meta information"})
	if err != nil {
		m.err = err
	}

	resp, err:=m.client.Client.GetUserAllCredentials(context.Background(), &pb.GetCredsRequest{Password: "12345"})
	if err != nil {
		m.err = err
	}
	_=resp
	
	res, err:=m.client.UploadFile(context.Background(), "/Users/alena/GoLang/gophkeeper/test_upload.log", "file1", "metainfo")
	if err != nil {
		m.err = err
	}
	_=res

	down, err:=m.client.DownloadFile(context.Background(), "file1", "/Users/alena/GoLang/gophkeeper/test_downlosd.log")
	if err != nil {
		m.err = err
	}
	_=down
}

func (m *model) register() {
	// Регистрация при нажатии Enter на последнем поле
	u := user{
		username: m.inputs[0].Value(),
		password: m.inputs[1].Value(),
	}
	
	if err := validateUser(u); err != nil {
		m.err = err
	}

	_, err:=m.client.Client.Register(context.Background(),&pb.Credentials{
		Login: u.username,
		Password: u.password,
	})
	if err != nil {
		m.err = err
	}

	// // Сброс формы и показ сообщения об успехе
	// m = InitialModel(m.client)
	// m.showSuccess = true
	// m.successMsg = "Пользователь успешно зарегистрирован!"
	// return nil
}