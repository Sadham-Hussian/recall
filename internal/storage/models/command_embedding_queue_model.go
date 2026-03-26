package models

import "time"

type CommandEmbeddingQueue struct {
	CommandExecutionID uint      `gorm:"column:command_execution_id;primaryKey"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime"`
}
