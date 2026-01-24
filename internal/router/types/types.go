package types

import "encoding/json"

// --- Anthropic Request Types ---

// AnthropicRequest represents the Anthropic Messages API request format.
type AnthropicRequest struct {
	Model         string             `json:"model"`
	Messages      []AnthropicMessage `json:"messages"`
	MaxTokens     int                `json:"max_tokens,omitempty"`
	Stream        bool               `json:"stream,omitempty"`
	System        json.RawMessage    `json:"system,omitempty"` // string or []ContentBlock
	Tools         []Tool             `json:"tools,omitempty"`
	ToolChoice    json.RawMessage    `json:"tool_choice,omitempty"`
	Temperature   *float64           `json:"temperature,omitempty"`
	TopP          *float64           `json:"top_p,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Metadata      map[string]any     `json:"metadata,omitempty"`
}

// AnthropicMessage represents a message in the Anthropic API.
type AnthropicMessage struct {
	Role    string          `json:"role"` // user, assistant
	Content json.RawMessage `json:"content"`
}

// ContentBlock represents a content block in Anthropic messages.
type ContentBlock struct {
	Type string `json:"type"` // text, image, tool_use, tool_result

	// text type
	Text string `json:"text,omitempty"`

	// image type
	Source *ImageSource `json:"source,omitempty"`

	// tool_use type
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// tool_result type
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"` // string or []ContentBlock
	IsError   bool            `json:"is_error,omitempty"`

	// cache_control
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// ImageSource represents an image source in Anthropic API.
type ImageSource struct {
	Type      string `json:"type"` // base64, url
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// CacheControl for prompt caching.
type CacheControl struct {
	Type string `json:"type"` // ephemeral
}

// Tool represents a tool definition in Anthropic API.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToolChoice represents tool_choice in Anthropic API.
type ToolChoice struct {
	Type string `json:"type"` // auto, any, tool, none
	Name string `json:"name,omitempty"`
}

// --- Anthropic Response Types ---

// AnthropicResponse represents the Anthropic Messages API response.
type AnthropicResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"` // message
	Role         string         `json:"role"` // assistant
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        *Usage         `json:"usage,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// --- Anthropic SSE Event Types ---

// MessageStartEvent is the first event in a stream.
type MessageStartEvent struct {
	Type    string             `json:"type"` // message_start
	Message *AnthropicResponse `json:"message"`
}

// ContentBlockStartEvent indicates a new content block.
type ContentBlockStartEvent struct {
	Type         string        `json:"type"` // content_block_start
	Index        int           `json:"index"`
	ContentBlock *ContentBlock `json:"content_block"`
}

// ContentBlockDeltaEvent contains incremental content.
type ContentBlockDeltaEvent struct {
	Type  string      `json:"type"` // content_block_delta
	Index int         `json:"index"`
	Delta *BlockDelta `json:"delta"`
}

// BlockDelta represents the delta content in a stream.
type BlockDelta struct {
	Type        string `json:"type"` // text_delta, input_json_delta
	Text        string `json:"text,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
}

// ContentBlockStopEvent indicates end of a content block.
type ContentBlockStopEvent struct {
	Type  string `json:"type"` // content_block_stop
	Index int    `json:"index"`
}

// MessageDeltaEvent contains message-level changes.
type MessageDeltaEvent struct {
	Type  string        `json:"type"` // message_delta
	Delta *MessageDelta `json:"delta"`
	Usage *Usage        `json:"usage,omitempty"`
}

// MessageDelta for message-level streaming updates.
type MessageDelta struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// MessageStopEvent indicates the end of the message stream.
type MessageStopEvent struct {
	Type string `json:"type"` // message_stop
}

// --- Anthropic Error Types ---

// AnthropicError represents an error response from Anthropic API.
type AnthropicError struct {
	Type  string       `json:"type"` // error
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Type    string `json:"type"` // invalid_request_error, authentication_error, etc.
	Message string `json:"message"`
}

// --- Helper Functions ---

// ParseContent parses the content field which can be a string or []ContentBlock.
func ParseContent(raw json.RawMessage) ([]ContentBlock, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// Try string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return []ContentBlock{{Type: "text", Text: s}}, nil
	}

	// Try []ContentBlock
	var blocks []ContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

// ParseSystem parses the system field which can be a string or []ContentBlock.
func ParseSystem(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}

	// Try string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}

	// Try []ContentBlock and concatenate text
	var blocks []ContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return "", err
	}

	var text string
	for _, b := range blocks {
		if b.Type == "text" {
			text += b.Text
		}
	}
	return text, nil
}
