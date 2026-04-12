package workflow

import (
	"fmt"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type WorkflowService struct {
	WorkflowRepo         *repositories.WorkflowRepository
	CommandExecutionRepo *repositories.CommandExecutionRepository
}

func NewWorkflowService() (*WorkflowService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &WorkflowService{
		WorkflowRepo:         repositories.NewWorkflowRepository(db),
		CommandExecutionRepo: repositories.NewCommandExecutionRepository(db),
	}, nil
}

func (s *WorkflowService) SaveFromCommands(name, description string, commands []string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands provided")
	}
	return s.save(name, description, commands)
}

func (s *WorkflowService) GetSessionCommands(sessionID string) ([]string, error) {
	commands, err := s.CommandExecutionRepo.GetCommandsBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands found for session %s", sessionID)
	}
	cmds := make([]string, len(commands))
	for i, c := range commands {
		cmds[i] = c.Command
	}
	return cmds, nil
}

func (s *WorkflowService) save(name, description string, commands []string) error {
	w := &models.Workflow{Name: name, Description: description}
	steps := make([]models.WorkflowStep, len(commands))
	for i, cmd := range commands {
		steps[i] = models.WorkflowStep{StepOrder: i + 1, Command: cmd}
	}
	return s.WorkflowRepo.Create(w, steps)
}

func (s *WorkflowService) List() ([]models.Workflow, error) {
	return s.WorkflowRepo.GetAll()
}

func (s *WorkflowService) Show(name string) (*models.Workflow, []models.WorkflowStep, error) {
	w, err := s.WorkflowRepo.GetByName(name)
	if err != nil {
		return nil, nil, err
	}
	steps, err := s.WorkflowRepo.GetSteps(w.ID)
	if err != nil {
		return nil, nil, err
	}
	return w, steps, nil
}

func (s *WorkflowService) Delete(name string) error {
	return s.WorkflowRepo.DeleteByName(name)
}
