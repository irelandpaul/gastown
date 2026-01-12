package inbox

import (
	"testing"
	"time"
)

func TestAgeStyle(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		timestamp time.Time
		want      string // We'll just check if it returns a non-empty style or check the color if possible
	}{
		{
			name:      "fresh (<1h)",
			timestamp: now.Add(-30 * time.Minute),
		},
		{
			name:      "recent (1h-24h)",
			timestamp: now.Add(-5 * time.Hour),
		},
		{
			name:      "old (1d-3d)",
			timestamp: now.Add(-48 * time.Hour),
		},
		{
			name:      "stale (>3d)",
			timestamp: now.Add(-5 * 24 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := AgeStyle(tt.timestamp)
			if style.GetForeground() == nil {
				t.Errorf("AgeStyle() returned style with no foreground color")
			}
		})
	}
}
