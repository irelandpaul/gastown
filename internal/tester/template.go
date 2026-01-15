package tester

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Note: Template is loaded from templates/tester-CLAUDE.md at runtime
// or falls back to inline template if not available.

// TesterTemplateData contains the data for rendering the tester CLAUDE.md.
type TesterTemplateData struct {
	// PersonaBlock is the full persona description.
	PersonaBlock string

	// PersonaName is the persona's name (e.g., "Sarah").
	PersonaName string

	// Goal is what the user is trying to accomplish.
	Goal string

	// ScenarioName is the scenario identifier.
	ScenarioName string

	// AppName is the application name (e.g., "ScreenCoach").
	AppName string

	// AppContext describes what the application does.
	AppContext string

	// SuccessCriteria lists the success criteria as a formatted string.
	SuccessCriteria string
}

// RenderTesterTemplate renders the tester CLAUDE.md template with the given data.
// It attempts to load from the templates directory, falling back to an inline template.
func RenderTesterTemplate(data *TesterTemplateData) (string, error) {
	return RenderTesterTemplateFromDir("", data)
}

// RenderTesterTemplateFromDir renders the template from a specific directory.
// If templateDir is empty, it uses the inline fallback template.
func RenderTesterTemplateFromDir(templateDir string, data *TesterTemplateData) (string, error) {
	if templateDir != "" {
		templatePath := filepath.Join(templateDir, "templates", "tester-CLAUDE.md")
		content, err := os.ReadFile(templatePath)
		if err == nil {
			return renderTemplateString(string(content), data)
		}
	}

	// Fallback to inline template
	return renderInlineTemplate(data)
}

// renderTemplateString renders a template string with the given data.
func renderTemplateString(tmplContent string, data *TesterTemplateData) (string, error) {
	// Use custom delimiters to avoid conflicts with existing {{ }}
	tmpl, err := template.New("tester").Delims("{{", "}}").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderInlineTemplate provides a fallback template if embed fails.
func renderInlineTemplate(data *TesterTemplateData) (string, error) {
	tmpl := `# Tester Agent

You are a **user** testing an application. NOT a QA engineer. NOT an AI assistant.
You ARE the person described in your persona. Think as they would think.

## Your Persona

{{.PersonaBlock}}

## Your Goal

{{.Goal}}

## How to Test

### 1. BE the User

Think like {{.PersonaName}}:
- What would confuse them?
- What would they expect to see?
- Where would they click first?
- What would frustrate them?

### 2. Think Aloud

As you navigate, speak your thoughts.

### 3. Document Observations

When confused or frustrated, document with severity (P0-P3) and confidence (high/medium/low).

### 4. Complete the Goal

Work toward your goal as {{.PersonaName}} would.

## Browser Control (Playwright MCP)

Use browser_* tools for navigation, clicking, typing, and screenshots.

## What {{.AppName}} Is

{{.AppContext}}

## Success Criteria

{{.SuccessCriteria}}

---

Now begin testing as {{.PersonaName}}.
`

	return renderTemplateString(tmpl, data)
}

// FormatPersonaBlock formats persona information for the template.
func FormatPersonaBlock(name, role, techComfort, patience, context string) string {
	var sb strings.Builder

	sb.WriteString("Name: " + name + "\n")
	sb.WriteString("Role: " + role + "\n")
	sb.WriteString("Tech Comfort: " + techComfort + "\n")
	if patience != "" {
		sb.WriteString("Patience: " + patience + "\n")
	}
	if context != "" {
		sb.WriteString("\nContext:\n" + context + "\n")
	}

	return sb.String()
}

// FormatSuccessCriteria formats success criteria as a bulleted list.
func FormatSuccessCriteria(criteria []string) string {
	var sb strings.Builder

	for _, c := range criteria {
		sb.WriteString("- " + c + "\n")
	}

	return sb.String()
}

// DefaultAppContext returns the default ScreenCoach app context.
func DefaultAppContext() string {
	return `ScreenCoach is a parental control app for managing children's screen time.
Key concepts:
- Parents create accounts and add children
- Each child has a profile with time limits
- Parents can set schedules and block apps
- Children use a separate app/extension

You're testing the Parent Portal - where parents manage settings.`
}

// ExamplePersonas provides example persona configurations.
var ExamplePersonas = map[string]struct {
	Name        string
	Role        string
	TechComfort string
	Patience    string
	Context     string
}{
	"sarah": {
		Name:        "Sarah",
		Role:        "Parent",
		TechComfort: "Low (needs clear guidance, confused by jargon)",
		Patience:    "Medium (will try a few times, then give up)",
		Context: `First-time user, not tech-savvy, has 2 kids (ages 8 and 12).
Found ScreenCoach through school recommendation.
Primary goal: limit screen time during homework hours.`,
	},
	"mike": {
		Name:        "Mike",
		Role:        "Parent",
		TechComfort: "High (works in IT, expects efficient interfaces)",
		Patience:    "Low (will abandon if frustrated)",
		Context: `Power user, wants granular control over settings.
Has 3 kids with different needs (gaming vs homework limits).
Looking for advanced scheduling and per-app controls.`,
	},
	"emma": {
		Name:        "Emma",
		Role:        "Parent",
		TechComfort: "Medium (uses apps daily, prefers mobile)",
		Patience:    "High (patient but dislikes complexity)",
		Context: `Mobile-first user, manages family on the go.
Single parent with 1 child (age 10).
Wants quick access to pause/resume and activity monitoring.`,
	},
	"rose": {
		Name:        "Rose",
		Role:        "Grandparent",
		TechComfort: "Very Low (struggles with technology)",
		Patience:    "High (willing to learn but easily confused)",
		Context: `Caring for grandchildren on weekends.
Uses a tablet, prefers large text and simple interfaces.
Wants to enforce basic rules without complex setup.`,
	},
}
