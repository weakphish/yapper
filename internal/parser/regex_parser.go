package parser

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/weakphish/yapper/internal/model"
)

// RegexNoteParser implements the NoteParser interface using a series of regular
// expressions that scan Markdown line-by-line. It serves as the v1 parser and
// keeps the rest of the system decoupled from any specific parsing strategy.
type RegexNoteParser struct {
	taskLinePattern     *regexp.Regexp
	logLinePattern      *regexp.Regexp
	tagPattern          *regexp.Regexp
	taskIDPattern       *regexp.Regexp
	logTimestampPattern *regexp.Regexp
}

// NewRegexNoteParser constructs a regex-driven NoteParser implementation.
func NewRegexNoteParser() NoteParser {
	return &RegexNoteParser{
		taskLinePattern:     regexp.MustCompile(`^\s*[-*]\s+\[([^\]])\]\s+(.+)$`),
		logLinePattern:      regexp.MustCompile(`^\s*[-*]\s+(.+)$`),
		tagPattern:          regexp.MustCompile(`#([[:alnum:]/_-]+)`),
		taskIDPattern:       regexp.MustCompile(`\[(T-[A-Za-z0-9_-]+)\]`),
		logTimestampPattern: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:[ T](\d{2}:\d{2}(?::\d{2})?))?(?:\s+-\s+|\s+)(.*)$`),
	}
}

// Parse translates the supplied note into structured data. The parser keeps
// scanning state minimal so that future implementations (Tree-sitter, AST, etc.)
// can replace it without affecting callers.
func (p *RegexNoteParser) Parse(ctx context.Context, note *model.Note) (*ParsedNote, error) {
	if note == nil {
		return nil, errors.New("note cannot be nil")
	}
	if err := ensureParserContext(ctx); err != nil {
		return nil, err
	}

	result := &ParsedNote{
		Note:       note,
		Tasks:      []model.Task{},
		LogEntries: []model.LogEntry{},
		Mentions:   []model.TaskMention{},
	}

	lines := strings.Split(note.Content, "\n")
	section := sectionNone
	for i, line := range lines {
		if err := ensureParserContext(ctx); err != nil {
			return nil, err
		}
		lineNumber := i + 1
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			// Section headings toggle parsing behavior; the simple regex parser
			// only cares about the "Tasks" and "Log" sections today.
			switch strings.ToLower(strings.TrimSpace(trimmed[3:])) {
			case "tasks":
				section = sectionTasks
			case "log":
				section = sectionLog
			default:
				section = sectionNone
			}
			continue
		}

		switch section {
		case sectionTasks:
			if task := p.parseTaskLine(note, line, lineNumber); task != nil {
				result.Tasks = append(result.Tasks, *task)
			}
		case sectionLog:
			entry, mentions := p.parseLogLine(note, line, lineNumber)
			if entry != nil {
				result.LogEntries = append(result.LogEntries, *entry)
			}
			if len(mentions) > 0 {
				result.Mentions = append(result.Mentions, mentions...)
			}
		default:
			if mentions := p.parseMentionsFromLine(line, lineNumber, nil, note); len(mentions) > 0 {
				result.Mentions = append(result.Mentions, mentions...)
			}
		}
	}

	return result, nil
}

// parseSection enumerates the parser states that correspond to the Markdown
// headings recognized by the regex parser.
type parseSection int

const (
	sectionNone parseSection = iota
	sectionTasks
	sectionLog
)

// parseTaskLine extracts a task from a Markdown bullet within the tasks section.
func (p *RegexNoteParser) parseTaskLine(note *model.Note, line string, lineNumber int) *model.Task {
	matches := p.taskLinePattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	status := parseTaskStatus(matches[1])
	body := strings.TrimSpace(matches[2])
	if body == "" {
		return nil
	}

	tags := p.extractTags(body)
	taskID := p.extractExplicitTaskID(body)
	if taskID == "" {
		taskID = model.TaskID(fmt.Sprintf("%s#%d", note.ID, lineNumber))
	}

	title := p.cleanContent(body)

	now := note.Date
	task := &model.Task{
		ID:        taskID,
		NoteID:    note.ID,
		Title:     title,
		Status:    status,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
		Line:      lineNumber,
	}
	if status == model.TaskStatusDone && !now.IsZero() {
		task.CompletedAt = &now
	}

	return task
}

// parseLogLine parses a single log bullet and returns the resulting LogEntry
// plus any task mentions that appear within the line.
func (p *RegexNoteParser) parseLogLine(note *model.Note, line string, lineNumber int) (*model.LogEntry, []model.TaskMention) {
	matches := p.logLinePattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, nil
	}

	body := strings.TrimSpace(matches[1])
	if body == "" {
		return nil, nil
	}

	timestamp, content := p.extractTimestamp(body, note.Date)
	tags := p.extractTags(content)
	refs := p.extractTaskIDs(content)
	cleanedContent := p.cleanContent(content)

	entry := &model.LogEntry{
		ID:        model.LogEntryID(fmt.Sprintf("%s#log#%d", note.ID, lineNumber)),
		NoteID:    note.ID,
		Line:      lineNumber,
		Timestamp: timestamp,
		Content:   cleanedContent,
		Tags:      tags,
		TaskRefs:  refs,
	}

	mentions := p.parseMentionsFromLine(content, lineNumber, tags, note)
	return entry, mentions
}

// parseMentionsFromLine extracts TaskMention objects from any `[T-xxxx]`
// references contained on the line.
func (p *RegexNoteParser) parseMentionsFromLine(line string, lineNumber int, tags []string, note *model.Note) []model.TaskMention {
	taskIDs := p.extractTaskIDs(line)
	if len(taskIDs) == 0 {
		return nil
	}

	if tags == nil {
		tags = p.extractTags(line)
	}

	context := p.cleanContent(line)
	mentions := make([]model.TaskMention, 0, len(taskIDs))
	for _, id := range taskIDs {
		mentions = append(mentions, model.TaskMention{
			TaskID:  id,
			NoteID:  note.ID,
			Line:    lineNumber,
			Context: context,
			Tags:    tags,
		})
	}
	return mentions
}

// extractTimestamp attempts to parse a log timestamp prefix and falls back to
// the note's date (or the zero value if absent).
func (p *RegexNoteParser) extractTimestamp(line string, fallback time.Time) (time.Time, string) {
	matches := p.logTimestampPattern.FindStringSubmatch(line)
	if matches == nil {
		return fallback, line
	}

	datePart := matches[1]
	timePart := matches[2]
	rest := strings.TrimSpace(matches[3])

	layout := "2006-01-02"
	if timePart != "" {
		layout = "2006-01-02 15:04"
		if strings.Count(timePart, ":") == 2 {
			layout = "2006-01-02 15:04:05"
		}
		datePart = fmt.Sprintf("%s %s", datePart, timePart)
	}

	parsed, err := time.Parse(layout, datePart)
	if err != nil {
		return fallback, line
	}
	if !fallback.IsZero() {
		parsed = parsed.In(fallback.Location())
	}
	return parsed, rest
}

// extractTags returns de-duplicated hashtag tokens from the provided input.
func (p *RegexNoteParser) extractTags(input string) []string {
	matches := p.tagPattern.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil
	}

	tags := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		tag := m[1]
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

// extractTaskIDs returns all task identifiers embedded in the line.
func (p *RegexNoteParser) extractTaskIDs(line string) []model.TaskID {
	matches := p.taskIDPattern.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil
	}

	ids := make([]model.TaskID, 0, len(matches))
	seen := make(map[model.TaskID]struct{}, len(matches))
	for _, m := range matches {
		id := model.TaskID(m[1])
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// extractExplicitTaskID returns the first task identifier declared inline on the
// task content, if any.
func (p *RegexNoteParser) extractExplicitTaskID(content string) model.TaskID {
	matches := p.taskIDPattern.FindStringSubmatch(content)
	if matches == nil {
		return ""
	}
	return model.TaskID(matches[1])
}

// cleanContent strips tags, task identifiers, and excess whitespace so titles
// and log entry content remain focused on the human-readable text.
func (p *RegexNoteParser) cleanContent(content string) string {
	withoutIDs := p.taskIDPattern.ReplaceAllString(content, "")
	withoutTags := p.tagPattern.ReplaceAllString(withoutIDs, "")
	fields := strings.Fields(withoutTags)
	return strings.Join(fields, " ")
}

// parseTaskStatus maps task checkbox characters to a TaskStatus value.
func parseTaskStatus(char string) model.TaskStatus {
	switch strings.TrimSpace(strings.ToLower(char)) {
	case "x":
		return model.TaskStatusDone
	case "~", "/":
		return model.TaskStatusInProgress
	case "!":
		return model.TaskStatusBlocked
	default:
		return model.TaskStatusTodo
	}
}

// ensureParserContext mirrors context checks inside the parser loop so parsing
// can be cancelled if needed.
func ensureParserContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
