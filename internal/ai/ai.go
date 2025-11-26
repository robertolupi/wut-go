package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
)

// FileSummary holds the summary and content of a processed file.
type FileSummary struct {
	Filename string
	Summary  string
	Content  string // Truncated content
}

// Summarizer defines the interface for summarizing content.
type Summarizer interface {
	Summarize(ctx context.Context, content, model, filename, contentType string) (*FileSummary, error)
	SummarizeAll(ctx context.Context, fileSummaries []FileSummary, model string) (string, error)
}

// OpenAISummarizer implements Summarizer using OpenAI.
type OpenAISummarizer struct {
	Client    *openai.Client
	MaxTokens int
}

// NewOpenAISummarizer creates a new OpenAISummarizer.
func NewOpenAISummarizer(client *openai.Client, maxTokens int) *OpenAISummarizer {
	return &OpenAISummarizer{Client: client, MaxTokens: maxTokens}
}

// Summarize generates a summary using the OpenAI API.
func (s *OpenAISummarizer) Summarize(ctx context.Context, content, model, filename, contentType string) (*FileSummary, error) {
	// Simple heuristic: 1 token ~= 4 chars
	maxChars := s.MaxTokens * 4
	truncatedContent := content
	if len(content) > maxChars {
		truncatedContent = content[:maxChars] + "\n...[TRUNCATED]..."
	}

	var messages []openai.ChatCompletionMessageParamUnion
	messages = append(messages, openai.SystemMessage("Describe the following file in one sentence"))
	messages = append(messages, openai.SystemMessage(fmt.Sprintf("The file name is called %s", filename)))
	messages = append(messages, openai.SystemMessage(fmt.Sprintf("The output of the /usr/bin/file command is: %s", contentType)))

	if strings.HasPrefix(contentType, "image/") {
		parts := []openai.ChatCompletionContentPartUnionParam{
			openai.TextContentPart("Describe this image."),
			openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL: fmt.Sprintf("data:%s;base64,%s", contentType, truncatedContent),
			}),
		}
		messages = append(messages, openai.UserMessage(parts))
	} else {
		messages = append(messages, openai.UserMessage(truncatedContent))
	}

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
	}

	completion, err := s.Client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return &FileSummary{
		Filename: filename,
		Summary:  completion.Choices[0].Message.Content,
		Content:  truncatedContent,
	}, nil
}

// SummarizeAll generates an overall summary from a list of file summaries.
func (s *OpenAISummarizer) SummarizeAll(ctx context.Context, fileSummaries []FileSummary, model string) (string, error) {
	var sb strings.Builder
	sb.WriteString("Here are the summaries and truncated contents of the files analyzed:\n\n")

	// Calculate available chars for content
	// Reserve some buffer for system prompt and summaries (e.g., 20% or fixed amount)
	// For simplicity, let's assume summaries + overhead take up to 2000 chars per file on average?
	// Or better, calculate dynamic budget.

	maxTotalChars := s.MaxTokens * 4
	// Estimate overhead: system prompt (~200 chars) + per-file summary (~500 chars) + headers
	overhead := 200 + len(fileSummaries)*500
	availableChars := maxTotalChars - overhead

	charsPerFile := 0
	if availableChars > 0 && len(fileSummaries) > 0 {
		charsPerFile = availableChars / len(fileSummaries)
	}

	for _, fs := range fileSummaries {
		sb.WriteString(fmt.Sprintf("--- File: %s ---\n", fs.Filename))
		sb.WriteString(fmt.Sprintf("Summary: %s\n", fs.Summary))

		truncatedContent := fs.Content
		if len(truncatedContent) > charsPerFile {
			if charsPerFile > 0 {
				truncatedContent = truncatedContent[:charsPerFile] + "\n...[TRUNCATED]..."
			} else {
				truncatedContent = "[CONTENT OMITTED DUE TO CONTEXT LIMIT]"
			}
		}
		sb.WriteString(fmt.Sprintf("Content:\n%s\n\n", truncatedContent))
	}

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are a helpful assistant. Provide a comprehensive summary of the provided files, highlighting the overall purpose and relationships between them."),
			openai.UserMessage(sb.String()),
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
