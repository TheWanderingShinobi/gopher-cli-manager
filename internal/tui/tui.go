package tui

import (
	"fmt"
	"strings"

	"github.com/TheWanderingShinobi/gopher-cli-manager/internal/database"
	"github.com/TheWanderingShinobi/gopher-cli-manager/pkg/models"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 0, 1, 2)
)

type model struct {
	db           *database.DB
	state        string
	list         list.Model
	inputs       []textinput.Model
	selectedCLI  models.Cli
	err          error
	confirmState string
}

func initialModel(db *database.DB) model {
	return model{
		db:    db,
		state: "menu",
		list:  list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "menu":
			return handleMenuInput(m, msg)
		case "list":
			return handleListInput(m, msg)
		case "add", "edit":
			return handleFormInput(m, msg)
		case "confirm":
			return handleConfirmInput(m, msg)
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case "menu":
		return appStyle.Render(m.viewMenu())
	case "list":
		return appStyle.Render(m.list.View())
	case "add", "edit":
		return appStyle.Render(m.viewForm())
	case "confirm":
		return appStyle.Render(m.viewConfirm())
	default:
		return "Error: Unknown state"
	}
}

func (m model) viewMenu() string {
	s := "MAIN MENU\n\n"
	items := []string{
		"v: View all CLIs",
		"s: Search for CLIs",
		"a: Add a new CLI",
		"q: Quit program",
	}
	hasRecords, _ := m.db.HasRecords()
	if hasRecords {
		items = append(items, "p: Purge database")
	}
	return s + strings.Join(items, "\n")
}

func (m model) viewForm() string {
	s := fmt.Sprintf("%s CLI\n\n", strings.ToUpper(m.state))
	for i := range m.inputs {
		s += m.inputs[i].View() + "\n"
	}
	s += "\nPress Enter to save, Esc to cancel"
	return s
}

func (m model) viewConfirm() string {
	return fmt.Sprintf("%s\n\nPress y to confirm, n to cancel", m.confirmState)
}

func handleMenuInput(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "v":
		return showCLIList(m, "")
	case "s":
		return initSearch(m)
	case "a":
		return initAddForm(m)
	case "p":
		hasRecords, _ := m.db.HasRecords()
		if hasRecords {
			m.state = "confirm"
			m.confirmState = "Are you sure you want to delete ALL the CLIs?"
			return m, nil
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func handleListInput(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		i, ok := m.list.SelectedItem().(cliItem)
		if ok {
			m.selectedCLI = i.cli
			return showCLIActions(m)
		}
	case "esc":
		return initialModel(m.db), nil
	}
	return m, nil
}

func handleFormInput(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.state == "add" {
			return saveCLI(m)
		} else if m.state == "edit" {
			return updateCLI(m)
		}
	case tea.KeyEsc:
		return initialModel(m.db), nil
	}

	var cmd tea.Cmd
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}
	return m, cmd
}

func handleConfirmInput(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if m.confirmState == "Are you sure you want to delete ALL the CLIs?" {
			m.db.DeleteAllRecords()
		} else {
			m.db.DeleteRecordById(m.selectedCLI.Id)
		}
		return initialModel(m.db), nil
	case "n":
		return initialModel(m.db), nil
	}
	return m, nil
}

func showCLIList(m model, search string) (tea.Model, tea.Cmd) {
	items := []list.Item{}
	clis, err := m.db.GetEntriesContainingText(search)
	if err != nil {
		m.err = err
		return m, nil
	}
	for _, cli := range clis {
		items = append(items, cliItem{cli: cli})
	}
	m.list.SetItems(items)
	m.state = "list"
	return m, nil
}

func initSearch(m model) (tea.Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "Enter search term"
	ti.Focus()
	m.inputs = []textinput.Model{ti}
	m.state = "search"
	return m, textinput.Blink
}

func initAddForm(m model) (tea.Model, tea.Cmd) {
	m.inputs = make([]textinput.Model, 3)
	for i := range m.inputs {
		t := textinput.New()
		switch i {
		case 0:
			t.Placeholder = "Name"
		case 1:
			t.Placeholder = "Description"
		case 2:
			t.Placeholder = "Path"
		}
		m.inputs[i] = t
	}
	m.inputs[0].Focus()
	m.state = "add"
	return m, textinput.Blink
}

func saveCLI(m model) (tea.Model, tea.Cmd) {
	cli := models.Cli{
		Name:        m.inputs[0].Value(),
		Description: m.inputs[1].Value(),
		Path:        m.inputs[2].Value(),
	}
	err := m.db.CreateCli(cli)
	if err != nil {
		m.err = err
		return m, nil
	}
	return initialModel(m.db), nil
}

func updateCLI(m model) (tea.Model, tea.Cmd) {
	m.selectedCLI.Name = m.inputs[0].Value()
	m.selectedCLI.Description = m.inputs[1].Value()
	m.selectedCLI.Path = m.inputs[2].Value()
	err := m.db.UpdateCli(m.selectedCLI)
	if err != nil {
		m.err = err
		return m, nil
	}
	return initialModel(m.db), nil
}

func showCLIActions(m model) (tea.Model, tea.Cmd) {
	items := []list.Item{
		cliActionItem{name: "Delete"},
		cliActionItem{name: "Edit"},
		cliActionItem{name: "Copy path to clipboard"},
		cliActionItem{name: "Back to menu"},
	}
	m.list.SetItems(items)
	return m, nil
}

type cliItem struct {
	cli models.Cli
}

func (i cliItem) Title() string       { return i.cli.Name }
func (i cliItem) Description() string { return i.cli.Description }
func (i cliItem) FilterValue() string { return i.cli.Name }

type cliActionItem struct {
	name string
}

func (i cliActionItem) Title() string       { return i.name }
func (i cliActionItem) Description() string { return "" }
func (i cliActionItem) FilterValue() string { return i.name }

func StartTea(db *database.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
