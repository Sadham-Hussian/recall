package models

type Workflow struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"default:''"`
	CreatedAt   string
	UpdatedAt   string
}

type WorkflowStep struct {
	ID         uint   `gorm:"primaryKey"`
	WorkflowID uint   `gorm:"not null"`
	StepOrder  int    `gorm:"not null"`
	Command    string `gorm:"not null"`
}
