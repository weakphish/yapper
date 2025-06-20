package model

import "time"

type TaskStatus int

const (
	Todo TaskStatus = iota
	InProgress
	Completed
)

type Task struct {
	ID          string
	Title       string
	Description string
	Status      TaskStatus
	CreatedAt   time.Time
	StartedAt   *time.Time // pointer to allow nullability in gorm
	CompletedAt *time.Time
	DependsOn   *Task
	Dependents  []*Task
}

func NewTask(id, title, description string) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      Todo,
		CreatedAt:   time.Now(),
	}
}
