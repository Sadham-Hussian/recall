# Recall

Recall is a local-first Go CLI that remembers what you ran in the terminal so you can search, replay, and reuse it later.

It records shell commands into SQLite, indexes them with FTS5 for fast lookup, groups them into sessions, learns common command sequences, and can optionally add semantic search with Ollama embeddings.

## Why Recall

Terminal history is useful, but it is usually:

- tied to a single shell history file
- hard to search when you only remember intent, not exact syntax
- bad at showing workflows instead of isolated commands
- not great at helping you resume interrupted work

Recall is an attempt to make command history feel more like memory.

## What v1 Can Do

- Automatically record commands from `zsh`, `bash`, and `fish`
- Import existing shell history into a local SQLite database
- Search command history with SQLite FTS5
- Fall back to fuzzy matching when exact search misses
- Rank results using frequency, recency, success rate, cwd/project context, and session context
- Group commands into sessions
- Replay previous sessions
- Suggest likely next commands based on historical command chains
- Run semantic command search with Ollama embeddings
- Run a `doctor` command to validate local setup

## How It Works

Recall stores command executions in SQLite and creates a full-text search index for fast keyword search.

Each recorded command can also:

- be linked to a session
- contribute to command-chain suggestions
- be queued for embedding generation

When embeddings are enabled, Recall uses Ollama to generate vectors for commands and combines semantic similarity with FTS results to answer intent-based queries like:

- `recall ask "find docker cleanup command"`
- `recall ask "how did I port-forward kubernetes service"`

## Tech Stack

- Go
- Cobra CLI
- SQLite
- SQLite FTS5
- GORM
- Ollama for local embeddings

## Project Structure

```text
cmd/                    Cobra commands
internal/config/        config loading and default config
internal/storage/       database setup, migrations, repositories, models
internal/services/      application services for recording, search, sessions, embeddings
internal/search/        hybrid ranking and FTS query helpers
internal/embedding/     embedding client and similarity helpers
internal/shell/         shell detection and history import
```

## Installation

### Prerequisites

- Go `1.24.1` or newer
- SQLite FTS5 support via the provided build tag
- Optional: Ollama if you want semantic search

### Build

```bash
go build -tags sqlite_fts5 -o recall .
```

### Optional install to a global path

```bash
make build
sudo mv recall /usr/local/bin/recall
```

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

### 5. Explore sessions

```bash
recall session
recall session --last 5
recall session replay <session_id>
recall continue
```

## Semantic Search With Ollama

Semantic search is optional and disabled by default.

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

### 4. Process the embedding queue

```bash
recall embed
```

### 5. Ask by intent

```bash
recall ask "find the command I used to tail docker logs"
recall ask "how did I port forward postgres"
```

## Commands

### Setup

- `recall init` - initialize config, database, and migrations
- `recall hook <shell>` - print shell integration script for `zsh`, `bash`, or `fish`
- `recall config` - ensure the config file exists
- `recall migrate` - run database migrations
- `recall doctor` - run local health checks
- `recall version` - print version

### Recording and history

- `recall history` - import shell history into the database
- `recall record` - internal hidden command used by shell hooks

### Querying

- `recall last` - show the most recent command
- `recall list` - list recent commands
- `recall search <query>` - keyword/full-text search with ranking
- `recall suggest <command>` - suggest likely next commands
- `recall ask <query>` - semantic search using embeddings

### Sessions

- `recall session` - show the current shell session
- `recall session --last <n>` - show the latest sessions
- `recall session replay <session_id>` - replay a recorded session
- `recall continue` - suggest the next command for the current shell workflow

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
```

## Example Workflows

### Find and rerun a command

```bash
recall search "kubectl port-forward" --interactive
```

### Recover a past workflow

```bash
recall session --last 3
recall session replay <session_id>
```

### Resume a likely next step

```bash
recall continue
```

## Doctor Checks

`recall doctor` verifies:

- config loads correctly
- database connection works
- required tables exist
- Ollama is reachable when embeddings are enabled
- the configured embedding model is available

## v1 Notes

This is a version 1 release. It is already useful for personal command recall, but it is still early.

Current constraints to be aware of:

- semantic search currently supports Ollama only
- command execution features are interactive and intended for trusted local usage
- imported shell history may not include cwd or exit code metadata
- ranking and session heuristics are practical, but still evolving
- docs and command UX will keep improving as the project matures

## Development

```bash
make build
make run
```

Build verification used for this repo:

```bash
go build -tags sqlite_fts5 ./...
```

## License

[MIT](LICENSE)
