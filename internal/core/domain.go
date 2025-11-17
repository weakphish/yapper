package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// VaultIndexManager coordinates the vault, parser, and index backend.
type VaultIndexManager struct {
	Vault  Vault
	Index  IndexStore
	parser NoteParser
}

// NewVaultIndexManager wires together the vault, index, and parser.
func NewVaultIndexManager(v Vault, idx IndexStore, parser NoteParser) *VaultIndexManager {
	return &VaultIndexManager{
		Vault:  v,
		Index:  idx,
		parser: parser,
	}
}

// FullReindex parses every note under the vault root.
func (m *VaultIndexManager) FullReindex() error {
	paths, err := m.Vault.ListNotePaths()
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := m.reindexSingle(path); err != nil {
			return err
		}
	}
	return nil
}

// ReindexNotePath reparses a single note path from disk.
func (m *VaultIndexManager) ReindexNotePath(path string) error {
	return m.reindexSingle(path)
}

func (m *VaultIndexManager) reindexSingle(path string) error {
	note, err := m.Vault.ReadNote(path)
	if err != nil {
		return err
	}
	parsed := m.parser.Parse(note)
	return m.Index.UpsertParsedNote(parsed)
}

// Domain exposes higher-level queries/actions on top of the index.
type Domain struct {
	IndexMgr *VaultIndexManager
}

// NewDomain constructs the domain entry point.
func NewDomain(mgr *VaultIndexManager) *Domain {
	return &Domain{IndexMgr: mgr}
}

// ReindexAll forces a full vault scan.
func (d *Domain) ReindexAll() error {
	return d.IndexMgr.FullReindex()
}

// ListTasks returns tasks matching the provided filter.
func (d *Domain) ListTasks(filter *TaskFilter) []Task {
	return d.IndexMgr.Index.ListTasks(filter)
}

// TaskDetail returns a task plus any mentions if known.
func (d *Domain) TaskDetail(id TaskID) (Task, []TaskMention, bool) {
	task, ok := d.IndexMgr.Index.GetTask(id)
	if !ok {
		return Task{}, nil, false
	}
	mentions := d.IndexMgr.Index.GetMentionsForTask(id)
	return task, mentions, true
}

// LogEntriesForTask returns log entries referencing the task.
func (d *Domain) LogEntriesForTask(id TaskID) []LogEntry {
	return d.IndexMgr.Index.GetLogEntriesForTask(id)
}

// ItemsForTag returns all tasks/log entries referencing the tag.
func (d *Domain) ItemsForTag(tag string) TagResult {
	return d.IndexMgr.Index.ItemsForTag(tag)
}

// NotesInRange lists note metadata whose dates fall inside the range.
func (d *Domain) NotesInRange(r *DateRange) []NoteMeta {
	return d.IndexMgr.Index.ListNotesByDate(r)
}

// ListTags returns the global tag listing.
func (d *Domain) ListTags() []string {
	return d.IndexMgr.Index.ListTags()
}

// ReadNote returns a fully materialized note.
func (d *Domain) ReadNote(id NoteID) (Note, bool) {
	return d.IndexMgr.Index.GetNote(id)
}

// WriteNote persists the markdown to disk and reindexes the note.
func (d *Domain) WriteNote(id NoteID, content string) (Note, error) {
	note, ok := d.IndexMgr.Index.GetNote(id)
	if !ok {
		return Note{}, fmt.Errorf("note %s not indexed", id)
	}
	if err := os.WriteFile(note.Path, []byte(content), 0o644); err != nil {
		return Note{}, fmt.Errorf("write note %s: %w", note.Path, err)
	}
	if err := d.IndexMgr.ReindexNotePath(note.Path); err != nil {
		return Note{}, err
	}
	if updated, ok := d.IndexMgr.Index.GetNote(id); ok {
		return updated, nil
	}
	return Note{}, fmt.Errorf("note %s unavailable after write", id)
}

// OpenDaily loads (or creates) the note for a given date.
func (d *Domain) OpenDaily(date Date) (Note, error) {
	rangeSel := DateRange{Start: date, End: date}
	for _, meta := range d.IndexMgr.Index.ListNotesByDate(&rangeSel) {
		if meta.Date != nil && !meta.Date.Time.IsZero() && meta.Date.Time.Equal(date.Time) {
			if note, ok := d.IndexMgr.Index.GetNote(meta.ID); ok {
				return note, nil
			}
		}
	}

	fileName := date.String()
	if fileName == "" {
		fileName = time.Now().Format("2006-01-02")
	}
	fileName += ".md"
	path := filepath.Join(d.IndexMgr.Vault.RootPath(), fileName)
	if err := ensureDailyTemplate(path, date); err != nil {
		return Note{}, err
	}
	if err := d.IndexMgr.ReindexNotePath(path); err != nil {
		return Note{}, err
	}
	if note, ok := d.IndexMgr.Index.GetNote(NoteID(path)); ok {
		return note, nil
	}
	return Note{}, fmt.Errorf("note %s unavailable after creation", path)
}

func ensureDailyTemplate(path string, date Date) error {
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		title := date.String()
		if title == "" {
			title = time.Now().Format("2006-01-02")
		}
		template := fmt.Sprintf("# %s\n\n## Tasks\n\n## Log\n", title)
		return os.WriteFile(path, []byte(template), 0o644)
	}
	return nil
}

// WeeklySummary aggregates activity across a date range.
func (d *Domain) WeeklySummary(r *DateRange) WeeklySummary {
	allTasks := d.IndexMgr.Index.ListTasks(&TaskFilter{})
	var newTasks []Task
	var completedTasks []Task
	start := r.Start.Time
	end := r.End.Time
	for _, task := range allTasks {
		if !task.CreatedAt.Before(start) && !task.CreatedAt.After(end) {
			newTasks = append(newTasks, task)
		}
		if task.ClosedAt != nil {
			closed := *task.ClosedAt
			if !closed.Before(start) && !closed.After(end) {
				completedTasks = append(completedTasks, task)
			}
		}
	}

	notes := d.IndexMgr.Index.ListNotesByDate(r)
	tagCounts := make(map[string]int)
	for _, task := range allTasks {
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
	}
	var topTags []TagCount
	for tag, count := range tagCounts {
		topTags = append(topTags, TagCount{Tag: tag, Count: count})
	}
	sort.Slice(topTags, func(i, j int) bool {
		return topTags[i].Count > topTags[j].Count
	})
	if len(topTags) > 10 {
		topTags = topTags[:10]
	}

	return WeeklySummary{
		NewTasks:       newTasks,
		CompletedTasks: completedTasks,
		Notes:          notes,
		TopTags:        topTags,
	}
}
