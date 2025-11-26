package fileutils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// execCommand allows mocking in tests
var execCommand = exec.Command

// GetContentType determines the content type of a file using the 'file' command.
func GetContentType(path string) (string, error) {
	// Use --mime-type to get standard MIME types like "text/plain"
	cmd := execCommand("file", "--mime-type", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("file command failed: %w", err)
	}
	// output example: "filename: text/plain"
	parts := strings.SplitN(string(output), ":", 2)
	if len(parts) != 2 {
		return strings.TrimSpace(parts[0]), nil
	}
	return strings.TrimSpace(parts[1]), nil
}

// ExtractTextFromPDF extracts text from a PDF file using 'pdftotext'.
func ExtractTextFromPDF(path string) (string, error) {
	cmd := execCommand("pdftotext", path, "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return string(output), nil
}

// ExtractBinaryInfo extracts metadata from a binary file using system tools.
func ExtractBinaryInfo(path string) (string, error) {
	var sb strings.Builder

	// File Info
	sb.WriteString("=== FILE INFO ===\n")
	out, _ := execCommand("file", path).CombinedOutput()
	sb.Write(out)

	// Shared Libraries
	sb.WriteString("\n=== SHARED LIBRARIES & FRAMEWORKS ===\n")
	out, _ = execCommand("otool", "-L", path).CombinedOutput()
	sb.Write(out)

	// Entitlements
	sb.WriteString("\n=== ENTITLEMENTS & SIGNING ===\n")
	out, _ = execCommand("codesign", "-d", "--entitlements", ":-", path).CombinedOutput()
	sb.Write(out)

	// Load Commands
	sb.WriteString("\n=== LOAD COMMANDS (Headers) ===\n")
	cmdStr := fmt.Sprintf("otool -l %q | grep -A 5 \"LC_VERSION_MIN\\|LC_BUILD_VERSION\\|LC_ENCRYPTION_INFO\"", path)
	out, _ = execCommand("sh", "-c", cmdStr).CombinedOutput()
	sb.Write(out)

	// External Symbols
	sb.WriteString("\n=== EXTERNAL SYMBOLS (Imports) ===\n")
	cmdStr = fmt.Sprintf("nm -u %q | c++filt | head -n 100", path)
	out, _ = execCommand("sh", "-c", cmdStr).CombinedOutput()
	sb.Write(out)

	// Interesting Strings
	sb.WriteString("\n=== INTERESTING STRINGS ===\n")
	cmdStr = fmt.Sprintf("strings %q | grep -E \"https?://|/usr/|/System/|/var/\" | head -n 50", path)
	out, _ = execCommand("sh", "-c", cmdStr).CombinedOutput()
	sb.Write(out)

	return sb.String(), nil
}

// ReadFileContent reads the content of a file based on its content type.
// It handles PDF text extraction automatically.
func ReadFileContent(filename string) (string, string, error) {
	st, err := os.Stat(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to stat file: %w", err)
	}
	if st.IsDir() {
		return "", "", fmt.Errorf("%s is a directory", filename)
	}
	if !st.Mode().IsRegular() {
		return "", "", fmt.Errorf("%s is not a regular file", filename)
	}

	cType, err := GetContentType(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to determine content type: %w", err)
	}

	var content string
	if strings.HasPrefix(cType, "application/pdf") {
		content, err = ExtractTextFromPDF(filename)
		if err != nil {
			return "", cType, fmt.Errorf("failed to extract text from PDF: %w", err)
		}
	} else if cType == "application/x-mach-binary" {
		content, err = ExtractBinaryInfo(filename)
		if err != nil {
			return "", cType, fmt.Errorf("failed to extract binary info: %w", err)
		}
	} else if strings.HasPrefix(cType, "text/") || strings.Contains(cType, "text") {
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return "", cType, fmt.Errorf("failed to read file: %w", err)
		}
		content = string(bytes)
	} else {
		// Attempt to read as text anyway if it's not PDF, or maybe return empty?
		bytes, err := os.ReadFile(filename)
		if err == nil {
			// Heuristic: if it has null bytes, it's binary.
			if !strings.Contains(string(bytes), "\x00") {
				content = string(bytes)
			}
		}
	}

	return content, cType, nil
}
