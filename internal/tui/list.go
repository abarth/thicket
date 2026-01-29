package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// FilterState holds the current filter settings.
type FilterState struct {
	Status *ticket.Status
}

// ListModel handles the ticket list view.
type ListModel struct {
	store   *storage.Store
	tickets []*ticket.Ticket
	cursor  int
	offset  int // scroll offset
	width   int
	height  int
	keys    KeyMap
	filters FilterState
	loading bool
	err     error
}

// NewListModel creates a new list model.
func NewListModel(store *storage.Store) ListModel {
	return ListModel{
		store: store,
		keys:  DefaultKeyMap(),
	}
}

// Init initializes the list model.
func (m ListModel) Init() tea.Cmd {
	return m.loadTickets()
}

// SetSize sets the dimensions for the list view.
func (m *ListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// hasFilters returns true if any filters are active.
func (m ListModel) hasFilters() bool {
	return m.filters.Status != nil
}

// Refresh reloads the ticket list.
func (m ListModel) Refresh() tea.Cmd {
	return m.loadTickets()
}

func (m ListModel) loadTickets() tea.Cmd {
	return func() tea.Msg {
		tickets, err := m.store.List(m.filters.Status)
		if err != nil {
			return TicketsLoadedMsg{Err: err}
		}
		return TicketsLoadedMsg{Tickets: tickets}
	}
}

// Update handles messages for the list view.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case TicketsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.tickets = msg.Tickets
		m.err = nil
		// Reset cursor if out of bounds
		if m.cursor >= len(m.tickets) {
			m.cursor = len(m.tickets) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.tickets)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case key.Matches(msg, m.keys.Top):
			m.cursor = 0
			m.offset = 0
		case key.Matches(msg, m.keys.Bottom):
			m.cursor = len(m.tickets) - 1
			m.ensureVisible()
		case key.Matches(msg, m.keys.Enter):
			if len(m.tickets) > 0 && m.cursor < len(m.tickets) {
				return m, func() tea.Msg {
					return ViewTicketMsg{ID: m.tickets[m.cursor].ID}
				}
			}
		case key.Matches(msg, m.keys.New):
			return m, func() tea.Msg {
				return CreateTicketMsg{}
			}
		case key.Matches(msg, m.keys.Edit):
			if len(m.tickets) > 0 && m.cursor < len(m.tickets) {
				t := m.tickets[m.cursor]
				return m, func() tea.Msg {
					return EditTicketMsg{Ticket: t}
				}
			}
		case key.Matches(msg, m.keys.Close):
			if len(m.tickets) > 0 && m.cursor < len(m.tickets) {
				t := m.tickets[m.cursor]
				if t.Status == ticket.StatusOpen {
					return m, m.closeTicket(t.ID)
				}
			}
		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, m.loadTickets()
		case key.Matches(msg, m.keys.FilterOpen):
			status := ticket.StatusOpen
			m.filters.Status = &status
			m.cursor = 0
			m.offset = 0
			return m, m.loadTickets()
		case key.Matches(msg, m.keys.FilterClosed):
			status := ticket.StatusClosed
			m.filters.Status = &status
			m.cursor = 0
			m.offset = 0
			return m, m.loadTickets()
		case key.Matches(msg, m.keys.FilterAll):
			m.filters.Status = nil
			m.cursor = 0
			m.offset = 0
			return m, m.loadTickets()
		}
	}

	return m, nil
}

func (m ListModel) closeTicket(id string) tea.Cmd {
	return func() tea.Msg {
		t, err := m.store.Get(id)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		t.Status = ticket.StatusClosed
		if err := m.store.Update(t); err != nil {
			return ErrorMsg{Err: err}
		}
		return TicketClosedMsg{ID: id}
	}
}

func (m *ListModel) ensureVisible() {
	visibleRows := m.visibleRows()
	if visibleRows <= 0 {
		return
	}

	// Scroll up if cursor is above visible area
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// Scroll down if cursor is below visible area
	if m.cursor >= m.offset+visibleRows {
		m.offset = m.cursor - visibleRows + 1
	}
}

func (m ListModel) visibleRows() int {
	// Account for header row
	rows := m.height - 2
	if rows < 1 {
		rows = 1
	}
	return rows
}

// View renders the list view.
func (m ListModel) View() string {
	if m.err != nil {
		return errorMsgStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.loading {
		return "Loading tickets..."
	}

	if len(m.tickets) == 0 {
		msg := "No tickets found."
		if m.hasFilters() {
			msg += " Try clearing filters with 'a'."
		}
		return helpStyle.Render(msg)
	}

	var b strings.Builder

	// Table header
	header := m.renderRow("", "ID", "PRI", "TYPE", "STATUS", "ASSIGNEE", "TITLE", false)
	b.WriteString(tableHeaderStyle.Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleRows := m.visibleRows()
	start := m.offset
	end := start + visibleRows
	if end > len(m.tickets) {
		end = len(m.tickets)
	}

	// Render visible tickets
	for i := start; i < end; i++ {
		t := m.tickets[i]
		selected := i == m.cursor

		cursor := "  "
		if selected {
			cursor = "> "
		}

		row := m.renderRow(
			cursor,
			t.ID,
			fmt.Sprintf("%d", t.Priority),
			string(t.Type),
			string(t.Status),
			t.Assignee,
			t.Title,
			selected,
		)

		if selected {
			b.WriteString(selectedRowStyle.Render(row))
		} else {
			b.WriteString(normalRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(m.tickets) > visibleRows {
		scrollInfo := fmt.Sprintf(" (%d-%d of %d)", start+1, end, len(m.tickets))
		b.WriteString(helpStyle.Render(scrollInfo))
	}

	return b.String()
}

func (m ListModel) renderRow(cursor, id, pri, typ, status, assignee, title string, selected bool) string {
	// Truncate fields
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	if assignee == "" {
		assignee = "-"
	}
	if len(assignee) > 10 {
		assignee = assignee[:7] + "..."
	}
	if typ == "" {
		typ = "-"
	}

	// Format with fixed widths
	return fmt.Sprintf("%s%-10s %3s  %-8s %-6s %-10s %s",
		cursor, id, pri, typ, status, assignee, title)
}
