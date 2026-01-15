// Package persona provides user persona definitions for AI User Testing.
// Personas define the characteristics of simulated users including their
// technical comfort level, patience, and device preferences.
package persona

// TechComfort represents a user's comfort level with technology.
type TechComfort string

const (
	// TechComfortLow represents users who struggle with technology.
	TechComfortLow TechComfort = "low"

	// TechComfortMedium represents users with average technical ability.
	TechComfortMedium TechComfort = "medium"

	// TechComfortHigh represents tech-savvy users.
	TechComfortHigh TechComfort = "high"
)

// Patience represents how patient a user is when encountering issues.
type Patience string

const (
	// PatienceLow represents users who give up quickly.
	PatienceLow Patience = "low"

	// PatienceMedium represents users with average patience.
	PatienceMedium Patience = "medium"

	// PatienceHigh represents very patient users.
	PatienceHigh Patience = "high"
)

// Device represents the type of device a user prefers.
type Device string

const (
	// DeviceDesktop represents desktop/laptop users.
	DeviceDesktop Device = "desktop"

	// DeviceMobile represents mobile phone users.
	DeviceMobile Device = "mobile"

	// DeviceTablet represents tablet users.
	DeviceTablet Device = "tablet"
)

// Persona defines the characteristics of a simulated user.
type Persona struct {
	// Name is the persona's name (e.g., "Sarah", "Mike").
	Name string `json:"name" yaml:"name"`

	// Role is the persona's role (e.g., "parent", "admin").
	Role string `json:"role" yaml:"role"`

	// Context provides background information about the persona.
	Context string `json:"context" yaml:"context"`

	// TechComfort indicates the persona's comfort with technology.
	TechComfort TechComfort `json:"tech_comfort" yaml:"tech_comfort"`

	// Patience indicates how patient the persona is.
	Patience Patience `json:"patience" yaml:"patience"`

	// Device is the persona's preferred device type.
	Device Device `json:"device" yaml:"device"`

	// Tags are optional labels for filtering personas.
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// IsValid checks if the persona has all required fields.
func (p *Persona) IsValid() bool {
	return p.Name != "" &&
		p.Role != "" &&
		p.Context != "" &&
		isValidTechComfort(p.TechComfort) &&
		isValidPatience(p.Patience) &&
		isValidDevice(p.Device)
}

func isValidTechComfort(tc TechComfort) bool {
	return tc == TechComfortLow || tc == TechComfortMedium || tc == TechComfortHigh
}

func isValidPatience(p Patience) bool {
	return p == PatienceLow || p == PatienceMedium || p == PatienceHigh
}

func isValidDevice(d Device) bool {
	return d == DeviceDesktop || d == DeviceMobile || d == DeviceTablet
}

// BehaviorHints returns guidance for the AI agent based on persona traits.
func (p *Persona) BehaviorHints() []string {
	var hints []string

	// Tech comfort hints
	switch p.TechComfort {
	case TechComfortLow:
		hints = append(hints,
			"Read all instructions carefully before acting",
			"Hesitate before clicking unfamiliar buttons",
			"Express confusion when technical terms appear",
			"May miss non-obvious UI elements",
		)
	case TechComfortMedium:
		hints = append(hints,
			"Follow standard UI patterns confidently",
			"Ask for help when truly stuck",
		)
	case TechComfortHigh:
		hints = append(hints,
			"Try shortcuts and advanced features",
			"Skim instructions quickly",
			"Expect responsive, modern UI",
		)
	}

	// Patience hints
	switch p.Patience {
	case PatienceLow:
		hints = append(hints,
			"Get frustrated after 2-3 failed attempts",
			"Abandon flow if too confusing",
			"Express impatience with slow loading",
		)
	case PatienceMedium:
		hints = append(hints,
			"Will retry a few times before giving up",
			"Willing to read help text if stuck",
		)
	case PatienceHigh:
		hints = append(hints,
			"Persist through multiple attempts",
			"Read all available help content",
			"Document issues without getting frustrated",
		)
	}

	// Device hints
	switch p.Device {
	case DeviceMobile:
		hints = append(hints,
			"Expect mobile-optimized interface",
			"Use touch gestures",
			"May have connectivity interruptions",
		)
	case DeviceTablet:
		hints = append(hints,
			"Expect touch-friendly interface",
			"May rotate between orientations",
		)
	}

	return hints
}
