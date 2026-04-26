package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"recall/internal/config"
	"recall/internal/generation"
	"recall/internal/services/command_execution"
	"recall/internal/services/explain"
	"recall/internal/services/session"
	"recall/internal/services/stats"
	"recall/internal/services/workflow"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func toJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func parseArgs(request mcp.CallToolRequest) map[string]any {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return make(map[string]any)
	}
	return args
}

func getInt(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		if n, ok := v.(float64); ok {
			return int(n)
		}
	}
	return defaultVal
}

func getString(args map[string]any, key, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return defaultVal
}

// logRecordDebug appends a single line per recall_record call to
// ~/.recall/mcp-debug.log capturing process and clientInfo details.
// Temporary — used to diagnose source mis-tagging across MCP clients.
func logRecordDebug(ctx context.Context, args map[string]any, resolvedSource string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".recall")
	_ = os.MkdirAll(dir, 0755)
	f, err := os.OpenFile(filepath.Join(dir, "mcp-debug.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	var clientName, clientVersion, sessionState string
	sess := server.ClientSessionFromContext(ctx)
	switch {
	case sess == nil:
		sessionState = "nil"
	default:
		if withInfo, ok := sess.(server.SessionWithClientInfo); ok {
			info := withInfo.GetClientInfo()
			clientName = info.Name
			clientVersion = info.Version
			sessionState = "with-info"
		} else {
			sessionState = "no-info"
		}
	}

	fmt.Fprintf(f,
		"[%s] pid=%d ppid=%d session=%s clientName=%q clientVersion=%q resolvedSource=%q argsSource=%q command=%q cwd=%q\n",
		time.Now().Format(time.RFC3339),
		os.Getpid(), os.Getppid(),
		sessionState, clientName, clientVersion,
		resolvedSource,
		getString(args, "source", ""),
		getString(args, "command", ""),
		getString(args, "cwd", ""),
	)
}

// clientNameFromContext returns the client's name as captured during the
// MCP initialize handshake (e.g. "claude-code", "cursor"). Returns "mcp"
// if the session doesn't expose client info.
func clientNameFromContext(ctx context.Context) string {
	sess := server.ClientSessionFromContext(ctx)
	if sess == nil {
		return "mcp"
	}
	withInfo, ok := sess.(server.SessionWithClientInfo)
	if !ok {
		return "mcp"
	}
	info := withInfo.GetClientInfo()
	if info.Name == "" {
		return "mcp"
	}
	// Normalize: lowercase, replace spaces with hyphens for consistency
	// e.g. "Claude Code" -> "claude-code"
	name := strings.ToLower(info.Name)
	name = strings.ReplaceAll(name, " ", "-")
	return name
}

// ── recall_search ────────────────────────────────────────────────────────────

func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := config.LoadConfig()
	args := parseArgs(request)
	query := getString(args, "query", "")
	limit := getInt(args, "limit", 20)

	svc, err := command_execution.NewCommandExecutionSearchService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	results, err := svc.Search(cfg, []string{query}, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(toJSON(results)), nil
}

// ── recall_list ──────────────────────────────────────────────────────────────

func handleList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(request)
	limit := getInt(args, "limit", 20)

	svc, err := command_execution.NewCommandExecutionService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	commands, err := svc.ListRecent(limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(toJSON(commands)), nil
}

// ── recall_record ────────────────────────────────────────────────────────────

func handleRecord(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := config.LoadConfig()
	args := parseArgs(request)
	command := getString(args, "command", "")
	exitCode := getInt(args, "exit_code", 0)
	cwd := getString(args, "cwd", "")
	// Default source = whatever client connected (claude-code, cursor, etc.)
	// Falls back to "mcp" if client info isn't exposed by the session.
	// defaultSource := clientNameFromContext(ctx)
	source := clientNameFromContext(ctx)

	// logRecordDebug(ctx, args, source)

	if command == "" {
		return mcp.NewToolResultError("command is required"), nil
	}

	svc, err := command_execution.NewCommandExecutionService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ts := time.Now().Unix()
	_, err = svc.RecordLiveCommandExecution(cfg, command, ts, cwd, exitCode, 0, "", source)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Recorded: %s (exit %d)", command, exitCode)), nil
}

// ── recall_session_list ──────────────────────────────────────────────────────

func handleSessionList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(request)
	limit := getInt(args, "limit", 10)

	svc, err := session.NewSessionService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sessions, err := svc.GetLastSessions(limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	type sessionSummary struct {
		SessionID    string `json:"session_id"`
		Name         string `json:"name,omitempty"`
		CommandCount int    `json:"command_count"`
		FirstCommand string `json:"first_command,omitempty"`
		LastCommand  string `json:"last_command,omitempty"`
	}

	summaries := make([]sessionSummary, len(sessions))
	for i, s := range sessions {
		summaries[i] = sessionSummary{
			SessionID:    s.SessionID,
			Name:         s.Name,
			CommandCount: len(s.Commands),
		}
		if len(s.Commands) > 0 {
			summaries[i].FirstCommand = s.Commands[0].Command
			summaries[i].LastCommand = s.Commands[len(s.Commands)-1].Command
		}
	}

	return mcp.NewToolResultText(toJSON(summaries)), nil
}

// ── recall_session_show ──────────────────────────────────────────────────────

func handleSessionShow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(request)
	input := getString(args, "session", "")

	svc, err := session.NewSessionService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sessionID, name, err := svc.ResolveSession(input)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	commands, err := svc.GetCommandsBySessionID(sessionID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := struct {
		SessionID string `json:"session_id"`
		Name      string `json:"name,omitempty"`
		Commands  []struct {
			Command   string `json:"command"`
			ExitCode  int    `json:"exit_code"`
			CWD       string `json:"cwd"`
			Timestamp int64  `json:"timestamp"`
		} `json:"commands"`
	}{
		SessionID: sessionID,
		Name:      name,
	}

	for _, c := range commands {
		result.Commands = append(result.Commands, struct {
			Command   string `json:"command"`
			ExitCode  int    `json:"exit_code"`
			CWD       string `json:"cwd"`
			Timestamp int64  `json:"timestamp"`
		}{c.Command, c.ExitCode, c.CWD, c.Timestamp})
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

// ── recall_stats ─────────────────────────────────────────────────────────────

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(request)
	days := getInt(args, "days", 0)

	svc, err := stats.NewStatsService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sinceTs int64
	if days > 0 {
		sinceTs = time.Now().AddDate(0, 0, -days).Unix()
	}

	overview, err := svc.Overview(sinceTs)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	topCmds, _ := svc.TopCommands(sinceTs, 10)
	cmdGroups, _ := svc.TopCommandGroups(sinceTs, 7)
	failed, _ := svc.MostFailed(sinceTs, 5, 5)
	topDirs, _ := svc.TopDirectories(sinceTs, 5)

	result := struct {
		Overview      any `json:"overview"`
		CommandGroups any `json:"command_groups"`
		TopCommands   any `json:"top_commands"`
		MostFailed    any `json:"most_failed"`
		TopDirs       any `json:"top_directories"`
	}{overview, cmdGroups, topCmds, failed, topDirs}

	return mcp.NewToolResultText(toJSON(result)), nil
}

// ── recall_workflow_list ─────────────────────────────────────────────────────

func handleWorkflowList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	svc, err := workflow.NewWorkflowService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	workflows, err := svc.List()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(toJSON(workflows)), nil
}

// ── recall_workflow_show ─────────────────────────────────────────────────────

func handleWorkflowShow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArgs(request)
	name := getString(args, "name", "")

	svc, err := workflow.NewWorkflowService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	w, steps, err := svc.Show(name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Steps       []string `json:"steps"`
	}{
		Name:        w.Name,
		Description: w.Description,
	}
	for _, s := range steps {
		result.Steps = append(result.Steps, s.Command)
	}

	return mcp.NewToolResultText(toJSON(result)), nil
}

// ── recall_suggest ───────────────────────────────────────────────────────────

func handleSuggest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := config.LoadConfig()
	args := parseArgs(request)
	command := getString(args, "command", "")
	limit := getInt(args, "limit", 5)

	svc, err := command_execution.NewCommandChainService()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	suggestions, err := svc.GetNextCommands(cfg, command, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(toJSON(suggestions)), nil
}

// ── recall_explain ───────────────────────────────────────────────────────────

func handleExplain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg := config.LoadConfig()
	args := parseArgs(request)
	command := getString(args, "command", "")

	if !cfg.Explain.IsExplainEnabled {
		return mcp.NewToolResultError("explain is disabled in recall config"), nil
	}

	gen := generation.NewOllamaGenerator(
		cfg.Explain.BaseURL,
		cfg.Explain.Model,
		cfg.Explain.TimeoutSeconds,
	)

	svc := explain.NewExplainService(gen)

	stream, err := svc.Explain(ctx, command)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer stream.Close()

	// Read the full streamed response (MCP tools return complete text, not streams)
	body, err := io.ReadAll(stream)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
