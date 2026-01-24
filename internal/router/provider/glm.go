package provider

import (
	"fmt"
	"net/http"

	"jikime-adk/internal/router/types"
)

// GLM provider (ZhipuAI) - uses OpenAI-compatible API format.
type GLM struct {
	*OpenAI
}

// NewGLM creates a new GLM provider.
func NewGLM(cfg *ProviderConfig) *GLM {
	return &GLM{OpenAI: NewOpenAI(cfg)}
}

func (g *GLM) Name() string { return "glm" }

func (g *GLM) Headers(apiKey string) map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
	}
}

// TransformRequest overrides the endpoint to use GLM's API URL.
func (g *GLM) TransformRequest(req *types.AnthropicRequest, model string) (*http.Request, error) {
	// Use parent's transform logic but override endpoint
	httpReq, err := g.OpenAI.TransformRequest(req, model)
	if err != nil {
		return nil, err
	}

	// Replace URL with GLM endpoint
	endpoint := g.endpoint()
	httpReq.URL, _ = httpReq.URL.Parse(endpoint)
	return httpReq, nil
}

func (g *GLM) endpoint() string {
	baseURL := g.cfg.BaseURL
	if baseURL == "" {
		if g.cfg.Region == "china" {
			baseURL = "https://open.bigmodel.cn/api/paas/v4"
		} else {
			baseURL = "https://api.z.ai/api/paas/v4"
		}
	}
	return baseURL + "/chat/completions"
}
