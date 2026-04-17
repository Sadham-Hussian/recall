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

## Positioning

If terminal history is raw logs, Recall is the beginning of a memory layer for terminal work.

You can think of v1 as:

- terminal memory for developers
- a local RAG-style retrieval layer for shell workflows
- an AI-powered command recall tool
- a foundation for future terminal chat and reasoning workflows

## Why Recall

Terminal history is useful, but it is usually:

- tied to a single shell history file
- hard to search when you remember intent but not exact syntax
- weak at showing workflows instead of isolated commands
- not designed to help resume interrupted work
- not built as a retrieval layer for AI-assisted tooling

Recall is an attempt to fix that by making command history structured, searchable, contextual, and eventually AI-native.

## What Recall Can Do

- Automatically record commands from `zsh`, `bash`, and `fish`
- Import existing shell history
- Search command history with full-text and fuzzy matching
- Rank results using frequency, recency, success rate, cwd/project context, and session context
- Group commands into sessions and replay them
- Suggest likely next commands based on historical command chains
- Save and run reusable command workflows
- Run semantic command search with Ollama embeddings
- Auto-process embeddings via a background daemon
- Self-upgrade to the latest release
- Tab-complete commands, subcommands, flags, and workflow names

## What Recall Is Becoming

The long-term direction for Recall is larger than command history search.

The vision is to turn Recall into an AI-powered terminal layer that can:

- understand user intent instead of relying only on exact command matches
- use retrieval over terminal history and workflows as context
- support chat-based interaction on top of local command memory
- reason about next steps in a terminal session
- become a practical bridge between raw shell usage and LLM-assisted execution

That future is not fully shipped yet, but v1 already includes the core pieces that make that direction possible: structured command storage, retrieval, sessions, command chains, and embeddings.

## How It Works Today

Recall stores command executions in SQLite and creates a full-text search index for fast keyword retrieval.

Each recorded command can also:

- be linked to a session
- contribute to command-chain suggestions
- be queued for embedding generation

When embeddings are enabled, Recall uses Ollama to generate vectors for commands and combines semantic similarity with FTS results to answer intent-based queries like:

- `recall ask "find docker cleanup command"`
- `recall ask "how did I port-forward kubernetes service"`

In other words, v1 already behaves like a lightweight local retrieval system for terminal actions.

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

### Setup

- `recall init` - initialize config, database, and migrations
- `recall hook <shell>` - print shell integration script for `zsh`, `bash`, or `fish`
- `recall install` - show instructions to enable shell integration
- `recall config` - ensure the config file exists
- `recall migrate` - run database migrations
- `recall doctor` - run local health checks
- `recall version` - print version
- `recall completion <shell>` - generate shell completion script for `bash`, `zsh`, or `fish`

### Recording and history

- `recall history` - import shell history into the database
- `recall record` - internal hidden command used by shell hooks

### Querying and retrieval

- `recall last` - show the most recent command
- `recall list` - list recent commands
- `recall search <query>` - keyword/full-text search with ranking
- `recall suggest <command>` - suggest likely next commands
- `recall ask <query>` - semantic search using embeddings
- `recall embed` - manually process the embedding queue

### Sessions

- `recall session` - show the current shell session
- `recall session --last <n>` - show the latest sessions
- `recall session replay <session_id>` - replay a recorded session
- `recall continue` - suggest the next command for the current shell workflow

### Workflows

- `recall workflow save <name>` - save a workflow (interactive or `--from-session <id>`)
- `recall workflow list` - list all saved workflows
- `recall workflow show <name>` - show steps in a workflow
- `recall workflow run <name>` - execute a saved workflow
- `recall workflow delete <name>` - delete a workflow

### Daemon

- `recall daemon install` - install and start as a system service (launchd/systemd)
- `recall daemon start` - start the daemon service
- `recall daemon stop` - stop the daemon service
- `recall daemon status` - show daemon service status
- `recall daemon run` - run the daemon in the foreground
- `recall daemon log` - tail the daemon log

### Upgrade

- `recall upgrade` - upgrade to the latest release
- `recall upgrade --check` - check for updates without installing
- `recall upgrade -y` - upgrade without confirmation

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

processor:
  batch_size: 5000

daemon:
  poll_interval_seconds: 30

upgrade:
  auto_check_enabled: true
  check_interval_hours: 24
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
