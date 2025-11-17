package core

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NoteParser consumes a note's markdown and emits structured entities.
type NoteParser interface {
	Parse(note Note) ParsedNote
}

// RegexMarkdownParser is the default backend modeled after the Rust version.
type RegexMarkdownParser struct{}

// NewRegexMarkdownParser builds a stateless parser instance.
func NewRegexMarkdownParser() *RegexMarkdownParser {
	return &RegexMarkdownParser{}
}

// Parse implements the NoteParser interface.
func (p *RegexMarkdownParser) Parse(note Note) ParsedNote {
	return parseNote(note)
}

var (
	taskLineExpr   = regexp.MustCompile(`^\s*-\s*\[(?P<mark> |x)\]\s+\[(?P<id>T-[0-9A-Za-z_-]+)\]\s*(?P<rest>.*)$`)
	timePrefixExpr = regexp.MustCompile(`^\s*-\s*([0-9]{1,2}:[0-9]{2}(?:\s?(?:am|pm|AM|PM))?)\s+(.*)$`)
	logTaskExpr    = regexp.MustCompile(`\[(T-[0-9A-Za-z_-]+)\]`)
)

type section int

const (
	sectionOther section = iota
	sectionTasks
	sectionLog
)

func parseNote(note Note) ParsedNote {
	lines := strings.Split(note.Content, "\n")
	timestampNow := time.Now().UTC()

	var tasks []Task
	var logEntries []LogEntry
	var mentions []TaskMention
	current := sectionOther

	for idx := 0; idx < len(lines); idx++ {
		rawLine := strings.TrimRight(lines[idx], "\r")
		lineNumber := idx + 1
		trimmedLeading := strings.TrimLeft(rawLine, " \t")
		if strings.HasPrefix(trimmedLeading, "## ") {
			heading := strings.TrimSpace(strings.TrimPrefix(trimmedLeading, "## "))
			switch strings.ToLower(heading) {
			case "tasks":
				current = sectionTasks
			case "log":
				current = sectionLog
			default:
				current = sectionOther
			}
			continue
		}

		switch current {
		case sectionTasks:
			if match := taskLineExpr.FindStringSubmatch(trimmedLeading); match != nil {
				continuation, consumed := collectContinuation(lines, idx+1)
				idx += consumed
				mark := match[1]
				taskID := TaskID(match[2])
				rest := strings.TrimSpace(match[3])
				combined := combineContinuation(rest, continuation)
				task := buildTask(note, taskID, mark, combined, timestampNow)
				tasks = append(tasks, task)
			}
		case sectionLog:
			entry, entryMentions, consumed := parseLogEntry(note, rawLine, lineNumber, lines, idx+1)
			if entry != nil {
				idx += consumed
				logEntries = append(logEntries, *entry)
				mentions = append(mentions, entryMentions...)
			}
		default:
		}
	}

	return ParsedNote{
		Note:       note,
		Tasks:      tasks,
		LogEntries: logEntries,
		Mentions:   mentions,
	}
}

func buildTask(note Note, id TaskID, mark string, body string, now time.Time) Task {
	status := TaskStatusOpen
	if mark == "x" || mark == "X" {
		status = TaskStatusDone
	}
	title, tags := splitTitleAndTags(body)
	task := Task{
		ID:        id,
		Title:     title,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      tags,
	}
	if status == TaskStatusDone {
		task.ClosedAt = &now
	}
	if note.ID != "" {
		idCopy := note.ID
		task.SourceNoteID = &idCopy
	}
	return task
}

func parseLogEntry(note Note, rawLine string, lineNumber int, lines []string, startIdx int) (*LogEntry, []TaskMention, int) {
	if !strings.HasPrefix(strings.TrimLeft(rawLine, " \t"), "- ") {
		return nil, nil, 0
	}

	continuation, consumed := collectContinuation(lines, startIdx)
	timeMatch := timePrefixExpr.FindStringSubmatch(rawLine)

	var timestamp *string
	var remainder string
	if timeMatch != nil {
		ts := strings.TrimSpace(timeMatch[1])
		timestamp = &ts
		remainder = strings.TrimSpace(timeMatch[2])
	} else {
		remainder = strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(rawLine, " \t"), "- "))
	}

	combined := combineContinuation(remainder, continuation)
	content, tags := splitTitleAndTags(combined)
	entryID := LogEntryID(string(note.ID) + ":" + strconv.Itoa(lineNumber))

	taskIDs, logMentions := extractTaskMentions(note.ID, entryID, combined)
	entry := &LogEntry{
		ID:         entryID,
		NoteID:     note.ID,
		LineNumber: lineNumber,
		Timestamp:  timestamp,
		ContentMD:  content,
		Tags:       tags,
		TaskIDs:    taskIDs,
	}

	return entry, logMentions, consumed
}

func extractTaskMentions(noteID NoteID, entryID LogEntryID, content string) ([]TaskID, []TaskMention) {
	var taskIDs []TaskID
	var mentions []TaskMention
	for _, match := range logTaskExpr.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		id := TaskID(match[1])
		taskIDs = append(taskIDs, id)
		entryIDCopy := entryID
		mentions = append(mentions, TaskMention{
			TaskID:     id,
			NoteID:     noteID,
			LogEntryID: &entryIDCopy,
			Excerpt:    buildExcerpt(content),
		})
	}
	return taskIDs, mentions
}

func combineContinuation(base string, continuation []string) string {
	combined := strings.TrimSpace(base)
	for _, extra := range continuation {
		if combined != "" {
			combined += "\n"
		}
		combined += strings.TrimSpace(extra)
	}
	return combined
}

func collectContinuation(lines []string, start int) ([]string, int) {
	var extras []string
	consumed := 0
	for i := start; i < len(lines); i++ {
		next := strings.TrimRight(lines[i], "\r")
		trimmed := strings.TrimLeft(next, " \t")
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "- ") || strings.TrimSpace(next) == "" {
			break
		}
		if strings.HasPrefix(next, " ") || strings.HasPrefix(next, "\t") {
			extras = append(extras, next)
			consumed++
		} else {
			break
		}
	}
	return extras, consumed
}

func splitTitleAndTags(input string) (string, []string) {
	var titleParts []string
	var tags []string
	for _, part := range strings.Fields(input) {
		if strings.HasPrefix(part, "#") && len(part) > 1 {
			tags = append(tags, part[1:])
			continue
		}
		titleParts = append(titleParts, part)
	}
	title := strings.TrimSpace(strings.Join(titleParts, " "))
	if title == "" {
		title = strings.TrimSpace(input)
	}
	return title, tags
}

func buildExcerpt(content string) string {
	const maxExcerpt = 120
	if len(content) <= maxExcerpt {
		return content
	}
	return content[:maxExcerpt] + "…"
}
