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
	SessionNameRepository      *repositories.SessionNameRepository
}

func NewSessionService() (*SessionService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &SessionService{
		CommandExecutionRepository: repositories.NewCommandExecutionRepository(db),
		CommandChainRepository:     repositories.NewCommandChainRepository(db),
		SessionNameRepository:      repositories.NewSessionNameRepository(db),
	}, nil
}

func (s *SessionService) GetCurrentSessionByShellPID(shellPID int) (string, string, []models.CommandExecution, error) {
	sessionID, err := s.CommandExecutionRepository.GetCurrentSessionByShellPID(shellPID)
	if err != nil || sessionID == "" {
		return "", "", nil, errors.New("No active session found")
	}

	commands, err := s.CommandExecutionRepository.GetCommandsBySessionID(sessionID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to fetch session commands: %v", err)
	}

	name, _ := s.SessionNameRepository.GetName(sessionID)

	return sessionID, name, commands, nil
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

	// Batch fetch names for all sessions
	names, _ := s.SessionNameRepository.GetNames(sessionOrder)

	var sessions []models.Session

	for _, sid := range sessionOrder {
		sessions = append(sessions, models.Session{
			SessionID: sid,
			Name:      names[sid],
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

// SetSessionName names or renames a session.
func (s *SessionService) SetSessionName(sessionID, name string) error {
	// Verify the session exists
	cmds, err := s.CommandExecutionRepository.GetCommandsBySessionID(sessionID)
	if err != nil {
		return err
	}
	if len(cmds) == 0 {
		return fmt.Errorf("session %s not found", sessionID)
	}
	return s.SessionNameRepository.SetName(sessionID, name)
}

// GetSessionName returns the name for a session.
func (s *SessionService) GetSessionName(sessionID string) (string, error) {
	return s.SessionNameRepository.GetName(sessionID)
}

// ResolveSession takes either a session ID or a session name and returns
// the canonical session ID and its name. Tries ID lookup first (cheap),
// then falls back to name lookup.
func (s *SessionService) ResolveSession(input string) (sessionID, name string, err error) {
	// Try as session ID first
	cmds, err := s.CommandExecutionRepository.GetCommandsBySessionID(input)
	if err == nil && len(cmds) > 0 {
		name, _ = s.SessionNameRepository.GetName(input)
		return input, name, nil
	}

	// Try as name
	sessionID, err = s.SessionNameRepository.GetSessionIDByName(input)
	if err != nil {
		return "", "", fmt.Errorf("no session found for '%s'", input)
	}
	return sessionID, input, nil
}
