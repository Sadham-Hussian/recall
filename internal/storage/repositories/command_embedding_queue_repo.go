package repositories

import "gorm.io/gorm"

type CommandEmbeddingQueueRepository struct {
	db *gorm.DB
}

func NewCommandEmbeddingQueueRepository(db *gorm.DB) *CommandEmbeddingQueueRepository {
	return &CommandEmbeddingQueueRepository{db: db}
}

func (r *CommandEmbeddingQueueRepository) Enqueue(commandExecutionID int64) error {
	return r.db.Exec(`
		INSERT OR IGNORE INTO embedding_queue (command_execution_id)
		VALUES (?)
	`, commandExecutionID).Error
}

func (r *CommandEmbeddingQueueRepository) DeleteFromQueue(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.Exec(`
		DELETE FROM embedding_queue
		WHERE command_execution_id IN ?
	`, ids).Error
}
