package fixer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinguoxing/mcp-gozero/internal/fixer"
)

func TestCleanupStyleConflicts(t *testing.T) {
	t.Run("cleanup_gozero_when_choosing_go_zero", func(t *testing.T) {
		// Create a temporary directory for this subtest
		tmpDir, err := os.MkdirTemp("", "style_conflict_test_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create internal/svc directory
		svcDir := filepath.Join(tmpDir, "internal", "svc")
		if err := os.MkdirAll(svcDir, 0755); err != nil {
			t.Fatalf("Failed to create svc dir: %v", err)
		}

		// Create both files
		goZeroFile := filepath.Join(svcDir, "service_context.go")
		gozeroFile := filepath.Join(svcDir, "servicecontext.go")

		if err := os.WriteFile(goZeroFile, []byte("package svc\ntype ServiceContext struct{}"), 0644); err != nil {
			t.Fatalf("Failed to create go_zero file: %v", err)
		}
		if err := os.WriteFile(gozeroFile, []byte("package svc\ntype ServiceContext struct{}"), 0644); err != nil {
			t.Fatalf("Failed to create gozero file: %v", err)
		}

		// Run cleanup with go_zero style
		if err := fixer.CleanupStyleConflicts(tmpDir, "go_zero"); err != nil {
			t.Fatalf("CleanupStyleConflicts failed: %v", err)
		}

		// Verify go_zero file exists
		if _, err := os.Stat(goZeroFile); os.IsNotExist(err) {
			t.Error("go_zero style file should exist")
		}

		// Verify gozero file was removed
		if _, err := os.Stat(gozeroFile); !os.IsNotExist(err) {
			t.Error("gozero style file should have been removed")
		}
	})

	t.Run("cleanup_go_zero_when_choosing_gozero", func(t *testing.T) {
		// Create a temporary directory for this subtest
		tmpDir, err := os.MkdirTemp("", "style_conflict_test_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create internal/svc directory
		svcDir := filepath.Join(tmpDir, "internal", "svc")
		if err := os.MkdirAll(svcDir, 0755); err != nil {
			t.Fatalf("Failed to create svc dir: %v", err)
		}

		// Create both files again
		goZeroFile := filepath.Join(svcDir, "service_context.go")
		gozeroFile := filepath.Join(svcDir, "servicecontext.go")

		if err := os.WriteFile(goZeroFile, []byte("package svc\ntype ServiceContext struct{}"), 0644); err != nil {
			t.Fatalf("Failed to create go_zero file: %v", err)
		}
		if err := os.WriteFile(gozeroFile, []byte("package svc\ntype ServiceContext struct{}"), 0644); err != nil {
			t.Fatalf("Failed to create gozero file: %v", err)
		}

		// Run cleanup with gozero style
		if err := fixer.CleanupStyleConflicts(tmpDir, "gozero"); err != nil {
			t.Fatalf("CleanupStyleConflicts failed: %v", err)
		}

		// Verify go_zero file was removed
		if _, err := os.Stat(goZeroFile); !os.IsNotExist(err) {
			t.Error("go_zero style file should have been removed")
		}

		// Verify gozero file exists
		if _, err := os.Stat(gozeroFile); os.IsNotExist(err) {
			t.Error("gozero style file should exist")
		}
	})
}

func TestDetectExistingStyle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "detect_style_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svcDir := filepath.Join(tmpDir, "internal", "svc")
	if err := os.MkdirAll(svcDir, 0755); err != nil {
		t.Fatalf("Failed to create svc dir: %v", err)
	}

	t.Run("detect_go_zero_style", func(t *testing.T) {
		goZeroFile := filepath.Join(svcDir, "service_context.go")
		if err := os.WriteFile(goZeroFile, []byte("package svc"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		style := fixer.DetectExistingStyle(tmpDir)
		if style != "go_zero" {
			t.Errorf("Expected 'go_zero', got '%s'", style)
		}

		os.Remove(goZeroFile)
	})

	t.Run("detect_gozero_style", func(t *testing.T) {
		gozeroFile := filepath.Join(svcDir, "servicecontext.go")
		if err := os.WriteFile(gozeroFile, []byte("package svc"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		style := fixer.DetectExistingStyle(tmpDir)
		if style != "gozero" {
			t.Errorf("Expected 'gozero', got '%s'", style)
		}

		os.Remove(gozeroFile)
	})

	t.Run("detect_no_style", func(t *testing.T) {
		style := fixer.DetectExistingStyle(tmpDir)
		if style != "" {
			t.Errorf("Expected empty string, got '%s'", style)
		}
	})
}

func TestValidateNoStyleConflicts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validate_conflicts_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	svcDir := filepath.Join(tmpDir, "internal", "svc")
	if err := os.MkdirAll(svcDir, 0755); err != nil {
		t.Fatalf("Failed to create svc dir: %v", err)
	}

	t.Run("no_conflicts", func(t *testing.T) {
		goZeroFile := filepath.Join(svcDir, "service_context.go")
		if err := os.WriteFile(goZeroFile, []byte("package svc"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		if err := fixer.ValidateNoStyleConflicts(tmpDir); err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		os.Remove(goZeroFile)
	})

	t.Run("with_conflicts", func(t *testing.T) {
		goZeroFile := filepath.Join(svcDir, "service_context.go")
		gozeroFile := filepath.Join(svcDir, "servicecontext.go")

		if err := os.WriteFile(goZeroFile, []byte("package svc"), 0644); err != nil {
			t.Fatalf("Failed to create go_zero file: %v", err)
		}
		if err := os.WriteFile(gozeroFile, []byte("package svc"), 0644); err != nil {
			t.Fatalf("Failed to create gozero file: %v", err)
		}

		if err := fixer.ValidateNoStyleConflicts(tmpDir); err == nil {
			t.Error("Expected error for conflicts, got nil")
		}

		os.Remove(goZeroFile)
		os.Remove(gozeroFile)
	})
}
