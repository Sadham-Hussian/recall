package command_execution

import (
	"recall/internal/config"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type CommandChainService struct {
	CommandChainRepository *repositories.CommandChainRepository
}

func NewCommandChainService() (*CommandChainService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}

	return &CommandChainService{
		CommandChainRepository: repositories.NewCommandChainRepository(db),
	}, nil
}

func (s *CommandChainService) GetNextCommands(cfg *config.Config, command string, limit int) ([]models.CommandChain, error) {
	if limit == 0 {
		limit = cfg.Search.SuggestionLimit
	}
	return s.CommandChainRepository.GetNextCommands(command, limit)
}
