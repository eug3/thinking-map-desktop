/*
 * LLM Client - OpenAI-compatible API client for AI calls
 * Supports any OpenAI-compatible endpoint (OpenAI, DeepSeek, etc.)
 */
package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LLMConfig holds the LLM configuration
type LLMConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatCompletionChoice represents a choice in the response
type ChatCompletionChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Delta   ChatMessage `json:"delta"`
}

// ChatCompletionResponse represents a non-streaming response
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Choices []ChatCompletionChoice `json:"choices"`
}

// LLMClient provides OpenAI-compatible API calls
type LLMClient struct {
	config     *LLMConfig
	httpClient *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(config *LLMConfig) *LLMClient {
	return &LLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// IsConfigured returns true if the LLM client has valid configuration
func (c *LLMClient) IsConfigured() bool {
	return c.config != nil &&
		c.config.BaseURL != "" &&
		c.config.APIKey != "" &&
		c.config.Model != ""
}

// UpdateConfig updates the LLM configuration
func (c *LLMClient) UpdateConfig(config *LLMConfig) {
	c.config = config
}

// StreamChat sends a streaming chat completion request
// The callback is called for each content chunk
func (c *LLMClient) StreamChat(messages []ChatMessage, callback func(chunk string) error) error {
	if !c.IsConfigured() {
		return fmt.Errorf("LLM is not configured. Please set base_url, api_key and model in Settings")
	}

	reqBody := ChatCompletionRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := strings.TrimRight(c.config.BaseURL, "/")
	url := baseURL + "/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	slog.Info("LLM streaming request", "url", url, "model", c.config.Model)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("Failed to parse SSE chunk", "error", err, "data", data)
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := callback(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	return nil
}

// Chat sends a non-streaming chat completion request
func (c *LLMClient) Chat(messages []ChatMessage) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("LLM is not configured. Please set base_url, api_key and model in Settings")
	}

	reqBody := ChatCompletionRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := strings.TrimRight(c.config.BaseURL, "/")
	url := baseURL + "/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Message.Content, nil
}
