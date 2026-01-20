package openai

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sonsonha/eng-noting/internal/domain/ai"
)

type Client struct {
	client *openai.Client
}

func NewClient(apiKey string) *Client {
	if apiKey == "" {
		return nil
	}
	return &Client{
		client: openai.NewClient(apiKey),
	}
}

func (c *Client) ExplainWord(systemPrompt, userPrompt string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("OpenAI client not initialized")
	}

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	return content, nil
}

// Ensure Client implements ai.Client interface
var _ ai.Client = (*Client)(nil)
