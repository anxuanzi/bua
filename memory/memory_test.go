package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	cfg := &Config{
		ShortTermLimit: 5,
	}

	m := NewManager(cfg)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.config.ShortTermLimit != 5 {
		t.Errorf("ShortTermLimit = %d, want 5", m.config.ShortTermLimit)
	}
}

func TestNewManager_DefaultLimit(t *testing.T) {
	m := NewManager(&Config{})
	if m.config.ShortTermLimit != 10 {
		t.Errorf("Default ShortTermLimit = %d, want 10", m.config.ShortTermLimit)
	}
}

func TestManager_StartTask(t *testing.T) {
	m := NewManager(&Config{})
	m.StartTask("Test task")

	ctx := m.GetTaskContext()
	if ctx == "" {
		t.Error("GetTaskContext() returned empty string after StartTask")
	}
}

func TestManager_AddObservation(t *testing.T) {
	m := NewManager(&Config{ShortTermLimit: 3})
	m.StartTask("Test task")

	// Add observations
	for i := 0; i < 5; i++ {
		m.AddObservation(&Observation{
			URL:   "https://example.com",
			Title: "Test",
		})
	}

	// Should have compacted
	obs := m.GetRecentObservations(0)
	if len(obs) > 3 {
		t.Errorf("Observations not compacted: got %d, want <= 3", len(obs))
	}
}

func TestManager_GetRecentObservations(t *testing.T) {
	m := NewManager(&Config{ShortTermLimit: 10})
	m.StartTask("Test task")

	// Add 5 observations
	for i := 0; i < 5; i++ {
		m.AddObservation(&Observation{
			URL:   "https://example.com",
			Title: "Test",
		})
	}

	// Get 3 most recent
	obs := m.GetRecentObservations(3)
	if len(obs) != 3 {
		t.Errorf("GetRecentObservations(3) returned %d, want 3", len(obs))
	}

	// Get all
	obs = m.GetRecentObservations(0)
	if len(obs) != 5 {
		t.Errorf("GetRecentObservations(0) returned %d, want 5", len(obs))
	}

	// Request more than available
	obs = m.GetRecentObservations(10)
	if len(obs) != 5 {
		t.Errorf("GetRecentObservations(10) returned %d, want 5", len(obs))
	}
}

func TestManager_LongTermMemory(t *testing.T) {
	m := NewManager(&Config{})

	entry := &LongTermEntry{
		Key:     "test-key",
		Type:    "pattern",
		Content: "Test content",
	}

	m.AddLongTermMemory(entry)

	// Retrieve it
	got, ok := m.GetLongTermMemory("test-key")
	if !ok {
		t.Fatal("GetLongTermMemory() returned false")
	}
	if got.Content != "Test content" {
		t.Errorf("Content = %s, want 'Test content'", got.Content)
	}
	if got.AccessCount != 1 {
		t.Errorf("AccessCount = %d, want 1", got.AccessCount)
	}

	// Not found case
	_, ok = m.GetLongTermMemory("nonexistent")
	if ok {
		t.Error("GetLongTermMemory('nonexistent') should return false")
	}
}

func TestManager_RecordSuccess(t *testing.T) {
	m := NewManager(&Config{})
	m.RecordSuccess("example.com", "click_login", "Successfully logged in")

	results := m.SearchLongTermMemory("login", "example.com")
	if len(results) == 0 {
		t.Error("RecordSuccess() entry not found in search")
	}
}

func TestManager_RecordFailure(t *testing.T) {
	m := NewManager(&Config{})
	m.RecordFailure("example.com", "click_submit", "Button not found")

	results := m.SearchLongTermMemory("submit", "example.com")
	if len(results) == 0 {
		t.Error("RecordFailure() entry not found in search")
	}
}

func TestManager_SaveLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "bua-memory-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager and add data
	m1 := NewManager(&Config{StorageDir: tmpDir})
	m1.AddLongTermMemory(&LongTermEntry{
		Key:     "test-key",
		Type:    "pattern",
		Content: "Test content",
	})

	// Save
	ctx := context.Background()
	if err := m1.Save(ctx); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	memFile := filepath.Join(tmpDir, "memory.json")
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Fatal("Memory file was not created")
	}

	// Create new manager and load
	m2 := NewManager(&Config{StorageDir: tmpDir})
	if err := m2.Load(ctx); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify data was loaded
	got, ok := m2.GetLongTermMemory("test-key")
	if !ok {
		t.Fatal("Loaded manager missing test-key")
	}
	if got.Content != "Test content" {
		t.Errorf("Loaded Content = %s, want 'Test content'", got.Content)
	}
}

func TestManager_Clear(t *testing.T) {
	m := NewManager(&Config{ShortTermLimit: 10})
	m.StartTask("Test task")
	m.AddObservation(&Observation{URL: "test"})
	m.AddLongTermMemory(&LongTermEntry{Key: "test"})

	m.Clear()

	obs := m.GetRecentObservations(0)
	if len(obs) != 0 {
		t.Errorf("Clear() didn't clear short-term: got %d observations", len(obs))
	}

	stats := m.Stats()
	if stats.LongTermCount != 0 {
		t.Errorf("Clear() didn't clear long-term: got %d entries", stats.LongTermCount)
	}
}

func TestManager_ClearShortTerm(t *testing.T) {
	m := NewManager(&Config{ShortTermLimit: 10})
	m.StartTask("Test task")
	m.AddObservation(&Observation{URL: "test"})
	m.AddLongTermMemory(&LongTermEntry{Key: "test"})

	m.ClearShortTerm()

	obs := m.GetRecentObservations(0)
	if len(obs) != 0 {
		t.Errorf("ClearShortTerm() didn't clear: got %d observations", len(obs))
	}

	// Long-term should still exist
	_, ok := m.GetLongTermMemory("test")
	if !ok {
		t.Error("ClearShortTerm() also cleared long-term")
	}
}

func TestManager_Stats(t *testing.T) {
	m := NewManager(&Config{ShortTermLimit: 10})
	m.StartTask("Test task")
	m.AddObservation(&Observation{URL: "test"})
	m.AddLongTermMemory(&LongTermEntry{Key: "test1"})
	m.AddLongTermMemory(&LongTermEntry{Key: "test2"})

	stats := m.Stats()

	if stats.ShortTermCount != 1 {
		t.Errorf("ShortTermCount = %d, want 1", stats.ShortTermCount)
	}
	if stats.ShortTermLimit != 10 {
		t.Errorf("ShortTermLimit = %d, want 10", stats.ShortTermLimit)
	}
	if stats.LongTermCount != 2 {
		t.Errorf("LongTermCount = %d, want 2", stats.LongTermCount)
	}
	if stats.TaskPrompt != "Test task" {
		t.Errorf("TaskPrompt = %s, want 'Test task'", stats.TaskPrompt)
	}
}

func TestObservation_Timestamp(t *testing.T) {
	m := NewManager(&Config{})
	m.StartTask("Test")

	// Add observation without timestamp
	m.AddObservation(&Observation{URL: "test"})

	obs := m.GetRecentObservations(1)
	if len(obs) != 1 {
		t.Fatal("Expected 1 observation")
	}

	// Timestamp should be set automatically
	if obs[0].Timestamp.IsZero() {
		t.Error("Timestamp was not automatically set")
	}
	if time.Since(obs[0].Timestamp) > time.Second {
		t.Error("Timestamp is too old")
	}
}
