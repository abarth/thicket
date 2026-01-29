package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/storage"
)

// viewState represents which view is currently active.
type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewCreate
	viewEdit
)

// Model is the root model for the TUI application.
type Model struct {
	// Current view state
	view viewState

	// Sub-models for each view
	list   ListModel
	detail DetailModel
	form   FormModel

	// Shared state
	store  *storage.Store
	config *config.Config
	keys   KeyMap
	width  int
	height int

	// Status message
	statusMsg   string
	statusError bool

	// Help visibility
	showHelp bool

	// File watching
	ticketsPath    string
	watcherChan    chan FileChangedMsg
	watcherCleanup func()
}

// New creates a new TUI model.
func New(store *storage.Store, cfg *config.Config, ticketsPath string) Model {
	// Set up file watcher with 100ms debounce
	watchChan, cleanup := WatchFile(ticketsPath, 100*time.Millisecond)()

	return Model{
		view:           viewList,
		list:           NewListModel(store),
		detail:         NewDetailModel(store),
		form:           NewFormModel(store, cfg.ProjectCode, nil),
		store:          store,
		config:         cfg,
		keys:           DefaultKeyMap(),
		ticketsPath:    ticketsPath,
		watcherChan:    watchChan,
		watcherCleanup: cleanup,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.list.Init(), m.waitForFileChange())
}

// waitForFileChange returns a command that waits for file change notifications.
func (m Model) waitForFileChange() tea.Cmd {
	return func() tea.Msg {
		_, ok := <-m.watcherChan
		if !ok {
			return nil // Channel closed, stop watching
		}
		return FileChangedMsg{}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate to sub-models
		m.list.SetSize(msg.Width, msg.Height-4) // Leave room for header and help
		m.detail.SetSize(msg.Width, msg.Height-4)
		m.form.SetSize(msg.Width, msg.Height-4)

	case tea.KeyMsg:
		// Clear status message on any key press
		m.statusMsg = ""
		m.statusError = false

		// Global key handling
		switch {
		case key.Matches(msg, m.keys.Quit):
			// Only quit from list view, or if Ctrl+C
			if (m.view == viewList && !m.list.IsSearching()) || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}

	// View transition messages
	case ViewTicketMsg:
		m.view = viewDetail
		m.detail.SetTicketID(msg.ID)
		return m, m.detail.LoadTicket()

	case BackToListMsg:
		m.view = viewList
		return m, m.list.Refresh()

	case CreateTicketMsg:
		m.view = viewCreate
		m.form = NewFormModel(m.store, m.config.ProjectCode, nil)
		m.form.SetSize(m.width, m.height-4)
		return m, m.form.Init()

	case EditTicketMsg:
		m.view = viewEdit
		m.form = NewFormModel(m.store, m.config.ProjectCode, msg.Ticket)
		m.form.SetSize(m.width, m.height-4)
		return m, m.form.Init()

	case TicketSavedMsg:
		m.view = viewList
		m.statusMsg = msg.Message
		return m, m.list.Refresh()

	case TicketClosedMsg:
		m.view = viewList
		m.statusMsg = fmt.Sprintf("Closed ticket %s", msg.ID)
		return m, m.list.Refresh()

	case TicketPriorityUpdatedMsg:
		m.statusMsg = fmt.Sprintf("Priority set to %d for %s", msg.NewPriority, msg.ID)
		return m, m.list.Refresh()

	case TicketTypeUpdatedMsg:
		m.statusMsg = fmt.Sprintf("Type set to %s for %s", msg.NewType, msg.ID)
		return m, m.list.Refresh()

	case RefreshListMsg:
		return m, m.list.Refresh()

	case FileChangedMsg:
		// File changed externally, sync from JSONL and refresh
		if err := m.store.SyncFromJSONL(); err != nil {
			m.statusMsg = fmt.Sprintf("Sync error: %v", err)
			m.statusError = true
		}
		// Continue watching for more changes
		return m, tea.Batch(m.list.Refresh(), m.waitForFileChange())

	case StatusMsg:
		m.statusMsg = msg.Message
		m.statusError = msg.IsError
		return m, nil

	case ErrorMsg:
		m.statusMsg = msg.Err.Error()
		m.statusError = true
		return m, nil
	}

	// Delegate to active view
	var cmd tea.Cmd
	switch m.view {
	case viewList:
		m.list, cmd = m.list.Update(msg)
	case viewDetail:
		m.detail, cmd = m.detail.Update(msg)
	case viewCreate, viewEdit:
		m.form, cmd = m.form.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Main content
	var content string
	switch m.view {
	case viewList:
		content = m.list.View()
	case viewDetail:
		content = m.detail.View()
	case viewCreate, viewEdit:
		content = m.form.View()
	}
	b.WriteString(content)

	// Help bar
	b.WriteString("\n")
	b.WriteString(m.renderHelpBar())

	return b.String()
}

func (m Model) renderHeader() string {
	var title string
	switch m.view {
	case viewList:
		count := len(m.list.tickets)
		openCount := 0
		for _, t := range m.list.tickets {
			if t.Status == "open" {
				openCount++
			}
		}
		title = fmt.Sprintf("Thicket - %d tickets (%d open)", count, openCount)
		if m.list.hasFilters() {
			title += " " + filterChipStyle.Render("filtered")
		}
	case viewDetail:
		title = fmt.Sprintf("Thicket - %s", m.detail.ticketID)
	case viewCreate:
		title = "Thicket - New Ticket"
	case viewEdit:
		title = fmt.Sprintf("Thicket - Edit %s", m.form.ticketID)
	}

	headerLine := titleStyle.Render(title)

	// Add status message if present
	if m.statusMsg != "" {
		style := statusMsgStyle
		if m.statusError {
			style = errorMsgStyle
		}
		headerLine += "  " + style.Render(m.statusMsg)
	}

	separator := strings.Repeat("─", m.width)

	return headerLine + "\n" + subtitleStyle.Render(separator)
}

func (m Model) renderHelpBar() string {
	separator := subtitleStyle.Render(strings.Repeat("─", m.width))

	var help string
	switch m.view {
	case viewList:
		help = ListHelp()
	case viewDetail:
		help = DetailHelp()
	case viewCreate, viewEdit:
		help = FormHelp()
	}

	return separator + "\n" + help
}

// Run starts the TUI application.
func Run(store *storage.Store, cfg *config.Config, ticketsPath string) error {
	model := New(store, cfg, ticketsPath)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	// Clean up the file watcher
	if model.watcherCleanup != nil {
		model.watcherCleanup()
	}
	return err
}

// renderStatusLine returns a styled status line.
func renderStatusLine(width int, left, right string) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}
