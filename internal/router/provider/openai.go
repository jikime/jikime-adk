package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"jikime-adk/internal/router/types"
)

// OpenAI provider implementation.
type OpenAI struct {
	cfg *ProviderConfig
}

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(cfg *ProviderConfig) *OpenAI {
	return &OpenAI{cfg: cfg}
}

func (o *OpenAI) Name() string { return "openai" }

func (o *OpenAI) Headers(apiKey string) map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
	}
}

// --- Request Transformation ---

// openaiRequest represents OpenAI chat completions request.
type openaiRequest struct {
	Model               string           `json:"model"`
	Messages            []openaiMessage  `json:"messages"`
	MaxTokens           *int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int             `json:"max_completion_tokens,omitempty"`
	Temperature         *float64         `json:"temperature,omitempty"`
	TopP                *float64         `json:"top_p,omitempty"`
	Stream              bool             `json:"stream"`
	Tools               []openaiTool     `json:"tools,omitempty"`
	ToolChoice          any              `json:"tool_choice,omitempty"`
	Stop                []string         `json:"stop,omitempty"`
	StreamOptions       *streamOptions   `json:"stream_options,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openaiMessage struct {
	Role       string         `json:"role"`
	Content    any            `json:"content"` // string or []openaiContentPart
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type openaiContentPart struct {
	Type     string          `json:"type"` // text, image_url
	Text     string          `json:"text,omitempty"`
	ImageURL *openaiImageURL `json:"image_url,omitempty"`
}

type openaiImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type openaiToolCall struct {
	Index    int                `json:"index"`
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"` // function
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string             `json:"type"` // function
	Function openaiToolDef      `json:"function"`
}

type openaiToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// modelMaxTokens returns the maximum output tokens for a given model.
func modelMaxTokens(model string) int {
	limits := map[string]int{
		"gpt-4o":            16384,
		"gpt-4o-mini":       16384,
		"gpt-4-turbo":       4096,
		"gpt-4":             8192,
		"gpt-3.5-turbo":     4096,
		"gpt-5.1":           100000,
		"gpt-5.1-codex":     100000,
		"o1":                100000,
		"o1-mini":           65536,
		"o1-preview":        32768,
		"o3":                100000,
		"o3-mini":           100000,
		"o4-mini":           100000,
	}

	// Check exact match first
	if limit, ok := limits[model]; ok {
		return limit
	}

	// Check prefix match (e.g., "gpt-4o-2024-08-06")
	for prefix, limit := range limits {
		if len(model) > len(prefix) && model[:len(prefix)+1] == prefix+"-" {
			return limit
		}
	}

	// Default: safe limit for unknown OpenAI models
	return 16384
}

// usesMaxCompletionTokens returns true if the model requires max_completion_tokens
// instead of max_tokens. Newer OpenAI models (o-series, gpt-5.x) use this parameter.
func usesMaxCompletionTokens(model string) bool {
	prefixes := []string{"o1", "o3", "o4", "gpt-5"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// TransformRequest converts Anthropic request to OpenAI format.
func (o *OpenAI) TransformRequest(req *types.AnthropicRequest, model string) (*http.Request, error) {
	oReq := &openaiRequest{
		Model:       model,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.StopSequences,
	}

	if req.MaxTokens > 0 {
		if usesMaxCompletionTokens(model) {
			// For newer models (o-series, gpt-5.x), don't set max_completion_tokens.
			// These models return 400 if the limit is too low for a complete response.
			// Let the model use its own default output limit.
		} else {
			maxTokens := req.MaxTokens
			limit := modelMaxTokens(model)
			if maxTokens > limit {
				maxTokens = limit
			}
			oReq.MaxTokens = &maxTokens
		}
	}

	if req.Stream {
		oReq.StreamOptions = &streamOptions{IncludeUsage: true}
	}

	// Convert system message
	sysText, err := types.ParseSystem(req.System)
	if err != nil {
		return nil, fmt.Errorf("parse system: %w", err)
	}

	var messages []openaiMessage
	if sysText != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: sysText})
	}

	// Convert messages
	for _, msg := range req.Messages {
		converted, err := o.convertMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("convert message: %w", err)
		}
		messages = append(messages, converted...)
	}
	oReq.Messages = messages

	// Convert tools
	if len(req.Tools) > 0 {
		oReq.Tools = o.convertTools(req.Tools)
	}

	// Convert tool_choice
	if len(req.ToolChoice) > 0 {
		oReq.ToolChoice = o.convertToolChoice(req.ToolChoice)
	}

	// Create HTTP request
	body, err := json.Marshal(oReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := o.endpoint()
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	return httpReq, nil
}

func (o *OpenAI) endpoint() string {
	baseURL := o.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return baseURL + "/chat/completions"
}

// convertMessage converts a single Anthropic message to OpenAI format.
func (o *OpenAI) convertMessage(msg types.AnthropicMessage) ([]openaiMessage, error) {
	blocks, err := types.ParseContent(msg.Content)
	if err != nil {
		return nil, err
	}

	var results []openaiMessage

	if msg.Role == "user" {
		result := openaiMessage{Role: "user"}

		// Check if we have mixed content types
		hasNonText := false
		for _, b := range blocks {
			if b.Type == "tool_result" {
				hasNonText = true
				break
			}
		}

		if hasNonText {
			// Split tool_results into separate messages
			for _, b := range blocks {
				switch b.Type {
				case "tool_result":
					toolMsg := openaiMessage{
						Role:       "tool",
						ToolCallID: b.ToolUseID,
					}
					// Parse tool result content
					content, _ := types.ParseContent(b.Content)
					if len(content) > 0 {
						toolMsg.Content = content[0].Text
					} else {
						toolMsg.Content = ""
					}
					results = append(results, toolMsg)
				case "text":
					results = append(results, openaiMessage{Role: "user", Content: b.Text})
				case "image":
					parts := o.convertImageBlock(b)
					results = append(results, openaiMessage{Role: "user", Content: parts})
				}
			}
		} else {
			// Simple text or image content
			parts := o.convertContentParts(blocks)
			if len(parts) == 1 && parts[0].Type == "text" {
				result.Content = parts[0].Text
			} else {
				result.Content = parts
			}
			results = append(results, result)
		}
	} else if msg.Role == "assistant" {
		result := openaiMessage{Role: "assistant"}

		var textParts []string
		var toolCalls []openaiToolCall

		for _, b := range blocks {
			switch b.Type {
			case "text":
				textParts = append(textParts, b.Text)
			case "tool_use":
				inputJSON, _ := json.Marshal(b.Input)
				if b.Input == nil {
					inputJSON = []byte("{}")
				}
				toolCalls = append(toolCalls, openaiToolCall{
					ID:   b.ID,
					Type: "function",
					Function: openaiToolFunction{
						Name:      b.Name,
						Arguments: string(inputJSON),
					},
				})
			}
		}

		if len(textParts) > 0 {
			combined := ""
			for _, t := range textParts {
				combined += t
			}
			result.Content = combined
		}
		if len(toolCalls) > 0 {
			result.ToolCalls = toolCalls
		}
		results = append(results, result)
	}

	return results, nil
}

// convertContentParts converts Anthropic content blocks to OpenAI content parts.
func (o *OpenAI) convertContentParts(blocks []types.ContentBlock) []openaiContentPart {
	var parts []openaiContentPart
	for _, b := range blocks {
		switch b.Type {
		case "text":
			parts = append(parts, openaiContentPart{Type: "text", Text: b.Text})
		case "image":
			parts = append(parts, o.convertImageBlock(b)...)
		}
	}
	return parts
}

// convertImageBlock converts an Anthropic image block to OpenAI format.
func (o *OpenAI) convertImageBlock(b types.ContentBlock) []openaiContentPart {
	if b.Source == nil {
		return nil
	}
	var url string
	if b.Source.Type == "base64" {
		url = fmt.Sprintf("data:%s;base64,%s", b.Source.MediaType, b.Source.Data)
	} else {
		url = b.Source.URL
	}
	return []openaiContentPart{{
		Type:     "image_url",
		ImageURL: &openaiImageURL{URL: url},
	}}
}

// convertTools converts Anthropic tools to OpenAI format.
func (o *OpenAI) convertTools(tools []types.Tool) []openaiTool {
	var result []openaiTool
	for _, t := range tools {
		result = append(result, openaiTool{
			Type: "function",
			Function: openaiToolDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	return result
}

// convertToolChoice converts Anthropic tool_choice to OpenAI format.
func (o *OpenAI) convertToolChoice(raw json.RawMessage) any {
	var tc types.ToolChoice
	if err := json.Unmarshal(raw, &tc); err != nil {
		return "auto"
	}
	switch tc.Type {
	case "auto":
		return "auto"
	case "any":
		return "required"
	case "none":
		return "none"
	case "tool":
		return map[string]any{
			"type":     "function",
			"function": map[string]string{"name": tc.Name},
		}
	default:
		return "auto"
	}
}

// --- Streaming Response Transformation ---

// openaiStreamChunk represents an OpenAI streaming chunk.
type openaiStreamChunk struct {
	ID      string               `json:"id"`
	Choices []openaiStreamChoice `json:"choices"`
	Usage   *openaiUsage         `json:"usage,omitempty"`
}

type openaiStreamChoice struct {
	Delta        openaiStreamDelta `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
	Index        int               `json:"index"`
}

type openaiStreamDelta struct {
	Role      string           `json:"role,omitempty"`
	Content   *string          `json:"content,omitempty"`
	ToolCalls []openaiToolCall `json:"tool_calls,omitempty"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TransformStreamChunk converts an OpenAI SSE chunk to Anthropic SSE events.
func (o *OpenAI) TransformStreamChunk(data []byte, state *StreamState) ([]SSEOutput, error) {
	var chunk openaiStreamChunk
	if err := json.Unmarshal(data, &chunk); err != nil {
		return nil, fmt.Errorf("parse chunk: %w", err)
	}

	var events []SSEOutput

	// Handle usage update (comes with stream_options.include_usage)
	if chunk.Usage != nil {
		state.InputTokens = chunk.Usage.PromptTokens
		state.OutputTokens = chunk.Usage.CompletionTokens
	}

	// Send message_start on first chunk
	if !state.Started {
		state.Started = true
		msgStart := &types.MessageStartEvent{
			Type: "message_start",
			Message: &types.AnthropicResponse{
				ID:    state.MessageID,
				Type:  "message",
				Role:  "assistant",
				Model: state.Model,
				Content: []types.ContentBlock{},
				Usage: &types.Usage{
					InputTokens:  state.InputTokens,
					OutputTokens: 0,
				},
			},
		}
		events = append(events, marshalSSE("message_start", msgStart))
	}

	if len(chunk.Choices) == 0 {
		return events, nil
	}

	choice := chunk.Choices[0]
	delta := choice.Delta

	// Handle text content
	if delta.Content != nil && *delta.Content != "" {
		if !state.TextStarted {
			state.TextStarted = true
			blockStart := &types.ContentBlockStartEvent{
				Type:  "content_block_start",
				Index: state.ContentIndex,
				ContentBlock: &types.ContentBlock{
					Type: "text",
					Text: "",
				},
			}
			events = append(events, marshalSSE("content_block_start", blockStart))
		}

		blockDelta := &types.ContentBlockDeltaEvent{
			Type:  "content_block_delta",
			Index: state.ContentIndex,
			Delta: &types.BlockDelta{
				Type: "text_delta",
				Text: *delta.Content,
			},
		}
		events = append(events, marshalSSE("content_block_delta", blockDelta))
	}

	// Handle tool calls
	for _, tc := range delta.ToolCalls {
		tcState, exists := state.ToolCalls[tc.Index]
		if !exists {
			// New tool call - close text block if open
			if state.TextStarted && len(state.ToolCalls) == 0 {
				events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
					Type:  "content_block_stop",
					Index: state.ContentIndex,
				}))
				state.ContentIndex++
			}

			tcState = &ToolCallState{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Index: state.ContentIndex + len(state.ToolCalls),
			}
			state.ToolCalls[tc.Index] = tcState
		}

		// Update tool call state
		if tc.ID != "" {
			tcState.ID = tc.ID
		}
		if tc.Function.Name != "" {
			tcState.Name = tc.Function.Name
		}
		tcState.Arguments += tc.Function.Arguments

		if !tcState.Started && tcState.Name != "" {
			tcState.Started = true
			blockStart := &types.ContentBlockStartEvent{
				Type:  "content_block_start",
				Index: tcState.Index,
				ContentBlock: &types.ContentBlock{
					Type: "tool_use",
					ID:   tcState.ID,
					Name: tcState.Name,
				},
			}
			events = append(events, marshalSSE("content_block_start", blockStart))
		}

		// Stream arguments
		if tc.Function.Arguments != "" && tcState.Started {
			blockDelta := &types.ContentBlockDeltaEvent{
				Type:  "content_block_delta",
				Index: tcState.Index,
				Delta: &types.BlockDelta{
					Type:        "input_json_delta",
					PartialJSON: tc.Function.Arguments,
				},
			}
			events = append(events, marshalSSE("content_block_delta", blockDelta))
		}
	}

	// Handle finish
	if choice.FinishReason != nil {
		state.Finished = true
		stopReason := convertFinishReason(*choice.FinishReason)

		// Close text block
		if state.TextStarted {
			events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: 0,
			}))
		}

		// Close tool call blocks
		for _, tc := range state.ToolCalls {
			if tc.Started {
				events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
					Type:  "content_block_stop",
					Index: tc.Index,
				}))
			}
		}

		// message_delta
		events = append(events, marshalSSE("message_delta", &types.MessageDeltaEvent{
			Type: "message_delta",
			Delta: &types.MessageDelta{
				StopReason: stopReason,
			},
			Usage: &types.Usage{
				OutputTokens: state.OutputTokens,
			},
		}))

		// message_stop
		events = append(events, marshalSSE("message_stop", &types.MessageStopEvent{
			Type: "message_stop",
		}))
	}

	return events, nil
}

// --- Non-streaming Response Transformation ---

// openaiResponse represents OpenAI chat completions response.
type openaiResponse struct {
	ID      string             `json:"id"`
	Choices []openaiChoice     `json:"choices"`
	Usage   *openaiUsage       `json:"usage,omitempty"`
	Model   string             `json:"model"`
}

type openaiChoice struct {
	Message      openaiMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// TransformResponse converts OpenAI response to Anthropic format.
func (o *OpenAI) TransformResponse(body []byte) (*types.AnthropicResponse, error) {
	var oResp openaiResponse
	if err := json.Unmarshal(body, &oResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	resp := &types.AnthropicResponse{
		ID:   generateMessageID(),
		Type: "message",
		Role: "assistant",
		Model: o.cfg.Model,
	}

	if len(oResp.Choices) > 0 {
		choice := oResp.Choices[0]
		resp.StopReason = convertFinishReason(choice.FinishReason)

		// Convert content
		if contentStr, ok := choice.Message.Content.(string); ok && contentStr != "" {
			resp.Content = append(resp.Content, types.ContentBlock{
				Type: "text",
				Text: contentStr,
			})
		}

		// Convert tool calls
		for _, tc := range choice.Message.ToolCalls {
			var input json.RawMessage
			if tc.Function.Arguments != "" {
				input = json.RawMessage(tc.Function.Arguments)
			} else {
				input = json.RawMessage("{}")
			}
			resp.Content = append(resp.Content, types.ContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
	}

	if oResp.Usage != nil {
		resp.Usage = &types.Usage{
			InputTokens:  oResp.Usage.PromptTokens,
			OutputTokens: oResp.Usage.CompletionTokens,
		}
	}

	return resp, nil
}

// --- Helper functions ---

// convertFinishReason maps OpenAI finish_reason to Anthropic stop_reason.
func convertFinishReason(reason string) string {
	switch reason {
	case "stop":
		return "end_turn"
	case "tool_calls":
		return "tool_use"
	case "length":
		return "max_tokens"
	case "content_filter":
		return "end_turn"
	default:
		return "end_turn"
	}
}

// marshalSSE creates an SSEOutput with JSON-encoded data.
func marshalSSE(event string, data any) SSEOutput {
	b, _ := json.Marshal(data)
	return SSEOutput{Event: event, Data: b}
}
