package librarian

import (
	"fmt"
	"strings"

	"github.com/steveyegge/gastown/internal/beads"
)

// Injector handles dynamic skill injection for beads.
// It extracts context from beads, matches skills, and builds enrichment.
type Injector struct {
	registry *SkillRegistry
	beads    *beads.Beads
	rigRoot  string
}

// NewInjector creates a new skill injector.
func NewInjector(townRoot, rigRoot string) *Injector {
	return &Injector{
		registry: NewSkillRegistry(townRoot),
		beads:    beads.New(rigRoot),
		rigRoot:  rigRoot,
	}
}

// InjectionResult contains the result of skill injection.
type InjectionResult struct {
	// MatchedSkills is the list of skills that matched the bead context
	MatchedSkills []*Skill

	// Enrichment is the generated enrichment markdown content
	Enrichment string

	// Stats contains statistics about the enrichment
	Stats EnrichmentStats

	// Context is the extracted bead context used for matching
	Context *BeadContext
}

// InjectForBead performs skill injection for a bead.
// It loads skills, extracts bead context, matches skills, and builds enrichment.
func (inj *Injector) InjectForBead(beadID string, depth EnrichmentDepth) (*InjectionResult, error) {
	// Load skills
	if err := inj.registry.LoadSkills(); err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}

	// Get bead information
	issue, err := inj.beads.Show(beadID)
	if err != nil {
		return nil, fmt.Errorf("fetching bead: %w", err)
	}

	// Extract context from bead
	ctx := inj.extractContext(issue)

	// Try to get parent labels if parent exists
	if ctx.ParentID != "" {
		parent, err := inj.beads.Show(ctx.ParentID)
		if err == nil {
			ctx.ParentLabel = parent.Labels
		}
	}

	// Match skills
	matchedSkills := inj.registry.MatchSkills(ctx)

	// Build enrichment
	builder := NewEnrichmentBuilder(inj.rigRoot, depth)

	// Inject all matched skills
	for _, skill := range matchedSkills {
		builder.InjectSkill(skill)
	}

	// Add context note about injected skills
	if len(matchedSkills) > 0 {
		skillNames := make([]string, len(matchedSkills))
		for i, s := range matchedSkills {
			skillNames[i] = s.Name
		}
		builder.AddContextNote(fmt.Sprintf("Skills injected: %s", strings.Join(skillNames, ", ")))
	}

	// Generate summary
	summary := generateSummary(issue, matchedSkills)

	return &InjectionResult{
		MatchedSkills: matchedSkills,
		Enrichment:    builder.Build(summary),
		Stats:         builder.Stats(),
		Context:       ctx,
	}, nil
}

// InjectForContext performs skill injection for a pre-extracted context.
// Useful when bead info is already available.
func (inj *Injector) InjectForContext(ctx *BeadContext, depth EnrichmentDepth) (*InjectionResult, error) {
	// Load skills
	if err := inj.registry.LoadSkills(); err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}

	// Match skills
	matchedSkills := inj.registry.MatchSkills(ctx)

	// Build enrichment
	builder := NewEnrichmentBuilder(inj.rigRoot, depth)

	// Inject all matched skills
	for _, skill := range matchedSkills {
		builder.InjectSkill(skill)
	}

	// Add context note about injected skills
	if len(matchedSkills) > 0 {
		skillNames := make([]string, len(matchedSkills))
		for i, s := range matchedSkills {
			skillNames[i] = s.Name
		}
		builder.AddContextNote(fmt.Sprintf("Skills injected: %s", strings.Join(skillNames, ", ")))
	}

	// Generate summary based on context
	summary := fmt.Sprintf("Context prepared for: %s", ctx.Title)

	return &InjectionResult{
		MatchedSkills: matchedSkills,
		Enrichment:    builder.Build(summary),
		Stats:         builder.Stats(),
		Context:       ctx,
	}, nil
}

// extractContext extracts BeadContext from a beads.Issue.
func (inj *Injector) extractContext(issue *beads.Issue) *BeadContext {
	return &BeadContext{
		ID:          issue.ID,
		Title:       issue.Title,
		Description: issue.Description,
		Labels:      issue.Labels,
		Type:        issue.Type,
		ParentID:    issue.Parent,
	}
}

// generateSummary generates a summary based on the bead and matched skills.
func generateSummary(issue *beads.Issue, skills []*Skill) string {
	if len(skills) == 0 {
		return fmt.Sprintf("Context for: %s", issue.Title)
	}

	// Build summary mentioning the task and skill domains
	domains := make([]string, 0, len(skills))
	seen := make(map[string]bool)
	for _, s := range skills {
		if !seen[s.Name] {
			domains = append(domains, s.Name)
			seen[s.Name] = true
		}
	}

	if len(domains) == 1 {
		return fmt.Sprintf("Context for %s with %s patterns.", issue.Title, domains[0])
	}

	return fmt.Sprintf("Context for %s covering: %s.", issue.Title, strings.Join(domains, ", "))
}

// ListSkills returns all available skills.
func (inj *Injector) ListSkills() ([]*Skill, error) {
	if err := inj.registry.LoadSkills(); err != nil {
		return nil, fmt.Errorf("loading skills: %w", err)
	}
	return inj.registry.AllSkills(), nil
}

// PreviewMatches returns skills that would match a given bead without building enrichment.
func (inj *Injector) PreviewMatches(beadID string) ([]*Skill, *BeadContext, error) {
	// Load skills
	if err := inj.registry.LoadSkills(); err != nil {
		return nil, nil, fmt.Errorf("loading skills: %w", err)
	}

	// Get bead information
	issue, err := inj.beads.Show(beadID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching bead: %w", err)
	}

	// Extract context
	ctx := inj.extractContext(issue)

	// Try to get parent labels if parent exists
	if ctx.ParentID != "" {
		parent, err := inj.beads.Show(ctx.ParentID)
		if err == nil {
			ctx.ParentLabel = parent.Labels
		}
	}

	// Match skills
	matchedSkills := inj.registry.MatchSkills(ctx)

	return matchedSkills, ctx, nil
}

// GetSkillsDir returns the path to the skills directory.
func (inj *Injector) GetSkillsDir() string {
	return inj.registry.SkillsDir()
}
