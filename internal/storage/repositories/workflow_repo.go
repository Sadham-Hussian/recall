package repositories

import (
	"recall/internal/storage/models"

	"gorm.io/gorm"
)

type WorkflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

func (r *WorkflowRepository) Create(w *models.Workflow, steps []models.WorkflowStep) error {
	tx := r.db.Begin()
	if err := tx.Create(w).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range steps {
		steps[i].WorkflowID = w.ID
	}
	if err := tx.Create(&steps).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *WorkflowRepository) GetByName(name string) (*models.Workflow, error) {
	var w models.Workflow
	err := r.db.Where("name = ?", name).First(&w).Error
	return &w, err
}

func (r *WorkflowRepository) GetAll() ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := r.db.Order("created_at DESC").Find(&workflows).Error
	return workflows, err
}

func (r *WorkflowRepository) GetSteps(workflowID uint) ([]models.WorkflowStep, error) {
	var steps []models.WorkflowStep
	err := r.db.Where("workflow_id = ?", workflowID).Order("step_order ASC").Find(&steps).Error
	return steps, err
}

func (r *WorkflowRepository) DeleteByName(name string) error {
	tx := r.db.Begin()
	var w models.Workflow
	if err := tx.Where("name = ?", name).First(&w).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("workflow_id = ?", w.ID).Delete(&models.WorkflowStep{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&w).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
