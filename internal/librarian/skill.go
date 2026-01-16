package librarian

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents a reusable skill definition that can be injected into beads.
// Skills contain domain knowledge, patterns, and file references that help
// agents understand how to approach specific types of work.
type Skill struct {
	// ID is the unique identifier for this skill (e.g., "go-testing")
	ID string `yaml:"id" json:"id"`

	// Name is a human-readable name for the skill
	Name string `yaml:"name" json:"name"`

	// Description explains what this skill covers
	Description string `yaml:"description" json:"description"`

	// Triggers define when this skill should be injected
	Triggers SkillTriggers `yaml:"triggers" json:"triggers"`

	// Content defines what gets injected into the enrichment
	Content SkillContent `yaml:"content" json:"content"`

	// Priority determines injection order (higher = earlier, default 0)
	Priority int `yaml:"priority,omitempty" json:"priority,omitempty"`

	// Exclusive means only one skill in this group can be injected
	Exclusive string `yaml:"exclusive,omitempty" json:"exclusive,omitempty"`
}

// SkillTriggers defines conditions for skill injection.
// A skill is triggered if ANY condition matches (OR logic).
type SkillTriggers struct {
	// Labels triggers on bead labels (e.g., "gt:testing", "domain:auth")
	Labels []string `yaml:"labels,omitempty" json:"labels,omitempty"`

	// TitlePatterns are regex patterns to match against bead titles
	TitlePatterns []string `yaml:"title_patterns,omitempty" json:"title_patterns,omitempty"`

	// DescriptionPatterns are regex patterns for bead descriptions
	DescriptionPatterns []string `yaml:"description_patterns,omitempty" json:"description_patterns,omitempty"`

	// Keywords are simple word matches (case-insensitive)
	Keywords []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`

	// ParentLabels triggers when parent bead has these labels
	ParentLabels []string `yaml:"parent_labels,omitempty" json:"parent_labels,omitempty"`

	// BeadTypes triggers on specific bead types
	BeadTypes []string `yaml:"bead_types,omitempty" json:"bead_types,omitempty"`
}

// SkillContent defines what content gets injected.
type SkillContent struct {
	// Files to include in "Files to Read" section
	Files []SkillFile `yaml:"files,omitempty" json:"files,omitempty"`

	// Patterns to include in "Key Patterns" section
	Patterns []SkillPattern `yaml:"patterns,omitempty" json:"patterns,omitempty"`

	// Documentation links to include
	Documentation []SkillDoc `yaml:"documentation,omitempty" json:"documentation,omitempty"`

	// ContextNotes to append to "Context Notes" section
	ContextNotes []string `yaml:"context_notes,omitempty" json:"context_notes,omitempty"`

	// PriorWorkQuery is a bd query to find related prior work
	PriorWorkQuery string `yaml:"prior_work_query,omitempty" json:"prior_work_query,omitempty"`
}

// SkillFile represents a file reference in a skill.
type SkillFile struct {
	// Path is relative from rig root (supports glob patterns)
	Path string `yaml:"path" json:"path"`

	// Lines optionally specifies a line range (e.g., "45-120")
	Lines string `yaml:"lines,omitempty" json:"lines,omitempty"`

	// Description explains why to read this file
	Description string `yaml:"description" json:"description"`

	// Optional means skip if file doesn't exist
	Optional bool `yaml:"optional,omitempty" json:"optional,omitempty"`
}

// SkillPattern represents a pattern in a skill.
type SkillPattern struct {
	// Name of the pattern
	Name string `yaml:"name" json:"name"`

	// Description of what the pattern is
	Description string `yaml:"description" json:"description"`

	// Example path to see the pattern in use
	Example string `yaml:"example,omitempty" json:"example,omitempty"`
}

// SkillDoc represents a documentation link.
type SkillDoc struct {
	// Title of the documentation
	Title string `yaml:"title" json:"title"`

	// URL of the documentation
	URL string `yaml:"url" json:"url"`

	// Description of what it covers
	Description string `yaml:"description" json:"description"`
}

// BeadContext contains the extracted context from a bead for skill matching.
type BeadContext struct {
	ID          string
	Title       string
	Description string
	Labels      []string
	Type        string
	ParentID    string
	ParentLabel []string // Labels from parent bead if available
}

// SkillRegistry manages skill definitions and matching.
type SkillRegistry struct {
	skills   []*Skill
	skillDir string
}

// NewSkillRegistry creates a new skill registry for a town.
func NewSkillRegistry(townRoot string) *SkillRegistry {
	return &SkillRegistry{
		skills:   make([]*Skill, 0),
		skillDir: filepath.Join(townRoot, "librarian", "skills"),
	}
}

// LoadSkills loads all skill definitions from the skills directory.
func (r *SkillRegistry) LoadSkills() error {
	// Check if skills directory exists
	if _, err := os.Stat(r.skillDir); os.IsNotExist(err) {
		// No skills directory is fine - just return empty
		return nil
	}

	// Walk the skills directory for YAML files
	return filepath.Walk(r.skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process YAML files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		// Load the skill
		skill, err := r.loadSkillFile(path)
		if err != nil {
			// Log warning but continue loading other skills
			fmt.Fprintf(os.Stderr, "Warning: failed to load skill %s: %v\n", path, err)
			return nil
		}

		r.skills = append(r.skills, skill)
		return nil
	})
}

// loadSkillFile loads a single skill from a YAML file.
func (r *SkillRegistry) loadSkillFile(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading skill file: %w", err)
	}

	var skill Skill
	if err := yaml.Unmarshal(data, &skill); err != nil {
		return nil, fmt.Errorf("parsing skill YAML: %w", err)
	}

	// Validate required fields
	if skill.ID == "" {
		skill.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if skill.Name == "" {
		skill.Name = skill.ID
	}

	return &skill, nil
}

// MatchSkills returns all skills that match the given bead context.
func (r *SkillRegistry) MatchSkills(ctx *BeadContext) []*Skill {
	// First pass: collect all matching skills
	var allMatched []*Skill
	for _, skill := range r.skills {
		if r.skillMatches(skill, ctx) {
			allMatched = append(allMatched, skill)
		}
	}

	// Sort by priority (higher first) BEFORE applying exclusive filtering
	sortSkillsByPriority(allMatched)

	// Second pass: filter exclusive groups (highest priority wins)
	var result []*Skill
	exclusiveGroups := make(map[string]bool)
	for _, skill := range allMatched {
		if skill.Exclusive != "" && exclusiveGroups[skill.Exclusive] {
			continue
		}
		result = append(result, skill)
		if skill.Exclusive != "" {
			exclusiveGroups[skill.Exclusive] = true
		}
	}

	return result
}

// skillMatches checks if a skill's triggers match the bead context.
func (r *SkillRegistry) skillMatches(skill *Skill, ctx *BeadContext) bool {
	triggers := skill.Triggers

	// Check labels
	for _, triggerLabel := range triggers.Labels {
		for _, beadLabel := range ctx.Labels {
			if matchLabel(triggerLabel, beadLabel) {
				return true
			}
		}
	}

	// Check title patterns
	for _, pattern := range triggers.TitlePatterns {
		if matchPattern(pattern, ctx.Title) {
			return true
		}
	}

	// Check description patterns
	for _, pattern := range triggers.DescriptionPatterns {
		if matchPattern(pattern, ctx.Description) {
			return true
		}
	}

	// Check keywords (case-insensitive in title or description)
	combinedText := strings.ToLower(ctx.Title + " " + ctx.Description)
	for _, keyword := range triggers.Keywords {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			return true
		}
	}

	// Check parent labels
	for _, triggerLabel := range triggers.ParentLabels {
		for _, parentLabel := range ctx.ParentLabel {
			if matchLabel(triggerLabel, parentLabel) {
				return true
			}
		}
	}

	// Check bead types
	for _, beadType := range triggers.BeadTypes {
		if strings.EqualFold(beadType, ctx.Type) {
			return true
		}
	}

	return false
}

// matchLabel checks if a trigger label matches a bead label.
// Supports wildcards (e.g., "domain:*" matches "domain:auth").
func matchLabel(trigger, label string) bool {
	if strings.HasSuffix(trigger, "*") {
		prefix := strings.TrimSuffix(trigger, "*")
		return strings.HasPrefix(label, prefix)
	}
	return strings.EqualFold(trigger, label)
}

// matchPattern checks if a regex pattern matches text.
func matchPattern(pattern, text string) bool {
	re, err := regexp.Compile("(?i)" + pattern) // Case-insensitive
	if err != nil {
		return false
	}
	return re.MatchString(text)
}

// sortSkillsByPriority sorts skills by priority (higher first).
func sortSkillsByPriority(skills []*Skill) {
	for i := 0; i < len(skills)-1; i++ {
		for j := i + 1; j < len(skills); j++ {
			if skills[j].Priority > skills[i].Priority {
				skills[i], skills[j] = skills[j], skills[i]
			}
		}
	}
}

// GetSkill returns a skill by ID.
func (r *SkillRegistry) GetSkill(id string) *Skill {
	for _, skill := range r.skills {
		if skill.ID == id {
			return skill
		}
	}
	return nil
}

// AllSkills returns all loaded skills.
func (r *SkillRegistry) AllSkills() []*Skill {
	return r.skills
}

// SkillsDir returns the skills directory path.
func (r *SkillRegistry) SkillsDir() string {
	return r.skillDir
}

// AddSkill adds a skill to the registry (useful for testing).
func (r *SkillRegistry) AddSkill(skill *Skill) {
	r.skills = append(r.skills, skill)
}
