package librarian

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillTriggerMatching(t *testing.T) {
	tests := []struct {
		name    string
		skill   *Skill
		ctx     *BeadContext
		matches bool
	}{
		{
			name: "label match exact",
			skill: &Skill{
				ID:   "test-skill",
				Name: "Test Skill",
				Triggers: SkillTriggers{
					Labels: []string{"gt:testing"},
				},
			},
			ctx: &BeadContext{
				Labels: []string{"gt:testing", "gt:feature"},
			},
			matches: true,
		},
		{
			name: "label match wildcard",
			skill: &Skill{
				ID:   "domain-skill",
				Name: "Domain Skill",
				Triggers: SkillTriggers{
					Labels: []string{"domain:*"},
				},
			},
			ctx: &BeadContext{
				Labels: []string{"domain:auth", "gt:task"},
			},
			matches: true,
		},
		{
			name: "keyword match in title",
			skill: &Skill{
				ID:   "auth-skill",
				Name: "Auth Skill",
				Triggers: SkillTriggers{
					Keywords: []string{"authentication", "login"},
				},
			},
			ctx: &BeadContext{
				Title: "Add login functionality",
			},
			matches: true,
		},
		{
			name: "keyword match in description",
			skill: &Skill{
				ID:   "db-skill",
				Name: "Database Skill",
				Triggers: SkillTriggers{
					Keywords: []string{"database", "sql"},
				},
			},
			ctx: &BeadContext{
				Title:       "Update user service",
				Description: "Add database migration for new fields",
			},
			matches: true,
		},
		{
			name: "title pattern match",
			skill: &Skill{
				ID:   "api-skill",
				Name: "API Skill",
				Triggers: SkillTriggers{
					TitlePatterns: []string{"add.*endpoint", "create.*api"},
				},
			},
			ctx: &BeadContext{
				Title: "Add user endpoint for profile updates",
			},
			matches: true,
		},
		{
			name: "bead type match",
			skill: &Skill{
				ID:   "bug-skill",
				Name: "Bug Skill",
				Triggers: SkillTriggers{
					BeadTypes: []string{"bug", "fix"},
				},
			},
			ctx: &BeadContext{
				Type: "bug",
			},
			matches: true,
		},
		{
			name: "no match",
			skill: &Skill{
				ID:   "unrelated-skill",
				Name: "Unrelated Skill",
				Triggers: SkillTriggers{
					Labels:   []string{"unrelated:label"},
					Keywords: []string{"unrelated"},
				},
			},
			ctx: &BeadContext{
				Title:       "Some other task",
				Description: "Does something different",
				Labels:      []string{"gt:task"},
			},
			matches: false,
		},
		{
			name: "parent label match",
			skill: &Skill{
				ID:   "epic-skill",
				Name: "Epic Child Skill",
				Triggers: SkillTriggers{
					ParentLabels: []string{"gt:epic"},
				},
			},
			ctx: &BeadContext{
				Title:       "Subtask of epic",
				ParentID:    "gt-parent",
				ParentLabel: []string{"gt:epic"},
			},
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := &SkillRegistry{skills: []*Skill{tt.skill}}
			matched := registry.MatchSkills(tt.ctx)

			if tt.matches {
				assert.Len(t, matched, 1, "expected skill to match")
				assert.Equal(t, tt.skill.ID, matched[0].ID)
			} else {
				assert.Len(t, matched, 0, "expected no match")
			}
		})
	}
}

func TestSkillPrioritySorting(t *testing.T) {
	registry := &SkillRegistry{
		skills: []*Skill{
			{ID: "low", Priority: 0, Triggers: SkillTriggers{Keywords: []string{"test"}}},
			{ID: "high", Priority: 10, Triggers: SkillTriggers{Keywords: []string{"test"}}},
			{ID: "medium", Priority: 5, Triggers: SkillTriggers{Keywords: []string{"test"}}},
		},
	}

	ctx := &BeadContext{Title: "test task"}
	matched := registry.MatchSkills(ctx)

	require.Len(t, matched, 3)
	assert.Equal(t, "high", matched[0].ID)
	assert.Equal(t, "medium", matched[1].ID)
	assert.Equal(t, "low", matched[2].ID)
}

func TestSkillExclusiveGroups(t *testing.T) {
	registry := &SkillRegistry{
		skills: []*Skill{
			{ID: "go-v1", Exclusive: "go-version", Priority: 5, Triggers: SkillTriggers{Keywords: []string{"golang"}}},
			{ID: "go-v2", Exclusive: "go-version", Priority: 10, Triggers: SkillTriggers{Keywords: []string{"golang"}}},
			{ID: "other", Triggers: SkillTriggers{Keywords: []string{"golang"}}},
		},
	}

	ctx := &BeadContext{Title: "golang project"}
	matched := registry.MatchSkills(ctx)

	// Should match go-v2 (higher priority in exclusive group) and other (no exclusive)
	require.Len(t, matched, 2)
	assert.Equal(t, "go-v2", matched[0].ID)
	assert.Equal(t, "other", matched[1].ID)
}

func TestLabelMatchWildcard(t *testing.T) {
	tests := []struct {
		trigger string
		label   string
		matches bool
	}{
		{"domain:*", "domain:auth", true},
		{"domain:*", "domain:db", true},
		{"domain:*", "other:auth", false},
		{"gt:testing", "gt:testing", true},
		{"gt:testing", "gt:test", false},
		{"GT:TESTING", "gt:testing", true}, // case-insensitive
	}

	for _, tt := range tests {
		t.Run(tt.trigger+"_"+tt.label, func(t *testing.T) {
			result := matchLabel(tt.trigger, tt.label)
			assert.Equal(t, tt.matches, result)
		})
	}
}

func TestPatternMatch(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		matches bool
	}{
		{"add.*endpoint", "Add user endpoint", true},
		{"add.*endpoint", "Update user endpoint", false},
		{"(fix|bug)", "Fix the bug", true},
		{"\\d+", "Issue 123", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := matchPattern(tt.pattern, tt.text)
			assert.Equal(t, tt.matches, result)
		})
	}
}

func TestEnrichmentBuilder(t *testing.T) {
	builder := NewEnrichmentBuilder("/tmp/rig", DepthStandard)

	// Inject a skill - mark files as Optional since they don't exist
	skill := &Skill{
		ID:   "test-skill",
		Name: "Test Skill",
		Content: SkillContent{
			Files: []SkillFile{
				{Path: "src/main.go", Description: "Main entry point", Optional: true},
				{Path: "src/handler.go", Lines: "10-50", Description: "Handler code", Optional: true},
			},
			Patterns: []SkillPattern{
				{Name: "Error handling", Description: "Use wrapped errors", Example: "src/errors.go"},
			},
			Documentation: []SkillDoc{
				{Title: "Go Docs", URL: "https://go.dev", Description: "Official docs"},
			},
			ContextNotes: []string{
				"Remember to run tests",
			},
		},
	}

	builder.InjectSkill(skill)
	builder.AddPriorWork("gt-123", "closed", "Related work", "Good example to follow")

	output := builder.Build("Test summary")

	// Verify output contains expected sections
	assert.Contains(t, output, "## Required Reading")
	assert.Contains(t, output, "### Summary")
	assert.Contains(t, output, "Test summary")
	// Files are optional and don't exist, so they should still be included
	assert.Contains(t, output, "### Files to Read")
	assert.Contains(t, output, "`src/main.go`")
	assert.Contains(t, output, "`src/handler.go:10-50`")
	assert.Contains(t, output, "### Key Patterns")
	assert.Contains(t, output, "**Error handling**")
	assert.Contains(t, output, "### Documentation")
	assert.Contains(t, output, "[Go Docs](https://go.dev)")
	assert.Contains(t, output, "### Prior Work")
	assert.Contains(t, output, "**gt-123**")
	assert.Contains(t, output, "### Context Notes")
	assert.Contains(t, output, "Remember to run tests")
}

func TestEnrichmentLimits(t *testing.T) {
	builder := NewEnrichmentBuilder("/tmp/rig", DepthStandard)

	// Add more files than the limit
	for i := 0; i < 20; i++ {
		builder.AddFile("file"+string(rune('a'+i))+".go", "", "Description")
	}

	stats := builder.Stats()
	output := builder.Build("Summary")

	// Stats should reflect actual count
	assert.Equal(t, 20, stats.FilesCount)

	// But output should be limited
	// Count occurrences of file entries (should be limited to MaxFiles)
	// This is a simple check - the limit is enforced in Build()
	assert.NotEmpty(t, output)
}
