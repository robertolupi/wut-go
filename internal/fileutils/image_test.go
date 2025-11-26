package fileutils

import (
	"encoding/base64"
	"os"
	"os/exec"
	"testing"
)

func TestReadFileContent_Image(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	// Create a dummy image file
	// This is a 1x1 pixel PNG
	imgData := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAAAAAA6fptVAAAACklEQVR4nGP6DwABBAEAAAAA"
	decoded, _ := base64.StdEncoding.DecodeString(imgData)
	err := os.WriteFile("test.png", decoded, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test.png")

	content, cType, err := ReadFileContent("test.png")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cType != "image/png" {
		t.Errorf("Expected image/png, got %s", cType)
	}

	// The content should be the base64 encoded string of the file content
	// Since we wrote the decoded bytes to the file, reading it back and encoding it
	// should give us the original base64 string (imgData)
	if content != imgData {
		t.Errorf("Expected content to be base64 string %q, got %q", imgData, content)
	}
}
