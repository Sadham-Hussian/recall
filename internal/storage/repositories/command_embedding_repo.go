package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
)

type CommandEmbeddingRepository struct {
	db *gorm.DB
}

func NewCommandEmbeddingRepository(db *gorm.DB) *CommandEmbeddingRepository {
	return &CommandEmbeddingRepository{db: db}
}

func (r *CommandEmbeddingRepository) InsertEmbedding(e models.CommandEmbedding) error {
	return r.db.Exec(`
		INSERT INTO command_embeddings
		(command_execution_id, model, dimensions, embedding)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(command_execution_id, model) DO NOTHING
	`, e.CommandExecutionID, e.Model, e.Dimensions, e.Embedding).Error
}

func (r *CommandEmbeddingRepository) FetchAllEmbeddings(model string, limit int) ([]models.EmbeddingSearchResult, error) {
	rows, err := r.db.Raw(`
		SELECT c.command, e.embedding
		FROM command_embeddings e
		JOIN command_executions c
		ON c.id = e.command_execution_id
		WHERE e.model = ?
		LIMIT ?
	`, model, limit).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.EmbeddingSearchResult

	for rows.Next() {
		var cmd string
		var blob []byte

		if err := rows.Scan(&cmd, &blob); err != nil {
			return nil, err
		}

		results = append(results, models.EmbeddingSearchResult{
			Command: cmd,
			Vector:  blob,
		})
	}

	return results, nil
}
