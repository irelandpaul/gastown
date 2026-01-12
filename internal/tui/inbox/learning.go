package inbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/steveyegge/gastown/internal/workspace"
)

// ClassificationRule defines a rule for overriding message type inference.
type ClassificationRule struct {
	ID             string      `json:"id"`
	SubjectPattern string      `json:"subject_pattern,omitempty"`
	BodyPattern    string      `json:"body_pattern,omitempty"`
	FromPattern    string      `json:"from_pattern,omitempty"`
	TargetType     MessageType `json:"target_type"`
}

// LearningSystem manages user-defined classification rules.
type LearningSystem struct {
	Rules []ClassificationRule `json:"rules"`
	path  string
}

// NewLearningSystem loads the learning system from the workspace config.
func NewLearningSystem(workDir string) *LearningSystem {
	townRoot, _ := workspace.FindFromCwd()
	if townRoot == "" {
		townRoot = workDir
	}
	path := filepath.Join(townRoot, "config", "inbox_rules.json")

	ls := &LearningSystem{
		Rules: make([]ClassificationRule, 0),
		path:  path,
	}
	ls.load()
	return ls
}

// load reads rules from disk.
func (ls *LearningSystem) load() {
	data, err := os.ReadFile(ls.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, ls)
}

// save writes rules to disk.
func (ls *LearningSystem) save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ls.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ls, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ls.path, data, 0644)
}

// Classify returns the overridden message type if a rule matches, or the original type.
func (ls *LearningSystem) Classify(msg Message) (MessageType, bool) {
	for _, rule := range ls.Rules {
		if ls.matches(rule, msg) {
			return rule.TargetType, true
		}
	}
	return msg.Type, false
}

// matches checks if a rule matches a message.
func (ls *LearningSystem) matches(rule ClassificationRule, msg Message) bool {
	if rule.SubjectPattern != "" && !strings.Contains(strings.ToLower(msg.Subject), strings.ToLower(rule.SubjectPattern)) {
		return false
	}
	if rule.BodyPattern != "" && !strings.Contains(strings.ToLower(msg.Body), strings.ToLower(rule.BodyPattern)) {
		return false
	}
	if rule.FromPattern != "" && !strings.Contains(strings.ToLower(msg.From), strings.ToLower(rule.FromPattern)) {
		return false
	}
	return true
}

// Learn adds a new rule based on a message and a target type.
func (ls *LearningSystem) Learn(msg Message, targetType MessageType) error {
	// Simple learning: use the subject as the pattern
	// We could make this smarter by extracting common prefixes
	pattern := msg.Subject
	// Strip "Re: " etc
	pattern = strings.TrimPrefix(pattern, "Re: ")
	pattern = strings.TrimPrefix(pattern, "RE: ")
	pattern = strings.TrimSpace(pattern)

	rule := ClassificationRule{
		ID:             msg.ID, // Use msg ID as initial rule ID
		SubjectPattern: pattern,
		TargetType:     targetType,
	}

	ls.Rules = append(ls.Rules, rule)
	return ls.save()
}
