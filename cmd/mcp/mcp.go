package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"recall/internal/config"
	mcpserver "recall/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server for AI coding agents",
}

// ── serve ────────────────────────────────────────────────────────────────────

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server (stdio transport)",
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		s := mcpserver.NewServer()
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("mcp server error: %v", err)
		}
	},
}

// ── setup ────────────────────────────────────────────────────────────────────

var setupCmd = &cobra.Command{
	Use:   "setup <client>",
	Short: "Configure an AI client to use recall's MCP server",
	Long: `Supported clients: claude-desktop, claude-code, cursor, windsurf, codex

Examples:
  recall mcp setup claude-desktop
  recall mcp setup claude-code
  recall mcp setup cursor`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := args[0]

		binaryPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("could not determine recall binary path: %w", err)
		}
		binaryPath, _ = filepath.EvalSymlinks(binaryPath)

		switch client {
		case "claude-desktop":
			return setupClaudeDesktop(binaryPath)
		case "claude-code":
			return setupClaudeCode(binaryPath)
		case "cursor":
			return setupCursor(binaryPath)
		case "windsurf":
			return setupWindsurf(binaryPath)
		case "codex":
			return setupCodex(binaryPath)
		default:
			return fmt.Errorf("unknown client: %s (supported: claude-desktop, claude-code, cursor, windsurf, codex)", client)
		}
	},
}

// ── Claude Desktop ───────────────────────────────────────────────────────────

func setupClaudeDesktop(binaryPath string) error {
	var configPath string
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "linux":
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "claude", "claude_desktop_config.json")
	default:
		return fmt.Errorf("claude desktop setup not supported on %s", runtime.GOOS)
	}

	return writeJSONMCPConfig(configPath, binaryPath, "claude-desktop")
}

// ── Claude Code ──────────────────────────────────────────────────────────────

func setupClaudeCode(binaryPath string) error {
	cmd := exec.Command("claude", "mcp", "add", "recall", "--", binaryPath, "mcp", "serve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude mcp add failed: %w (is claude-code installed?)", err)
	}
	fmt.Println("✔ Recall MCP server added to Claude Code")
	printRecordGuidance("claude-code")
	return nil
}

// ── Cursor ───────────────────────────────────────────────────────────────────

func setupCursor(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".cursor", "mcp.json")
	return writeJSONMCPConfig(configPath, binaryPath, "cursor")
}

// ── Windsurf ─────────────────────────────────────────────────────────────────

func setupWindsurf(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
	return writeJSONMCPConfig(configPath, binaryPath, "windsurf")
}

// ── Codex ────────────────────────────────────────────────────────────────────

func setupCodex(binaryPath string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codex", "mcp.json")
	return writeJSONMCPConfig(configPath, binaryPath, "codex")
}

// ── Shared config writer ─────────────────────────────────────────────────────

func writeJSONMCPConfig(configPath, binaryPath, clientName string) error {
	// Read existing config or start fresh
	var configMap map[string]any

	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &configMap)
	}
	if configMap == nil {
		configMap = make(map[string]any)
	}

	// Ensure mcpServers key exists
	servers, ok := configMap["mcpServers"].(map[string]any)
	if !ok {
		servers = make(map[string]any)
	}

	// Add recall entry
	servers["recall"] = map[string]any{
		"command": binaryPath,
		"args":    []string{"mcp", "serve"},
	}
	configMap["mcpServers"] = servers

	// Write back
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return err
	}

	fmt.Printf("✔ Recall MCP server configured for %s\n", clientName)
	fmt.Printf("  Config: %s\n", configPath)
	fmt.Printf("  Restart %s to activate.\n", clientName)
	printRecordGuidance(clientName)
	return nil
}

// ── Registration ─────────────────────────────────────────────────────────────

func GetMCPCmd() *cobra.Command {
	mcpCmd.AddCommand(serveCmd)
	mcpCmd.AddCommand(setupCmd)
	return mcpCmd
}

// printRecordGuidance prints whether the client needs explicit recall_record
// configuration based on whether it runs commands in the user's interactive
// shell (where recall's hook already captures them) or its own subshell.
func printRecordGuidance(clientName string) {
	interactiveShellClients := map[string]bool{
		"windsurf": true,
		"cursor":   true,
	}

	fmt.Println()
	if interactiveShellClients[clientName] {
		fmt.Println("ℹ This client runs commands in your interactive terminal, which is")
		fmt.Println("  already captured by recall's shell hook. Do NOT instruct the agent")
		fmt.Println("  to call recall_record — it would create duplicate entries.")
	} else {
		fmt.Println("ℹ This client runs commands in a non-interactive subshell, which")
		fmt.Println("  recall's shell hook does NOT capture. To record agent-executed")
		fmt.Println("  commands, instruct the agent (via system prompt / CLAUDE.md /")
		fmt.Println("  equivalent rules file) to call the `recall_record` MCP tool")
		fmt.Println("  after each command it runs.")
	}
}
