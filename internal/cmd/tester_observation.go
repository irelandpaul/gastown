package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ObservationType represents the type of UX observation
type ObservationType string

const (
	// ObservationConfusion indicates the user was confused by UI element or flow
	ObservationConfusion ObservationType = "confusion"
	// ObservationFriction indicates unnecessary friction in the user experience
	ObservationFriction ObservationType = "friction"
	// ObservationBlocked indicates the user could not proceed (blocking issue)
	ObservationBlocked ObservationType = "blocked"
	// ObservationBug indicates a functional bug was encountered
	ObservationBug ObservationType = "bug"
)

// ValidObservationTypes returns all valid observation types
func ValidObservationTypes() []ObservationType {
	return []ObservationType{
		ObservationConfusion,
		ObservationFriction,
		ObservationBlocked,
		ObservationBug,
	}
}

// IsValidObservationType checks if a string is a valid observation type
func IsValidObservationType(t string) bool {
	for _, valid := range ValidObservationTypes() {
		if strings.EqualFold(t, string(valid)) {
			return true
		}
	}
	return false
}

// Severity represents the priority level of an observation
type Severity string

const (
	// SeverityP0 is critical - user cannot complete goal
	SeverityP0 Severity = "P0"
	// SeverityP1 is significant friction - user likely to abandon
	SeverityP1 Severity = "P1"
	// SeverityP2 is minor friction - noticeable but not blocking
	SeverityP2 Severity = "P2"
	// SeverityP3 is a nitpick - improvement opportunity
	SeverityP3 Severity = "P3"
)

// SeverityDescriptions maps severity levels to their descriptions
var SeverityDescriptions = map[Severity]string{
	SeverityP0: "Blocking - user cannot complete goal",
	SeverityP1: "Significant friction - user likely to abandon",
	SeverityP2: "Minor friction - noticeable but not blocking",
	SeverityP3: "Nitpick - improvement opportunity",
}

// ValidSeverities returns all valid severity levels
func ValidSeverities() []Severity {
	return []Severity{SeverityP0, SeverityP1, SeverityP2, SeverityP3}
}

// IsValidSeverity checks if a string is a valid severity level
func IsValidSeverity(s string) bool {
	upper := strings.ToUpper(s)
	for _, valid := range ValidSeverities() {
		if upper == string(valid) {
			return true
		}
	}
	return false
}

// NormalizeSeverity converts a severity string to the canonical form (P0, P1, etc.)
func NormalizeSeverity(s string) Severity {
	upper := strings.ToUpper(s)
	for _, valid := range ValidSeverities() {
		if upper == string(valid) {
			return valid
		}
	}
	return SeverityP3 // Default to lowest severity
}

// SeverityRequiresAction returns true if the severity level requires automatic action
func (s Severity) RequiresAction() bool {
	return s == SeverityP0 || s == SeverityP1
}

// Confidence represents the agent's confidence in an observation
type Confidence string

const (
	// ConfidenceHigh means the agent is very confident this is a real issue
	ConfidenceHigh Confidence = "high"
	// ConfidenceMedium means the agent is fairly confident but not certain
	ConfidenceMedium Confidence = "medium"
	// ConfidenceLow means the agent is uncertain and recommends human review
	ConfidenceLow Confidence = "low"
)

// ValidConfidences returns all valid confidence levels
func ValidConfidences() []Confidence {
	return []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
}

// IsValidConfidence checks if a string is a valid confidence level
func IsValidConfidence(c string) bool {
	lower := strings.ToLower(c)
	for _, valid := range ValidConfidences() {
		if lower == string(valid) {
			return true
		}
	}
	return false
}

// NormalizeConfidence converts a confidence string to the canonical form
func NormalizeConfidence(c string) Confidence {
	lower := strings.ToLower(c)
	for _, valid := range ValidConfidences() {
		if lower == string(valid) {
			return valid
		}
	}
	return ConfidenceLow // Default to low confidence
}

// RequiresReview returns true if the confidence level requires human review
func (c Confidence) RequiresReview() bool {
	return c == ConfidenceLow || c == ConfidenceMedium
}

// Observation represents a UX observation made during AI user testing.
// This is the full observation format matching the spec.
type Observation struct {
	// Type of observation (confusion, friction, blocked, bug)
	Type ObservationType `json:"type"`

	// Severity level (P0-P3)
	Severity Severity `json:"severity"`

	// Confidence in the observation (high, medium, low)
	Confidence Confidence `json:"confidence"`

	// Timestamp in the test run (e.g., "00:23")
	Timestamp string `json:"timestamp"`

	// Location in the app where the observation occurred
	Location string `json:"location"`

	// Description of the observation
	Description string `json:"description"`

	// Screenshot filename if one was captured
	Screenshot string `json:"screenshot,omitempty"`

	// Validated is set to true/false after human review (nil = not reviewed)
	Validated *bool `json:"validated"`

	// FalsePositive is set to true if the observation was incorrect (nil = not reviewed)
	FalsePositive *bool `json:"false_positive"`
}

// NewObservation creates a new observation with required fields
func NewObservation(obsType ObservationType, severity Severity, confidence Confidence, description string) *Observation {
	return &Observation{
		Type:        obsType,
		Severity:    severity,
		Confidence:  confidence,
		Description: description,
	}
}

// WithTimestamp sets the timestamp for when the observation occurred
func (o *Observation) WithTimestamp(timestamp string) *Observation {
	o.Timestamp = timestamp
	return o
}

// WithLocation sets the location in the app
func (o *Observation) WithLocation(location string) *Observation {
	o.Location = location
	return o
}

// WithScreenshot sets the screenshot filename
func (o *Observation) WithScreenshot(screenshot string) *Observation {
	o.Screenshot = screenshot
	return o
}

// Validate checks that the observation has valid field values
func (o *Observation) Validate() error {
	if !IsValidObservationType(string(o.Type)) {
		return fmt.Errorf("invalid observation type: %s (valid: %v)", o.Type, ValidObservationTypes())
	}
	if !IsValidSeverity(string(o.Severity)) {
		return fmt.Errorf("invalid severity: %s (valid: %v)", o.Severity, ValidSeverities())
	}
	if !IsValidConfidence(string(o.Confidence)) {
		return fmt.Errorf("invalid confidence: %s (valid: %v)", o.Confidence, ValidConfidences())
	}
	if o.Description == "" {
		return fmt.Errorf("observation description is required")
	}
	return nil
}

// RequiresBeadCreation returns true if this observation should trigger automatic bead creation
func (o *Observation) RequiresBeadCreation() bool {
	return o.Severity.RequiresAction()
}

// RequiresHumanReview returns true if this observation should be flagged for human review
func (o *Observation) RequiresHumanReview() bool {
	return o.Confidence.RequiresReview() || o.Severity.RequiresAction()
}

// InfraError represents an infrastructure error during testing
type InfraError struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Attempt   int       `json:"attempt"`
}

// ObservationResult contains the full result of a test run with observations.
// This matches the spec format for observations.json output.
type ObservationResult struct {
	// Scenario name
	Scenario string `json:"scenario"`

	// Persona name
	Persona string `json:"persona"`

	// Whether the test completed (even if with issues)
	Completed bool `json:"completed"`

	// Duration in seconds
	DurationSeconds int `json:"duration_seconds"`

	// List of observations made during the test
	Observations []Observation `json:"observations"`

	// Success criteria that were met
	SuccessCriteriaMet []string `json:"success_criteria_met"`

	// Success criteria that were not met
	SuccessCriteriaFailed []string `json:"success_criteria_failed"`

	// Overall experience summary
	OverallExperience string `json:"overall_experience"`

	// Number of retry attempts
	RetryCount int `json:"retry_count"`

	// Infrastructure errors encountered
	InfrastructureErrors []InfraError `json:"infrastructure_errors"`

	// Run metadata
	RunID     string    `json:"run_id,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Model     string    `json:"model,omitempty"`
}

// NewObservationResult creates a new observation result
func NewObservationResult(scenario, persona string) *ObservationResult {
	return &ObservationResult{
		Scenario:              scenario,
		Persona:               persona,
		Observations:          make([]Observation, 0),
		SuccessCriteriaMet:    make([]string, 0),
		SuccessCriteriaFailed: make([]string, 0),
		InfrastructureErrors:  make([]InfraError, 0),
		StartTime:             time.Now(),
	}
}

// AddObservation adds an observation to the result
func (r *ObservationResult) AddObservation(obs Observation) {
	r.Observations = append(r.Observations, obs)
}

// AddInfraError records an infrastructure error
func (r *ObservationResult) AddInfraError(errType, message string, attempt int) {
	r.InfrastructureErrors = append(r.InfrastructureErrors, InfraError{
		Type:      errType,
		Message:   message,
		Timestamp: time.Now(),
		Attempt:   attempt,
	})
}

// Complete marks the test as complete and calculates duration
func (r *ObservationResult) Complete() {
	r.Completed = true
	r.EndTime = time.Now()
	r.DurationSeconds = int(r.EndTime.Sub(r.StartTime).Seconds())
}

// CountBySeverity returns the count of observations by severity level
func (r *ObservationResult) CountBySeverity() map[Severity]int {
	counts := make(map[Severity]int)
	for _, obs := range r.Observations {
		counts[obs.Severity]++
	}
	return counts
}

// HasBlockingIssues returns true if there are P0 or P1 observations
func (r *ObservationResult) HasBlockingIssues() bool {
	for _, obs := range r.Observations {
		if obs.Severity.RequiresAction() {
			return true
		}
	}
	return false
}

// PendingReview returns observations that need human review
func (r *ObservationResult) PendingReview() []Observation {
	var pending []Observation
	for _, obs := range r.Observations {
		if obs.Validated == nil && obs.RequiresHumanReview() {
			pending = append(pending, obs)
		}
	}
	return pending
}

// WriteToFile writes the observation result to a JSON file
func (r *ObservationResult) WriteToFile(dir string) error {
	path := filepath.Join(dir, "observations.json")

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling observations: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing observations file: %w", err)
	}

	return nil
}

// LoadObservationResult loads an observation result from a JSON file
func LoadObservationResult(path string) (*ObservationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var result ObservationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &result, nil
}

// ParseObservationFromAgent parses an observation from agent output.
// Expected format: "[OBSERVATION] P2/high confusion at homepage: Signup button hard to find"
func ParseObservationFromAgent(line string) (*Observation, error) {
	// Remove [OBSERVATION] prefix if present
	line = strings.TrimPrefix(line, "[OBSERVATION]")
	line = strings.TrimSpace(line)

	// Expected format: "P2/high confusion at homepage: description"
	// or: "P2/high/confusion/homepage: description"

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format: expected 'metadata: description'")
	}

	metadata := strings.TrimSpace(parts[0])
	description := strings.TrimSpace(parts[1])

	if description == "" {
		return nil, fmt.Errorf("description is required")
	}

	// Parse metadata: "P2/high confusion at homepage" or "P2/high/confusion/homepage"
	obs := &Observation{
		Description: description,
		Confidence:  ConfidenceMedium, // default
		Severity:    SeverityP2,       // default
		Type:        ObservationFriction, // default
	}

	// Look for severity (P0-P3)
	for _, sev := range ValidSeverities() {
		if strings.Contains(strings.ToUpper(metadata), string(sev)) {
			obs.Severity = sev
			break
		}
	}

	// Look for confidence
	lowerMeta := strings.ToLower(metadata)
	for _, conf := range ValidConfidences() {
		if strings.Contains(lowerMeta, string(conf)) {
			obs.Confidence = conf
			break
		}
	}

	// Look for type
	for _, t := range ValidObservationTypes() {
		if strings.Contains(lowerMeta, string(t)) {
			obs.Type = t
			break
		}
	}

	// Look for location (after "at " or final segment)
	if idx := strings.Index(lowerMeta, " at "); idx >= 0 {
		obs.Location = strings.TrimSpace(metadata[idx+4:])
	}

	return obs, nil
}

// FormatObservationForOutput formats an observation for terminal output
func FormatObservationForOutput(obs Observation, withTimestamp bool) string {
	var parts []string

	if withTimestamp && obs.Timestamp != "" {
		parts = append(parts, fmt.Sprintf("[%s]", obs.Timestamp))
	}

	parts = append(parts, fmt.Sprintf("%s/%s %s", obs.Severity, obs.Confidence, obs.Type))

	if obs.Location != "" {
		parts = append(parts, fmt.Sprintf("at %s", obs.Location))
	}

	parts = append(parts, "-", obs.Description)

	return strings.Join(parts, " ")
}
