package ai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Summarizer defines the interface for summarizing content.
type Summarizer interface {
	Summarize(ctx context.Context, content, model, filename, contentType string) (string, error)
}

// OpenAISummarizer implements Summarizer using OpenAI.
type OpenAISummarizer struct {
	Client *openai.Client
}

// NewOpenAISummarizer creates a new OpenAISummarizer.
func NewOpenAISummarizer(client *openai.Client) *OpenAISummarizer {
	return &OpenAISummarizer{Client: client}
}

// Summarize generates a summary using the OpenAI API.
func (s *OpenAISummarizer) Summarize(ctx context.Context, content, model, filename, contentType string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Describe the following file in one sentence"),
			openai.SystemMessage(fmt.Sprintf("The file name is called %s", filename)),
			openai.SystemMessage(fmt.Sprintf("The output of the /usr/bin/file command is: %s", contentType)),
			openai.UserMessage(content),
		},
		Model: model,
	}

	completion, err := s.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return completion.Choices[0].Message.Content, nil
}
