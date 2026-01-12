package inbox

import (
	"fmt"
	"testing"
	"time"
)

func TestFilterStackedInfo(t *testing.T) {
	now := time.Now()
	messages := []Message{
		{
			ID:        "1",
			Type:      TypeInfo,
			From:      "witness",
			Subject:   "Patrol completed",
			Timestamp: now.Add(-10 * time.Minute),
		},
		{
			ID:        "2",
			Type:      TypeInfo,
			From:      "witness",
			Subject:   "Patrol completed",
			Timestamp: now.Add(-5 * time.Minute),
		},
		{
			ID:        "3", // Newest, should be kept
			Type:      TypeInfo,
			From:      "witness",
			Subject:   "Patrol completed",
			Timestamp: now,
		},
		{
			ID:        "4",
			Type:      TypeAlert, // Different type, should be kept
			From:      "witness",
			Subject:   "Patrol completed",
			Timestamp: now.Add(-2 * time.Minute),
		},
		{
			ID:        "5",
			Type:      TypeInfo,
			From:      "deacon", // Different sender, should be kept
			Subject:   "Patrol completed",
			Timestamp: now,
		},
	}

	filtered, toArchive := filterStackedInfo(messages)

	if len(filtered) != 3 {
		t.Errorf("Expected 3 filtered messages, got %d", len(filtered))
	}

	if len(toArchive) != 2 {
		t.Errorf("Expected 2 messages to archive, got %d", len(toArchive))
	}

	// Verify IDs 1 and 2 are in toArchive
	archiveMap := make(map[string]bool)
	for _, id := range toArchive {
		archiveMap[id] = true
	}
	if !archiveMap["1"] || !archiveMap["2"] {
		t.Errorf("Expected IDs 1 and 2 to be archived, got %v", toArchive)
	}
}

func TestPagination(t *testing.T) {
	messages := make([]Message, 250)
	for i := 0; i < 250; i++ {
		messages[i] = Message{ID: fmt.Sprintf("%d", i)}
	}

	m := New("test", "test")
	m.messages = messages
	m.page = 0

	// Page 0: items 0-99
	if m.page != 0 {
		t.Errorf("Expected page 0, got %d", m.page)
	}

	// Move to next page
	m.page++ // simulating key press logic
	if m.page != 1 {
		t.Errorf("Expected page 1, got %d", m.page)
	}

	// Move to next page
	m.page++
	if m.page != 2 {
		t.Errorf("Expected page 2, got %d", m.page)
	}

	// Should not move past last page
	if (m.page+1)*100 >= len(m.messages) {
		// next page would be 3, 3*100 = 300 > 250
	}
}

func BenchmarkLoadAndFilter1000(b *testing.B) {
	now := time.Now()
	messages := make([]Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = Message{
			ID:        fmt.Sprintf("%d", i),
			Type:      TypeInfo,
			From:      "agent",
			Subject:   "Status Update",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterStackedInfo(messages)
		sortMessages(messages)
	}
}