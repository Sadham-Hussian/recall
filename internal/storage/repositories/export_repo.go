package repositories

import (
	"fmt"
	"recall/internal/storage/models"

	"gorm.io/gorm"
)

type ExportRepository struct {
	db *gorm.DB
}

func NewExportRepository(db *gorm.DB) *ExportRepository {
	return &ExportRepository{db: db}
}

// AllCommands returns all command executions, optionally filtered by sinceTs.
func (r *ExportRepository) AllCommands(sinceTs int64) ([]models.ExportCommand, error) {
	var commands []models.ExportCommand
	q := r.db.Table("command_executions").
		Select("command, timestamp, cwd, exit_code, shell_pid, session_id").
		Order("timestamp ASC")
	if sinceTs > 0 {
		q = q.Where("timestamp >= ?", sinceTs)
	}
	err := q.Find(&commands).Error
	return commands, err
}

// AllChains returns all command chains.
func (r *ExportRepository) AllChains() ([]models.ExportCommandChain, error) {
	var chains []models.ExportCommandChain
	err := r.db.Table("command_chains").
		Select("prev_command, next_command, session_id, occurrence_count").
		Find(&chains).Error
	return chains, err
}

// AllWorkflows returns all workflows with their steps inlined.
func (r *ExportRepository) AllWorkflows() ([]models.ExportWorkflow, error) {
	var workflows []models.Workflow
	if err := r.db.Order("name ASC").Find(&workflows).Error; err != nil {
		return nil, err
	}

	result := make([]models.ExportWorkflow, 0, len(workflows))
	for _, w := range workflows {
		var steps []models.WorkflowStep
		r.db.Where("workflow_id = ?", w.ID).Order("step_order ASC").Find(&steps)

		cmds := make([]string, len(steps))
		for i, s := range steps {
			cmds[i] = s.Command
		}
		result = append(result, models.ExportWorkflow{
			Name:        w.Name,
			Description: w.Description,
			Steps:       cmds,
		})
	}
	return result, nil
}

// AllSessionNames returns all session name mappings.
func (r *ExportRepository) AllSessionNames() ([]models.ExportSessionName, error) {
	var names []models.ExportSessionName
	err := r.db.Table("session_names").
		Select("session_id, name").
		Find(&names).Error
	return names, err
}

// WipeAll deletes all data from all tables in the correct order
// (respecting foreign key dependencies).
func (r *ExportRepository) WipeAll() error {
	tables := []string{
		"command_embeddings",
		"embedding_queue",
		"command_chains",
		"workflow_steps",
		"workflows",
		"session_names",
		"command_executions_fts",
		"command_executions",
	}
	for _, t := range tables {
		if err := r.db.Exec("DELETE FROM " + t).Error; err != nil {
			return fmt.Errorf("wipe %s: %w", t, err)
		}
	}
	return nil
}
