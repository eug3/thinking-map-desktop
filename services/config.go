/*
 * Config Service - Manages application configuration (LLM settings)
 * Stores config as JSON file in the data directory
 */
package services

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig holds the complete application configuration
type AppConfig struct {
	LLM LLMConfig `json:"llm"`
}

// ConfigService manages configuration persistence
type ConfigService struct {
	configPath string
	config     AppConfig
	mu         sync.RWMutex
}

// NewConfigService creates a new config service and loads existing config
func NewConfigService(dataDir string) *ConfigService {
	configPath := filepath.Join(dataDir, "config.json")
	cs := &ConfigService{
		configPath: configPath,
	}
	cs.load()
	return cs
}

// GetConfig returns the current configuration
func (s *ConfigService) GetConfig() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetLLMConfig returns the LLM configuration
func (s *ConfigService) GetLLMConfig() LLMConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.LLM
}

// SaveLLMConfig saves the LLM configuration
func (s *ConfigService) SaveLLMConfig(cfg LLMConfig) error {
	s.mu.Lock()
	s.config.LLM = cfg
	s.mu.Unlock()
	return s.save()
}

// load reads config from disk
func (s *ConfigService) load() {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No config file found, using defaults", "path", s.configPath)
			return
		}
		slog.Error("Failed to read config file", "error", err)
		return
	}

	if err := json.Unmarshal(data, &s.config); err != nil {
		slog.Error("Failed to parse config file", "error", err)
	}
}

// save writes config to disk
func (s *ConfigService) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.config, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0600)
}
