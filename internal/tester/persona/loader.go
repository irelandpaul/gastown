package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads a persona from a YAML file.
func LoadFromFile(path string) (*Persona, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read persona file: %w", err)
	}

	var p Persona
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse persona YAML: %w", err)
	}

	if !p.IsValid() {
		return nil, fmt.Errorf("invalid persona: missing required fields")
	}

	return &p, nil
}

// LoadFromDir loads all personas from a directory.
func LoadFromDir(dir string) ([]*Persona, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read persona directory: %w", err)
	}

	var personas []*Persona
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		p, err := LoadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", entry.Name(), err)
		}

		personas = append(personas, p)
	}

	return personas, nil
}

// Resolve resolves a persona reference to a Persona.
// It first checks built-in personas by name, then tries to load from file.
func Resolve(ref string) (*Persona, error) {
	// Check built-in personas first
	if p := Get(ref); p != nil {
		return p, nil
	}

	// Check if it's a file path
	if strings.HasSuffix(ref, ".yaml") || strings.HasSuffix(ref, ".yml") {
		return LoadFromFile(ref)
	}

	return nil, fmt.Errorf("unknown persona: %s (not a built-in persona or file path)", ref)
}

// ToYAML converts a persona to YAML format.
func (p *Persona) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

// SaveToFile saves a persona to a YAML file.
func (p *Persona) SaveToFile(path string) error {
	data, err := p.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to marshal persona: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
