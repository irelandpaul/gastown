package inbox

import (
	"reflect"
	"testing"
)

func TestExtractReferences(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "single reference",
			body: "Please check gt-123 for details.",
			want: []string{"gt-123"},
		},
		{
			name: "multiple references",
			body: "See gt-123 and hq-cv-abc. Also bd-xyz.",
			want: []string{"gt-123", "hq-cv-abc", "bd-xyz"},
		},
		{
			name: "reference with punctuation",
			body: "Done with sc-456; now starting gt-789!",
			want: []string{"sc-456", "gt-789"},
		},
		{
			name: "reference in parentheses",
			body: "The issue (gt-abc123) is blocked.",
			want: []string{"gt-abc123"},
		},
		{
			name: "duplicates",
			body: "gt-123 is related to gt-123.",
			want: []string{"gt-123"},
		},
		{
			name: "no references",
			body: "No beads here, just text.",
			want: nil,
		},
		{
			name: "invalid references",
			body: "a-1 is too short. g-123 is too short. this-is-too-long-prefix-123",
			want: nil,
		},
		{
			name: "complex references",
			body: "Check gt-gastown-witness and bd-beads-polecat-rictus.",
			want: []string{"gt-gastown-witness", "bd-beads-polecat-rictus"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractReferences(tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}
