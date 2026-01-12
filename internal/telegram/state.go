package telegram

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// State represents the persistent state of the Telegram bridge.
type State struct {
	LastUpdateID int    `json:"last_update_id"`
	MsgMap       map[int]string `json:"msg_map"` // Telegram MessageID -> Beads ID
	
	mu sync.RWMutex
	path string
}

// NewState creates a new empty state.
func NewState(path string) *State {
	return &State{
		MsgMap: make(map[int]string),
		path:   path,
	}
}

// LoadState loads the state from disk.
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewState(path), nil
		}
		return nil, err
	}

	state := NewState(path)
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	if state.MsgMap == nil {
		state.MsgMap = make(map[int]string)
	}
	return state, nil
}

// Save persists the state to disk.
func (s *State) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0600)
}

// AddMapping records a mapping from Telegram message ID to beads ID.
func (s *State) AddMapping(tgID int, beadsID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MsgMap[tgID] = beadsID
}

// GetBeadsID returns the beads ID for a given Telegram message ID.
func (s *State) GetBeadsID(tgID int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.MsgMap[tgID]
}

// SetLastUpdateID updates the last processed Telegram update ID.
func (s *State) SetLastUpdateID(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastUpdateID = id
}
