package index

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/weakphish/yapper/internal/model"
	"github.com/weakphish/yapper/internal/parser"
)

// InMemoryIndexStore implements IndexStore with plain Go maps guarded by a RW
// mutex. It prioritizes correctness and determinism over absolute performance.
type InMemoryIndexStore struct {
	mu sync.RWMutex

	notes    map[model.NoteID]*model.Note
	tasks    map[model.TaskID]model.Task
	logs     map[model.LogEntryID]model.LogEntry
	mentions map[model.TaskID][]model.TaskMention
	tagIndex map[string]*tagBucket
	noteData map[model.NoteID]*noteSnapshot
}

type tagBucket struct {
	taskIDs  map[model.TaskID]struct{}
	logIDs   map[model.LogEntryID]struct{}
	mentions map[string]model.TaskMention
}

type noteSnapshot struct {
	TaskIDs     []model.TaskID
	LogIDs      []model.LogEntryID
	Mentions    []model.TaskMention
	TaskTags    map[string][]model.TaskID
	LogTags     map[string][]model.LogEntryID
	MentionTags map[string][]string
}

// NewInMemoryIndexStore constructs a ready-to-use in-memory store instance.
func NewInMemoryIndexStore() *InMemoryIndexStore {
	return &InMemoryIndexStore{
		notes:    make(map[model.NoteID]*model.Note),
		tasks:    make(map[model.TaskID]model.Task),
		logs:     make(map[model.LogEntryID]model.LogEntry),
		mentions: make(map[model.TaskID][]model.TaskMention),
		tagIndex: make(map[string]*tagBucket),
		noteData: make(map[model.NoteID]*noteSnapshot),
	}
}

// UpsertParsedNote replaces any previously indexed state for parsed.Note.Note.
func (s *InMemoryIndexStore) UpsertParsedNote(ctx context.Context, parsed *parser.ParsedNote) error {
	if parsed == nil || parsed.Note == nil {
		return errors.New("parsed note cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeNoteLocked(parsed.Note.ID)

	s.notes[parsed.Note.ID] = parsed.Note

	snap := &noteSnapshot{
		TaskTags:    make(map[string][]model.TaskID),
		LogTags:     make(map[string][]model.LogEntryID),
		MentionTags: make(map[string][]string),
	}

	for _, task := range parsed.Tasks {
		taskCopy := task
		s.tasks[task.ID] = taskCopy
		snap.TaskIDs = append(snap.TaskIDs, task.ID)
		for _, tag := range normalizeTags(task.Tags) {
			s.addTagTask(tag, taskCopy)
			snap.TaskTags[tag] = append(snap.TaskTags[tag], task.ID)
		}
	}

	for _, logEntry := range parsed.LogEntries {
		entryCopy := logEntry
		s.logs[entryCopy.ID] = entryCopy
		snap.LogIDs = append(snap.LogIDs, entryCopy.ID)
		for _, tag := range normalizeTags(entryCopy.Tags) {
			s.addTagLog(tag, entryCopy)
			snap.LogTags[tag] = append(snap.LogTags[tag], entryCopy.ID)
		}
	}

	for _, mention := range parsed.Mentions {
		mentionCopy := mention
		s.mentions[mention.TaskID] = append(s.mentions[mention.TaskID], mentionCopy)
		snap.Mentions = append(snap.Mentions, mentionCopy)
		for _, tag := range normalizeTags(mention.Tags) {
			key := mentionKey(mentionCopy)
			b := s.ensureTagBucket(tag)
			b.mentions[key] = mentionCopy
			snap.MentionTags[tag] = append(snap.MentionTags[tag], key)
		}
	}

	s.noteData[parsed.Note.ID] = snap

	return nil
}

// RemoveNote drops every indexed entity derived from noteID.
func (s *InMemoryIndexStore) RemoveNote(ctx context.Context, noteID model.NoteID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeNoteLocked(noteID)
	delete(s.notes, noteID)
	delete(s.noteData, noteID)
	return nil
}

func (s *InMemoryIndexStore) removeNoteLocked(noteID model.NoteID) {
	snap, ok := s.noteData[noteID]
	if !ok {
		return
	}
	for _, id := range snap.TaskIDs {
		delete(s.tasks, id)
	}
	for _, id := range snap.LogIDs {
		delete(s.logs, id)
	}
	for _, mention := range snap.Mentions {
		taskMentions := s.mentions[mention.TaskID]
		filtered := taskMentions[:0]
		for _, existing := range taskMentions {
			if mentionKey(existing) != mentionKey(mention) {
				filtered = append(filtered, existing)
			}
		}
		if len(filtered) == 0 {
			delete(s.mentions, mention.TaskID)
		} else {
			s.mentions[mention.TaskID] = filtered
		}
	}
	for tag, ids := range snap.TaskTags {
		b := s.tagIndex[tag]
		if b == nil {
			continue
		}
		for _, id := range ids {
			delete(b.taskIDs, id)
		}
		s.cleanupTag(tag)
	}
	for tag, ids := range snap.LogTags {
		b := s.tagIndex[tag]
		if b == nil {
			continue
		}
		for _, id := range ids {
			delete(b.logIDs, id)
		}
		s.cleanupTag(tag)
	}
	for tag, keys := range snap.MentionTags {
		b := s.tagIndex[tag]
		if b == nil {
			continue
		}
		for _, key := range keys {
			delete(b.mentions, key)
		}
		s.cleanupTag(tag)
	}
}

// GetTask returns a task if it exists.
func (s *InMemoryIndexStore) GetTask(ctx context.Context, id model.TaskID) (model.Task, bool, error) {
	if err := ctx.Err(); err != nil {
		return model.Task{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok, nil
}

// ListTasks returns the tasks that satisfy the provided filter.
func (s *InMemoryIndexStore) ListTasks(ctx context.Context, filter TaskFilter) ([]model.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.Task
	for _, task := range s.tasks {
		if !matchTaskFilter(task, filter) {
			continue
		}
		result = append(result, task)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// GetLogEntriesForTask returns log entries referencing the task ID.
func (s *InMemoryIndexStore) GetLogEntriesForTask(ctx context.Context, id model.TaskID) ([]model.LogEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var entries []model.LogEntry
	for _, entry := range s.logs {
		for _, ref := range entry.TaskRefs {
			if ref == id {
				entries = append(entries, entry)
				break
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	return entries, nil
}

// GetMentionsForTask returns mentions associated with the task.
func (s *InMemoryIndexStore) GetMentionsForTask(ctx context.Context, id model.TaskID) ([]model.TaskMention, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	mentions := append([]model.TaskMention(nil), s.mentions[id]...)
	sort.Slice(mentions, func(i, j int) bool {
		if mentions[i].NoteID == mentions[j].NoteID {
			return mentions[i].Line < mentions[j].Line
		}
		return mentions[i].NoteID < mentions[j].NoteID
	})
	return mentions, nil
}

// ListNotes lists notes filtered by date range and ordered by newest first.
func (s *InMemoryIndexStore) ListNotes(ctx context.Context, filter NoteFilter) ([]model.Note, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var notes []model.Note
	for _, note := range s.notes {
		if !matchNoteFilter(note, filter) {
			continue
		}
		notes = append(notes, *note)
	}
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Date.After(notes[j].Date)
	})
	return notes, nil
}

// ListTags returns the sorted set of tags.
func (s *InMemoryIndexStore) ListTags(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var tags []string
	for tag := range s.tagIndex {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

// ItemsForTag returns every entity referencing the provided tag.
func (s *InMemoryIndexStore) ItemsForTag(ctx context.Context, tag string) (TagItems, bool, error) {
	if err := ctx.Err(); err != nil {
		return TagItems{}, false, err
	}
	tag = normalizeTag(tag)
	if tag == "" {
		return TagItems{}, false, errors.New("tag cannot be empty")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	bucket, ok := s.tagIndex[tag]
	if !ok {
		return TagItems{}, false, nil
	}

	items := TagItems{Tag: tag}
	for id := range bucket.taskIDs {
		if task, ok := s.tasks[id]; ok {
			items.Tasks = append(items.Tasks, task)
		}
	}
	for id := range bucket.logIDs {
		if entry, ok := s.logs[id]; ok {
			items.LogEntries = append(items.LogEntries, entry)
		}
	}
	for _, mention := range bucket.mentions {
		items.Mentions = append(items.Mentions, mention)
	}

	sort.Slice(items.Tasks, func(i, j int) bool { return items.Tasks[i].ID < items.Tasks[j].ID })
	sort.Slice(items.LogEntries, func(i, j int) bool { return items.LogEntries[i].ID < items.LogEntries[j].ID })
	sort.Slice(items.Mentions, func(i, j int) bool {
		if items.Mentions[i].TaskID == items.Mentions[j].TaskID {
			if items.Mentions[i].NoteID == items.Mentions[j].NoteID {
				return items.Mentions[i].Line < items.Mentions[j].Line
			}
			return items.Mentions[i].NoteID < items.Mentions[j].NoteID
		}
		return items.Mentions[i].TaskID < items.Mentions[j].TaskID
	})

	return items, true, nil
}

func (s *InMemoryIndexStore) addTagTask(tag string, task model.Task) {
	if tag == "" {
		return
	}
	b := s.ensureTagBucket(tag)
	b.taskIDs[task.ID] = struct{}{}
}

func (s *InMemoryIndexStore) addTagLog(tag string, entry model.LogEntry) {
	if tag == "" {
		return
	}
	b := s.ensureTagBucket(tag)
	b.logIDs[entry.ID] = struct{}{}
}

func (s *InMemoryIndexStore) ensureTagBucket(tag string) *tagBucket {
	tag = normalizeTag(tag)
	if tag == "" {
		return &tagBucket{
			taskIDs:  make(map[model.TaskID]struct{}),
			logIDs:   make(map[model.LogEntryID]struct{}),
			mentions: make(map[string]model.TaskMention),
		}
	}
	b, ok := s.tagIndex[tag]
	if !ok {
		b = &tagBucket{
			taskIDs:  make(map[model.TaskID]struct{}),
			logIDs:   make(map[model.LogEntryID]struct{}),
			mentions: make(map[string]model.TaskMention),
		}
		s.tagIndex[tag] = b
	}
	return b
}

func (s *InMemoryIndexStore) cleanupTag(tag string) {
	b := s.tagIndex[tag]
	if b == nil {
		return
	}
	if len(b.taskIDs) == 0 && len(b.logIDs) == 0 && len(b.mentions) == 0 {
		delete(s.tagIndex, tag)
	}
}

func matchTaskFilter(task model.Task, filter TaskFilter) bool {
	if len(filter.Statuses) > 0 && !containsStatus(filter.Statuses, task.Status) {
		return false
	}
	if len(filter.Tags) > 0 && !hasAnyTag(task.Tags, filter.Tags) {
		return false
	}
	if len(filter.NoteIDs) > 0 && !containsNoteID(filter.NoteIDs, task.NoteID) {
		return false
	}
	return true
}

func containsStatus(statuses []model.TaskStatus, status model.TaskStatus) bool {
	for _, s := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

func hasAnyTag(tags []string, targets []string) bool {
	normTags := normalizeTags(tags)
	for _, target := range targets {
		if normTagsContains(normTags, target) {
			return true
		}
	}
	return false
}

func normTagsContains(tags []string, candidate string) bool {
	candidate = normalizeTag(candidate)
	for _, tag := range tags {
		if tag == candidate {
			return true
		}
	}
	return false
}

func containsNoteID(ids []model.NoteID, noteID model.NoteID) bool {
	for _, id := range ids {
		if id == noteID {
			return true
		}
	}
	return false
}

func matchNoteFilter(note *model.Note, filter NoteFilter) bool {
	if note == nil {
		return false
	}
	if filter.Start != nil && note.Date.Before(*filter.Start) {
		return false
	}
	if filter.End != nil && note.Date.After(*filter.End) {
		return false
	}
	return true
}

func normalizeTags(tags []string) []string {
	var result []string
	seen := make(map[string]struct{})
	for _, tag := range tags {
		norm := normalizeTag(tag)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, norm)
	}
	sort.Strings(result)
	return result
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func mentionKey(m model.TaskMention) string {
	return fmt.Sprintf("%s|%s|%d|%s", m.TaskID, m.NoteID, m.Line, m.Context)
}
