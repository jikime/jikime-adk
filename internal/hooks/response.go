package hooks

import (
	"encoding/json"
	"os"
)

// HookResponse represents the standard response format for Claude Code hooks
type HookResponse struct {
	Continue       bool              `json:"continue"`
	SystemMessage  string            `json:"systemMessage,omitempty"`
	StopReason     string            `json:"stopReason,omitempty"`
	Decision       string            `json:"decision,omitempty"`
	Reason         string            `json:"reason,omitempty"`
	Performance    map[string]bool   `json:"performance,omitempty"`
	ErrorDetails   map[string]string `json:"error_details,omitempty"`
}

// HookInput represents the input received from Claude Code
type HookInput struct {
	SessionID     string                 `json:"session_id,omitempty"`
	ToolName      string                 `json:"tool_name,omitempty"`
	ToolInput     map[string]interface{} `json:"tool_input,omitempty"`
	ToolOutput    string                 `json:"tool_output,omitempty"`
	FilePath      string                 `json:"file_path,omitempty"`
	Content       string                 `json:"content,omitempty"`
	ContextWindow map[string]interface{} `json:"context_window,omitempty"`
}

// ReadInput reads and parses JSON input from stdin
func ReadInput() (*HookInput, error) {
	var input HookInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// Return empty input if parsing fails (non-fatal)
		return &HookInput{}, nil
	}
	return &input, nil
}

// WriteResponse writes the hook response to stdout
func WriteResponse(response HookResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}

// SuccessResponse creates a successful hook response
func SuccessResponse(message string) HookResponse {
	return HookResponse{
		Continue:      true,
		SystemMessage: message,
		Performance: map[string]bool{
			"go_hook": true,
		},
	}
}

// ErrorResponse creates an error hook response (but continues)
func ErrorResponse(err error) HookResponse {
	return HookResponse{
		Continue:      true,
		SystemMessage: "Hook encountered an error - continuing",
		ErrorDetails: map[string]string{
			"error": err.Error(),
		},
	}
}

// BlockResponse creates a blocking hook response
func BlockResponse(reason string) HookResponse {
	return HookResponse{
		Continue:   false,
		Decision:   "block",
		Reason:     reason,
		StopReason: reason,
	}
}

// AllowResponse creates an allowing hook response
func AllowResponse(message string) HookResponse {
	return HookResponse{
		Continue:      true,
		Decision:      "allow",
		SystemMessage: message,
	}
}
