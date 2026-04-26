# Recall

Recall is an AI-powered, local-first terminal CLI that turns your shell history into a searchable memory and retrieval layer.

Today, Recall records commands, groups them into sessions, learns command sequences, supports reusable workflows, and optionally adds semantic retrieval with Ollama embeddings. Over time, the goal is to evolve Recall into a deeper AI-native terminal workflow system with chat, intent detection, reasoning, and smarter command assistance.

## Installation

### macOS

For both Apple Silicon and Intel macOS:

```bash
brew install Sadham-Hussian/recall/recall
```

### Linux

Linux install currently supports `amd64` only:

```bash
curl -sSL https://raw.githubusercontent.com/Sadham-Hussian/recall/master/install.sh | bash
```

### Build from source

#### Prerequisites

- Go `1.24.1` or newer
- GCC (CGO is required for SQLite)
- SQLite FTS5 support via the provided build tag
- Optional: Ollama if you want semantic retrieval

#### Build

```bash
make install
```

### Upgrading

**From v1.1.0 onwards:**

```bash
recall upgrade
```

**From v1.0.0 (one-time bootstrap):**

Re-run your original install method — `install.sh`, `brew upgrade Sadham-Hussian/recall/recall`, or `git pull && make install`. After reaching v1.1.0, all future upgrades go through `recall upgrade`.

## Why Recall

Shell history is tied to one file per shell, hard to search by intent, and weak at showing workflows instead of isolated commands. Recall makes command history structured, searchable, and contextual — and exposes it as a retrieval layer for AI-assisted tooling.

## What Recall Can Do

- Automatically record commands from `zsh`, `bash`, and `fish`
- Import existing shell history
- Search command history with full-text and fuzzy matching
- Rank results using frequency, recency, success rate, cwd/project context, and session context
- Group commands into sessions, name them, and replay them
- Suggest likely next commands based on historical command chains
- Save and run reusable command workflows
- Run semantic command search with Ollama embeddings
- Auto-process embeddings via a background daemon
- Self-upgrade to the latest release
- Tab-complete commands, subcommands, flags, and workflow names
- Plug into AI coding agents (Claude Code, Cursor, Windsurf, Codex, Claude Desktop) via MCP
- AI-powered command explanation via local Ollama LLM
- Export and import command history as JSON
- Usage statistics — top commands, top directories, most-failed commands
- Configurable ignore list for noisy or sensitive commands

## How It Works

Recall stores command executions in SQLite with a full-text search index for keyword retrieval. Each command is linked to a session, contributes to command-chain suggestions, and (if embeddings are enabled) is queued for vectorization. Semantic queries combine FTS and vector similarity to answer intent-based questions like `recall ask "how did I port-forward postgres"`.

## Quick Start

### 1. Initialize Recall

```bash
recall init
```

This creates the default config at `~/.recall/config.yaml`, ensures the database exists, and runs migrations.

### 2. Enable shell integration

For `zsh`:

```bash
eval "$(recall hook zsh)"
```

For `bash`:

```bash
eval "$(recall hook bash)"
```

For `fish`:

```fish
recall hook fish | source
```

To persist it, add the matching command to your shell rc file.

### 3. Import existing history

```bash
recall history
```

### 4. Search your command history

```bash
recall search docker
recall search kubectl logs --full
recall list --limit 20
recall last
```

### 5. Use semantic retrieval

```bash
recall ask "find the command I used to tail docker logs"
recall ask "how did I port forward postgres"
```

### 6. Explore sessions and workflow memory

```bash
recall session
recall session --last 5
recall session replay <session_id>
recall session name <session_id> "deploy debug"
recall continue
```

### 7. Workflows

Save recurring command sequences and replay them later:

```bash
# Interactive save — type commands one by one, type 'save' to finish
recall workflow save deploy

# Save from a session — pick commands by number
recall workflow save debug --from-session <session_id>

# List, show, run, delete
recall workflow list
recall workflow show deploy
recall workflow run deploy
recall workflow delete deploy
```

### 8. Shell completions (optional)

Enable Tab-completion of `recall` subcommands, flags, and saved workflow names:

**Zsh:**

```zsh
recall completion zsh > "${fpath[1]}/_recall"
autoload -U compinit && compinit
```

**Bash:**

```bash
# Requires bash-completion package: apt install bash-completion (if not installed)
mkdir -p ~/.local/share/bash-completion/completions
recall completion bash > ~/.local/share/bash-completion/completions/recall
```

**Fish:**

```fish
recall completion fish > ~/.config/fish/completions/recall.fish
```

## MCP Integration (AI Coding Agents)

Recall ships an MCP server (`recall mcp serve`) so AI coding agents can read and write your terminal history through the Model Context Protocol. Local stdio transport — no network.

### Setup

One command per supported client:

```bash
recall mcp setup claude-code
recall mcp setup claude-desktop
recall mcp setup cursor
recall mcp setup windsurf
recall mcp setup codex
```

Each setup command prints a tailored note about whether you should configure that agent to call `recall_record` (see below).

### Available tools

| Tool | What it does |
|---|---|
| `recall_search` | Full-text + fuzzy search of command history |
| `recall_list` | List recent commands |
| `recall_record` | Record a command into history |
| `recall_session_list` / `recall_session_show` | Browse sessions |
| `recall_stats` | Usage statistics |
| `recall_workflow_list` / `recall_workflow_show` | Saved workflows |
| `recall_suggest` | Suggest the next command from chain history |
| `recall_explain` | AI-powered command explanation |

### Should I configure my agent to call `recall_record`?

Depends on whether the agent runs commands in your interactive terminal or its own subshell:

- **Interactive-shell agents** (Cursor terminal mode, Windsurf Cascade): your shell hook already captures their commands. **Do not** configure them to call `recall_record` — it creates duplicate entries.
- **Non-interactive subshell agents** (Claude Code, Codex, Claude Desktop with shell MCP): the shell hook does not fire on those subshells. **Configure** the agent (system prompt, `CLAUDE.md`, or equivalent rules file) to call `recall_record` after each command it runs.

The `source` of each MCP-recorded command is auto-detected from the MCP handshake (`clientInfo.name`) — agents cannot override it.

## AI Command Explanation

Get a plain-English explanation of a shell command using a local Ollama LLM.

### 1. Pull a model

```bash
ollama pull llama3.2
```

### 2. Enable in config

In `~/.recall/config.yaml`:

```yaml
explain:
  is_explain_enabled: true
  provider: "ollama"
  model: "llama3.2"
  base_url: "http://localhost:11434"
  timeout_seconds: 30
```

### 3. Run

```bash
recall explain "kubectl get pods -n prod -o jsonpath='{.items[*].metadata.name}'"
```

The same explanation is also available to AI agents via the `recall_explain` MCP tool.

## Stats

Show usage statistics — overview, top commands, top command groups, most-failed commands, and top directories.

```bash
recall stats                    # all-time
recall stats --days 7           # last 7 days
recall stats --format md        # markdown output
recall stats --format json      # JSON output
```

## Export / Import

Back up your command history, migrate between machines, or share data via JSON snapshots.

```bash
recall export                       # write JSON to stdout
recall export -o recall.json        # write to a file
recall export --days 30             # export only last 30 days

recall import recall.json           # merge into existing data
recall import recall.json --replace # wipe existing data first
```

## Background Embedding Daemon

When semantic retrieval is enabled, Recall can automatically process the embedding queue in the background via a daemon. This replaces the manual `recall embed` step.

### Install as a system service

```bash
recall daemon install
```

This installs and starts a background service via launchd (macOS) or systemd (Linux). The daemon polls for unprocessed commands and generates embeddings automatically.

### Manage the daemon

```bash
recall daemon status    # show service status
recall daemon stop      # stop the service
recall daemon start     # start the service
recall daemon run       # run in foreground (for debugging)
```

Log file: `~/.recall/daemon.log`

## Self-Update

Recall checks for new releases in the background and prints a one-line notice when a newer version is available.

```bash
recall upgrade              # download, verify checksum, and swap binary
recall upgrade --check      # check only, don't install
recall upgrade -y           # skip confirmation
```

Homebrew users should use `brew upgrade Sadham-Hussian/recall/recall` instead.

Set `auto_check_enabled: false` in `~/.recall/config.yaml` under `upgrade:` to disable the background version check.

## Troubleshooting

### Homebrew shows wrong version or "already installed"

If `brew upgrade` reports `recall 64 already installed` or shows an incorrect version, the Cellar has stale data from an older formula. Fix with a clean reinstall:

```bash
brew uninstall --force Sadham-Hussian/recall/recall
rm -rf "$(brew --cellar)/recall"
brew install Sadham-Hussian/recall/recall
brew list --versions recall    # should show the correct version
```

### CGO / SQLite errors on Linux

If you see `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work`, install GCC and rebuild:

```bash
apt-get install -y gcc
CGO_ENABLED=1 make build
```

## Semantic Retrieval With Ollama

Semantic retrieval is optional and disabled by default.

### 1. Start Ollama

```bash
ollama serve
```

### 2. Pull an embedding model

```bash
ollama pull nomic-embed-text
```

### 3. Enable embeddings in config

Update `~/.recall/config.yaml`:

```yaml
embedding:
  is_embed_enabled: true
  embedding_provider: "ollama"
  ollama_embedding_model: "nomic-embed-text"
  ollama_embedding_base_url: "http://localhost:11434"
  ollama_http_timeout_in_sec: 10
```

### 4. Process embeddings

With the daemon running (`recall daemon install`), embeddings are processed automatically. Or process manually:

```bash
recall embed
```

### 5. Query by intent

```bash
recall ask "find the command I used to tail docker logs"
recall ask "how did I port forward postgres"
```

## Commands

Run `recall <command> --help` for full flags. Quick reference:

| Command | Purpose |
|---|---|
| **Setup** | |
| `recall init` | initialize config, database, and migrations |
| `recall hook <shell>` | print shell integration for `zsh`, `bash`, or `fish` |
| `recall config` / `recall migrate` / `recall doctor` | ensure config / run migrations / health check |
| `recall completion <shell>` | generate shell completion script |
| `recall version` | print version |
| **History & Search** | |
| `recall history` | import existing shell history |
| `recall last` / `recall list` | show most recent / list recent commands |
| `recall search <query>` | full-text + fuzzy search |
| `recall ask <query>` | semantic search (Ollama) |
| `recall suggest <command>` | suggest likely next commands |
| `recall embed` | manually process embedding queue |
| **Sessions & Workflows** | |
| `recall session [--last N]` | show current or recent sessions |
| `recall session replay <id>` | replay a recorded session |
| `recall session name <id> [label]` | name a session or show its name |
| `recall continue` | suggest next command for the current workflow |
| `recall workflow save <name>` | save a workflow (interactive or `--from-session <id>`) |
| `recall workflow list` / `show` / `run` / `delete` | manage saved workflows |
| **AI Integrations** | |
| `recall mcp serve` | start MCP server (stdio, used by AI clients) |
| `recall mcp setup <client>` | configure `claude-code`, `claude-desktop`, `cursor`, `windsurf`, `codex` |
| `recall explain <command>` | AI-powered command explanation (Ollama) |
| **Insights & Backup** | |
| `recall stats [--days N] [--format md\|json]` | usage statistics |
| `recall export [-o file] [--days N]` | export history as JSON |
| `recall import <file> [--replace]` | import history from JSON |
| **Daemon & Upgrade** | |
| `recall daemon install` / `start` / `stop` / `status` / `run` / `log` | manage embedding daemon |
| `recall upgrade [--check] [-y]` | upgrade to the latest release |

## Default Config

```yaml
embedding:
  is_embed_enabled: false
  embedding_provider: "ollama"
  ollama_embedding_model: "nomic-embed-text"
  ollama_embedding_base_url: "http://localhost:11434"
  ollama_http_timeout_in_sec: 10

database:
  path: "~/.recall/recall.db"

search:
  top_k: 10
  default_search_limit: 100
  fuzzy_search_limit: 500
  suggestion_limit: 10
  fts_search_limit: 100
  semantic_candidate_limit: 10000

session:
  gap_seconds: 600
  autocomplete_limit: 100

processor:
  batch_size: 5000

daemon:
  poll_interval_seconds: 30

upgrade:
  auto_check_enabled: true
  check_interval_hours: 24

ignore:
  commands:
    - cd
    - ls
    - clear
    - pwd
    - exit
    - whoami
    - history
    - mkdir
    - touch
    - nano
  patterns:
    - "^export .*(TOKEN|SECRET|PASSWORD|API_KEY|CREDENTIALS)="
    - "^curl .*-H.*(Authorization|Bearer)"

explain:
  is_explain_enabled: false
  provider: "ollama"
  model: "llama3.2"
  base_url: "http://localhost:11434"
  timeout_seconds: 30
```

## Notes

Recall already works as a useful local retrieval system for terminal commands and workflows, but the bigger AI-native product direction is still ahead.

Current constraints to be aware of:

- semantic retrieval currently supports Ollama only
- chat, explicit reasoning, and richer LLM intent detection are future-facing
- command execution features are interactive and intended for trusted local usage
- imported shell history may not include cwd or exit code metadata
- ranking and session heuristics are practical, but still evolving

## Who is this for

- Developers working heavily in terminal
- DevOps / SRE workflows
- Anyone tired of searching shell history

## License

[MIT](LICENSE)
