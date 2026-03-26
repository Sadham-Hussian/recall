package session

import (
	"errors"
	"fmt"
	"recall/internal/config"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type SessionService struct {
	CommandExecutionRepository *repositories.CommandExecutionRepository
	CommandChainRepository     *repositories.CommandChainRepository
}

func NewSessionService() (*SessionService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &SessionService{
		CommandExecutionRepository: repositories.NewCommandExecutionRepository(db),
		CommandChainRepository:     repositories.NewCommandChainRepository(db),
	}, nil
}

func (s *SessionService) GetCurrentSessionByShellPID(shellPID int) (string, []models.CommandExecution, error) {
	sessionID, err := s.CommandExecutionRepository.GetCurrentSessionByShellPID(shellPID)
	if err != nil || sessionID == "" {
		return "", nil, errors.New("No active session found")
	}

	commands, err := s.CommandExecutionRepository.GetCommandsBySessionID(sessionID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch session commands: %v", err)
	}

	return sessionID, commands, nil
}

func (s *SessionService) GetLastCommand(cfg *config.Config, lastExec *models.CommandExecution, shellPID int, ts int64) (string, error) {

	var sessionID string

	if lastExec == nil {
		// first command in shell
		sessionID = generateSessionID(shellPID, ts)
	} else {
		gap := ts - lastExec.Timestamp
		if gap <= int64(cfg.Session.GapSeconds) {
			sessionID = lastExec.SessionID
		} else {
			sessionID = generateSessionID(shellPID, ts)
		}
	}
	return sessionID, nil
}

func generateSessionID(shellPID int, ts int64) string {
	return fmt.Sprintf("%d_%d", shellPID, ts)
}

func (s *SessionService) GetLastSessions(limit int) ([]models.Session, error) {

	executions, err := s.CommandExecutionRepository.
		GetLastSessionCommands(limit)

	if err != nil {
		return nil, err
	}

	sessionOrder := []string{}
	sessionMap := map[string][]models.CommandExecution{}

	for _, e := range executions {

		if _, exists := sessionMap[e.SessionID]; !exists {
			sessionOrder = append(sessionOrder, e.SessionID)
		}

		sessionMap[e.SessionID] = append(sessionMap[e.SessionID], e)
	}

	var sessions []models.Session

	for _, sid := range sessionOrder {
		sessions = append(sessions, models.Session{
			SessionID: sid,
			Commands:  sessionMap[sid],
		})
	}

	return sessions, nil
}

func (s *SessionService) GetCommandsBySessionID(
	sessionID string,
) ([]models.CommandExecution, error) {

	commands, err := s.CommandExecutionRepository.
		GetCommandsBySessionID(sessionID)

	if err != nil {
		return nil, err
	}

	return commands, nil
}

func (s *SessionService) GetNextCommandSuggestion(
	shellPID int,
) (*models.CommandChain, error) {

	lastCommand, err := s.CommandExecutionRepository.
		GetLastCommandByShellPID(shellPID)

	if err != nil {
		return nil, err
	}

	if lastCommand == nil {
		return nil, nil
	}

	results, err := s.CommandChainRepository.
		GetNextCommands(lastCommand.Command, 1)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}
