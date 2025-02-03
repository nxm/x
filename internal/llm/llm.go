package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
    Role    string
    Content string
}

type CompletionRequest struct {
    Messages    []Message
    MaxTokens   int
    Temperature float32
}

type CompletionResponse struct {
    Content string
    Usage   TokenUsage
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

type OpenAIProvider struct {
    apiKey     string
    modelName  string
    apiBaseURL string
}

func NewOpenAIProvider(apiKey, modelName string) *OpenAIProvider {
    return &OpenAIProvider{
        apiKey:     apiKey,
        modelName:  modelName,
        apiBaseURL: "https://api.openai.com/v1",
    }
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
    type OpenAIMessage struct {
        Role    string `json:"role"`
        Content string `json:"content"`
    }

    type OpenAIRequest struct {
        Model       string         `json:"model"`
        Messages    []OpenAIMessage `json:"messages"`
        MaxTokens   int            `json:"max_tokens,omitempty"`
        Temperature float32        `json:"temperature,omitempty"`
    }

    type OpenAIChoice struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    }

    type OpenAIResponse struct {
        Choices []OpenAIChoice `json:"choices"`
        Usage   struct {
            PromptTokens     int `json:"prompt_tokens"`
            CompletionTokens int `json:"completion_tokens"`
            TotalTokens      int `json:"total_tokens"`
        } `json:"usage"`
    }

    messages := make([]OpenAIMessage, len(req.Messages))
    for i, msg := range req.Messages {
        messages[i] = OpenAIMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }

    openAIReq := OpenAIRequest{
        Model:       p.modelName,
        Messages:    messages,
        MaxTokens:   req.MaxTokens,
        Temperature: req.Temperature,
    }

    jsonBody, err := json.Marshal(openAIReq)
    if err != nil {
        return CompletionResponse{}, fmt.Errorf("marshaling request: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(
        ctx,
        "POST",
        fmt.Sprintf("%s/chat/completions", p.apiBaseURL),
        bytes.NewReader(jsonBody),
    )
    if err != nil {
        return CompletionResponse{}, fmt.Errorf("creating request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return CompletionResponse{}, fmt.Errorf("making request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return CompletionResponse{}, fmt.Errorf("reading response body: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
        return CompletionResponse{}, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
    }

    var openAIResp OpenAIResponse
    if err := json.Unmarshal(body, &openAIResp); err != nil {
        return CompletionResponse{}, fmt.Errorf("parsing response: %w", err)
    }

    if len(openAIResp.Choices) == 0 {
        return CompletionResponse{}, errors.New("no completions returned")
    }

    return CompletionResponse{
        Content: openAIResp.Choices[0].Message.Content,
        Usage: TokenUsage{
            PromptTokens:     openAIResp.Usage.PromptTokens,
            CompletionTokens: openAIResp.Usage.CompletionTokens,
            TotalTokens:      openAIResp.Usage.TotalTokens,
        },
    }, nil
}

type Client struct {
    provider Provider
}

func NewClient(provider Provider) *Client {
    return &Client{provider: provider}
}

func (c *Client) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
    return c.provider.Complete(ctx, req)
}
