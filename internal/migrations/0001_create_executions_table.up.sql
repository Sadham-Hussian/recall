CREATE TABLE IF NOT EXISTS command_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    cwd TEXT,
    exit_code INTEGER,
    shell_pid INTEGER,
    session_id TEXT,
    UNIQUE(command, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_timestamp ON command_executions(timestamp);
CREATE INDEX IF NOT EXISTS idx_shell_session ON command_executions(shell_pid, timestamp DESC);

CREATE VIRTUAL TABLE IF NOT EXISTS command_executions_fts
USING fts5(command, content='command_executions', content_rowid='id');


CREATE TABLE IF NOT EXISTS command_chains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    prev_command TEXT NOT NULL,
    next_command TEXT NOT NULL,
    session_id TEXT NOT NULL,
    occurrence_count INTEGER DEFAULT 1
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_chain
ON command_chains(prev_command, next_command, session_id);

CREATE TABLE IF NOT EXISTS recall_metadata (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE IF NOT EXISTS embedding_queue (
    command_execution_id INTEGER PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (command_execution_id)
        REFERENCES command_executions(id)
        ON DELETE CASCADE
);

CREATE TABLE command_embeddings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    command_execution_id INTEGER NOT NULL,
    model TEXT NOT NULL,
    dimensions INTEGER NOT NULL,
    embedding BLOB NOT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (command_execution_id)
        REFERENCES command_executions(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_command_embeddings_unique
ON command_embeddings(command_execution_id, model);

CREATE INDEX idx_command_embeddings_command_execution_id
ON command_embeddings(command_execution_id);

CREATE INDEX idx_command_embeddings_model
ON command_embeddings(model);