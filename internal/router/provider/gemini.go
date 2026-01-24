package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"jikime-adk/internal/router/types"
)

// Gemini provider implementation.
type Gemini struct {
	cfg *ProviderConfig
}

// NewGemini creates a new Gemini provider.
func NewGemini(cfg *ProviderConfig) *Gemini {
	return &Gemini{cfg: cfg}
}

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) Headers(apiKey string) map[string]string {
	// Gemini uses API key in URL, not in headers
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// --- Gemini Request Types ---

type geminiRequest struct {
	Contents          []geminiContent       `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	Tools             []geminiToolSet       `json:"tools,omitempty"`
	ToolConfig        *geminiToolConfig     `json:"toolConfig,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                `json:"text,omitempty"`
	InlineData       *geminiInlineData     `json:"inlineData,omitempty"`
	FunctionCall     *geminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResp   `json:"functionResponse,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type geminiFunctionResp struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiToolSet struct {
	FunctionDeclarations []geminiFuncDecl `json:"functionDeclarations,omitempty"`
}

type geminiFuncDecl struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type geminiToolConfig struct {
	FunctionCallingConfig *geminiFuncCallConfig `json:"functionCallingConfig,omitempty"`
}

type geminiFuncCallConfig struct {
	Mode                 string   `json:"mode"` // AUTO, NONE, ANY
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

type geminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// geminiMaxTokens returns the maximum output tokens for a given Gemini model.
func geminiMaxTokens(model string) int {
	limits := map[string]int{
		"gemini-2.0-flash":       8192,
		"gemini-2.0-flash-lite":  8192,
		"gemini-1.5-pro":         8192,
		"gemini-1.5-flash":       8192,
		"gemini-2.5-pro":         65536,
		"gemini-2.5-flash":       65536,
	}

	if limit, ok := limits[model]; ok {
		return limit
	}

	// Check prefix match for versioned models
	for prefix, limit := range limits {
		if len(model) > len(prefix) && model[:len(prefix)+1] == prefix+"-" {
			return limit
		}
	}

	return 8192
}

// --- Request Transformation ---

func (g *Gemini) TransformRequest(req *types.AnthropicRequest, model string) (*http.Request, error) {
	gReq := &geminiRequest{}

	// System instruction
	sysText, _ := types.ParseSystem(req.System)
	if sysText != "" {
		gReq.SystemInstruction = &geminiContent{
			Parts: []geminiPart{{Text: sysText}},
		}
	}

	// Convert messages
	contents, err := g.convertMessages(req.Messages)
	if err != nil {
		return nil, fmt.Errorf("convert messages: %w", err)
	}
	gReq.Contents = contents

	// Convert tools
	if len(req.Tools) > 0 {
		gReq.Tools = g.convertTools(req.Tools)
	}

	// Convert tool choice
	if len(req.ToolChoice) > 0 {
		gReq.ToolConfig = g.convertToolChoice(req.ToolChoice)
	}

	// Generation config
	gReq.GenerationConfig = &geminiGenerationConfig{
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if req.MaxTokens > 0 {
		maxTokens := req.MaxTokens
		limit := geminiMaxTokens(model)
		if maxTokens > limit {
			maxTokens = limit
		}
		gReq.GenerationConfig.MaxOutputTokens = &maxTokens
	}
	if len(req.StopSequences) > 0 {
		gReq.GenerationConfig.StopSequences = req.StopSequences
	}

	// Create HTTP request
	body, err := json.Marshal(gReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := g.endpoint(model, req.Stream)
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	return httpReq, nil
}

func (g *Gemini) endpoint(model string, stream bool) string {
	baseURL := g.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	method := "generateContent"
	if stream {
		method = "streamGenerateContent"
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:%s?key=%s", baseURL, model, method, g.cfg.APIKey)
	if stream {
		url += "&alt=sse"
	}
	return url
}

// convertMessages converts Anthropic messages to Gemini contents.
func (g *Gemini) convertMessages(msgs []types.AnthropicMessage) ([]geminiContent, error) {
	var contents []geminiContent

	for _, msg := range msgs {
		blocks, err := types.ParseContent(msg.Content)
		if err != nil {
			return nil, err
		}

		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		var parts []geminiPart

		for _, b := range blocks {
			switch b.Type {
			case "text":
				parts = append(parts, geminiPart{Text: b.Text})

			case "image":
				if b.Source != nil && b.Source.Type == "base64" {
					parts = append(parts, geminiPart{
						InlineData: &geminiInlineData{
							MimeType: b.Source.MediaType,
							Data:     b.Source.Data,
						},
					})
				}

			case "tool_use":
				var args map[string]any
				if b.Input != nil {
					json.Unmarshal(b.Input, &args)
				}
				parts = append(parts, geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: b.Name,
						Args: args,
					},
				})

			case "tool_result":
				resultContent, _ := types.ParseContent(b.Content)
				resultText := ""
				if len(resultContent) > 0 {
					resultText = resultContent[0].Text
				}

				// Find the tool name from the tool_use_id
				toolName := g.findToolName(msgs, b.ToolUseID)

				parts = append(parts, geminiPart{
					FunctionResponse: &geminiFunctionResp{
						Name: toolName,
						Response: map[string]any{
							"result": resultText,
						},
					},
				})
			}
		}

		if len(parts) > 0 {
			contents = append(contents, geminiContent{Role: role, Parts: parts})
		}
	}

	return contents, nil
}

// findToolName searches for the tool name by tool_use_id.
func (g *Gemini) findToolName(msgs []types.AnthropicMessage, toolUseID string) string {
	for _, msg := range msgs {
		blocks, _ := types.ParseContent(msg.Content)
		for _, b := range blocks {
			if b.Type == "tool_use" && b.ID == toolUseID {
				return b.Name
			}
		}
	}
	return "unknown"
}

// convertTools converts Anthropic tools to Gemini functionDeclarations.
func (g *Gemini) convertTools(tools []types.Tool) []geminiToolSet {
	var decls []geminiFuncDecl
	for _, t := range tools {
		params := cleanupParameters(t.InputSchema)
		decls = append(decls, geminiFuncDecl{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  params,
		})
	}
	return []geminiToolSet{{FunctionDeclarations: decls}}
}

// unsupportedSchemaFields lists JSON Schema fields not supported by Gemini API.
var unsupportedSchemaFields = []string{
	"$schema",
	"additionalProperties",
	"exclusiveMinimum",
	"exclusiveMaximum",
	"propertyNames",
}

// cleanupParameters cleans JSON schema for Gemini compatibility (recursive).
func cleanupParameters(schema json.RawMessage) json.RawMessage {
	if len(schema) == 0 {
		return schema
	}

	var obj map[string]any
	if err := json.Unmarshal(schema, &obj); err != nil {
		return schema
	}

	cleanupSchemaObject(obj)

	cleaned, _ := json.Marshal(obj)
	return cleaned
}

// cleanupSchemaObject recursively removes unsupported fields from a schema object.
func cleanupSchemaObject(obj map[string]any) {
	// Remove unsupported fields at this level
	for _, field := range unsupportedSchemaFields {
		delete(obj, field)
	}

	// Recurse into "properties"
	if props, ok := obj["properties"].(map[string]any); ok {
		for _, v := range props {
			if propObj, ok := v.(map[string]any); ok {
				cleanupSchemaObject(propObj)
			}
		}
	}

	// Recurse into "items"
	if items, ok := obj["items"].(map[string]any); ok {
		cleanupSchemaObject(items)
	}

	// Recurse into "allOf", "anyOf", "oneOf"
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		if arr, ok := obj[key].([]any); ok {
			for _, item := range arr {
				if itemObj, ok := item.(map[string]any); ok {
					cleanupSchemaObject(itemObj)
				}
			}
		}
	}
}

// convertToolChoice converts Anthropic tool_choice to Gemini format.
func (g *Gemini) convertToolChoice(raw json.RawMessage) *geminiToolConfig {
	var tc types.ToolChoice
	if err := json.Unmarshal(raw, &tc); err != nil {
		return nil
	}

	config := &geminiToolConfig{
		FunctionCallingConfig: &geminiFuncCallConfig{},
	}

	switch tc.Type {
	case "auto":
		config.FunctionCallingConfig.Mode = "AUTO"
	case "none":
		config.FunctionCallingConfig.Mode = "NONE"
	case "any":
		config.FunctionCallingConfig.Mode = "ANY"
	case "tool":
		config.FunctionCallingConfig.Mode = "ANY"
		config.FunctionCallingConfig.AllowedFunctionNames = []string{tc.Name}
	}

	return config
}

// --- Streaming Response Transformation ---

type geminiStreamResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage      `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

func (g *Gemini) TransformStreamChunk(data []byte, state *StreamState) ([]SSEOutput, error) {
	var resp geminiStreamResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse gemini chunk: %w", err)
	}

	var events []SSEOutput

	// Update usage
	if resp.UsageMetadata != nil {
		state.InputTokens = resp.UsageMetadata.PromptTokenCount
		state.OutputTokens = resp.UsageMetadata.CandidatesTokenCount
	}

	// Send message_start on first chunk
	if !state.Started {
		state.Started = true
		msgStart := &types.MessageStartEvent{
			Type: "message_start",
			Message: &types.AnthropicResponse{
				ID:      state.MessageID,
				Type:    "message",
				Role:    "assistant",
				Model:   state.Model,
				Content: []types.ContentBlock{},
				Usage: &types.Usage{
					InputTokens:  state.InputTokens,
					OutputTokens: 0,
				},
			},
		}
		events = append(events, marshalSSE("message_start", msgStart))
	}

	if len(resp.Candidates) == 0 {
		return events, nil
	}

	candidate := resp.Candidates[0]

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			// Text content
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
					Text: part.Text,
				},
			}
			events = append(events, marshalSSE("content_block_delta", blockDelta))

		} else if part.FunctionCall != nil {
			// Function call
			if state.TextStarted && len(state.ToolCalls) == 0 {
				events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
					Type:  "content_block_stop",
					Index: state.ContentIndex,
				}))
				state.ContentIndex++
			}

			tcIndex := len(state.ToolCalls)
			blockIndex := state.ContentIndex + tcIndex

			// Generate tool call ID
			toolID := fmt.Sprintf("toolu_%s", randomID(20))

			tcState := &ToolCallState{
				ID:      toolID,
				Name:    part.FunctionCall.Name,
				Index:   blockIndex,
				Started: true,
			}
			state.ToolCalls[tcIndex] = tcState

			// content_block_start
			blockStart := &types.ContentBlockStartEvent{
				Type:  "content_block_start",
				Index: blockIndex,
				ContentBlock: &types.ContentBlock{
					Type: "tool_use",
					ID:   toolID,
					Name: part.FunctionCall.Name,
				},
			}
			events = append(events, marshalSSE("content_block_start", blockStart))

			// Stream the full arguments as a single delta
			argsJSON, _ := json.Marshal(part.FunctionCall.Args)
			if argsJSON == nil {
				argsJSON = []byte("{}")
			}

			blockDelta := &types.ContentBlockDeltaEvent{
				Type:  "content_block_delta",
				Index: blockIndex,
				Delta: &types.BlockDelta{
					Type:        "input_json_delta",
					PartialJSON: string(argsJSON),
				},
			}
			events = append(events, marshalSSE("content_block_delta", blockDelta))

			// content_block_stop
			events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: blockIndex,
			}))
		}
	}

	// Handle finish
	if candidate.FinishReason != "" {
		state.Finished = true
		stopReason := convertGeminiFinishReason(candidate.FinishReason)

		if state.TextStarted && len(state.ToolCalls) == 0 {
			events = append(events, marshalSSE("content_block_stop", &types.ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: 0,
			}))
		}

		events = append(events, marshalSSE("message_delta", &types.MessageDeltaEvent{
			Type: "message_delta",
			Delta: &types.MessageDelta{
				StopReason: stopReason,
			},
			Usage: &types.Usage{
				OutputTokens: state.OutputTokens,
			},
		}))

		events = append(events, marshalSSE("message_stop", &types.MessageStopEvent{
			Type: "message_stop",
		}))
	}

	return events, nil
}

// --- Non-streaming Response ---

func (g *Gemini) TransformResponse(body []byte) (*types.AnthropicResponse, error) {
	var gResp geminiStreamResponse
	if err := json.Unmarshal(body, &gResp); err != nil {
		return nil, fmt.Errorf("parse gemini response: %w", err)
	}

	resp := &types.AnthropicResponse{
		ID:   generateMessageID(),
		Type: "message",
		Role: "assistant",
		Model: g.cfg.Model,
	}

	if len(gResp.Candidates) > 0 {
		candidate := gResp.Candidates[0]
		resp.StopReason = convertGeminiFinishReason(candidate.FinishReason)

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				resp.Content = append(resp.Content, types.ContentBlock{
					Type: "text",
					Text: part.Text,
				})
			} else if part.FunctionCall != nil {
				argsJSON, _ := json.Marshal(part.FunctionCall.Args)
				resp.Content = append(resp.Content, types.ContentBlock{
					Type:  "tool_use",
					ID:    fmt.Sprintf("toolu_%s", randomID(20)),
					Name:  part.FunctionCall.Name,
					Input: argsJSON,
				})
			}
		}
	}

	if gResp.UsageMetadata != nil {
		resp.Usage = &types.Usage{
			InputTokens:  gResp.UsageMetadata.PromptTokenCount,
			OutputTokens: gResp.UsageMetadata.CandidatesTokenCount,
		}
	}

	return resp, nil
}

// convertGeminiFinishReason maps Gemini finish reasons to Anthropic stop reasons.
func convertGeminiFinishReason(reason string) string {
	switch strings.ToUpper(reason) {
	case "STOP":
		return "end_turn"
	case "MAX_TOKENS":
		return "max_tokens"
	case "SAFETY":
		return "end_turn"
	case "RECITATION":
		return "end_turn"
	default:
		return "end_turn"
	}
}
