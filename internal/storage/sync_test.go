package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abarth/thicket/internal/config"
	"github.com/abarth/thicket/internal/ticket"
)

func setupTestProject(t *testing.T) (config.Paths, func()) {
	t.Helper()
	dir := t.TempDir()

	// Create .thicket directory
	thicketDir := filepath.Join(dir, ".thicket")
	if err := os.MkdirAll(thicketDir, 0755); err != nil {
		t.Fatalf("Failed to create .thicket directory: %v", err)
	}

	// Create empty tickets file
	ticketsPath := filepath.Join(thicketDir, "tickets.jsonl")
	if err := os.WriteFile(ticketsPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create tickets file: %v", err)
	}

	paths := config.GetPaths(dir)
	return paths, func() {}
}

func TestStore_Open(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
}

func TestStore_AddAndGet(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil")
	}

	if got.ID != tk.ID {
		t.Errorf("Get().ID = %q, want %q", got.ID, tk.ID)
	}
	if got.Title != tk.Title {
		t.Errorf("Get().Title = %q, want %q", got.Title, tk.Title)
	}
}

func TestStore_Update(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Original title", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	newTitle := "Updated title"
	if err := tk.Update(&newTitle, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Fatalf("ticket.Update() error = %v", err)
	}

	if err := store.Update(tk); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Title != "Updated title" {
		t.Errorf("Get().Title = %q, want 'Updated title'", got.Title)
	}
}

func TestStore_List(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Add some tickets
	for i := 0; i < 3; i++ {
		tk, err := ticket.New("TH", "Ticket", "", ticket.TypeTask, i, nil, "")
		if err != nil {
			t.Fatalf("ticket.New() error = %v", err)
		}
		if err := store.Add(tk); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	// List all
	all, err := store.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 3 {
		t.Errorf("List() returned %d tickets, want 3", len(all))
	}

	// List open only
	open := ticket.StatusOpen
	openTickets, err := store.List(&open)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(openTickets) != 3 {
		t.Errorf("List(open) returned %d tickets, want 3", len(openTickets))
	}
}

func TestStore_SyncFromJSONL(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Pre-populate JSONL with tickets
	now := time.Now().UTC()
	tickets := []*ticket.Ticket{
		{ID: "TH-111111", Title: "First", Status: ticket.StatusOpen, Priority: 1, Created: now, Updated: now},
		{ID: "TH-222222", Title: "Second", Status: ticket.StatusOpen, Priority: 2, Created: now, Updated: now},
	}
	if err := WriteJSONL(paths.Tickets, tickets); err != nil {
		t.Fatalf("WriteJSONL() error = %v", err)
	}

	// Open store - should sync from JSONL
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Verify tickets were loaded
	all, err := store.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("List() returned %d tickets, want 2", len(all))
	}

	got, err := store.Get("TH-111111")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil || got.Title != "First" {
		t.Errorf("Get() = %+v, want ticket with Title=First", got)
	}
}

func TestStore_SyncOnReopen(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Create initial store and add a ticket
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tk, _ := ticket.New("TH", "Initial", "", ticket.TypeTask, 1, nil, "")
	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	store.Close()

	// Externally modify the JSONL file
	now := time.Now().UTC()
	externalTickets := []*ticket.Ticket{
		{ID: "TH-external", Title: "External", Status: ticket.StatusOpen, Priority: 0, Created: now, Updated: now},
	}
	if err := WriteJSONL(paths.Tickets, externalTickets); err != nil {
		t.Fatalf("WriteJSONL() error = %v", err)
	}

	// Reopen store - should sync from modified JSONL
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store2.Close()

	// Should only have the external ticket
	all, err := store2.List(nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("List() returned %d tickets, want 1", len(all))
	}
	if all[0].ID != "TH-external" {
		t.Errorf("List()[0].ID = %q, want TH-external", all[0].ID)
	}
}

func TestStore_AddAndGetComments(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Add a comment
	c, err := ticket.NewComment(tk.ID, "Test comment")
	if err != nil {
		t.Fatalf("ticket.NewComment() error = %v", err)
	}

	if err := store.AddComment(c); err != nil {
		t.Fatalf("AddComment() error = %v", err)
	}

	// Get comments for ticket
	comments, err := store.GetComments(tk.ID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("GetComments() returned %d comments, want 1", len(comments))
	}

	if comments[0].ID != c.ID {
		t.Errorf("comments[0].ID = %q, want %q", comments[0].ID, c.ID)
	}
	if comments[0].Content != "Test comment" {
		t.Errorf("comments[0].Content = %q, want 'Test comment'", comments[0].Content)
	}
}

func TestStore_CommentsForNonexistentTicket(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Get comments for nonexistent ticket should return empty slice
	comments, err := store.GetComments("TH-999999")
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("GetComments() returned %d comments, want 0", len(comments))
	}
}

func TestStore_MultipleComments(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, _ := ticket.New("TH", "Test ticket", "", ticket.TypeTask, 1, nil, "")
	store.Add(tk)

	// Add multiple comments
	for i := 0; i < 3; i++ {
		c, _ := ticket.NewComment(tk.ID, "Comment")
		if err := store.AddComment(c); err != nil {
			t.Fatalf("AddComment() error = %v", err)
		}
	}

	comments, err := store.GetComments(tk.ID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 3 {
		t.Errorf("GetComments() returned %d comments, want 3", len(comments))
	}
}

func TestStore_SyncCommentsOnReopen(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Create initial store and add a ticket with comment
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tk, _ := ticket.New("TH", "Test", "", ticket.TypeTask, 1, nil, "")
	store.Add(tk)

	c, _ := ticket.NewComment(tk.ID, "Original comment")
	store.AddComment(c)
	store.Close()

	// Reopen store - should sync comments from JSONL
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store2.Close()

	comments, err := store2.GetComments(tk.ID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("GetComments() returned %d comments, want 1", len(comments))
	}

	if comments[0].Content != "Original comment" {
		t.Errorf("comments[0].Content = %q, want 'Original comment'", comments[0].Content)
	}
}

func TestStore_AddAndGetDependencies(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Create two tickets
	tk1, _ := ticket.New("TH", "Blocker ticket", "", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Blocked ticket", "", ticket.TypeTask, 2, nil, "")
	store.Add(tk1)
	store.Add(tk2)

	// Add a blocked_by dependency
	dep, err := ticket.NewDependency(tk2.ID, tk1.ID, ticket.DependencyBlockedBy)
	if err != nil {
		t.Fatalf("ticket.NewDependency() error = %v", err)
	}

	if err := store.AddDependency(dep); err != nil {
		t.Fatalf("AddDependency() error = %v", err)
	}

	// Verify blockers
	blockers, err := store.GetBlockers(tk2.ID)
	if err != nil {
		t.Fatalf("GetBlockers() error = %v", err)
	}
	if len(blockers) != 1 {
		t.Fatalf("GetBlockers() returned %d, want 1", len(blockers))
	}
	if blockers[0].ID != tk1.ID {
		t.Errorf("blockers[0].ID = %q, want %q", blockers[0].ID, tk1.ID)
	}

	// Verify blocking
	blocking, err := store.GetBlocking(tk1.ID)
	if err != nil {
		t.Fatalf("GetBlocking() error = %v", err)
	}
	if len(blocking) != 1 {
		t.Fatalf("GetBlocking() returned %d, want 1", len(blocking))
	}
	if blocking[0].ID != tk2.ID {
		t.Errorf("blocking[0].ID = %q, want %q", blocking[0].ID, tk2.ID)
	}
}

func TestStore_CircularDependencyPrevention(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Create three tickets for testing transitive cycle
	tk1, _ := ticket.New("TH", "Ticket 1", "", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Ticket 2", "", ticket.TypeTask, 2, nil, "")
	tk3, _ := ticket.New("TH", "Ticket 3", "", ticket.TypeTask, 3, nil, "")
	store.Add(tk1)
	store.Add(tk2)
	store.Add(tk3)

	// Create chain: tk1 -> tk2 -> tk3 (tk1 blocked by tk2, tk2 blocked by tk3)
	dep1, _ := ticket.NewDependency(tk1.ID, tk2.ID, ticket.DependencyBlockedBy)
	dep2, _ := ticket.NewDependency(tk2.ID, tk3.ID, ticket.DependencyBlockedBy)
	store.AddDependency(dep1)
	store.AddDependency(dep2)

	// Try to create cycle: tk3 -> tk1 should fail
	dep3, _ := ticket.NewDependency(tk3.ID, tk1.ID, ticket.DependencyBlockedBy)
	err = store.AddDependency(dep3)
	if err != ticket.ErrCircularDependency {
		t.Errorf("AddDependency() error = %v, want ErrCircularDependency", err)
	}
}

func TestStore_DuplicateDependencyPrevention(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk1, _ := ticket.New("TH", "Ticket 1", "", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Ticket 2", "", ticket.TypeTask, 2, nil, "")
	store.Add(tk1)
	store.Add(tk2)

	dep, _ := ticket.NewDependency(tk1.ID, tk2.ID, ticket.DependencyBlockedBy)
	store.AddDependency(dep)

	// Try to add same dependency again
	dep2, _ := ticket.NewDependency(tk1.ID, tk2.ID, ticket.DependencyBlockedBy)
	err = store.AddDependency(dep2)
	if err != ticket.ErrDuplicateDependency {
		t.Errorf("AddDependency() error = %v, want ErrDuplicateDependency", err)
	}
}

func TestStore_CreatedFromDependency(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	parent, _ := ticket.New("TH", "Parent ticket", "", ticket.TypeTask, 1, nil, "")
	child, _ := ticket.New("TH", "Child ticket", "", ticket.TypeTask, 2, nil, "")
	store.Add(parent)
	store.Add(child)

	dep, _ := ticket.NewDependency(child.ID, parent.ID, ticket.DependencyCreatedFrom)
	store.AddDependency(dep)

	got, err := store.GetCreatedFrom(child.ID)
	if err != nil {
		t.Fatalf("GetCreatedFrom() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetCreatedFrom() returned nil")
	}
	if got.ID != parent.ID {
		t.Errorf("GetCreatedFrom().ID = %q, want %q", got.ID, parent.ID)
	}
}

func TestStore_IsBlocked(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	blocker, _ := ticket.New("TH", "Blocker", "", ticket.TypeTask, 1, nil, "")
	blocked, _ := ticket.New("TH", "Blocked", "", ticket.TypeTask, 2, nil, "")
	store.Add(blocker)
	store.Add(blocked)

	// Not blocked yet
	isBlocked, _ := store.IsBlocked(blocked.ID)
	if isBlocked {
		t.Error("IsBlocked() = true, want false (no blockers)")
	}

	// Add blocker
	dep, _ := ticket.NewDependency(blocked.ID, blocker.ID, ticket.DependencyBlockedBy)
	store.AddDependency(dep)

	// Now blocked
	isBlocked, _ = store.IsBlocked(blocked.ID)
	if !isBlocked {
		t.Error("IsBlocked() = false, want true (has open blocker)")
	}

	// Close the blocker
	blocker.Close()
	store.Update(blocker)

	// No longer blocked
	isBlocked, _ = store.IsBlocked(blocked.ID)
	if isBlocked {
		t.Error("IsBlocked() = true, want false (blocker is closed)")
	}
}

func TestStore_SyncDependenciesOnReopen(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tk1, _ := ticket.New("TH", "Ticket 1", "", ticket.TypeTask, 1, nil, "")
	tk2, _ := ticket.New("TH", "Ticket 2", "", ticket.TypeTask, 2, nil, "")
	store.Add(tk1)
	store.Add(tk2)

	dep, _ := ticket.NewDependency(tk2.ID, tk1.ID, ticket.DependencyBlockedBy)
	store.AddDependency(dep)
	store.Close()

	// Reopen store - should sync dependencies from JSONL
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store2.Close()

	blockers, err := store2.GetBlockers(tk2.ID)
	if err != nil {
		t.Fatalf("GetBlockers() error = %v", err)
	}
	if len(blockers) != 1 {
		t.Fatalf("GetBlockers() returned %d, want 1", len(blockers))
	}
}

func TestStore_UpdatePreservesCommentsAndDependencies(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// 1. Add a ticket
	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 2. Add a comment
	comment := &ticket.Comment{
		ID:       "TH-c123456",
		TicketID: tk.ID,
		Content:  "Test comment",
	}
	if err := store.AddComment(comment); err != nil {
		t.Fatalf("AddComment() error = %v", err)
	}

	// 3. Add a dependency
	tk2, err := ticket.New("TH", "Blocked ticket", "Description", ticket.TypeTask, 1, nil, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}
	if err := store.Add(tk2); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	dep := &ticket.Dependency{
		ID:           "TH-d123456",
		FromTicketID: tk2.ID,
		ToTicketID:   tk.ID,
		Type:         ticket.DependencyBlockedBy,
	}
	if err := store.AddDependency(dep); err != nil {
		t.Fatalf("AddDependency() error = %v", err)
	}

	// 4. Update the ticket
	newTitle := "Updated title"
	if err := tk.Update(&newTitle, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Fatalf("tk.Update() error = %v", err)
	}
	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	store.Close()

	// MUST DELETE CACHE TO FORCE RE-SYNC FROM JSONL
	if err := os.Remove(paths.Cache); err != nil {
		t.Fatalf("Failed to remove cache: %v", err)
	}

	// 5. Open a new store (re-sync from JSONL)
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() (again) error = %v", err)
	}
	defer store2.Close()

	// 6. Check if comment is still there
	comments, err := store2.GetComments(tk.ID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(comments) == 0 {
		t.Error("Comment was lost after update")
	} else {
		if comments[0].Content != "Test comment" {
			t.Errorf("Comment content lost. Got %q, want %q", comments[0].Content, "Test comment")
		}
		if comments[0].TicketID != tk.ID {
			t.Errorf("Comment TicketID lost. Got %q, want %q", comments[0].TicketID, tk.ID)
		}
	}

	// 7. Check if dependency is still there
	deps, err := store2.db.GetDependenciesFrom(tk2.ID)
	if err != nil {
		t.Fatalf("GetDependenciesFrom() error = %v", err)
	}
	if len(deps) == 0 {
		t.Error("Dependency was lost after update")
	} else {
		if deps[0].ToTicketID != tk.ID {
			t.Errorf("Dependency ToTicketID lost. Got %q, want %q", deps[0].ToTicketID, tk.ID)
		}
		if deps[0].Type != ticket.DependencyBlockedBy {
			t.Errorf("Dependency type lost. Got %q, want %q", deps[0].Type, ticket.DependencyBlockedBy)
		}
	}
}

func TestStore_AddWithLabels(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, []string{"bug", "urgent"}, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil")
	}

	if len(got.Labels) != 2 {
		t.Errorf("Get().Labels length = %d, want 2", len(got.Labels))
	}
}

func TestStore_UpdateLabels(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	tk, err := ticket.New("TH", "Test ticket", "Description", ticket.TypeTask, 1, []string{"initial"}, "")
	if err != nil {
		t.Fatalf("ticket.New() error = %v", err)
	}

	if err := store.Add(tk); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Add a new label
	if err := tk.Update(nil, nil, nil, nil, nil, []string{"added"}, nil, nil); err != nil {
		t.Fatalf("ticket.Update() error = %v", err)
	}

	if err := store.Update(tk); err != nil {
		t.Fatalf("store.Update() error = %v", err)
	}

	got, err := store.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(got.Labels) != 2 {
		t.Errorf("Get().Labels length = %d, want 2", len(got.Labels))
	}
}

func TestStore_ListByLabel(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Create tickets with different labels
	tk1, _ := ticket.New("TH", "Bug ticket", "", ticket.TypeTask, 1, []string{"bug"}, "")
	tk2, _ := ticket.New("TH", "Feature ticket", "", ticket.TypeTask, 2, []string{"feature"}, "")
	tk3, _ := ticket.New("TH", "Bug and feature", "", ticket.TypeTask, 3, []string{"bug", "feature"}, "")
	store.Add(tk1)
	store.Add(tk2)
	store.Add(tk3)

	// List by "bug" label
	bugs, err := store.ListByLabel("bug", nil)
	if err != nil {
		t.Fatalf("ListByLabel() error = %v", err)
	}
	if len(bugs) != 2 {
		t.Errorf("ListByLabel(bug) returned %d tickets, want 2", len(bugs))
	}

	// List by "feature" label
	features, err := store.ListByLabel("feature", nil)
	if err != nil {
		t.Fatalf("ListByLabel() error = %v", err)
	}
	if len(features) != 2 {
		t.Errorf("ListByLabel(feature) returned %d tickets, want 2", len(features))
	}

	// List by "nonexistent" label
	none, err := store.ListByLabel("nonexistent", nil)
	if err != nil {
		t.Fatalf("ListByLabel() error = %v", err)
	}
	if len(none) != 0 {
		t.Errorf("ListByLabel(nonexistent) returned %d tickets, want 0", len(none))
	}
}

func TestStore_LabelsPreservedOnReopen(t *testing.T) {
	paths, cleanup := setupTestProject(t)
	defer cleanup()

	// Create store and add ticket with labels
	store, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tk, _ := ticket.New("TH", "Test", "", ticket.TypeTask, 1, []string{"label1", "label2"}, "")
	store.Add(tk)
	store.Close()

	// Delete the cache to force re-sync from JSONL
	if err := os.Remove(paths.Cache); err != nil {
		t.Fatalf("Failed to remove cache: %v", err)
	}

	// Reopen store
	store2, err := Open(paths)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store2.Close()

	got, err := store2.Get(tk.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(got.Labels) != 2 {
		t.Errorf("Labels not preserved after reopen. Got %d labels, want 2", len(got.Labels))
	}
}
