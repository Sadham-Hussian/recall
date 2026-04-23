package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
)

type CommandChainRepository struct {
	db *gorm.DB
}

func NewCommandChainRepository(db *gorm.DB) *CommandChainRepository {
	return &CommandChainRepository{db: db}
}

func (r *CommandChainRepository) Upsert(prev, next, session string) error {

	sql := `
	INSERT INTO command_chains
	(prev_command, next_command, session_id, occurrence_count)
	VALUES (?, ?, ?, 1)
	ON CONFLICT(prev_command, next_command, session_id)
	DO UPDATE SET occurrence_count = occurrence_count + 1;
	`

	return r.db.Exec(sql, prev, next, session).Error
}

func (r *CommandChainRepository) GetNextCommands(
	command string,
	limit int,
) ([]models.CommandChain, error) {

	sql := `
	SELECT next_command, SUM(occurrence_count) as occurrence_count
	FROM command_chains
	WHERE prev_command = ?
	GROUP BY next_command
	ORDER BY occurrence_count DESC
	LIMIT ?;
	`

	var result []models.CommandChain

	err := r.db.Raw(sql, command, limit).Scan(&result).Error

	return result, err
}

// UpsertWithCount inserts a chain or adds to its occurrence count.
// Used during import to preserve the original count rather than
// incrementing by 1.
func (r *CommandChainRepository) UpsertWithCount(prev, next, sessionID string, count int) error {
	sql := `
		INSERT INTO command_chains (prev_command, next_command, session_id, occurrence_count)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(prev_command, next_command, session_id)
		DO UPDATE SET occurrence_count = occurrence_count + ?
	`
	return r.db.Exec(sql, prev, next, sessionID, count, count).Error
}
