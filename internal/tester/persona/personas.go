package persona

// Built-in personas for AI User Testing.
// These personas represent common user archetypes for testing ScreenCoach.

// Sarah represents a low-tech parent who found ScreenCoach through school.
var Sarah = Persona{
	Name: "Sarah",
	Role: "parent",
	Context: `First-time user, not tech-savvy, has 2 kids (ages 8 and 12).
Found ScreenCoach through school recommendation.
Works part-time, usually accesses the app in evenings.
Wants simple solutions, gets overwhelmed by too many options.`,
	TechComfort: TechComfortLow,
	Patience:    PatienceMedium,
	Device:      DeviceDesktop,
	Tags:        []string{"low-tech", "new-user", "parent"},
}

// Mike represents a tech-savvy parent who expects efficiency.
var Mike = Persona{
	Name: "Mike",
	Role: "parent",
	Context: `Software developer, has 1 kid (age 10).
Researched multiple screen time apps before choosing ScreenCoach.
Expects modern, responsive UI and efficient workflows.
Will abandon app if it feels outdated or clunky.`,
	TechComfort: TechComfortHigh,
	Patience:    PatienceLow,
	Device:      DeviceDesktop,
	Tags:        []string{"tech-savvy", "power-user", "parent"},
}

// Emma represents a mobile-first parent with medium tech skills.
var Emma = Persona{
	Name: "Emma",
	Role: "parent",
	Context: `Busy professional, has 3 kids (ages 6, 9, 14).
Primarily uses phone for everything, rarely on desktop.
Wants quick actions and notifications, not detailed dashboards.
Often multitasking while using the app.`,
	TechComfort: TechComfortMedium,
	Patience:    PatienceLow,
	Device:      DeviceMobile,
	Tags:        []string{"mobile-first", "busy", "parent"},
}

// Rose represents a very low-tech grandparent user.
var Rose = Persona{
	Name: "Rose",
	Role: "parent",
	Context: `Grandmother helping with grandchildren, age 68.
Very uncomfortable with technology, needs clear guidance.
Will read every instruction carefully before proceeding.
Tends to call for help rather than experiment.`,
	TechComfort: TechComfortLow,
	Patience:    PatienceHigh,
	Device:      DeviceDesktop,
	Tags:        []string{"very-low-tech", "senior", "cautious"},
}

// BuiltInPersonas contains all built-in personas indexed by name.
var BuiltInPersonas = map[string]*Persona{
	"sarah": &Sarah,
	"mike":  &Mike,
	"emma":  &Emma,
	"rose":  &Rose,
}

// Get returns a built-in persona by name (case-insensitive).
// Returns nil if not found.
func Get(name string) *Persona {
	return BuiltInPersonas[normalize(name)]
}

// List returns all built-in persona names.
func List() []string {
	return []string{"sarah", "mike", "emma", "rose"}
}

// normalize converts a name to lowercase for case-insensitive lookup.
func normalize(name string) string {
	// Simple lowercase conversion
	result := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
