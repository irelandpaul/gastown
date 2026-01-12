package inbox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLearningSystem(t *testing.T) {
	tmpDir := t.TempDir()
	rulesPath := filepath.Join(tmpDir, "config", "inbox_rules.json")
	
	// Create config dir
	os.MkdirAll(filepath.Dir(rulesPath), 0755)

	ls := &LearningSystem{
		Rules: make([]ClassificationRule, 0),
		path:  rulesPath,
	}

	msg := Message{
		ID:      "msg-1",
		Subject: "Test Proposal",
		Body:    "Please approve this test.",
		From:    "agent-1",
		Type:    TypeInfo, // Initially inferred as INFO
	}

	// 1. Initial classification should be the message's original type
	gotType, ok := ls.Classify(msg)
	if ok {
		t.Errorf("ls.Classify() should not have matched any rules yet")
	}
	if gotType != TypeInfo {
		t.Errorf("ls.Classify() = %v, want %v", gotType, TypeInfo)
	}

	// 2. Learn a new classification
	err := ls.Learn(msg, TypeProposal)
	if err != nil {
		t.Fatalf("ls.Learn() failed: %v", err)
	}

	// 3. Classification should now be TypeProposal
	gotType, ok = ls.Classify(msg)
	if !ok {
		t.Errorf("ls.Classify() should have matched a rule")
	}
	if gotType != TypeProposal {
		t.Errorf("ls.Classify() = %v, want %v", gotType, TypeProposal)
	}

	// 4. Verify persistence
	ls2 := &LearningSystem{
		path: rulesPath,
	}
	ls2.load()
	if len(ls2.Rules) != 1 {
		t.Errorf("ls2.Rules count = %d, want 1", len(ls2.Rules))
	}
	if ls2.Rules[0].TargetType != TypeProposal {
		t.Errorf("ls2.Rules[0].TargetType = %v, want %v", ls2.Rules[0].TargetType, TypeProposal)
	}
}
