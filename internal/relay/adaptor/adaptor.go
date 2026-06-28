package adaptor

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChatRequest is the OpenAI-compatible chat completion request shape.
type ChatRequest struct {
	Model               string        `json:"model"`
	Messages            []Message     `json:"messages"`
	Stream              bool          `json:"stream,omitempty"`
	Temperature         *float64      `json:"temperature,omitempty"`
	MaxTokens           *int          `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int          `json:"max_completion_tokens,omitempty"`
	TopP                *float64      `json:"top_p,omitempty"`
	Tools               []interface{} `json:"tools,omitempty"`
	ToolChoice          interface{}   `json:"tool_choice,omitempty"`
	ResponseFormat      interface{}   `json:"response_format,omitempty"`
	Stop                interface{}   `json:"stop,omitempty"`
	PresencePenalty     *float64      `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64      `json:"frequency_penalty,omitempty"`
	Seed                *int          `json:"seed,omitempty"`
	User                string        `json:"user,omitempty"`
	N                   *int          `json:"n,omitempty"`
	Logprobs            *bool         `json:"logprobs,omitempty"`
	TopLogprobs         *int          `json:"top_logprobs,omitempty"`
	StreamOptions       interface{}   `json:"stream_options,omitempty"`
}

// Message is an OpenAI-compatible chat message. Content may be a string or an array for multimodal requests.
type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"`
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  interface{} `json:"tool_calls,omitempty"`
}

// ChatResponse is the OpenAI-compatible chat completion response shape.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice is an OpenAI-compatible response choice.
type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason *string  `json:"finish_reason,omitempty"`
}

// Usage is an OpenAI-compatible token usage block.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CacheTokens      int `json:"cache_tokens,omitempty"`
	ReasoningTokens  int `json:"reasoning_tokens,omitempty"`
}

// StreamChunk is an OpenAI-compatible chat completion stream chunk.
type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Adaptor converts between the public OpenAI-compatible API and provider APIs.
type Adaptor interface {
	GetName() string
	TransformRequest(req *ChatRequest) (interface{}, error)
	TransformResponse(resp interface{}) (*ChatResponse, error)
	TransformStreamChunk(chunk []byte) (*StreamChunk, error)
	GetEndpoint(baseURL string, model string) string
	GetHeaders(apiKey string) map[string]string
}

// IsOpenAICompatible reports whether the provider natively speaks OpenAI chat completions.
func IsOpenAICompatible(ad Adaptor) bool {
	name := ad.GetName()
	return name == "openai" || name == "deepseek"
}

// OpenAIAdaptor is a pass-through adaptor for OpenAI-compatible providers.
type OpenAIAdaptor struct{}

func NewOpenAIAdaptor() *OpenAIAdaptor                                          { return &OpenAIAdaptor{} }
func (a *OpenAIAdaptor) GetName() string                                        { return "openai" }
func (a *OpenAIAdaptor) TransformRequest(req *ChatRequest) (interface{}, error) { return req, nil }
func (a *OpenAIAdaptor) TransformResponse(resp interface{}) (*ChatResponse, error) {
	data, err := responseBytes(resp)
	if err != nil {
		return nil, err
	}
	var result ChatResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
func (a *OpenAIAdaptor) TransformStreamChunk(chunk []byte) (*StreamChunk, error) {
	var result StreamChunk
	if err := json.Unmarshal(chunk, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
func (a *OpenAIAdaptor) GetEndpoint(baseURL string, model string) string {
	return strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
}
func (a *OpenAIAdaptor) GetHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
		"Accept":        "application/json, text/event-stream",
	}
}

// ClaudeAdaptor converts OpenAI chat requests to Anthropic Messages API.
type ClaudeAdaptor struct{}

func NewClaudeAdaptor() *ClaudeAdaptor   { return &ClaudeAdaptor{} }
func (a *ClaudeAdaptor) GetName() string { return "claude" }
func (a *ClaudeAdaptor) TransformRequest(req *ChatRequest) (interface{}, error) {
	system := ""
	messages := make([]map[string]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		content := contentToText(msg.Content)
		if msg.Role == "system" {
			system = content
			continue
		}
		role := msg.Role
		if role == "tool" || role == "function" {
			role = "user"
		}
		messages = append(messages, map[string]string{"role": role, "content": content})
	}
	maxTokens := 4096
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	if req.MaxCompletionTokens != nil {
		maxTokens = *req.MaxCompletionTokens
	}
	claudeReq := map[string]interface{}{"model": req.Model, "messages": messages, "max_tokens": maxTokens, "stream": req.Stream}
	if system != "" {
		claudeReq["system"] = system
	}
	if req.Temperature != nil {
		claudeReq["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		claudeReq["top_p"] = *req.TopP
	}
	if req.Stop != nil {
		claudeReq["stop_sequences"] = req.Stop
	}
	return claudeReq, nil
}
func (a *ClaudeAdaptor) TransformResponse(resp interface{}) (*ChatResponse, error) {
	data, err := responseBytes(resp)
	if err != nil {
		return nil, err
	}
	var claudeResp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &claudeResp); err != nil {
		return nil, err
	}
	content := ""
	if len(claudeResp.Content) > 0 {
		content = claudeResp.Content[0].Text
	}
	stop := "stop"
	return &ChatResponse{
		ID:      claudeResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   claudeResp.Model,
		Choices: []Choice{{Index: 0, Message: &Message{Role: "assistant", Content: content}, FinishReason: &stop}},
		Usage:   Usage{PromptTokens: claudeResp.Usage.InputTokens, CompletionTokens: claudeResp.Usage.OutputTokens, TotalTokens: claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens},
	}, nil
}
func (a *ClaudeAdaptor) TransformStreamChunk(chunk []byte) (*StreamChunk, error) {
	var c struct {
		Type  string `json:"type"`
		Delta struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := json.Unmarshal(chunk, &c); err != nil {
		return nil, err
	}
	if c.Type != "content_block_delta" || c.Delta.Text == "" {
		return nil, nil
	}
	return &StreamChunk{ID: "chatcmpl-claude", Object: "chat.completion.chunk", Created: time.Now().Unix(), Choices: []Choice{{Index: 0, Delta: &Message{Role: "assistant", Content: c.Delta.Text}}}}, nil
}
func (a *ClaudeAdaptor) GetEndpoint(baseURL string, model string) string {
	return strings.TrimSuffix(baseURL, "/") + "/v1/messages"
}
func (a *ClaudeAdaptor) GetHeaders(apiKey string) map[string]string {
	return map[string]string{"x-api-key": apiKey, "Content-Type": "application/json", "Accept": "application/json, text/event-stream", "anthropic-version": "2023-06-01"}
}

// GeminiAdaptor converts OpenAI chat requests to Gemini generateContent.
type GeminiAdaptor struct{}

func NewGeminiAdaptor() *GeminiAdaptor   { return &GeminiAdaptor{} }
func (a *GeminiAdaptor) GetName() string { return "gemini" }
func (a *GeminiAdaptor) TransformRequest(req *ChatRequest) (interface{}, error) {
	contents := make([]map[string]interface{}, 0, len(req.Messages))
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		if msg.Role == "system" {
			role = "user"
		}
		contents = append(contents, map[string]interface{}{"role": role, "parts": []map[string]string{{"text": contentToText(msg.Content)}}})
	}
	generationConfig := map[string]interface{}{}
	if req.Temperature != nil {
		generationConfig["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		generationConfig["topP"] = *req.TopP
	}
	if req.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *req.MaxTokens
	}
	if req.MaxCompletionTokens != nil {
		generationConfig["maxOutputTokens"] = *req.MaxCompletionTokens
	}
	geminiReq := map[string]interface{}{"contents": contents}
	if len(generationConfig) > 0 {
		geminiReq["generationConfig"] = generationConfig
	}
	return geminiReq, nil
}
func (a *GeminiAdaptor) TransformResponse(resp interface{}) (*ChatResponse, error) {
	data, err := responseBytes(resp)
	if err != nil {
		return nil, err
	}
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(data, &geminiResp); err != nil {
		return nil, err
	}
	content := ""
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		content = geminiResp.Candidates[0].Content.Parts[0].Text
	}
	stop := "stop"
	return &ChatResponse{
		ID:      "chatcmpl-gemini",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Choices: []Choice{{Index: 0, Message: &Message{Role: "assistant", Content: content}, FinishReason: &stop}},
		Usage:   Usage{PromptTokens: geminiResp.UsageMetadata.PromptTokenCount, CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount, TotalTokens: geminiResp.UsageMetadata.TotalTokenCount},
	}, nil
}
func (a *GeminiAdaptor) TransformStreamChunk(chunk []byte) (*StreamChunk, error) {
	var c struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(chunk, &c); err != nil {
		return nil, err
	}
	if len(c.Candidates) == 0 || len(c.Candidates[0].Content.Parts) == 0 {
		return nil, nil
	}
	return &StreamChunk{ID: "chatcmpl-gemini", Object: "chat.completion.chunk", Created: time.Now().Unix(), Choices: []Choice{{Index: 0, Delta: &Message{Role: "assistant", Content: c.Candidates[0].Content.Parts[0].Text}}}}, nil
}
func (a *GeminiAdaptor) GetEndpoint(baseURL string, model string) string {
	return strings.TrimSuffix(baseURL, "/") + "/v1beta/models/" + model + ":generateContent"
}
func (a *GeminiAdaptor) GetHeaders(apiKey string) map[string]string {
	return map[string]string{"Content-Type": "application/json", "Accept": "application/json", "x-goog-api-key": apiKey}
}

// DeepSeekAdaptor is OpenAI-compatible.
type DeepSeekAdaptor struct{ OpenAIAdaptor }

func NewDeepSeekAdaptor() *DeepSeekAdaptor { return &DeepSeekAdaptor{} }
func (a *DeepSeekAdaptor) GetName() string { return "deepseek" }

// GetAdaptor returns an adaptor by channel type.
func GetAdaptor(channelType int) (Adaptor, error) {
	switch channelType {
	case 1:
		return NewOpenAIAdaptor(), nil
	case 2:
		return NewClaudeAdaptor(), nil
	case 3:
		return NewGeminiAdaptor(), nil
	case 4:
		return NewDeepSeekAdaptor(), nil
	default:
		return nil, fmt.Errorf("unsupported channel type: %d", channelType)
	}
}

func responseBytes(resp interface{}) ([]byte, error) {
	switch v := resp.(type) {
	case []byte:
		return v, nil
	case json.RawMessage:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

func contentToText(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}
