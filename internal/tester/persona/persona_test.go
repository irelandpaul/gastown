package persona

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltInPersonas(t *testing.T) {
	// Verify Sarah
	if Sarah.Name != "Sarah" {
		t.Errorf("expected Sarah.Name=Sarah, got %s", Sarah.Name)
	}
	if Sarah.TechComfort != TechComfortLow {
		t.Errorf("expected Sarah.TechComfort=low, got %s", Sarah.TechComfort)
	}
	if Sarah.Patience != PatienceMedium {
		t.Errorf("expected Sarah.Patience=medium, got %s", Sarah.Patience)
	}
	if !Sarah.IsValid() {
		t.Error("expected Sarah to be valid")
	}

	// Verify Mike
	if Mike.Name != "Mike" {
		t.Errorf("expected Mike.Name=Mike, got %s", Mike.Name)
	}
	if Mike.TechComfort != TechComfortHigh {
		t.Errorf("expected Mike.TechComfort=high, got %s", Mike.TechComfort)
	}
	if Mike.Patience != PatienceLow {
		t.Errorf("expected Mike.Patience=low, got %s", Mike.Patience)
	}
	if !Mike.IsValid() {
		t.Error("expected Mike to be valid")
	}

	// Verify Emma
	if Emma.Name != "Emma" {
		t.Errorf("expected Emma.Name=Emma, got %s", Emma.Name)
	}
	if Emma.TechComfort != TechComfortMedium {
		t.Errorf("expected Emma.TechComfort=medium, got %s", Emma.TechComfort)
	}
	if Emma.Device != DeviceMobile {
		t.Errorf("expected Emma.Device=mobile, got %s", Emma.Device)
	}
	if !Emma.IsValid() {
		t.Error("expected Emma to be valid")
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		expected *Persona
	}{
		{"sarah", &Sarah},
		{"Sarah", &Sarah},
		{"SARAH", &Sarah},
		{"mike", &Mike},
		{"Mike", &Mike},
		{"emma", &Emma},
		{"rose", &Rose},
		{"unknown", nil},
	}

	for _, tt := range tests {
		result := Get(tt.name)
		if tt.expected == nil {
			if result != nil {
				t.Errorf("Get(%s) expected nil, got %v", tt.name, result)
			}
		} else {
			if result == nil {
				t.Errorf("Get(%s) expected persona, got nil", tt.name)
			} else if result.Name != tt.expected.Name {
				t.Errorf("Get(%s) expected %s, got %s", tt.name, tt.expected.Name, result.Name)
			}
		}
	}
}

func TestList(t *testing.T) {
	names := List()
	if len(names) != 4 {
		t.Errorf("expected 4 personas, got %d", len(names))
	}

	expected := map[string]bool{"sarah": true, "mike": true, "emma": true, "rose": true}
	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected persona in list: %s", name)
		}
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name    string
		persona Persona
		valid   bool
	}{
		{
			name: "valid persona",
			persona: Persona{
				Name:        "Test",
				Role:        "parent",
				Context:     "Test context",
				TechComfort: TechComfortMedium,
				Patience:    PatienceMedium,
				Device:      DeviceDesktop,
			},
			valid: true,
		},
		{
			name: "missing name",
			persona: Persona{
				Role:        "parent",
				Context:     "Test context",
				TechComfort: TechComfortMedium,
				Patience:    PatienceMedium,
				Device:      DeviceDesktop,
			},
			valid: false,
		},
		{
			name: "invalid tech comfort",
			persona: Persona{
				Name:        "Test",
				Role:        "parent",
				Context:     "Test context",
				TechComfort: "invalid",
				Patience:    PatienceMedium,
				Device:      DeviceDesktop,
			},
			valid: false,
		},
		{
			name: "invalid patience",
			persona: Persona{
				Name:        "Test",
				Role:        "parent",
				Context:     "Test context",
				TechComfort: TechComfortMedium,
				Patience:    "invalid",
				Device:      DeviceDesktop,
			},
			valid: false,
		},
		{
			name: "invalid device",
			persona: Persona{
				Name:        "Test",
				Role:        "parent",
				Context:     "Test context",
				TechComfort: TechComfortMedium,
				Patience:    PatienceMedium,
				Device:      "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.persona.IsValid() != tt.valid {
				t.Errorf("expected IsValid()=%v, got %v", tt.valid, tt.persona.IsValid())
			}
		})
	}
}

func TestBehaviorHints(t *testing.T) {
	// Low tech should have more hints
	hints := Sarah.BehaviorHints()
	if len(hints) == 0 {
		t.Error("expected Sarah to have behavior hints")
	}

	// High tech should have different hints
	mikeHints := Mike.BehaviorHints()
	if len(mikeHints) == 0 {
		t.Error("expected Mike to have behavior hints")
	}

	// Mobile should have device-specific hints
	emmaHints := Emma.BehaviorHints()
	hasMobileHint := false
	for _, h := range emmaHints {
		if h == "Use touch gestures" || h == "Expect mobile-optimized interface" {
			hasMobileHint = true
			break
		}
	}
	if !hasMobileHint {
		t.Error("expected Emma to have mobile-specific hints")
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid persona file
	validYAML := `name: TestUser
role: parent
context: Test context for testing purposes.
tech_comfort: medium
patience: high
device: tablet
tags:
  - test
  - example
`
	validPath := filepath.Join(tmpDir, "valid.yaml")
	os.WriteFile(validPath, []byte(validYAML), 0644)

	p, err := LoadFromFile(validPath)
	if err != nil {
		t.Fatalf("failed to load valid persona: %v", err)
	}

	if p.Name != "TestUser" {
		t.Errorf("expected Name=TestUser, got %s", p.Name)
	}
	if p.TechComfort != TechComfortMedium {
		t.Errorf("expected TechComfort=medium, got %s", p.TechComfort)
	}
	if p.Patience != PatienceHigh {
		t.Errorf("expected Patience=high, got %s", p.Patience)
	}
	if p.Device != DeviceTablet {
		t.Errorf("expected Device=tablet, got %s", p.Device)
	}
	if len(p.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(p.Tags))
	}

	// Test invalid file
	invalidYAML := `name: Invalid
role: parent
context: Missing other fields
`
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	os.WriteFile(invalidPath, []byte(invalidYAML), 0644)

	_, err = LoadFromFile(invalidPath)
	if err == nil {
		t.Error("expected error for invalid persona file")
	}

	// Test non-existent file
	_, err = LoadFromFile(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create persona files
	personas := []string{
		`name: User1
role: parent
context: First test user.
tech_comfort: low
patience: low
device: desktop`,
		`name: User2
role: parent
context: Second test user.
tech_comfort: high
patience: high
device: mobile`,
	}

	os.WriteFile(filepath.Join(tmpDir, "user1.yaml"), []byte(personas[0]), 0644)
	os.WriteFile(filepath.Join(tmpDir, "user2.yml"), []byte(personas[1]), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a persona"), 0644)

	loaded, err := LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load personas from dir: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 personas, got %d", len(loaded))
	}
}

func TestResolve(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a custom persona file
	customYAML := `name: Custom
role: admin
context: Custom persona for testing.
tech_comfort: high
patience: medium
device: desktop`
	customPath := filepath.Join(tmpDir, "custom.yaml")
	os.WriteFile(customPath, []byte(customYAML), 0644)

	// Test resolving built-in
	p, err := Resolve("sarah")
	if err != nil {
		t.Fatalf("failed to resolve sarah: %v", err)
	}
	if p.Name != "Sarah" {
		t.Errorf("expected Sarah, got %s", p.Name)
	}

	// Test resolving file path
	p, err = Resolve(customPath)
	if err != nil {
		t.Fatalf("failed to resolve custom: %v", err)
	}
	if p.Name != "Custom" {
		t.Errorf("expected Custom, got %s", p.Name)
	}

	// Test unknown persona
	_, err = Resolve("unknown")
	if err == nil {
		t.Error("expected error for unknown persona")
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()

	p := Persona{
		Name:        "Saved",
		Role:        "parent",
		Context:     "A saved persona.",
		TechComfort: TechComfortLow,
		Patience:    PatienceHigh,
		Device:      DeviceDesktop,
		Tags:        []string{"saved"},
	}

	path := filepath.Join(tmpDir, "subdir", "saved.yaml")
	err := p.SaveToFile(path)
	if err != nil {
		t.Fatalf("failed to save persona: %v", err)
	}

	// Load it back
	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("failed to load saved persona: %v", err)
	}

	if loaded.Name != p.Name {
		t.Errorf("expected Name=%s, got %s", p.Name, loaded.Name)
	}
	if loaded.TechComfort != p.TechComfort {
		t.Errorf("expected TechComfort=%s, got %s", p.TechComfort, loaded.TechComfort)
	}
}

func TestToYAML(t *testing.T) {
	data, err := Sarah.ToYAML()
	if err != nil {
		t.Fatalf("failed to convert to YAML: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty YAML output")
	}

	// Should contain key fields
	yaml := string(data)
	if !contains(yaml, "name: Sarah") {
		t.Error("YAML should contain name")
	}
	if !contains(yaml, "tech_comfort: low") {
		t.Error("YAML should contain tech_comfort")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
