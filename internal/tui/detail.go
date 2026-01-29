package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// DetailModel handles the ticket detail view.
type DetailModel struct {
	store     *storage.Store
	ticketID  string
	ticket    *ticket.Ticket
	comments  []*ticket.Comment
	blockedBy []*ticket.Ticket
	blocking  []*ticket.Ticket
	width     int
	height    int
	scrollY   int
	keys      KeyMap
	loading   bool
	err       error

	// Comment input mode
	commenting    bool
	commentInput  textarea.Model
}

// NewDetailModel creates a new detail model.
func NewDetailModel(store *storage.Store) DetailModel {
	ti := textarea.New()
	ti.Placeholder = "Enter your comment..."
	ti.SetWidth(60)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false

	return DetailModel{
		store:        store,
		keys:         DefaultKeyMap(),
		commentInput: ti,
	}
}

// SetSize sets the dimensions for the detail view.
func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.commentInput.SetWidth(width - 4)
}

// SetTicketID sets the ticket to display.
func (m *DetailModel) SetTicketID(id string) {
	m.ticketID = id
	m.ticket = nil
	m.comments = nil
	m.blockedBy = nil
	m.blocking = nil
	m.scrollY = 0
	m.commenting = false
}

// LoadTicket loads the ticket details.
func (m DetailModel) LoadTicket() tea.Cmd {
	return func() tea.Msg {
		t, err := m.store.Get(m.ticketID)
		if err != nil {
			return TicketLoadedMsg{Err: err}
		}

		comments, _ := m.store.GetComments(m.ticketID)
		blockedBy, _ := m.store.GetBlockers(m.ticketID)
		blocking, _ := m.store.GetBlocking(m.ticketID)

		return TicketLoadedMsg{
			Ticket:    t,
			Comments:  comments,
			BlockedBy: blockedBy,
			Blocking:  blocking,
		}
	}
}

// Update handles messages for the detail view.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TicketLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.ticket = msg.Ticket
		m.comments = msg.Comments
		m.blockedBy = msg.BlockedBy
		m.blocking = msg.Blocking
		m.err = nil
		return m, nil

	case tea.KeyMsg:
		// If in commenting mode, handle differently
		if m.commenting {
			switch msg.String() {
			case "esc":
				m.commenting = false
				m.commentInput.Reset()
				return m, nil
			case "ctrl+s":
				content := strings.TrimSpace(m.commentInput.Value())
				if content != "" {
					m.commenting = false
					return m, m.saveComment(content)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.commentInput, cmd = m.commentInput.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg {
				return BackToListMsg{}
			}
		case key.Matches(msg, m.keys.Up):
			if m.scrollY > 0 {
				m.scrollY--
			}
		case key.Matches(msg, m.keys.Down):
			m.scrollY++
		case key.Matches(msg, m.keys.Edit):
			if m.ticket != nil {
				t := m.ticket
				return m, func() tea.Msg {
					return EditTicketMsg{Ticket: t}
				}
			}
		case key.Matches(msg, m.keys.Close):
			if m.ticket != nil && m.ticket.Status == ticket.StatusOpen {
				return m, m.closeTicket()
			}
		case key.Matches(msg, m.keys.Comment):
			if m.ticket != nil {
				m.commenting = true
				m.commentInput.Focus()
				return m, nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m DetailModel) closeTicket() tea.Cmd {
	return func() tea.Msg {
		t, err := m.store.Get(m.ticketID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		t.Status = ticket.StatusClosed
		if err := m.store.Update(t); err != nil {
			return ErrorMsg{Err: err}
		}
		return TicketClosedMsg{ID: m.ticketID}
	}
}

func (m DetailModel) saveComment(content string) tea.Cmd {
	return func() tea.Msg {
		comment, err := ticket.NewComment(m.ticketID, content)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		if err := m.store.AddComment(comment); err != nil {
			return ErrorMsg{Err: err}
		}
		return CommentSavedMsg{TicketID: m.ticketID}
	}
}

// View renders the detail view.
func (m DetailModel) View() string {
	if m.err != nil {
		return errorMsgStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.loading || m.ticket == nil {
		return "Loading ticket..."
	}

	var b strings.Builder
	t := m.ticket

	// Build all content lines
	var lines []string

	lines = append(lines, m.renderField("ID", t.ID))
	lines = append(lines, m.renderField("Title", t.Title))

	typ := string(t.Type)
	if typ == "" {
		typ = "-"
	}
	lines = append(lines, m.renderField("Type", typ))
	lines = append(lines, m.renderField("Status", string(t.Status)))
	lines = append(lines, m.renderField("Priority", fmt.Sprintf("%d", t.Priority)))

	assignee := t.Assignee
	if assignee == "" {
		assignee = "(unassigned)"
	}
	lines = append(lines, m.renderField("Assignee", assignee))

	labels := strings.Join(t.Labels, ", ")
	if labels == "" {
		labels = "(none)"
	}
	lines = append(lines, m.renderField("Labels", labels))

	lines = append(lines, m.renderField("Created", t.Created.Format(time.RFC3339)))
	lines = append(lines, m.renderField("Updated", t.Updated.Format(time.RFC3339)))

	// Blocked by
	if len(m.blockedBy) > 0 {
		lines = append(lines, "")
		lines = append(lines, subtitleStyle.Render("Blocked by:"))
		for _, blocker := range m.blockedBy {
			status := ""
			if blocker.Status == ticket.StatusClosed {
				status = " [closed]"
			}
			lines = append(lines, fmt.Sprintf("  - %s: %s%s", blocker.ID, blocker.Title, status))
		}
	}

	// Blocking
	if len(m.blocking) > 0 {
		lines = append(lines, "")
		lines = append(lines, subtitleStyle.Render("Blocking:"))
		for _, blocked := range m.blocking {
			lines = append(lines, fmt.Sprintf("  - %s: %s", blocked.ID, blocked.Title))
		}
	}

	// Description
	if t.Description != "" {
		lines = append(lines, "")
		lines = append(lines, subtitleStyle.Render("Description:"))
		// Wrap description lines
		for _, line := range strings.Split(t.Description, "\n") {
			lines = append(lines, "  "+line)
		}
	}

	// Comments
	if len(m.comments) > 0 {
		lines = append(lines, "")
		lines = append(lines, subtitleStyle.Render("Comments:"))
		for _, c := range m.comments {
			timestamp := c.Created.Format("2006-01-02 15:04")
			lines = append(lines, fmt.Sprintf("  [%s] %s", timestamp, c.Content))
		}
	}

	// Apply scroll offset
	visibleLines := m.height - 2
	if visibleLines < 1 {
		visibleLines = 10
	}

	// Clamp scroll
	maxScroll := len(lines) - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}

	start := m.scrollY
	end := start + visibleLines
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}

	// Show scroll indicator
	if len(lines) > visibleLines {
		scrollInfo := fmt.Sprintf(" (line %d-%d of %d)", start+1, end, len(lines))
		b.WriteString(helpStyle.Render(scrollInfo))
	}

	// Comment input
	if m.commenting {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Add comment (ctrl+s to save, esc to cancel):"))
		b.WriteString("\n")
		b.WriteString(m.commentInput.View())
	}

	return b.String()
}

func (m DetailModel) renderField(label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}
