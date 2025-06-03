package model

import "log/slog"

// IsComplete returns true if the task is complete
func (b *Block) IsComplete() bool {
	if !b.IsTask() {
		return false
	}

	return b.completed
}

// ToggleTask toggles the completion status of a task
// Returns true if the toggle was successful
func (b *Block) ToggleTask() bool {
	if !b.IsTask() {
		slog.Warn("attempted to toggle a non-task block", "id", b.id.String(), "type", b.blockType)
		return false
	}

	// Toggle the completion status
	b.completed = !b.completed
	if b.completed {
		slog.Info("marking task as complete", "id", b.id.String())
	} else {
		slog.Info("marking task as incomplete", "id", b.id.String())
	}

	return true
}

// SetCompleted sets the task's completion status
func (b *Block) SetCompleted(completed bool) {
	if !b.IsTask() {
		slog.Warn("attempted to set completion status on non-task block", "id", b.id.String())
		return
	}

	if b.completed != completed {
		b.completed = completed
		if completed {
			slog.Info("task marked as complete", "id", b.id.String())
		} else {
			slog.Info("task marked as incomplete", "id", b.id.String())
		}
	}
}
