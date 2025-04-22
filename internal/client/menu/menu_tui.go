package menu

// import (
// 	"fmt"
// 	"os"

// 	"github.com/charmbracelet/bubbles/list"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// )

// type item struct {
// 	title       string
// 	description string
// }

// func (i item) Title() string       { return i.title }
// func (i item) Description() string { return i.description }
// func (i item) FilterValue() string { return i.title }

// type model struct {
// 	list     list.Model
// 	choice   string
// 	quitting bool
// }

// func initialModel() model {
// 	items := []list.Item{
// 		item{title: "Регистрация", description: "Регистрация нового пользователя"},
// 		item{title: "Вход", description: "Вход в систему для доступа к функционалу"},
// 		item{title: "Добавить данные логин/пароль", description: "Записать в хранилище данные"},
// 	}

// 	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
// 	l.Title = "Главное меню"
// 	l.SetShowStatusBar(false)
// 	l.SetFilteringEnabled(false)
// 	l.Styles.Title = titleStyle
// 	l.Styles.PaginationStyle = paginationStyle
// 	l.Styles.HelpStyle = helpStyle

// 	return model{list: l}
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		m.list.SetWidth(msg.Width)
// 		return m, nil

// 	case tea.KeyMsg:
// 		switch keypress := msg.String(); keypress {
// 		case "ctrl+c", "q":
// 			m.quitting = true
// 			return m, tea.Quit

// 		case "enter":
// 			i, ok := m.list.SelectedItem().(item)
// 			if ok {
// 				m.choice = string(i.title)
// 			}
// 			return m, tea.Quit
// 		}
// 	}

// 	var cmd tea.Cmd
// 	m.list, cmd = m.list.Update(msg)
// 	return m, cmd
// }

// func (m model) View() string {
// 	if m.choice != "" {
// 		return quitMessageStyle.Render(fmt.Sprintf("Вы выбрали: %s", m.choice))
// 	}
// 	if m.quitting {
// 		return quitMessageStyle.Render("До свидания!")
// 	}
// 	return "\n" + m.list.View()
// }

// // Стили
// var (
// 	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
// 	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
// 	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
// 	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
// 	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
// 	quitMessageStyle  = lipgloss.NewStyle().Margin(1, 0, 2, 4)
// )

// func RunMainMenu() {
// 	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

// 	if _, err := p.Run(); err != nil {
// 		fmt.Printf("Ошибка: %v", err)
// 		os.Exit(1)
// 	}
// }


import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Состояния приложения
type state int

const (
	menuState state = iota
	formState
	infoState
)

// Главная модель
type model struct {
	state     state
	menu      menuModel
	form      formModel
	info      infoModel
	quitting  bool
}

// Модель меню
type menuModel struct {
	list   list.Model
	choice string
}

// Модель формы
type formModel struct {
	textInput textinput.Model
	value     string
}

// Модель информационного экрана
type infoModel struct {
	content string
}

func initialModel() model {
	return model{
		state:    menuState,
		menu:     newMenuModel(),
		form:     newFormModel(),
		info:     newInfoModel(),
		quitting: false,
	}
}

func newMenuModel() menuModel {
	items := []list.Item{
		item{title: "Форма ввода", description: "Перейти к форме ввода текста"},
		item{title: "Информация", description: "Показать информационный экран"},
		item{title: "Выход", description: "Завершить работу программы"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Главное меню"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle

	return menuModel{list: l}
}

func newFormModel() formModel {
	ti := textinput.New()
	ti.Placeholder = "Введите текст..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return formModel{
		textInput: ti,
	}
}

func newInfoModel() infoModel {
	return infoModel{
		content: "Добро пожаловать в информационный раздел!\n\nЗдесь может быть любая информация.\n\nНажмите 'q' для возврата в меню.",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.state {
	case menuState:
		return m.updateMenu(msg)
	case formState:
		return m.updateForm(msg)
	case infoState:
		return m.updateInfo(msg)
	}

	return m, cmd
}

func (m *model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.menu.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Обработка выбора в меню
	var cmd tea.Cmd
	m.menu.list, cmd = m.menu.list.Update(msg)

	// Проверяем выбор элемента
	if i, ok := m.menu.list.SelectedItem().(item); ok {
		switch i.title {
		case "Форма ввода":
			m.state = formState
			m.form = newFormModel() // Сбрасываем форму при каждом входе
			return m, textinput.Blink
		case "Информация":
			m.state = infoState
			return m, nil
		case "Выход":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m *model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = menuState
			return m, nil
		case "enter":
			m.form.value = m.form.textInput.Value()
			// Можно что-то сделать с введенным значением
			m.state = menuState
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.form.textInput, cmd = m.form.textInput.Update(msg)
	return m, cmd
}

func (m *model) updateInfo(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			m.state = menuState
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return quitMessageStyle.Render("До свидания!")
	}

	switch m.state {
	case menuState:
		return "\n" + m.menu.list.View()
	case formState:
		return formStyle.Render(
			fmt.Sprintf(
				"Форма ввода\n\n%s\n\n(Enter - подтвердить, Esc - отмена)",
				m.form.textInput.View(),
			),
		)
	case infoState:
		return infoStyle.Render(m.info.content)
	default:
		return ""
	}
}

// Стили
var (
	titleStyle     = lipgloss.NewStyle().MarginLeft(2)
	formStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	infoStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4).Border(lipgloss.RoundedBorder())
	quitMessageStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

func RunMainMenu() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Ошибка: %v", err)
		os.Exit(1)
	}
}