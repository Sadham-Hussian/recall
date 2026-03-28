package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
)

type CommandExecutionSearchRepository struct {
	db *gorm.DB
}

func NewCommandExecutionSearchRepository(db *gorm.DB) *CommandExecutionSearchRepository {
	return &CommandExecutionSearchRepository{db: db}
}

func (r *CommandExecutionSearchRepository) HybridSearch(query string, limit int) ([]models.HybridSearchResult, error) {

	sql := `
	SELECT
		e.command,
		COUNT(*) as count,
		MAX(e.timestamp) as last_timestamp,
		MAX(e.cwd) as cwd,
		SUM(CASE WHEN e.exit_code = 0 THEN 1 ELSE 0 END) as success_count,
		MAX(e.session_id) as session_id
	FROM command_executions_fts
	JOIN command_executions e 
		ON e.id = command_executions_fts.rowid
	WHERE command_executions_fts MATCH ? 
	AND e.command NOT LIKE 'recall%' 
	AND e.command NOT LIKE './recall%'
	GROUP BY e.command
	LIMIT ?;
	`

	var results []models.HybridSearchResult
	err := r.db.Raw(sql, query, limit*5).Scan(&results).Error
	return results, err
}

func (r *CommandExecutionSearchRepository) FuzzySearch(query string, limit int) ([]models.HybridSearchResult, error) {

	sql := `
	SELECT
		command,
		COUNT(*) as count,
		MAX(timestamp) as last_timestamp,
		MAX(cwd) as cwd,
		SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) as success_count,
		MAX(session_id) as session_id
	FROM command_executions
	GROUP BY command
	LIMIT ?;
	`

	var result []models.HybridSearchResult

	err := r.db.Raw(sql, limit).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *CommandExecutionSearchRepository) FTSSearch(query string, limit int) ([]models.FTSResult, error) {

	sql := `
		SELECT
			e.command,
			bm25(command_executions_fts) as rank
		FROM command_executions_fts
		JOIN command_executions e 
			ON e.id = command_executions_fts.rowid
		WHERE command_executions_fts MATCH ?
		AND e.command NOT LIKE 'recall%'
		AND e.command NOT LIKE './recall%'
		ORDER BY rank
		LIMIT ?;
	`

	var results []models.FTSResult
	err := r.db.Raw(sql, query, limit).Scan(&results).Error
	return results, err
}
