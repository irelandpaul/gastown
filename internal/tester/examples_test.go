package tester

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExampleScenarios validates that all example scenario files parse correctly.
func TestExampleScenarios(t *testing.T) {
	// Find the project root by looking for go.mod
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to find project root
	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Skip("Could not find project root")
			return
		}
		root = parent
	}

	examplesDir := filepath.Join(root, "docs", "examples", "scenarios")

	// Check if examples directory exists
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		t.Skip("Examples directory does not exist")
		return
	}

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("Failed to read examples directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(examplesDir, entry.Name())
			s, err := ParseScenarioFile(path)
			if err != nil {
				t.Errorf("Failed to parse %s: %v", entry.Name(), err)
				return
			}

			// Verify required fields are present
			if s.Scenario == "" {
				t.Error("Scenario name is empty")
			}
			if s.Persona == "" {
				t.Error("Persona is empty")
			}
			if s.Goal == "" {
				t.Error("Goal is empty")
			}
			if len(s.SuccessCriteria) == 0 {
				t.Error("Success criteria are empty")
			}
			if s.Environment.URL == "" {
				t.Error("Environment URL is empty")
			}

			t.Logf("Parsed scenario: %s (persona: %s)", s.Scenario, s.Persona)
		})
	}
}
