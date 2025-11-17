package core

import (
	"sort"
	"strings"
	"time"
)

// IndexStore abstracts the storage layer for parsed vault data.
type IndexStore interface {
	UpsertParsedNote(parsed ParsedNote) error
	RemoveNote(id NoteID) error

	GetTask(id TaskID) (Task, bool)
	ListTasks(filter *TaskFilter) []Task
	GetLogEntriesForTask(id TaskID) []LogEntry
	GetMentionsForTask(id TaskID) []TaskMention

	ListNotesByDate(rangeSel *DateRange) []NoteMeta
	GetNote(id NoteID) (Note, bool)

	ListTags() []string
	ItemsForTag(tag string) TagResult
}

// InMemoryIndexStore implements IndexStore with map-based lookups.
type InMemoryIndexStore struct {
	data VaultIndex
}

// NewInMemoryIndex builds a fresh, empty index store.
func NewInMemoryIndex() *InMemoryIndexStore {
	return &InMemoryIndexStore{
		data: NewVaultIndex(),
	}
}

// UpsertParsedNote implements IndexStore.
func (s *InMemoryIndexStore) UpsertParsedNote(parsed ParsedNote) error {
	noteID := parsed.Note.ID
	_ = s.RemoveNote(noteID)

	meta := NoteMeta{
		ID:    parsed.Note.ID,
		Path:  parsed.Note.Path,
		Title: parsed.Note.Title,
		Date:  parsed.Note.Date,
	}
	s.data.Notes[noteID] = meta
	s.data.NoteContent[noteID] = parsed.Note

	var taskIDs []TaskID
	for _, task := range parsed.Tasks {
		for _, tag := range task.Tags {
			s.data.TagsToTasks[tag] = append(s.data.TagsToTasks[tag], task.ID)
		}
		taskIDs = append(taskIDs, task.ID)
		s.data.Tasks[task.ID] = task
	}
	if len(taskIDs) > 0 {
		s.data.NoteToTaskIDs[noteID] = append([]TaskID{}, taskIDs...)
	}

	var logIDs []LogEntryID
	for _, entry := range parsed.LogEntries {
		for _, tag := range entry.Tags {
			s.data.TagsToLogEntries[tag] = append(s.data.TagsToLogEntries[tag], entry.ID)
		}
		for _, taskID := range entry.TaskIDs {
			s.data.TaskRefsByTask[taskID] = append(s.data.TaskRefsByTask[taskID], entry.ID)
		}
		logIDs = append(logIDs, entry.ID)
		s.data.LogEntries[entry.ID] = entry
	}
	if len(logIDs) > 0 {
		s.data.NoteToLogEntryIDs[noteID] = append([]LogEntryID{}, logIDs...)
	}

	for _, mention := range parsed.Mentions {
		s.data.MentionsByTask[mention.TaskID] = append(s.data.MentionsByTask[mention.TaskID], mention)
	}

	return nil
}

// RemoveNote deletes everything derived from a note.
func (s *InMemoryIndexStore) RemoveNote(id NoteID) error {
	delete(s.data.Notes, id)
	delete(s.data.NoteContent, id)

	if ids, ok := s.data.NoteToTaskIDs[id]; ok {
		for _, taskID := range ids {
			s.removeTask(taskID)
		}
		delete(s.data.NoteToTaskIDs, id)
	}
	if ids, ok := s.data.NoteToLogEntryIDs[id]; ok {
		for _, entryID := range ids {
			s.removeLogEntry(entryID)
		}
		delete(s.data.NoteToLogEntryIDs, id)
	}

	for taskID, mentions := range s.data.MentionsByTask {
		filtered := mentions[:0]
		for _, mention := range mentions {
			if mention.NoteID != id {
				filtered = append(filtered, mention)
			}
		}
		if len(filtered) == 0 {
			delete(s.data.MentionsByTask, taskID)
		} else {
			s.data.MentionsByTask[taskID] = filtered
		}
	}

	return nil
}

func (s *InMemoryIndexStore) removeTask(id TaskID) {
	task, ok := s.data.Tasks[id]
	if ok {
		for _, tag := range task.Tags {
			s.data.TagsToTasks[tag] = removeTaskID(s.data.TagsToTasks[tag], id)
			if len(s.data.TagsToTasks[tag]) == 0 {
				delete(s.data.TagsToTasks, tag)
			}
		}
	}
	delete(s.data.Tasks, id)
	delete(s.data.MentionsByTask, id)
	delete(s.data.TaskRefsByTask, id)
}

func (s *InMemoryIndexStore) removeLogEntry(id LogEntryID) {
	entry, ok := s.data.LogEntries[id]
	if ok {
		for _, tag := range entry.Tags {
			s.data.TagsToLogEntries[tag] = removeLogEntryID(s.data.TagsToLogEntries[tag], id)
			if len(s.data.TagsToLogEntries[tag]) == 0 {
				delete(s.data.TagsToLogEntries, tag)
			}
		}
		for _, taskID := range entry.TaskIDs {
			s.data.TaskRefsByTask[taskID] = removeLogEntryID(s.data.TaskRefsByTask[taskID], id)
			if len(s.data.TaskRefsByTask[taskID]) == 0 {
				delete(s.data.TaskRefsByTask, taskID)
			}
		}
	}
	delete(s.data.LogEntries, id)
}

// GetTask implements IndexStore.
func (s *InMemoryIndexStore) GetTask(id TaskID) (Task, bool) {
	task, ok := s.data.Tasks[id]
	return task, ok
}

// ListTasks implements IndexStore.
func (s *InMemoryIndexStore) ListTasks(filter *TaskFilter) []Task {
	if filter == nil {
		filter = &TaskFilter{}
	}
	var tasks []Task
	for _, task := range s.data.Tasks {
		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}
		if len(filter.Tags) > 0 {
			if !taskHasTags(task, filter.Tags) {
				continue
			}
		}
		if filter.TextSearch != nil {
			text := strings.ToLower(*filter.TextSearch)
			title := strings.ToLower(task.Title)
			match := strings.Contains(title, text)
			if !match && task.DescriptionMD != nil {
				match = strings.Contains(strings.ToLower(*task.DescriptionMD), text)
			}
			if !match {
				continue
			}
		}
		if filter.TouchedSince != nil && !filter.TouchedSince.Time.IsZero() {
			cutoff := filter.TouchedSince.Time
			if task.UpdatedAt.Before(cutoff) {
				if task.ClosedAt == nil || task.ClosedAt.Before(cutoff) {
					continue
				}
			}
		}
		tasks = append(tasks, task)
	}
	return tasks
}

func taskHasTags(task Task, tags []string) bool {
	tagSet := make(map[string]struct{}, len(task.Tags))
	for _, tag := range task.Tags {
		tagSet[tag] = struct{}{}
	}
	for _, desired := range tags {
		if _, ok := tagSet[desired]; !ok {
			return false
		}
	}
	return true
}

// GetLogEntriesForTask implements IndexStore.
func (s *InMemoryIndexStore) GetLogEntriesForTask(id TaskID) []LogEntry {
	var entries []LogEntry
	for _, entryID := range s.data.TaskRefsByTask[id] {
		if entry, ok := s.data.LogEntries[entryID]; ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// GetMentionsForTask implements IndexStore.
func (s *InMemoryIndexStore) GetMentionsForTask(id TaskID) []TaskMention {
	return append([]TaskMention{}, s.data.MentionsByTask[id]...)
}

// ListNotesByDate implements IndexStore.
func (s *InMemoryIndexStore) ListNotesByDate(rangeSel *DateRange) []NoteMeta {
	if rangeSel == nil {
		return nil
	}
	var notes []NoteMeta
	for _, meta := range s.data.Notes {
		if meta.Date == nil || meta.Date.Time.IsZero() {
			continue
		}
		if meta.Date.Time.Before(rangeSel.Start.Time) || meta.Date.Time.After(rangeSel.End.Time) {
			continue
		}
		notes = append(notes, meta)
	}
	sort.Slice(notes, func(i, j int) bool {
		var left, right time.Time
		if notes[i].Date != nil {
			left = notes[i].Date.Time
		}
		if notes[j].Date != nil {
			right = notes[j].Date.Time
		}
		return left.Before(right)
	})
	return notes
}

// GetNote implements IndexStore.
func (s *InMemoryIndexStore) GetNote(id NoteID) (Note, bool) {
	note, ok := s.data.NoteContent[id]
	return note, ok
}

// ListTags implements IndexStore.
func (s *InMemoryIndexStore) ListTags() []string {
	tagSet := make(map[string]struct{})
	for tag := range s.data.TagsToTasks {
		tagSet[tag] = struct{}{}
	}
	for tag := range s.data.TagsToLogEntries {
		tagSet[tag] = struct{}{}
	}
	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

// ItemsForTag implements IndexStore.
func (s *InMemoryIndexStore) ItemsForTag(tag string) TagResult {
	var tasks []Task
	for _, id := range s.data.TagsToTasks[tag] {
		if task, ok := s.data.Tasks[id]; ok {
			tasks = append(tasks, task)
		}
	}

	var logEntries []LogEntry
	for _, id := range s.data.TagsToLogEntries[tag] {
		if entry, ok := s.data.LogEntries[id]; ok {
			logEntries = append(logEntries, entry)
		}
	}

	return TagResult{
		Tag:        tag,
		Tasks:      tasks,
		LogEntries: logEntries,
	}
}

func removeTaskID(haystack []TaskID, needle TaskID) []TaskID {
	result := haystack[:0]
	for _, id := range haystack {
		if id != needle {
			result = append(result, id)
		}
	}
	return result
}

func removeLogEntryID(haystack []LogEntryID, needle LogEntryID) []LogEntryID {
	result := haystack[:0]
	for _, id := range haystack {
		if id != needle {
			result = append(result, id)
		}
	}
	return result
}
