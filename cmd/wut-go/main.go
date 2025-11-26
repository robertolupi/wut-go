package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"wut-go/internal/ai"
	"wut-go/internal/fileutils"
)

func main() {
	modelFlag := flag.String("model", "openai/gpt-oss-20b", "Model Name")
	baseURLFlag := flag.String("base-url", "http://localhost:1234/v1", "OpenAI-compatible URL")
	apiKeyFlag := flag.String("api-key", "", "API key")
	verboseFlag := flag.Bool("verbose", false, "Verbose output")
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

	summarizer := ai.NewOpenAISummarizer(&client)

	for _, filename := range files {
		if *verboseFlag {
			fmt.Printf("Processing %s...\n", filename)
		}
		content, cType, err := fileutils.ReadFileContent(filename)
		if err != nil {
			// We print the error, which includes "is a directory" or "not a regular file" details
			fmt.Printf("Skipping %s: %s\n", filename, err)
			continue
		}

		if *verboseFlag {
			fmt.Printf("Content type: %s\n", cType)
			fmt.Printf("Content length: %d\n", len(content))
			fmt.Printf("Content: %s\n", content)
		}

		guess, err := summarizer.Summarize(context.Background(), content, *modelFlag, filename, cType)
		if err != nil {
			fmt.Printf("Failed to guess file %s: %s\n", filename, err)
			continue
		}

		fmt.Printf("%s: %s\n", filename, guess)
	}
}
