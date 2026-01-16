package librarian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EnrichmentDepth specifies how thorough the enrichment should be.
type EnrichmentDepth string

const (
	DepthQuick    EnrichmentDepth = "quick"
	DepthStandard EnrichmentDepth = "standard"
	DepthDeep     EnrichmentDepth = "deep"
)

// EnrichmentLimits defines maximum sizes for enrichment sections.
var EnrichmentLimits = struct {
	MaxFiles          int
	MaxPriorBeads     int
	MaxDocs           int
	MaxPatterns       int
	MaxContextNotes   int
	MaxTotalSizeBytes int
}{
	MaxFiles:          10,
	MaxPriorBeads:     5,
	MaxDocs:           5,
	MaxPatterns:       5,
	MaxContextNotes:   1024, // bytes
	MaxTotalSizeBytes: 10240,
}

// EnrichmentBuilder builds enrichment content from matched skills.
type EnrichmentBuilder struct {
	files        []fileEntry
	patterns     []patternEntry
	docs         []docEntry
	contextNotes []string
	priorWork    []priorWorkEntry
	depth        EnrichmentDepth
	startTime    time.Time
	rigRoot      string
}

type fileEntry struct {
	path        string
	lines       string
	description string
}

type patternEntry struct {
	name        string
	description string
	example     string
}

type docEntry struct {
	title       string
	url         string
	description string
}

type priorWorkEntry struct {
	id       string
	status   string
	title    string
	learning string
}

// NewEnrichmentBuilder creates a new enrichment builder.
func NewEnrichmentBuilder(rigRoot string, depth EnrichmentDepth) *EnrichmentBuilder {
	return &EnrichmentBuilder{
		files:        make([]fileEntry, 0),
		patterns:     make([]patternEntry, 0),
		docs:         make([]docEntry, 0),
		contextNotes: make([]string, 0),
		priorWork:    make([]priorWorkEntry, 0),
		depth:        depth,
		startTime:    time.Now(),
		rigRoot:      rigRoot,
	}
}

// InjectSkill adds content from a skill to the enrichment.
func (b *EnrichmentBuilder) InjectSkill(skill *Skill) {
	content := skill.Content

	// Add files
	for _, f := range content.Files {
		// Check if file exists (unless optional)
		if !f.Optional {
			fullPath := filepath.Join(b.rigRoot, f.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				// Skip non-existent files unless optional
				continue
			}
		}

		b.files = append(b.files, fileEntry{
			path:        f.Path,
			lines:       f.Lines,
			description: f.Description,
		})
	}

	// Add patterns
	for _, p := range content.Patterns {
		b.patterns = append(b.patterns, patternEntry{
			name:        p.Name,
			description: p.Description,
			example:     p.Example,
		})
	}

	// Add documentation
	for _, d := range content.Documentation {
		b.docs = append(b.docs, docEntry{
			title:       d.Title,
			url:         d.URL,
			description: d.Description,
		})
	}

	// Add context notes
	b.contextNotes = append(b.contextNotes, content.ContextNotes...)
}

// AddFile adds a file to the enrichment.
func (b *EnrichmentBuilder) AddFile(path, lines, description string) {
	b.files = append(b.files, fileEntry{
		path:        path,
		lines:       lines,
		description: description,
	})
}

// AddPattern adds a pattern to the enrichment.
func (b *EnrichmentBuilder) AddPattern(name, description, example string) {
	b.patterns = append(b.patterns, patternEntry{
		name:        name,
		description: description,
		example:     example,
	})
}

// AddDoc adds a documentation link to the enrichment.
func (b *EnrichmentBuilder) AddDoc(title, url, description string) {
	b.docs = append(b.docs, docEntry{
		title:       title,
		url:         url,
		description: description,
	})
}

// AddPriorWork adds a prior work reference to the enrichment.
func (b *EnrichmentBuilder) AddPriorWork(id, status, title, learning string) {
	b.priorWork = append(b.priorWork, priorWorkEntry{
		id:       id,
		status:   status,
		title:    title,
		learning: learning,
	})
}

// AddContextNote adds a context note to the enrichment.
func (b *EnrichmentBuilder) AddContextNote(note string) {
	b.contextNotes = append(b.contextNotes, note)
}

// Build generates the enrichment markdown content.
func (b *EnrichmentBuilder) Build(summary string) string {
	var sb strings.Builder
	elapsed := time.Since(b.startTime)

	// Header
	sb.WriteString("## Required Reading\n\n")
	sb.WriteString(fmt.Sprintf("> Enriched by Librarian on %s | Depth: %s | Time: %s\n\n",
		time.Now().Format("2006-01-02"),
		b.depth,
		formatDuration(elapsed)))

	// Summary
	if summary != "" {
		sb.WriteString("### Summary\n")
		sb.WriteString(summary + "\n\n")
	}

	// Files to Read (limited)
	if len(b.files) > 0 {
		sb.WriteString("### Files to Read\n")
		files := b.files
		if len(files) > EnrichmentLimits.MaxFiles {
			files = files[:EnrichmentLimits.MaxFiles]
		}
		for _, f := range files {
			path := f.path
			if f.lines != "" {
				path = fmt.Sprintf("%s:%s", f.path, f.lines)
			}
			sb.WriteString(fmt.Sprintf("- `%s` - %s\n", path, f.description))
		}
		sb.WriteString("\n")
	}

	// Prior Work (limited)
	if len(b.priorWork) > 0 {
		sb.WriteString("### Prior Work\n")
		work := b.priorWork
		if len(work) > EnrichmentLimits.MaxPriorBeads {
			work = work[:EnrichmentLimits.MaxPriorBeads]
		}
		for _, w := range work {
			sb.WriteString(fmt.Sprintf("- **%s** (%s): \"%s\" - %s\n",
				w.id, w.status, w.title, w.learning))
		}
		sb.WriteString("\n")
	}

	// Documentation (limited)
	if len(b.docs) > 0 {
		sb.WriteString("### Documentation\n")
		docs := b.docs
		if len(docs) > EnrichmentLimits.MaxDocs {
			docs = docs[:EnrichmentLimits.MaxDocs]
		}
		for _, d := range docs {
			sb.WriteString(fmt.Sprintf("- [%s](%s) - %s\n", d.title, d.url, d.description))
		}
		sb.WriteString("\n")
	}

	// Key Patterns (limited)
	if len(b.patterns) > 0 {
		sb.WriteString("### Key Patterns\n")
		patterns := b.patterns
		if len(patterns) > EnrichmentLimits.MaxPatterns {
			patterns = patterns[:EnrichmentLimits.MaxPatterns]
		}
		for _, p := range patterns {
			entry := fmt.Sprintf("- **%s**: %s", p.name, p.description)
			if p.example != "" {
				entry += fmt.Sprintf(" (see `%s`)", p.example)
			}
			sb.WriteString(entry + "\n")
		}
		sb.WriteString("\n")
	}

	// Context Notes (limited by size)
	if len(b.contextNotes) > 0 {
		sb.WriteString("### Context Notes\n")
		totalSize := 0
		for _, note := range b.contextNotes {
			if totalSize+len(note) > EnrichmentLimits.MaxContextNotes {
				break
			}
			sb.WriteString("- " + note + "\n")
			totalSize += len(note)
		}
	}

	result := sb.String()

	// Enforce total size limit
	if len(result) > EnrichmentLimits.MaxTotalSizeBytes {
		result = result[:EnrichmentLimits.MaxTotalSizeBytes]
		// Try to end at a newline
		if idx := strings.LastIndex(result, "\n"); idx > len(result)/2 {
			result = result[:idx+1]
		}
		result += "\n[Truncated due to size limit]\n"
	}

	return result
}

// Stats returns statistics about the enrichment content.
func (b *EnrichmentBuilder) Stats() EnrichmentStats {
	return EnrichmentStats{
		FilesCount:      len(b.files),
		PriorBeadsCount: len(b.priorWork),
		DocsCount:       len(b.docs),
		PatternsCount:   len(b.patterns),
		Depth:           string(b.depth),
	}
}

// EnrichmentStats contains statistics about enrichment content.
type EnrichmentStats struct {
	FilesCount      int    `json:"files_count"`
	PriorBeadsCount int    `json:"prior_beads_count"`
	DocsCount       int    `json:"docs_count"`
	PatternsCount   int    `json:"patterns_count"`
	Depth           string `json:"depth"`
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
