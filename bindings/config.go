/*
 * Config Binding - Exposes configuration operations to the frontend
 */
package bindings

import (
	"context"

	"github.com/PGshen/thinking-map-desktop/services"
)

// ConfigBinding exposes config operations to the frontend
type ConfigBinding struct {
	configService *services.ConfigService
	onLLMChanged  func(cfg services.LLMConfig) // callback when LLM config changes
}

// InitConfigBinding initializes the config binding with its dependencies
func InitConfigBinding(b *ConfigBinding, configService *services.ConfigService, onLLMChanged func(cfg services.LLMConfig)) {
	b.configService = configService
	b.onLLMChanged = onLLMChanged
}

// LLMConfigResponse is the response for LLM config (api_key is masked)
type LLMConfigResponse struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// GetLLMConfig returns the current LLM configuration
func (b *ConfigBinding) GetLLMConfig(ctx context.Context) (LLMConfigResponse, error) {
	cfg := b.configService.GetLLMConfig()
	return LLMConfigResponse{
		BaseURL: cfg.BaseURL,
		APIKey:  maskAPIKey(cfg.APIKey),
		Model:   cfg.Model,
	}, nil
}

// SaveLLMConfigRequest is the request to save LLM config
type SaveLLMConfigRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// SaveLLMConfig saves the LLM configuration
func (b *ConfigBinding) SaveLLMConfig(ctx context.Context, req SaveLLMConfigRequest) error {
	// If API key is masked (unchanged), keep the original
	existing := b.configService.GetLLMConfig()
	apiKey := req.APIKey
	if isMasked(apiKey) {
		apiKey = existing.APIKey
	}

	cfg := services.LLMConfig{
		BaseURL: req.BaseURL,
		APIKey:  apiKey,
		Model:   req.Model,
	}

	if err := b.configService.SaveLLMConfig(cfg); err != nil {
		return err
	}

	// Notify that LLM config changed
	if b.onLLMChanged != nil {
		b.onLLMChanged(cfg)
	}

	return nil
}

// TestLLMConfig tests the LLM configuration by making a simple API call
func (b *ConfigBinding) TestLLMConfig(ctx context.Context, req SaveLLMConfigRequest) (string, error) {
	// If API key is masked, use the existing one
	existing := b.configService.GetLLMConfig()
	apiKey := req.APIKey
	if isMasked(apiKey) {
		apiKey = existing.APIKey
	}

	client := services.NewLLMClient(&services.LLMConfig{
		BaseURL: req.BaseURL,
		APIKey:  apiKey,
		Model:   req.Model,
	})

	result, err := client.Chat([]services.ChatMessage{
		{Role: "user", Content: "Say 'OK' if you can hear me."},
	})
	if err != nil {
		return "", err
	}

	return result, nil
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func isMasked(key string) bool {
	return key != "" && len(key) > 4 && key[4:8] == "****"
}
