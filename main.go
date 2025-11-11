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

func guessContent(client openai.Client, content string, model string, filename string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.SystemMessage(fmt.Sprintf("The file name is called %s", filename)),
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

func isTextFile(path string) bool {
	// Use the file utility to determine if the file contains text
	cmd := exec.Command("file", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "text")
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

		if !isTextFile(filename) {
			fmt.Printf("File %s is not a text file\n", filename)
			continue
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Failed to read file %s: %s\n", filename, err)
			continue
		}

		guess, err := guessContent(client, string(content), *modelFlag, filename)
		if err != nil {
			fmt.Printf("Failed to guess file %s: %s\n", filename, err)
			continue
		}

		fmt.Printf("%s: %s\n", filename, guess)
	}
}
