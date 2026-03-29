# Recall

Recall is an AI-powered, local-first terminal CLI that turns your shell history into a searchable memory and retrieval layer.

Today, Recall records commands, groups them into sessions, learns command sequences, and optionally adds semantic retrieval with Ollama embeddings. Over time, the goal is to evolve Recall into a deeper AI-native terminal workflow system with chat, intent detection, reasoning, and smarter command assistance.

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
- SQLite FTS5 support via the provided build tag
- Optional: Ollama if you want semantic retrieval

#### Build

```bash
make install
```

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

## What v1 Can Do

- Automatically record commands from `zsh`, `bash`, and `fish`
- Import existing shell history
- Search command history
- Fall back to fuzzy matching when exact search misses
- Rank results using frequency, recency, success rate, cwd/project context, and session context
- Group commands into sessions
- Replay previous sessions
- Suggest likely next commands based on historical command chains
- Run semantic command search with Ollama embeddings

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

### 4. Process the embedding queue manually

```bash
recall embed
```

In v1, embedding generation is a manual step. The current flow is to record commands first and then run `recall embed` when you want to process the pending embedding queue.

Later, Recall will likely move this into a background daemon or worker so embeddings can be processed automatically.

### 5. Query by intent

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

## v1 Notes

This is a version 1 release.

Recall already works as a useful local retrieval system for terminal commands and workflows, but the bigger AI-native product direction is still ahead.

Current constraints to be aware of:

- semantic retrieval currently supports Ollama only
- embedding generation is manual via `recall embed` in v1
- chat, explicit reasoning, and richer LLM intent detection are future-facing, not fully shipped in v1
- command execution features are interactive and intended for trusted local usage
- imported shell history may not include cwd or exit code metadata
- ranking and session heuristics are practical, but still evolving

## 👤 Who is this for

- Developers working heavily in terminal
- DevOps / SRE workflows
- Anyone tired of searching shell history

## License

[MIT](LICENSE)
