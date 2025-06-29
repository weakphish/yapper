package model

type Note struct {
	ID           int
	Title        string
	Content      string
	RelatedTasks []Task `gorm:"many2many:note_tasks;"`
}
