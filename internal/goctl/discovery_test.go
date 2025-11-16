package goctl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverGoctl(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*testing.T) func()
		expectError bool
		checkPath   bool
	}{
		{
			name: "goctl in PATH",
			setupFunc: func(t *testing.T) func() {
				// Check if goctl exists in PATH (skip if not available)
				if _, err := exec.LookPath("goctl"); err != nil {
					t.Skip("goctl not found in PATH, skipping test")
				}
				return func() {} // No cleanup needed
			},
			expectError: false,
			checkPath:   true,
		},
		{
			name: "GOCTL_PATH environment variable",
			setupFunc: func(t *testing.T) func() {
				// Create temporary executable
				tmpDir := t.TempDir()
				goctlPath := filepath.Join(tmpDir, "goctl")

				// Create a simple executable script
				content := []byte("#!/bin/sh\necho 'test goctl'")
				if err := os.WriteFile(goctlPath, content, 0755); err != nil {
					t.Fatalf("failed to create test executable: %v", err)
				}

				// Set environment variable
				oldPath := os.Getenv("GOCTL_PATH")
				os.Setenv("GOCTL_PATH", goctlPath)

				return func() {
					os.Setenv("GOCTL_PATH", oldPath)
				}
			},
			expectError: false,
			checkPath:   true,
		},
		{
			name: "goctl in common location",
			setupFunc: func(t *testing.T) func() {
				// Check if goctl exists in ~/go/bin (skip if not available)
				homeDir, err := os.UserHomeDir()
				if err != nil {
					t.Skip("cannot determine home directory")
				}

				goctlPath := filepath.Join(homeDir, "go", "bin", "goctl")
				if _, err := os.Stat(goctlPath); err != nil {
					t.Skip("goctl not found in ~/go/bin, skipping test")
				}

				// Clear GOCTL_PATH to force common path search
				oldPath := os.Getenv("GOCTL_PATH")
				os.Unsetenv("GOCTL_PATH")

				return func() {
					if oldPath != "" {
						os.Setenv("GOCTL_PATH", oldPath)
					}
				}
			},
			expectError: false,
			checkPath:   true,
		},
		{
			name: "goctl not found",
			setupFunc: func(t *testing.T) func() {
				// Save original environment
				oldGoctlPath := os.Getenv("GOCTL_PATH")
				oldPath := os.Getenv("PATH")
				oldGoPath := os.Getenv("GOPATH")
				oldHome := os.Getenv("HOME")

				// Clear all paths to force not found
				os.Unsetenv("GOCTL_PATH")
				os.Setenv("PATH", "/nonexistent")
				os.Setenv("GOPATH", "/nonexistent")
				os.Setenv("HOME", "/nonexistent")

				return func() {
					if oldGoctlPath != "" {
						os.Setenv("GOCTL_PATH", oldGoctlPath)
					} else {
						os.Unsetenv("GOCTL_PATH")
					}
					os.Setenv("PATH", oldPath)
					if oldGoPath != "" {
						os.Setenv("GOPATH", oldGoPath)
					} else {
						os.Unsetenv("GOPATH")
					}
					os.Setenv("HOME", oldHome)
				}
			},
			expectError: true,
			checkPath:   false,
		},
		{
			name: "GOCTL_PATH points to non-executable",
			setupFunc: func(t *testing.T) func() {
				// Create temporary non-executable file
				tmpDir := t.TempDir()
				goctlPath := filepath.Join(tmpDir, "goctl")

				// Create a non-executable file
				if err := os.WriteFile(goctlPath, []byte("not executable"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				// Set environment variable and clear other paths
				oldGoctlPath := os.Getenv("GOCTL_PATH")
				oldPath := os.Getenv("PATH")
				oldGoPath := os.Getenv("GOPATH")
				oldHome := os.Getenv("HOME")

				os.Setenv("GOCTL_PATH", goctlPath)
				os.Setenv("PATH", "/nonexistent")
				os.Setenv("GOPATH", "/nonexistent")
				os.Setenv("HOME", "/nonexistent")

				return func() {
					if oldGoctlPath != "" {
						os.Setenv("GOCTL_PATH", oldGoctlPath)
					} else {
						os.Unsetenv("GOCTL_PATH")
					}
					os.Setenv("PATH", oldPath)
					if oldGoPath != "" {
						os.Setenv("GOPATH", oldGoPath)
					} else {
						os.Unsetenv("GOPATH")
					}
					os.Setenv("HOME", oldHome)
				}
			},
			expectError: true,
			checkPath:   false,
		},
		{
			name: "GOPATH/bin fallback",
			setupFunc: func(t *testing.T) func() {
				// Create temporary GOPATH with goctl
				tmpDir := t.TempDir()
				binDir := filepath.Join(tmpDir, "bin")
				if err := os.MkdirAll(binDir, 0755); err != nil {
					t.Fatalf("failed to create bin directory: %v", err)
				}

				goctlPath := filepath.Join(binDir, "goctl")
				content := []byte("#!/bin/sh\necho 'test goctl'")
				if err := os.WriteFile(goctlPath, content, 0755); err != nil {
					t.Fatalf("failed to create test executable: %v", err)
				}

				// Save and set environment
				oldGoctlPath := os.Getenv("GOCTL_PATH")
				oldGoPath := os.Getenv("GOPATH")
				oldPath := os.Getenv("PATH")

				os.Unsetenv("GOCTL_PATH")
				os.Setenv("GOPATH", tmpDir)
				os.Setenv("PATH", "/nonexistent") // Clear PATH to test GOPATH fallback

				return func() {
					if oldGoctlPath != "" {
						os.Setenv("GOCTL_PATH", oldGoctlPath)
					}
					if oldGoPath != "" {
						os.Setenv("GOPATH", oldGoPath)
					} else {
						os.Unsetenv("GOPATH")
					}
					os.Setenv("PATH", oldPath)
				}
			},
			expectError: false,
			checkPath:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			path, err := DiscoverGoctl()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, but got none")
				}
				// Verify error message is actionable
				if err != nil && !strings.Contains(err.Error(), "go install") {
					t.Errorf("error message should contain installation instructions, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.checkPath && err == nil {
				if path == "" {
					t.Errorf("expected non-empty path, got empty string")
				}
				if !filepath.IsAbs(path) {
					t.Errorf("expected absolute path, got: %s", path)
				}
			}
		})
	}
}

func TestDiscoverGoctlPriority(t *testing.T) {
	// This test verifies that GOCTL_PATH takes priority over other methods
	tmpDir := t.TempDir()
	goctlPath := filepath.Join(tmpDir, "custom_goctl")

	// Create custom executable
	content := []byte("#!/bin/sh\necho 'custom goctl'")
	if err := os.WriteFile(goctlPath, content, 0755); err != nil {
		t.Fatalf("failed to create test executable: %v", err)
	}

	// Set GOCTL_PATH
	oldPath := os.Getenv("GOCTL_PATH")
	os.Setenv("GOCTL_PATH", goctlPath)
	defer func() {
		if oldPath != "" {
			os.Setenv("GOCTL_PATH", oldPath)
		} else {
			os.Unsetenv("GOCTL_PATH")
		}
	}()

	path, err := DiscoverGoctl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != goctlPath {
		t.Errorf("expected GOCTL_PATH to take priority, got: %s, want: %s", path, goctlPath)
	}
}

func TestDiscoverGoctlCache(t *testing.T) {
	// Test that multiple calls return consistent results
	path1, err1 := DiscoverGoctl()
	path2, err2 := DiscoverGoctl()

	// Both should have the same error state
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("inconsistent error state: err1=%v, err2=%v", err1, err2)
	}

	// If successful, paths should be identical
	if err1 == nil && path1 != path2 {
		t.Errorf("inconsistent paths: path1=%s, path2=%s", path1, path2)
	}
}
