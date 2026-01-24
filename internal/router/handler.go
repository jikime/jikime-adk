package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"jikime-adk/internal/router/provider"
	"jikime-adk/internal/router/types"
)

// handleMessages handles POST /v1/messages requests.
func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "invalid_request_error", "Method not allowed")
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req types.AnthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	// Resolve provider
	providerName, _ := s.config.ResolveProvider(0, false)

	// Use model from request (set by Claude Code via ANTHROPIC_DEFAULT_*_MODEL)
	model := req.Model
	if model == "" {
		// Fallback to config model if request doesn't specify one
		if p, ok := s.config.Providers[providerName]; ok {
			model = p.Model
		}
	}

	// Get provider config
	provCfg, ok := s.config.Providers[providerName]
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "api_error",
			fmt.Sprintf("Provider '%s' not configured", providerName))
		return
	}

	// Create provider instance if different from default
	prov := s.provider
	if providerName != s.config.Router.Provider {
		pCfg := toProviderConfig(&provCfg)
		var err error
		prov, err = provider.NewProvider(providerName, pCfg)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "api_error",
				fmt.Sprintf("Failed to create provider: %v", err))
			return
		}
	}

	s.logger.Printf("-> %s/%s (stream=%v)", providerName, model, req.Stream)

	// Transform and forward request
	provReq, err := prov.TransformRequest(&req, model)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid_request_error",
			fmt.Sprintf("Transform error: %v", err))
		return
	}

	// Add provider headers
	for k, v := range prov.Headers(provCfg.APIKey) {
		provReq.Header.Set(k, v)
	}

	// Forward to provider
	client := &http.Client{}
	resp, err := client.Do(provReq)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "api_error",
			fmt.Sprintf("Provider request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	// Check provider response status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		s.logger.Printf("<- Provider error (%d): %s", resp.StatusCode, string(respBody))
		s.writeError(w, resp.StatusCode, "api_error",
			fmt.Sprintf("Provider returned %d: %s", resp.StatusCode, string(respBody)))
		return
	}

	if req.Stream {
		s.handleStreamResponse(w, resp, prov, model)
	} else {
		s.handleSyncResponse(w, resp, prov)
	}
}

// handleStreamResponse processes a streaming response from the provider.
func (s *Server) handleStreamResponse(w http.ResponseWriter, resp *http.Response, prov provider.Provider, model string) {
	SetSSEHeaders(w)
	w.WriteHeader(http.StatusOK)

	sseWriter := NewSSEWriter(w)
	sseReader := NewSSEReader(resp.Body)
	state := provider.NewStreamState(model)

	for {
		_, data, err := sseReader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			s.logger.Printf("SSE read error: %v", err)
			break
		}

		if data == "[DONE]" {
			if state.Started && !state.Finished {
				s.sendStreamEnd(sseWriter, state)
			}
			break
		}

		events, err := prov.TransformStreamChunk([]byte(data), state)
		if err != nil {
			s.logger.Printf("Transform chunk error: %v", err)
			continue
		}

		for _, evt := range events {
			sseWriter.WriteRawEvent(evt.Event, evt.Data)
		}
	}
}

// handleSyncResponse processes a non-streaming response from the provider.
func (s *Server) handleSyncResponse(w http.ResponseWriter, resp *http.Response, prov provider.Provider) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.writeError(w, http.StatusBadGateway, "api_error", "Failed to read provider response")
		return
	}

	anthropicResp, err := prov.TransformResponse(body)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "api_error",
			fmt.Sprintf("Transform response error: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anthropicResp)
}

// sendStreamEnd sends the final stream events.
func (s *Server) sendStreamEnd(sw *SSEWriter, state *provider.StreamState) {
	if state.TextStarted {
		sw.WriteEvent("content_block_stop", &types.ContentBlockStopEvent{
			Type:  "content_block_stop",
			Index: 0,
		})
	}

	for _, tc := range state.ToolCalls {
		if tc.Started {
			sw.WriteEvent("content_block_stop", &types.ContentBlockStopEvent{
				Type:  "content_block_stop",
				Index: tc.Index,
			})
		}
	}

	sw.WriteEvent("message_delta", &types.MessageDeltaEvent{
		Type: "message_delta",
		Delta: &types.MessageDelta{
			StopReason: "end_turn",
		},
		Usage: &types.Usage{
			OutputTokens: state.OutputTokens,
		},
	})

	sw.WriteEvent("message_stop", &types.MessageStopEvent{
		Type: "message_stop",
	})
}

// writeError writes an Anthropic-format error response.
func (s *Server) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&types.AnthropicError{
		Type: "error",
		Error: &types.ErrorDetail{
			Type:    errType,
			Message: message,
		},
	})
}
