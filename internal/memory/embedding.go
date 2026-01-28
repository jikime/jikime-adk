package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// NewEmbeddingProvider creates a provider based on config.
// Returns nil if provider is "none" or no API key is available.
func NewEmbeddingProvider(cfg EmbeddingConfig) (EmbeddingProvider, error) {
	switch cfg.Provider {
	case "none", "":
		return nil, nil
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("openai: API key required")
		}
		return newOpenAIProvider(cfg), nil
	case "gemini":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("gemini: API key required")
		}
		return newGeminiProvider(cfg), nil
	case "auto":
		return autoDetectProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown embedding provider: %s", cfg.Provider)
	}
}

// LoadEmbeddingConfig returns embedding configuration from environment variables.
func LoadEmbeddingConfig() EmbeddingConfig {
	cfg := EmbeddingConfig{
		Provider: os.Getenv("JIKIME_EMBEDDING_PROVIDER"),
	}

	if cfg.Provider == "" {
		cfg.Provider = "auto"
	}

	cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("GEMINI_API_KEY")
	}

	cfg.Model = os.Getenv("JIKIME_EMBEDDING_MODEL")
	cfg.BaseURL = os.Getenv("JIKIME_EMBEDDING_BASE_URL")

	return cfg
}

func autoDetectProvider(cfg EmbeddingConfig) (EmbeddingProvider, error) {
	// Try OpenAI first
	openaiKey := cfg.APIKey
	if openaiKey == "" {
		openaiKey = os.Getenv("OPENAI_API_KEY")
	}
	if openaiKey != "" {
		c := cfg
		c.Provider = "openai"
		c.APIKey = openaiKey
		return newOpenAIProvider(c), nil
	}

	// Try Gemini
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		c := cfg
		c.Provider = "gemini"
		c.APIKey = geminiKey
		return newGeminiProvider(c), nil
	}

	// No provider available
	return nil, nil
}

// --- OpenAI Provider ---

type openAIProvider struct {
	apiKey  string
	model   string
	baseURL string
	dims    int
}

func newOpenAIProvider(cfg EmbeddingConfig) *openAIProvider {
	model := cfg.Model
	if model == "" {
		model = "text-embedding-3-small"
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	dims := cfg.Dims
	if dims == 0 {
		dims = 1536
	}
	return &openAIProvider{
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: baseURL,
		dims:    dims,
	}
}

func (p *openAIProvider) ID() string    { return "openai" }
func (p *openAIProvider) Model() string { return p.model }
func (p *openAIProvider) Dims() int     { return p.dims }

func (p *openAIProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	results, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("openai: empty embedding response")
	}
	return results[0], nil
}

func (p *openAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	body := map[string]interface{}{
		"model": p.model,
		"input": texts,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}

	embeddings := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}
	return embeddings, nil
}

// --- Gemini Provider ---

type geminiProvider struct {
	apiKey  string
	model   string
	baseURL string
	dims    int
}

func newGeminiProvider(cfg EmbeddingConfig) *geminiProvider {
	model := cfg.Model
	if model == "" {
		model = "text-embedding-004"
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1"
	}
	dims := cfg.Dims
	if dims == 0 {
		dims = 768
	}
	return &geminiProvider{
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: baseURL,
		dims:    dims,
	}
}

func (p *geminiProvider) ID() string    { return "gemini" }
func (p *geminiProvider) Model() string { return p.model }
func (p *geminiProvider) Dims() int     { return p.dims }

func (p *geminiProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	results, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("gemini: empty embedding response")
	}
	return results[0], nil
}

func (p *geminiProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	// Gemini batchEmbedContents API
	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Parts []part `json:"parts"`
	}
	type embedReq struct {
		Model   string  `json:"model"`
		Content content `json:"content"`
	}

	requests := make([]embedReq, len(texts))
	for i, t := range texts {
		requests[i] = embedReq{
			Model: "models/" + p.model,
			Content: content{
				Parts: []part{{Text: t}},
			},
		}
	}

	body := map[string]interface{}{
		"requests": requests,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models/%s:batchEmbedContents?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embeddings []struct {
			Values []float32 `json:"values"`
		} `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w", err)
	}

	embeddings := make([][]float32, len(result.Embeddings))
	for i, e := range result.Embeddings {
		embeddings[i] = e.Values
	}
	return embeddings, nil
}
