package models

type CommandEmbedding struct {
	ID                 uint   `gorm:"primaryKey"`
	CommandExecutionID uint   `gorm:"column:command_execution_id"`
	Model              string `gorm:"column:model"`
	Dimensions         int    `gorm:"column:dimensions"`
	Embedding          []byte `gorm:"column:embedding;type:blob"`
	CreatedAt          string `gorm:"column:created_at;autoCreateTime"`
}

type EmbeddingSearchResult struct {
	Command string `gorm:"column:command"`
	Vector  []byte `gorm:"column:vector"`
}

type SearchResult struct {
	Command string
	Score   float64
}
