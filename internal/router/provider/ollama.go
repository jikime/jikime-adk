package provider

import (
	"net/http"

	"jikime-adk/internal/router/types"
)

// Ollama provider - uses OpenAI-compatible API at localhost.
type Ollama struct {
	*OpenAI
}

// NewOllama creates a new Ollama provider.
func NewOllama(cfg *ProviderConfig) *Ollama {
	return &Ollama{OpenAI: NewOpenAI(cfg)}
}

func (ol *Ollama) Name() string { return "ollama" }

// Headers returns headers without Authorization (local server).
func (ol *Ollama) Headers(_ string) map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// TransformRequest overrides to use Ollama's OpenAI-compatible endpoint.
func (ol *Ollama) TransformRequest(req *types.AnthropicRequest, model string) (*http.Request, error) {
	httpReq, err := ol.OpenAI.TransformRequest(req, model)
	if err != nil {
		return nil, err
	}

	endpoint := ol.endpoint()
	httpReq.URL, _ = httpReq.URL.Parse(endpoint)
	return httpReq, nil
}

func (ol *Ollama) endpoint() string {
	baseURL := ol.cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return baseURL + "/v1/chat/completions"
}
