package command_execution

import (
	"errors"
	"fmt"
	"recall/internal/shell"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
	"strings"
)

type CommandExecutionService struct {
	CommandExecutionRepository      *repositories.CommandExecutionRepository
	CommandChainRepository          *repositories.CommandChainRepository
	CommandEmbeddingQueueRepository *repositories.CommandEmbeddingQueueRepository
}

func NewCommandExecutionService() (*CommandExecutionService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &CommandExecutionService{
		CommandExecutionRepository:      repositories.NewCommandExecutionRepository(db),
		CommandChainRepository:          repositories.NewCommandChainRepository(db),
		CommandEmbeddingQueueRepository: repositories.NewCommandEmbeddingQueueRepository(db),
	}, nil
}

func (s *CommandExecutionService) Last() (*models.CommandExecution, error) {
	execution, err := s.CommandExecutionRepository.Last()
	if err != nil {
		return nil, err
	}
	return execution, nil
}

func (s *CommandExecutionService) ListRecent(limit int) ([]models.CommandExecution, error) {
	executions, err := s.CommandExecutionRepository.ListRecent(limit)
	if err != nil {
		return nil, err
	}
	return executions, nil
}

func (s *CommandExecutionService) RecordLiveCommandExecution(cmdStr string,
	timestamp int64, cwd string, exitCode int, shellPID int, sessionID string) (*models.CommandExecution, error) {

	execution := &models.CommandExecution{
		Command:   strings.TrimSpace(cmdStr),
		Timestamp: timestamp,
		CWD:       cwd,
		ExitCode:  exitCode,
		ShellPID:  shellPID,
		SessionID: sessionID,
	}
	err := s.CommandExecutionRepository.InsertWithFTS(execution)
	if err != nil {
		return nil, err
	}

	prevCmd, err := s.CommandExecutionRepository.
		GetPreviousCommandByID(sessionID, execution.ID)

	if err != nil {
		return nil, err
	}

	if prevCmd != "" {

		err = s.CommandChainRepository.Upsert(prevCmd, execution.Command, sessionID)
		if err != nil {
			return nil, err
		}
	}

	if execution.ID > 0 {
		err = s.CommandEmbeddingQueueRepository.Enqueue(int64(execution.ID))
		if err != nil {
			return nil, err
		}
	}

	return execution, nil
}

func (s *CommandExecutionService) RecordCommandHistory() (int, error) {
	sh, err := shell.Detect()
	if err != nil {
		return 0, errors.New("failed to detect shell : " + err.Error())
	}

	fmt.Println("Detected shell:", sh.Name())

	entries, err := sh.ReadHistory()
	if err != nil {
		return 0, errors.New("failed to read history: " + err.Error())
	}

	imported := 0

	for _, e := range entries {

		commandExecutionModel := &models.CommandExecution{
			Command:   e.Command,
			Timestamp: e.Timestamp.Unix(),
			CWD:       "", // unknown from history file
			ExitCode:  0,  // unknown
		}

		err := s.CommandExecutionRepository.InsertWithFTS(commandExecutionModel)

		if err == nil {
			imported++
		}

		// Add to embedding queue
		err = s.CommandEmbeddingQueueRepository.Enqueue(int64(commandExecutionModel.ID))
		if err != nil {
			return 0, err
		}
	}
	return imported, nil
}
