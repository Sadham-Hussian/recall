package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommandExecutionRepository struct {
	db *gorm.DB
}

func NewCommandExecutionRepository(db *gorm.DB) *CommandExecutionRepository {
	return &CommandExecutionRepository{db: db}
}

func (r *CommandExecutionRepository) InsertIgnore(e *models.CommandExecution) error {
	return r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(e).Error
}

func (r *CommandExecutionRepository) InsertWithFTS(e *models.CommandExecution) error {

	tx := r.db.Begin()

	// Insert into main table with conflict ignore
	err := tx.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(e).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// If ID is zero → it was ignored (duplicate)
	if e.ID == 0 {
		tx.Rollback()
		return nil
	}

	// Insert into FTS only if new row was created
	err = tx.Exec(
		`INSERT INTO command_executions_fts(rowid, command) VALUES (?, ?)`,
		e.ID,
		e.Command,
	).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *CommandExecutionRepository) ListRecent(limit int) ([]models.CommandExecution, error) {
	var executions []models.CommandExecution

	err := r.db.
		Order("timestamp DESC").
		Limit(limit).
		Find(&executions).Error

	return executions, err
}

func (r *CommandExecutionRepository) Last() (*models.CommandExecution, error) {
	var execution models.CommandExecution

	err := r.db.
		Order("timestamp DESC").
		Limit(1).
		First(&execution).Error

	if err != nil {
		return nil, err
	}

	return &execution, nil
}

func (r *CommandExecutionRepository) Search(query string, limit int) ([]models.CommandExecution, error) {

	sql := `
	SELECT e.*
	FROM command_executions_fts f
	JOIN command_executions e ON e.id = f.rowid
	WHERE command_executions_fts MATCH ?
	ORDER BY rank
	LIMIT ?;
	`

	var results []models.CommandExecution
	err := r.db.Raw(sql, query, limit).Scan(&results).Error

	return results, err
}

func (r *CommandExecutionRepository) GetLastCommand(shellPID int) (*models.CommandExecution, error) {

	var exec models.CommandExecution

	err := r.db.
		Where("shell_pid = ?", shellPID).
		Order("timestamp DESC").
		Limit(1).
		Find(&exec).Error

	if err != nil {
		return nil, err
	}

	if exec.ID == 0 {
		return nil, nil
	}

	return &exec, nil
}

func (r *CommandExecutionRepository) GetCurrentSessionID() (string, error) {

	var sessionID string

	err := r.db.
		Table("command_executions").
		Select("session_id").
		Order("timestamp DESC").
		Limit(1).
		Scan(&sessionID).Error

	return sessionID, err
}

func (r *CommandExecutionRepository) GetPreviousCommandByID(
	sessionID string,
	currentID uint,
) (string, error) {

	var command string

	err := r.db.
		Table("command_executions").
		Select("command").
		Where("session_id = ? AND id < ?", sessionID, currentID).
		Order("id DESC").
		Limit(1).
		Scan(&command).Error

	return command, err
}

func (r *CommandExecutionRepository) GetCurrentSessionByShellPID(
	shellPID int,
) (string, error) {

	var sessionID string

	err := r.db.
		Table("command_executions").
		Select("session_id").
		Where("shell_pid = ?", shellPID).
		Order("id DESC").
		Limit(1).
		Scan(&sessionID).Error

	return sessionID, err
}

func (r *CommandExecutionRepository) GetCommandsBySessionID(
	sessionID string,
) ([]models.CommandExecution, error) {

	var commands []models.CommandExecution

	err := r.db.
		Table("command_executions").
		Where("session_id = ?", sessionID).
		Order("id ASC").
		Find(&commands).Error

	return commands, err
}

func (r *CommandExecutionRepository) GetLastSessionCommands(
	limit int,
) ([]models.CommandExecution, error) {

	var executions []models.CommandExecution

	err := r.db.Raw(`
		SELECT *
		FROM command_executions
		WHERE session_id IN (
			SELECT session_id
			FROM command_executions
			WHERE session_id != ''
			GROUP BY session_id
			ORDER BY MAX(id) DESC
			LIMIT ?
		)
		ORDER BY id ASC
	`, limit).Scan(&executions).Error

	return executions, err
}

func (r *CommandExecutionRepository) GetLastCommandByShellPID(
	shellPID int,
) (*models.CommandExecution, error) {

	var execution models.CommandExecution

	err := r.db.
		Where("shell_pid = ?", shellPID).
		Order("id DESC").
		First(&execution).Error

	if err != nil {
		return nil, err
	}

	return &execution, nil
}

func (r *CommandExecutionRepository) FetchForEmbedding(limit int) ([]models.CommandExecution, error) {
	var commands []models.CommandExecution

	err := r.db.
		Model(&models.CommandExecution{}).
		Select("command_executions.id, command_executions.command").
		Joins("JOIN embedding_queue q ON q.command_execution_id = command_executions.id").
		Order("command_executions.id ASC").
		Limit(limit).
		Find(&commands).Error

	if err != nil {
		return nil, err
	}

	return commands, nil
}
