package cmd

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMultipartMessage_SingleTextAttachment(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("hello attachment"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := buildMultipartMessage("to@example.com", "Subject", "Body text", "", "", []string{filePath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "multipart/mixed") {
		t.Error("expected Content-Type to contain multipart/mixed")
	}

	if !strings.Contains(output, "text/plain") {
		t.Error("expected output to contain text/plain part")
	}

	if !strings.Contains(output, "text/html") {
		t.Error("expected output to contain text/html part")
	}

	if !strings.Contains(output, "test.txt") {
		t.Error("expected output to contain attachment filename test.txt")
	}

	if !strings.Contains(output, "Content-Disposition: attachment") {
		t.Error("expected output to contain Content-Disposition attachment header")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte("hello attachment"))
	if !strings.Contains(output, encoded) {
		t.Error("expected output to contain base64-encoded file content")
	}
}

func TestBuildMultipartMessage_MultipleAttachments(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "first.txt")
	if err := os.WriteFile(file1, []byte("content one"), 0644); err != nil {
		t.Fatalf("failed to create first temp file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "second.txt")
	if err := os.WriteFile(file2, []byte("content two"), 0644); err != nil {
		t.Fatalf("failed to create second temp file: %v", err)
	}

	result, err := buildMultipartMessage("to@example.com", "Subject", "Body text", "", "", []string{file1, file2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "first.txt") {
		t.Error("expected output to contain first attachment filename")
	}

	if !strings.Contains(output, "second.txt") {
		t.Error("expected output to contain second attachment filename")
	}
}

func TestBuildMultipartMessage_WithCCAndBCC(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dummy.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := buildMultipartMessage("to@example.com", "Subject", "Body text", "cc@example.com", "bcc@example.com", []string{filePath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := string(result)

	if !strings.Contains(output, "Cc: cc@example.com") {
		t.Error("expected output to contain Cc header")
	}

	if !strings.Contains(output, "Bcc: bcc@example.com") {
		t.Error("expected output to contain Bcc header")
	}
}

func TestBuildMultipartMessage_NonexistentFileReturnsError(t *testing.T) {
	t.Parallel()
	_, err := buildMultipartMessage("to@example.com", "Subject", "Body text", "", "", []string{"/nonexistent/path/nofile.txt"})
	if err == nil {
		t.Fatal("expected error for nonexistent attachment file, got nil")
	}

	if !strings.Contains(err.Error(), "nofile.txt") {
		t.Errorf("expected error to reference the missing file name, got: %v", err)
	}
}
