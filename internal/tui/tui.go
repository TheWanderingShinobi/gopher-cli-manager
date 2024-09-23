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
	db             *database.DB
	state          string
	list           list.Model
	inputs         []textinput.Model
	selectedCLI    models.Cli
	err            error
	confirmState   string
	successMessage string
}

func initialModel(db *database.DB) model {
	return model{
		db:             db,
		state:          "menu",
		list:           list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		err:            nil,
		successMessage: "",
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
		case "search":
			return handleSearchInput(m, msg)
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil
	case errMsg:
		m.err = fmt.Errorf(string(msg))
		return m, nil
	case successMsg:
		m.successMessage = string(msg)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var content string

	switch m.state {
	case "menu":
		content = m.viewMenu()
	case "list":
		content = m.list.View()
	case "add", "edit":
		content = m.viewForm()
	case "confirm":
		content = m.viewConfirm()
	case "search":
		content = m.viewSearch()
	default:
		content = "Error: Unknown state"
	}

	if m.err != nil {
		content += "\n\nError: " + m.err.Error()
	}

	if m.successMessage != "" {
		content += "\n\nSuccess: " + m.successMessage
	}

	return appStyle.Render(content)
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

func (m model) viewSearch() string {
	return fmt.Sprintf("Search CLIs\n\n%s\n\nPress Enter to search, Esc to cancel", m.inputs[0].View())
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
	current := -1
	for i, input := range m.inputs {
		if input.Focused() {
			current = i
			break
		}
	}

	switch msg.Type {
	case tea.KeyEnter:
		if m.state == "add" {
			return saveCLI(m)
		} else if m.state == "edit" {
			return updateCLI(m)
		}
	case tea.KeyEsc:
		return initialModel(m.db), nil
	case tea.KeyUp, tea.KeyShiftTab:
		if current > 0 {
			m.inputs[current].Blur()
			m.inputs[current-1].Focus()
		}
	case tea.KeyDown, tea.KeyTab:
		if current < len(m.inputs)-1 {
			m.inputs[current].Blur()
			m.inputs[current+1].Focus()
		}
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
			err := m.db.DeleteAllRecords()
			if err != nil {
				return m, showErrorMsg(fmt.Sprintf("Failed to delete all records: %v", err))
			}
			return initialModel(m.db), showSuccessMsg("All CLIs deleted successfully")
		} else {
			err := m.db.DeleteRecordById(m.selectedCLI.Id)
			if err != nil {
				return m, showErrorMsg(fmt.Sprintf("Failed to delete CLI: %v", err))
			}
			return initialModel(m.db), showSuccessMsg(fmt.Sprintf("CLI '%s' deleted successfully", m.selectedCLI.Name))
		}
	case "n":
		return initialModel(m.db), nil
	}
	return m, nil
}

func handleSearchInput(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return showCLIList(m, m.inputs[0].Value())
	case tea.KeyEsc:
		return initialModel(m.db), nil
	}

	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func showCLIList(m model, search string) (tea.Model, tea.Cmd) {
	clis, err := m.db.GetEntriesContainingText(search)
	if err != nil {
		return m, showErrorMsg(fmt.Sprintf("Failed to fetch CLIs: %v", err))
	}

	items := make([]list.Item, len(clis))
	for i, cli := range clis {
		items[i] = cliItem{cli: cli}
	}

	m.list.SetItems(items)
	m.state = "list"

	// Debug logging
	fmt.Printf("Fetched %d CLIs\n", len(clis))
	for _, cli := range clis {
		fmt.Printf("CLI: %+v\n", cli)
	}

	if len(items) == 0 {
		return m, showErrorMsg("No CLIs found")
	}

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
			t.Focus()
		case 1:
			t.Placeholder = "Description"
		case 2:
			t.Placeholder = "Path"
		}
		m.inputs[i] = t
	}
	m.state = "add"
	return m, nil
}

func saveCLI(m model) (tea.Model, tea.Cmd) {
	cli := models.Cli{
		Name:        m.inputs[0].Value(),
		Description: m.inputs[1].Value(),
		Path:        m.inputs[2].Value(),
	}

	if cli.Name == "" || cli.Description == "" || cli.Path == "" {
		return m, showErrorMsg("All fields must be filled")
	}

	err := m.db.CreateCli(cli)
	if err != nil {
		return m, showErrorMsg(fmt.Sprintf("Failed to save CLI: %v", err))
	}

	return initialModel(m.db), showSuccessMsg(fmt.Sprintf("CLI '%s' saved successfully", cli.Name))
}

func updateCLI(m model) (tea.Model, tea.Cmd) {
	m.selectedCLI.Name = m.inputs[0].Value()
	m.selectedCLI.Description = m.inputs[1].Value()
	m.selectedCLI.Path = m.inputs[2].Value()

	if m.selectedCLI.Name == "" || m.selectedCLI.Description == "" || m.selectedCLI.Path == "" {
		return m, showErrorMsg("All fields must be filled")
	}

	err := m.db.UpdateCli(m.selectedCLI)
	if err != nil {
		return m, showErrorMsg(fmt.Sprintf("Failed to update CLI: %v", err))
	}

	return initialModel(m.db), showSuccessMsg(fmt.Sprintf("CLI '%s' updated successfully", m.selectedCLI.Name))
}

func showCLIActions(m model) (tea.Model, tea.Cmd) {
	items := []list.Item{
		cliActionItem{name: "Delete", action: deleteCLIAction},
		cliActionItem{name: "Edit", action: editCLIAction},
		cliActionItem{name: "Copy path to clipboard", action: copyCLIPathAction},
		cliActionItem{name: "Back to menu", action: backToMenuAction},
	}
	m.list.SetItems(items)
	m.state = "list"
	return m, nil
}

func deleteCLIAction(m model) (tea.Model, tea.Cmd) {
	m.state = "confirm"
	m.confirmState = fmt.Sprintf("Are you sure you want to delete the CLI '%s'?", m.selectedCLI.Name)
	return m, nil
}

func editCLIAction(m model) (tea.Model, tea.Cmd) {
	m.state = "edit"
	m.inputs = make([]textinput.Model, 3)
	m.inputs[0] = textinput.New()
	m.inputs[0].SetValue(m.selectedCLI.Name)
	m.inputs[0].Focus()
	m.inputs[1] = textinput.New()
	m.inputs[1].SetValue(m.selectedCLI.Description)
	m.inputs[2] = textinput.New()
	m.inputs[2].SetValue(m.selectedCLI.Path)
	return m, nil
}

func copyCLIPathAction(m model) (tea.Model, tea.Cmd) {
	err := clipboard.WriteAll(m.selectedCLI.Path)
	if err != nil {
		return m, showErrorMsg(fmt.Sprintf("Failed to copy path to clipboard: %v", err))
	}
	return initialModel(m.db), showSuccessMsg(fmt.Sprintf("Path for CLI '%s' copied to clipboard", m.selectedCLI.Name))
}

func backToMenuAction(m model) (tea.Model, tea.Cmd) {
	return initialModel(m.db), nil
}

type cliItem struct {
	cli models.Cli
}

func (i cliItem) Title() string       { return i.cli.Name }
func (i cliItem) Description() string { return i.cli.Description }
func (i cliItem) FilterValue() string { return i.cli.Name }

type cliActionItem struct {
	name   string
	action func(model) (tea.Model, tea.Cmd)
}

func (i cliActionItem) Title() string       { return i.name }
func (i cliActionItem) Description() string { return "" }
func (i cliActionItem) FilterValue() string { return i.name }

func showErrorMsg(msg string) tea.Cmd {
	return func() tea.Msg {
		return errMsg(msg)
	}
}

func showSuccessMsg(msg string) tea.Cmd {
	return func() tea.Msg {
		return successMsg(msg)
	}
}

type errMsg string
type successMsg string

func StartTea(db *database.DB) {
	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}

