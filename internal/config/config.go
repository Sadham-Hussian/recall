package config

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

//go:embed default_config.yaml
var defaultConfigFS embed.FS

type Config struct {
	Embedding struct {
		IsEmbedEnabled         bool   `mapstructure:"is_embed_enabled"`
		EmbeddingProvider      string `mapstructure:"embedding_provider"`
		OllamaEmbeddingModel   string `mapstructure:"ollama_embedding_model"`
		OllamaEmbeddingBaseURL string `mapstructure:"ollama_embedding_base_url"`
		OllamaHttpTimeoutInSec int    `mapstructure:"ollama_http_timeout_in_sec"`
	} `mapstructure:"embedding"`

	Database struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	Search struct {
		TopK                   int `mapstructure:"top_k"`
		DefaultSearchLimit     int `mapstructure:"default_search_limit"`
		FuzzySearchLimit       int `mapstructure:"fuzzy_search_limit"`
		SuggestionLimit        int `mapstructure:"suggestion_limit"`
		FTSSearchLimit         int `mapstructure:"fts_search_limit"`
		SemanticCandidateLimit int `mapstructure:"semantic_candidate_limit"`
	} `mapstructure:"search"`

	Session struct {
		GapSeconds int `mapstructure:"gap_seconds"`
	} `mapstructure:"session"`

	Processor struct {
		BatchSize int `mapstructure:"batch_size"`
	} `mapstructure:"processor"`

	Daemon struct {
		PollIntervalSeconds int `mapstructure:"poll_interval_seconds"`
	} `mapstructure:"daemon"`

	Upgrade struct {
		AutoCheckEnabled   bool `mapstructure:"auto_check_enabled"`
		CheckIntervalHours int  `mapstructure:"check_interval_hours"`
	} `mapstructure:"upgrade"`
}

var AppConfig *Config

func LoadConfig() *Config {
	if AppConfig != nil {
		return AppConfig
	}
	path, err := ensureConfigFile()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	AppConfig = &config
	return &config
}

func getUserConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".recall", "config.yaml"), nil
}

func getEmbeddedDefaultConfig() ([]byte, error) {
	data, err := defaultConfigFS.ReadFile("default_config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}
	return data, nil
}

func copyDefaultConfig(dst string) error {
	data, err := getEmbeddedDefaultConfig()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Config not found. Created default config at:", dst)
	fmt.Println("You can edit this file to customize recall.")

	return nil
}

func ensureConfigFile() (string, error) {
	path, err := getUserConfigPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := copyDefaultConfig(path)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

func ReloadConfig() (*Config, error) {
	path, err := getUserConfigPath()
	if err != nil {
		return AppConfig, err
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return AppConfig, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return AppConfig, err
	}

	AppConfig = &cfg
	return AppConfig, nil
}

func GetUserConfigPath() (string, error) {
	return getUserConfigPath()
}

func EnsureConfigFile() (string, error) {
	return ensureConfigFile()
}
