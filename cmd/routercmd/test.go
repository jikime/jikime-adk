package routercmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	router "jikime-adk/internal/router"
	"jikime-adk/internal/router/types"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <provider>",
		Short: "Send a test request to the router",
		Long: `Send a test request to a specific provider via the router.

Example:
  jikime router test openai
  jikime router test gemini`,
		Args: cobra.ExactArgs(1),
		RunE: runTest,
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	cfg, err := router.LoadConfig()
	if err != nil {
		return err
	}

	// Check if provider exists
	provCfg, ok := cfg.Providers[providerName]
	if !ok {
		return fmt.Errorf("provider '%s' not found in config", providerName)
	}

	// Check if router is running
	pid := readPID()
	if pid == 0 || !processExists(pid) {
		return fmt.Errorf("router is not running. Start it with 'jikime router start'")
	}

	// URL includes provider path: http://localhost:8787/{provider}/v1/messages
	addr := fmt.Sprintf("http://%s:%d/%s", cfg.Router.Host, cfg.Router.Port, providerName)

	// Create test request
	testReq := &types.AnthropicRequest{
		Model:     provCfg.Model,
		MaxTokens: 100,
		Stream:    false,
		Messages: []types.AnthropicMessage{
			{
				Role:    "user",
				Content: json.RawMessage(`"Say hello in one word."`),
			},
		},
	}

	body, _ := json.Marshal(testReq)

	fmt.Println()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("  Testing %s -> %s/%s\n",
		cyan(addr), cyan(providerName), cyan(provCfg.Model))

	start := time.Now()
	resp, err := http.Post(addr+"/v1/messages", "application/json", bytes.NewReader(body))
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var anthropicResp types.AnthropicResponse
		json.Unmarshal(respBody, &anthropicResp)

		color.Green("  Success (%dms)", elapsed.Milliseconds())
		if len(anthropicResp.Content) > 0 && anthropicResp.Content[0].Type == "text" {
			text := anthropicResp.Content[0].Text
			if len(text) > 100 {
				text = text[:100] + "..."
			}
			fmt.Printf("  Response: %s\n", text)
		}
		if anthropicResp.Usage != nil {
			fmt.Printf("  Tokens: in=%d, out=%d\n",
				anthropicResp.Usage.InputTokens, anthropicResp.Usage.OutputTokens)
		}
	} else {
		color.Red("  Failed (%d) - %dms", resp.StatusCode, elapsed.Milliseconds())
		fmt.Printf("  %s\n", string(respBody))
	}

	fmt.Println()
	return nil
}
