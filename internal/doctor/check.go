package doctor

import (
	"io"
	"net/http"
	"recall/internal/config"
	"recall/internal/storage"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func checkConfig(cfg *config.Config) CheckResult {
	if cfg == nil {
		return CheckResult{
			Name: "Config",
			OK:   false,
			Fix:  "Run: recall init",
		}
	}

	return CheckResult{
		Name: "Config loaded",
		OK:   true,
	}
}

func checkDB(cfg *config.Config) CheckResult {
	path, err := storage.DBPath(cfg)
	if err != nil {
		return CheckResult{
			Name:    "Database path",
			OK:      false,
			Message: err.Error(),
		}
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return CheckResult{
			Name:    "Database connection",
			OK:      false,
			Message: err.Error(),
		}
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	return CheckResult{
		Name:    "Database connected",
		OK:      true,
		Message: path,
	}
}

func checkTables(cfg *config.Config) CheckResult {
	path, _ := storage.DBPath(cfg)

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return CheckResult{Name: "Tables", OK: false}
	}

	var count int64

	// command_executions
	db.Raw(`
		SELECT count(*) FROM sqlite_master 
		WHERE type='table' AND name='command_executions'
	`).Scan(&count)

	if count == 0 {
		return CheckResult{
			Name: "command_executions table",
			OK:   false,
			Fix:  "Run: recall init",
		}
	}

	// FTS table
	db.Raw(`
		SELECT count(*) FROM sqlite_master 
		WHERE type='table' AND name='command_executions_fts'
	`).Scan(&count)

	if count == 0 {
		return CheckResult{
			Name: "FTS5 table",
			OK:   false,
			Fix:  "Run migrations / recall init",
		}
	}

	return CheckResult{
		Name: "Database tables",
		OK:   true,
	}
}

func checkEmbeddingProvider(cfg *config.Config) CheckResult {

	if cfg.Embedding.EmbeddingProvider != "ollama" {
		return CheckResult{
			Name: "Embedding provider",
			OK:   true,
		}
	}

	url := cfg.Embedding.OllamaEmbeddingBaseURL

	resp, err := http.Get(url)
	if err != nil {
		return CheckResult{
			Name: "Ollama",
			OK:   false,
			Fix:  "Run: ollama serve",
		}
	}
	defer resp.Body.Close()

	return CheckResult{
		Name: "Ollama reachable",
		OK:   true,
	}
}

func checkEmbeddingModel(cfg *config.Config) CheckResult {
	if cfg.Embedding.EmbeddingProvider != "ollama" {
		return CheckResult{
			Name: "Embedding model",
			OK:   true,
		}
	}

	url := cfg.Embedding.OllamaEmbeddingBaseURL + "/api/tags"

	resp, err := http.Get(url)
	if err != nil {
		return CheckResult{
			Name: "Embedding model",
			OK:   false,
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if !strings.Contains(string(body), cfg.Embedding.OllamaEmbeddingModel) {
		return CheckResult{
			Name: "Embedding model",
			OK:   false,
			Fix:  "Run: ollama pull " + cfg.Embedding.OllamaEmbeddingModel,
		}
	}

	return CheckResult{
		Name: "Embedding model available",
		OK:   true,
	}
}
