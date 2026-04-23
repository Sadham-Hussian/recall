package models

import "fmt"

type ExportData struct {
	Version       int                  `json:"version"`
	ExportedAt    string               `json:"exported_at"`
	RecallVersion string               `json:"recall_version"`
	Commands      []ExportCommand      `json:"commands"`
	CommandChains []ExportCommandChain `json:"command_chains"`
	Workflows     []ExportWorkflow     `json:"workflows"`
	SessionNames  []ExportSessionName  `json:"session_names"`
}

type ExportCommand struct {
	Command   string `json:"command"`
	Timestamp int64  `json:"timestamp"`
	CWD       string `json:"cwd"`
	ExitCode  int    `json:"exit_code"`
	ShellPID  int    `json:"shell_pid"`
	SessionID string `json:"session_id"`
}

type ExportCommandChain struct {
	PrevCommand     string `json:"prev_command"`
	NextCommand     string `json:"next_command"`
	SessionID       string `json:"session_id"`
	OccurrenceCount int    `json:"occurrence_count"`
}

type ExportWorkflow struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

type ExportSessionName struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
}

type ImportResult struct {
	CommandsImported     int
	CommandErrors        int
	ChainsImported       int
	ChainErrors          int
	WorkflowsImported    int
	WorkflowErrors       int
	SessionNamesImported int
	SessionNameErrors    int
}

func (r *ImportResult) String() string {
	s := fmt.Sprintf("Imported: %d commands, %d chains, %d workflows, %d session names",
		r.CommandsImported, r.ChainsImported, r.WorkflowsImported, r.SessionNamesImported)
	errors := r.CommandErrors + r.ChainErrors + r.WorkflowErrors + r.SessionNameErrors
	if errors > 0 {
		s += fmt.Sprintf(" (%d errors skipped)", errors)
	}
	return s
}
