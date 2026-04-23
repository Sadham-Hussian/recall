package export

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"recall/cmd/setup"
	"recall/internal/storage"
	"recall/internal/storage/models"
	"recall/internal/storage/repositories"
)

type ExportService struct {
	ExportRepo                *repositories.ExportRepository
	CommandExecutionRepo      *repositories.CommandExecutionRepository
	WorkflowRepo              *repositories.WorkflowRepository
	SessionNameRepo           *repositories.SessionNameRepository
	CommandChainRepo          *repositories.CommandChainRepository
	CommandEmbeddingQueueRepo *repositories.CommandEmbeddingQueueRepository
}

func NewExportService() (*ExportService, error) {
	db, err := storage.NewDB()
	if err != nil {
		return nil, err
	}
	return &ExportService{
		ExportRepo:                repositories.NewExportRepository(db),
		CommandExecutionRepo:      repositories.NewCommandExecutionRepository(db),
		WorkflowRepo:              repositories.NewWorkflowRepository(db),
		SessionNameRepo:           repositories.NewSessionNameRepository(db),
		CommandChainRepo:          repositories.NewCommandChainRepository(db),
		CommandEmbeddingQueueRepo: repositories.NewCommandEmbeddingQueueRepository(db),
	}, nil
}

// Export gathers all data and writes JSON to the writer.
func (s *ExportService) Export(w io.Writer, sinceTs int64) error {
	commands, err := s.ExportRepo.AllCommands(sinceTs)
	if err != nil {
		return fmt.Errorf("export commands: %w", err)
	}

	chains, err := s.ExportRepo.AllChains()
	if err != nil {
		return fmt.Errorf("export chains: %w", err)
	}

	workflows, err := s.ExportRepo.AllWorkflows()
	if err != nil {
		return fmt.Errorf("export workflows: %w", err)
	}

	sessionNames, err := s.ExportRepo.AllSessionNames()
	if err != nil {
		return fmt.Errorf("export session names: %w", err)
	}

	data := models.ExportData{
		Version:       1,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		RecallVersion: setup.Version,
		Commands:      commands,
		CommandChains: chains,
		Workflows:     workflows,
		SessionNames:  sessionNames,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// Import reads JSON from the reader and inserts data into the database.
// When replace is true, existing data is wiped first.
func (s *ExportService) Import(r io.Reader, replace bool) (*models.ImportResult, error) {
	var data models.ExportData
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if data.Version != 1 {
		return nil, fmt.Errorf("unsupported export version: %d", data.Version)
	}

	if replace {
		if err := s.ExportRepo.WipeAll(); err != nil {
			return nil, fmt.Errorf("wipe failed: %w", err)
		}
	}

	result := &models.ImportResult{}

	// Import commands
	for _, cmd := range data.Commands {
		exec := &models.CommandExecution{
			Command:   cmd.Command,
			Timestamp: cmd.Timestamp,
			CWD:       cmd.CWD,
			ExitCode:  cmd.ExitCode,
			ShellPID:  cmd.ShellPID,
			SessionID: cmd.SessionID,
		}
		if err := s.CommandExecutionRepo.InsertWithFTS(exec); err != nil {
			result.CommandErrors++
			continue
		}
		result.CommandsImported++

		// Enqueue for embedding
		if exec.ID > 0 {
			_ = s.CommandEmbeddingQueueRepo.Enqueue(int64(exec.ID))
		}
	}

	// Import command chains
	for _, chain := range data.CommandChains {
		if err := s.CommandChainRepo.UpsertWithCount(
			chain.PrevCommand, chain.NextCommand, chain.SessionID, chain.OccurrenceCount,
		); err != nil {
			result.ChainErrors++
			continue
		}
		result.ChainsImported++
	}

	// Import workflows
	for _, wf := range data.Workflows {
		w := &models.Workflow{Name: wf.Name, Description: wf.Description}
		steps := make([]models.WorkflowStep, len(wf.Steps))
		for i, cmd := range wf.Steps {
			steps[i] = models.WorkflowStep{StepOrder: i + 1, Command: cmd}
		}
		if err := s.WorkflowRepo.Create(w, steps); err != nil {
			result.WorkflowErrors++
			continue
		}
		result.WorkflowsImported++
	}

	// Import session names
	for _, sn := range data.SessionNames {
		if err := s.SessionNameRepo.SetName(sn.SessionID, sn.Name); err != nil {
			result.SessionNameErrors++
			continue
		}
		result.SessionNamesImported++
	}

	return result, nil
}
