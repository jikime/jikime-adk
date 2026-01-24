package provider

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"jikime-adk/internal/router/types"
)

// Provider defines the interface for LLM API providers.
type Provider interface {
	// Name returns the provider identifier.
	Name() string

	// TransformRequest converts an Anthropic request to the provider's HTTP request.
	TransformRequest(req *types.AnthropicRequest, model string) (*http.Request, error)

	// TransformStreamChunk converts a provider's SSE chunk to Anthropic SSE events.
	TransformStreamChunk(data []byte, state *StreamState) ([]SSEOutput, error)

	// TransformResponse converts a non-streaming provider response to Anthropic format.
	TransformResponse(body []byte) (*types.AnthropicResponse, error)

	// Headers returns the required headers for the provider API.
	Headers(apiKey string) map[string]string
}

// SSEOutput represents a single SSE event to send to the client.
type SSEOutput struct {
	Event string // event type (message_start, content_block_delta, etc.)
	Data  []byte // JSON-encoded event data
}

// StreamState tracks the state of a streaming response transformation.
type StreamState struct {
	MessageID    string
	Model        string
	ContentIndex int
	Started      bool
	TextStarted  bool
	Finished     bool // true if finish events (message_stop) already sent
	ToolCalls    map[int]*ToolCallState
	InputTokens  int
	OutputTokens int
}

// ToolCallState tracks the state of a tool call being streamed.
type ToolCallState struct {
	ID        string
	Name      string
	Arguments string
	Index     int  // Content block index in Anthropic format
	Started   bool
}

// NewStreamState creates a new StreamState.
func NewStreamState(model string) *StreamState {
	return &StreamState{
		MessageID: generateMessageID(),
		Model:     model,
		ToolCalls: make(map[int]*ToolCallState),
	}
}

// ProviderConfig contains provider-specific settings.
type ProviderConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Region  string
}

// NewProvider creates a provider instance by name.
func NewProvider(name string, cfg *ProviderConfig) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAI(cfg), nil
	case "gemini":
		return NewGemini(cfg), nil
	case "glm":
		return NewGLM(cfg), nil
	case "ollama":
		return NewOllama(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// generateMessageID creates a unique message ID in Anthropic format.
func generateMessageID() string {
	return fmt.Sprintf("msg_%s", randomID(24))
}

// randomID generates a random hex string.
func randomID(n int) string {
	b := make([]byte, n/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}
