package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abarth/thicket/internal/storage"
	"github.com/abarth/thicket/internal/ticket"
)

// formField represents which field is currently focused.
type formField int

const (
	fieldTitle formField = iota
	fieldDescription
	fieldType
	fieldPriority
	fieldAssignee
	fieldLabels
	fieldCount
)

// validTypes are the allowed ticket types.
var validTypes = map[ticket.Type]bool{
	ticket.TypeBug:     true,
	ticket.TypeFeature: true,
	ticket.TypeTask:    true,
	ticket.TypeEpic:    true,
	ticket.TypeCleanup: true,
}

// FormModel handles the create/edit ticket form.
type FormModel struct {
	store       *storage.Store
	projectCode string
	ticketID    string
	isNew       bool
	width       int
	height      int
	keys        KeyMap
	focus       formField

	// Form inputs
	title       textinput.Model
	description textinput.Model
	ticketType  textinput.Model
	priority    textinput.Model
	assignee    textinput.Model
	labels      textinput.Model

	// Validation errors
	errors map[formField]string
}

// NewFormModel creates a new form model.
func NewFormModel(store *storage.Store, projectCode string, t *ticket.Ticket) FormModel {
	m := FormModel{
		store:       store,
		projectCode: projectCode,
		isNew:       t == nil,
		keys:        DefaultKeyMap(),
		errors:      make(map[formField]string),
	}

	// Initialize text inputs
	m.title = textinput.New()
	m.title.Placeholder = "Ticket title"
	m.title.CharLimit = 200
	m.title.Width = 50

	m.description = textinput.New()
	m.description.Placeholder = "Description (optional)"
	m.description.CharLimit = 1000
	m.description.Width = 50

	m.ticketType = textinput.New()
	m.ticketType.Placeholder = "bug, feature, task, epic, cleanup"
	m.ticketType.CharLimit = 20
	m.ticketType.Width = 30

	m.priority = textinput.New()
	m.priority.Placeholder = "1-5 (1=highest)"
	m.priority.CharLimit = 1
	m.priority.Width = 10

	m.assignee = textinput.New()
	m.assignee.Placeholder = "Assignee (optional)"
	m.assignee.CharLimit = 50
	m.assignee.Width = 30

	m.labels = textinput.New()
	m.labels.Placeholder = "Comma-separated labels (optional)"
	m.labels.CharLimit = 200
	m.labels.Width = 50

	// Pre-populate for editing
	if t != nil {
		m.ticketID = t.ID
		m.title.SetValue(t.Title)
		m.description.SetValue(t.Description)
		m.ticketType.SetValue(string(t.Type))
		m.priority.SetValue(strconv.Itoa(t.Priority))
		m.assignee.SetValue(t.Assignee)
		m.labels.SetValue(strings.Join(t.Labels, ", "))
	} else {
		// Defaults for new ticket
		m.priority.SetValue("2")
		m.ticketType.SetValue("task")
	}

	// Focus first field
	m.focusField(fieldTitle)

	return m
}

// Init initializes the form model.
func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize sets the dimensions for the form view.
func (m *FormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	inputWidth := width - 20
	if inputWidth > 60 {
		inputWidth = 60
	}
	m.title.Width = inputWidth
	m.description.Width = inputWidth
	m.labels.Width = inputWidth
}

func (m *FormModel) focusField(f formField) {
	// Blur all
	m.title.Blur()
	m.description.Blur()
	m.ticketType.Blur()
	m.priority.Blur()
	m.assignee.Blur()
	m.labels.Blur()

	m.focus = f

	// Focus the target
	switch f {
	case fieldTitle:
		m.title.Focus()
	case fieldDescription:
		m.description.Focus()
	case fieldType:
		m.ticketType.Focus()
	case fieldPriority:
		m.priority.Focus()
	case fieldAssignee:
		m.assignee.Focus()
	case fieldLabels:
		m.labels.Focus()
	}
}

func (m *FormModel) nextField() {
	next := m.focus + 1
	if next >= fieldCount {
		next = 0
	}
	m.focusField(next)
}

func (m *FormModel) prevField() {
	prev := m.focus - 1
	if prev < 0 {
		prev = fieldCount - 1
	}
	m.focusField(prev)
}

// Update handles messages for the form view.
func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg {
				return BackToListMsg{}
			}
		case key.Matches(msg, m.keys.Save):
			return m, m.save()
		case key.Matches(msg, m.keys.NextField):
			m.nextField()
			return m, nil
		case key.Matches(msg, m.keys.PrevField):
			m.prevField()
			return m, nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.focus {
	case fieldTitle:
		m.title, cmd = m.title.Update(msg)
	case fieldDescription:
		m.description, cmd = m.description.Update(msg)
	case fieldType:
		m.ticketType, cmd = m.ticketType.Update(msg)
	case fieldPriority:
		m.priority, cmd = m.priority.Update(msg)
	case fieldAssignee:
		m.assignee, cmd = m.assignee.Update(msg)
	case fieldLabels:
		m.labels, cmd = m.labels.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m FormModel) validate() map[formField]string {
	errors := make(map[formField]string)

	title := strings.TrimSpace(m.title.Value())
	if title == "" {
		errors[fieldTitle] = "Title is required"
	}

	pri := strings.TrimSpace(m.priority.Value())
	if pri != "" {
		p, err := strconv.Atoi(pri)
		if err != nil || p < 1 || p > 5 {
			errors[fieldPriority] = "Priority must be 1-5"
		}
	}

	typ := strings.TrimSpace(m.ticketType.Value())
	if typ != "" {
		if !validTypes[ticket.Type(typ)] {
			errors[fieldType] = "Invalid type"
		}
	}

	return errors
}

func (m FormModel) save() tea.Cmd {
	return func() tea.Msg {
		// Validate
		errors := m.validate()
		if len(errors) > 0 {
			// Return first error
			for _, e := range errors {
				return ErrorMsg{Err: fmt.Errorf("%s", e)}
			}
		}

		title := strings.TrimSpace(m.title.Value())
		description := strings.TrimSpace(m.description.Value())
		typ := strings.TrimSpace(m.ticketType.Value())
		pri := strings.TrimSpace(m.priority.Value())
		assignee := strings.TrimSpace(m.assignee.Value())
		labelsStr := strings.TrimSpace(m.labels.Value())

		priority := 2
		if pri != "" {
			priority, _ = strconv.Atoi(pri)
		}

		var labels []string
		if labelsStr != "" {
			for _, l := range strings.Split(labelsStr, ",") {
				l = strings.TrimSpace(l)
				if l != "" {
					labels = append(labels, l)
				}
			}
		}

		issueType := ticket.TypeTask
		if typ != "" {
			issueType = ticket.Type(typ)
		}

		if m.isNew {
			// Create new ticket
			t, err := ticket.New(m.projectCode, title, description, issueType, priority, labels, assignee)
			if err != nil {
				return ErrorMsg{Err: err}
			}

			if err := m.store.Add(t); err != nil {
				return ErrorMsg{Err: err}
			}

			return TicketSavedMsg{
				ID:      t.ID,
				IsNew:   true,
				Message: fmt.Sprintf("Created ticket %s", t.ID),
			}
		}

		// Update existing ticket
		t, err := m.store.Get(m.ticketID)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		t.Title = title
		t.Description = description
		t.Type = issueType
		t.Priority = priority
		t.Assignee = assignee
		t.Labels = labels

		if err := m.store.Update(t); err != nil {
			return ErrorMsg{Err: err}
		}

		return TicketSavedMsg{
			ID:      t.ID,
			IsNew:   false,
			Message: fmt.Sprintf("Updated ticket %s", t.ID),
		}
	}
}

// View renders the form view.
func (m FormModel) View() string {
	var b strings.Builder

	// Render each field
	b.WriteString(m.renderField("Title", m.title, fieldTitle))
	b.WriteString("\n")
	b.WriteString(m.renderField("Description", m.description, fieldDescription))
	b.WriteString("\n")
	b.WriteString(m.renderField("Type", m.ticketType, fieldType))
	b.WriteString("\n")
	b.WriteString(m.renderField("Priority", m.priority, fieldPriority))
	b.WriteString("\n")
	b.WriteString(m.renderField("Assignee", m.assignee, fieldAssignee))
	b.WriteString("\n")
	b.WriteString(m.renderField("Labels", m.labels, fieldLabels))

	return b.String()
}

func (m FormModel) renderField(label string, input textinput.Model, field formField) string {
	style := labelStyle.Copy().Width(14)
	labelText := style.Render(label + ":")

	inputView := input.View()

	// Show error if any
	if err, ok := m.errors[field]; ok {
		inputView += "  " + errorMsgStyle.Render(err)
	}

	return labelText + " " + inputView
}
