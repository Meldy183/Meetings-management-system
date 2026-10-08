package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpNeedsNoDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var stdout, stderr bytes.Buffer
	if err := run(context.Background(), []string{"-help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "read-only") {
		t.Fatal("help should describe the read-only export")
	}
}

func TestRunRequiresExplicitDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	path := filepath.Join(t.TempDir(), "participants.xlsx")
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{"-out", path}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("expected explicit connection requirement, got %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("failed export should not create output: %v", err)
	}
}

func TestRunRejectsExistingOutputBeforeConnecting(t *testing.T) {
	t.Setenv("DATABASE_URL", "not a valid DSN")
	path := filepath.Join(t.TempDir(), "existing.xlsx")
	if err := os.WriteFile(path, []byte("existing data"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := run(context.Background(), []string{"-out", path}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "output already exists") {
		t.Fatalf("expected output protection before any database connection, got %v", err)
	}
}
