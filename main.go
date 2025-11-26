package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const systemPrompt = "Describe the following file in one sentence"

func guessContent(client openai.Client, content string, model string, filename string, contentType string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.SystemMessage(fmt.Sprintf("The file name is called %s", filename)),
			openai.SystemMessage(fmt.Sprintf("The output of the /usr/bin/file command is: %s", contentType)),
			openai.UserMessage(content),
		},
		Model: model,
	}

	completion, err := client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		return "", err
	}
	return completion.Choices[0].Message.Content, nil
}

func contentType(path string) (string, error) {
	// Use the file utility to determine if the file contains text
	cmd := exec.Command("file", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(string(output), ":", 2)
	if len(parts) != 2 {
		return parts[0], nil
	}
	return parts[1], nil
}

func extractTextContentFromPdf(path string) (string, error) {
	cmd := exec.Command("pdftotext", path, "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func main() {
	modelFlag := flag.String("model", "openai/gpt-oss-20b", "Model Name")
	baseURLFlag := flag.String("base-url", "http://localhost:1234/v1", "OpenAI-compatible URL")
	apiKeyFlag := flag.String("api-key", "", "API key")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("Please specify at least one file")
		os.Exit(1)
	}

	client := openai.NewClient(
		option.WithBaseURL(*baseURLFlag),
		option.WithAPIKey(*apiKeyFlag),
	)

	for _, filename := range files {
		st, err := os.Stat(filename)
		if err != nil {
			fmt.Printf("Failed to stat file %s: %s\n", filename, err)
			continue
		}
		if st.IsDir() {
			fmt.Printf("%s is a directory.\n", filename)
			continue
		}
		if !st.Mode().IsRegular() {
			fmt.Printf("File %s is not a file\n", filename)
			continue
		}

		contentType, err := contentType(filename)
		if err != nil {
			fmt.Printf("Failed to determine content type of file %s: %s\n", filename, err)
			continue
		}
		var content string
		if strings.HasPrefix(contentType, "application/pdf") {
			content, err = extractTextContentFromPdf(filename)
			if err != nil {
				fmt.Printf("Failed to extract text from PDF file %s: %s\n", filename, err)
			}
		} else if strings.HasPrefix(contentType, "text/") {
			bytes, err := os.ReadFile(filename)
			if err != nil {
				fmt.Printf("Failed to read file %s: %s\n", filename, err)
			}
			content = string(bytes)
		}

		guess, err := guessContent(client, string(content), *modelFlag, filename, contentType)
		if err != nil {
			fmt.Printf("Failed to guess file %s: %s\n", filename, err)
		}

		fmt.Printf("%s: %s\n", filename, guess)
	}
}
