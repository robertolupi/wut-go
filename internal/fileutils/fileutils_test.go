package fileutils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestHelperProcess isn't a real test. It's used as a helper process for execCommand.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd := args[0]
	switch cmd {
	case "file":
		// Mock 'file --mime-type path'
		// Expect args: --mime-type path
		if len(args) >= 2 && args[1] == "--mime-type" {
			path := args[len(args)-1]
			if strings.Contains(path, "binary") {
				fmt.Printf("%s: application/octet-stream\n", path)
			} else {
				fmt.Printf("%s: text/plain\n", path)
			}
		} else {
			fmt.Printf("%s: unknown\n", args[len(args)-1])
		}
	case "pdftotext":
		// Mock 'pdftotext path -'
		fmt.Print("Extracted PDF Content")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", cmd)
		os.Exit(2)
	}
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{" -test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGetContentType(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	ctype, err := GetContentType("dummy.txt")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ctype != "text/plain" {
		t.Errorf("Expected text/plain, got %s", ctype)
	}
}

func TestExtractTextFromPDF(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	content, err := ExtractTextFromPDF("dummy.pdf")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if content != "Extracted PDF Content" {
		t.Errorf("Expected 'Extracted PDF Content', got %s", content)
	}
}

func TestReadFileContent_Binary(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Create a dummy binary file
	err := os.WriteFile("test_binary.bin", []byte("Binary\x00Content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_binary.bin")

	content, cType, err := ReadFileContent("test_binary.bin")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cType != "application/octet-stream" {
		t.Errorf("Expected application/octet-stream, got %s", cType)
	}
	if content != "" {
		t.Errorf("Expected empty content for binary file, got %q", content)
	}
}
