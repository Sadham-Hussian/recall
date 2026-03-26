package models

type CommandChain struct {
	ID              uint   `gorm:"primaryKey"`
	PreviousCommand string `gorm:"column:prev_command"`
	NextCommand     string `gorm:"column:next_command"`
	SessionID       string `gorm:"column:session_id"`
	OccurrenceCount int    `gorm:"column:occurrence_count"`
}
