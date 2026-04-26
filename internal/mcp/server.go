package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"recall",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s)

	return s
}

func registerTools(s *server.MCPServer) {
	// Search
	s.AddTool(mcp.NewTool("recall_search",
		mcp.WithDescription("Search command history using full-text and fuzzy matching"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
	), handleSearch)

	// List recent commands
	s.AddTool(mcp.NewTool("recall_list",
		mcp.WithDescription("List recent commands from terminal history"),
		mcp.WithNumber("limit", mcp.Description("Number of commands (default 20)")),
		mcp.WithString("source", mcp.Description("Filter by source: shell-hook, claude-code, cursor, etc.")),
	), handleList)

	// Record a command
	s.AddTool(mcp.NewTool("recall_record",
		mcp.WithDescription("Record a command execution to recall history. Call this after running shell commands so the user's history stays complete."),
		mcp.WithString("command", mcp.Required(), mcp.Description("The command that was executed")),
		mcp.WithNumber("exit_code", mcp.Required(), mcp.Description("Exit code of the command")),
		mcp.WithString("cwd", mcp.Required(), mcp.Description("Working directory where the command was run")),
		mcp.WithString("source", mcp.Description("Source identifier, e.g. claude-code, cursor (default: mcp)")),
	), handleRecord)

	// Session list
	s.AddTool(mcp.NewTool("recall_session_list",
		mcp.WithDescription("List recent command sessions with names"),
		mcp.WithNumber("limit", mcp.Description("Number of sessions (default 10)")),
	), handleSessionList)

	// Session show
	s.AddTool(mcp.NewTool("recall_session_show",
		mcp.WithDescription("Show commands in a specific session"),
		mcp.WithString("session", mcp.Required(), mcp.Description("Session ID or name")),
	), handleSessionShow)

	// Stats
	s.AddTool(mcp.NewTool("recall_stats",
		mcp.WithDescription("Show usage statistics for terminal history"),
		mcp.WithNumber("days", mcp.Description("Limit stats to last N days (default: all time)")),
	), handleStats)

	// Workflow list
	s.AddTool(mcp.NewTool("recall_workflow_list",
		mcp.WithDescription("List all saved command workflows"),
	), handleWorkflowList)

	// Workflow show
	s.AddTool(mcp.NewTool("recall_workflow_show",
		mcp.WithDescription("Show steps in a saved workflow"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Workflow name")),
	), handleWorkflowShow)

	// Suggest
	s.AddTool(mcp.NewTool("recall_suggest",
		mcp.WithDescription("Suggest the next command based on command chain history"),
		mcp.WithString("command", mcp.Required(), mcp.Description("The current/last command")),
		mcp.WithNumber("limit", mcp.Description("Number of suggestions (default 5)")),
	), handleSuggest)

	// Explain
	s.AddTool(mcp.NewTool("recall_explain",
		mcp.WithDescription("Explain what a shell command does using AI"),
		mcp.WithString("command", mcp.Required(), mcp.Description("The command to explain")),
	), handleExplain)
}
