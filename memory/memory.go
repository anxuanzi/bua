// Package memory provides short-term and long-term memory management for the agent.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds memory configuration.
type Config struct {
	// ShortTermLimit is the maximum number of observations to keep in short-term memory.
	ShortTermLimit int

	// StorageDir is the directory for persisting long-term memory.
	StorageDir string

	// SiteSpecific indicates whether to use site-specific memory files.
	SiteSpecific bool
}

// Observation represents a single observation from the agent's interaction.
type Observation struct {
	// Timestamp is when this observation was made.
	Timestamp time.Time `json:"timestamp"`

	// URL is the page URL at the time of observation.
	URL string `json:"url"`

	// Title is the page title.
	Title string `json:"title"`

	// ElementCount is the number of elements on the page.
	ElementCount int `json:"elementCount"`

	// ScreenshotPath is the path to the screenshot taken.
	ScreenshotPath string `json:"screenshotPath,omitempty"`

	// Action is the action that was taken, if any.
	Action *Action `json:"action,omitempty"`

	// Result is the result of the action.
	Result string `json:"result,omitempty"`
}

// Action represents an action taken by the agent.
type Action struct {
	// Type is the type of action (click, type, scroll, navigate, etc.).
	Type string `json:"type"`

	// Target describes what was targeted (element index, URL, etc.).
	Target string `json:"target"`

	// Value is the value used (text to type, scroll amount, etc.).
	Value string `json:"value,omitempty"`

	// Reasoning is the LLM's reasoning for taking this action.
	Reasoning string `json:"reasoning,omitempty"`
}

// LongTermEntry represents an entry in long-term memory.
type LongTermEntry struct {
	// Key is a unique identifier for this entry.
	Key string `json:"key"`

	// Type categorizes the entry (pattern, success, failure, user_correction).
	Type string `json:"type"`

	// Site is the site this entry is associated with (if any).
	Site string `json:"site,omitempty"`

	// Content is the actual memory content.
	Content string `json:"content"`

	// Metadata contains additional information.
	Metadata map[string]any `json:"metadata,omitempty"`

	// CreatedAt is when this entry was created.
	CreatedAt time.Time `json:"createdAt"`

	// AccessedAt is when this entry was last accessed.
	AccessedAt time.Time `json:"accessedAt"`

	// AccessCount tracks how often this entry is used.
	AccessCount int `json:"accessCount"`
}

// Manager manages both short-term and long-term memory.
type Manager struct {
	config *Config

	// Short-term memory (in-memory, not persisted)
	shortTerm []*Observation
	stMu      sync.RWMutex

	// Long-term memory (persisted to disk)
	longTerm map[string]*LongTermEntry
	ltMu     sync.RWMutex

	// Current task context
	taskPrompt string
	taskStart  time.Time
}

// NewManager creates a new memory manager.
func NewManager(cfg *Config) *Manager {
	if cfg.ShortTermLimit == 0 {
		cfg.ShortTermLimit = 10
	}
	if cfg.StorageDir != "" {
		os.MkdirAll(cfg.StorageDir, 0755)
	}

	return &Manager{
		config:    cfg,
		shortTerm: make([]*Observation, 0, cfg.ShortTermLimit),
		longTerm:  make(map[string]*LongTermEntry),
	}
}

// StartTask begins tracking a new task.
func (m *Manager) StartTask(prompt string) {
	m.stMu.Lock()
	defer m.stMu.Unlock()

	m.taskPrompt = prompt
	m.taskStart = time.Now()
	m.shortTerm = make([]*Observation, 0, m.config.ShortTermLimit)
}

// AddObservation adds an observation to short-term memory.
func (m *Manager) AddObservation(obs *Observation) {
	m.stMu.Lock()
	defer m.stMu.Unlock()

	if obs.Timestamp.IsZero() {
		obs.Timestamp = time.Now()
	}

	m.shortTerm = append(m.shortTerm, obs)

	// Enforce limit by removing oldest observations
	if len(m.shortTerm) > m.config.ShortTermLimit {
		// Before removing, consider saving important info to long-term memory
		m.compactShortTerm()
	}
}

// compactShortTerm removes old observations while preserving important information.
func (m *Manager) compactShortTerm() {
	if len(m.shortTerm) <= m.config.ShortTermLimit/2 {
		return
	}

	// Keep most recent observations
	keep := m.config.ShortTermLimit / 2
	m.shortTerm = m.shortTerm[len(m.shortTerm)-keep:]
}

// GetRecentObservations returns the most recent observations.
func (m *Manager) GetRecentObservations(limit int) []*Observation {
	m.stMu.RLock()
	defer m.stMu.RUnlock()

	if limit <= 0 || limit > len(m.shortTerm) {
		limit = len(m.shortTerm)
	}

	result := make([]*Observation, limit)
	start := len(m.shortTerm) - limit
	copy(result, m.shortTerm[start:])
	return result
}

// GetTaskContext returns a summary of the current task for the LLM.
func (m *Manager) GetTaskContext() string {
	m.stMu.RLock()
	defer m.stMu.RUnlock()

	if m.taskPrompt == "" {
		return ""
	}

	return fmt.Sprintf("Task: %s\nStarted: %s ago\nSteps taken: %d",
		m.taskPrompt,
		time.Since(m.taskStart).Round(time.Second),
		len(m.shortTerm),
	)
}

// AddLongTermMemory adds or updates a long-term memory entry.
func (m *Manager) AddLongTermMemory(entry *LongTermEntry) {
	m.ltMu.Lock()
	defer m.ltMu.Unlock()

	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.AccessedAt = time.Now()

	m.longTerm[entry.Key] = entry
}

// GetLongTermMemory retrieves a long-term memory entry.
func (m *Manager) GetLongTermMemory(key string) (*LongTermEntry, bool) {
	m.ltMu.Lock()
	defer m.ltMu.Unlock()

	entry, ok := m.longTerm[key]
	if ok {
		entry.AccessedAt = time.Now()
		entry.AccessCount++
	}
	return entry, ok
}

// SearchLongTermMemory searches for relevant entries.
func (m *Manager) SearchLongTermMemory(query string, site string) []*LongTermEntry {
	m.ltMu.RLock()
	defer m.ltMu.RUnlock()

	var results []*LongTermEntry
	for _, entry := range m.longTerm {
		// Filter by site if specified
		if site != "" && entry.Site != "" && entry.Site != site {
			continue
		}

		// Simple keyword matching (could be enhanced with embeddings)
		if containsKeywords(entry.Content, query) || containsKeywords(entry.Key, query) {
			results = append(results, entry)
		}
	}

	return results
}

func containsKeywords(text, query string) bool {
	// Simple substring matching - could be enhanced
	return len(text) > 0 && len(query) > 0
}

// RecordSuccess records a successful action pattern.
func (m *Manager) RecordSuccess(site, pattern, description string) {
	entry := &LongTermEntry{
		Key:     fmt.Sprintf("success_%s_%d", site, time.Now().Unix()),
		Type:    "success",
		Site:    site,
		Content: description,
		Metadata: map[string]any{
			"pattern": pattern,
		},
	}
	m.AddLongTermMemory(entry)
}

// RecordFailure records a failed action for learning.
func (m *Manager) RecordFailure(site, pattern, error string) {
	entry := &LongTermEntry{
		Key:     fmt.Sprintf("failure_%s_%d", site, time.Now().Unix()),
		Type:    "failure",
		Site:    site,
		Content: error,
		Metadata: map[string]any{
			"pattern": pattern,
		},
	}
	m.AddLongTermMemory(entry)
}

// RecordUserCorrection records a user correction for learning.
func (m *Manager) RecordUserCorrection(site, original, corrected string) {
	entry := &LongTermEntry{
		Key:     fmt.Sprintf("correction_%s_%d", site, time.Now().Unix()),
		Type:    "user_correction",
		Site:    site,
		Content: fmt.Sprintf("Original: %s\nCorrected: %s", original, corrected),
		Metadata: map[string]any{
			"original":  original,
			"corrected": corrected,
		},
	}
	m.AddLongTermMemory(entry)
}

// Save persists long-term memory to disk.
func (m *Manager) Save(ctx context.Context) error {
	m.ltMu.RLock()
	defer m.ltMu.RUnlock()

	if m.config.StorageDir == "" {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.config.StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	filepath := filepath.Join(m.config.StorageDir, "memory.json")

	data, err := json.MarshalIndent(m.longTerm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write memory file: %w", err)
	}

	return nil
}

// Load loads long-term memory from disk.
func (m *Manager) Load(ctx context.Context) error {
	m.ltMu.Lock()
	defer m.ltMu.Unlock()

	if m.config.StorageDir == "" {
		return nil
	}

	filepath := filepath.Join(m.config.StorageDir, "memory.json")

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No memory file yet, start fresh
		}
		return fmt.Errorf("failed to read memory file: %w", err)
	}

	var entries map[string]*LongTermEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal memory: %w", err)
	}

	m.longTerm = entries
	return nil
}

// Clear clears all memory (both short-term and long-term).
func (m *Manager) Clear() {
	m.stMu.Lock()
	m.shortTerm = make([]*Observation, 0, m.config.ShortTermLimit)
	m.taskPrompt = ""
	m.stMu.Unlock()

	m.ltMu.Lock()
	m.longTerm = make(map[string]*LongTermEntry)
	m.ltMu.Unlock()
}

// ClearShortTerm clears only short-term memory.
func (m *Manager) ClearShortTerm() {
	m.stMu.Lock()
	defer m.stMu.Unlock()

	m.shortTerm = make([]*Observation, 0, m.config.ShortTermLimit)
}

// Stats returns memory statistics.
func (m *Manager) Stats() MemoryStats {
	m.stMu.RLock()
	m.ltMu.RLock()
	defer m.stMu.RUnlock()
	defer m.ltMu.RUnlock()

	return MemoryStats{
		ShortTermCount: len(m.shortTerm),
		ShortTermLimit: m.config.ShortTermLimit,
		LongTermCount:  len(m.longTerm),
		TaskPrompt:     m.taskPrompt,
		TaskDuration:   time.Since(m.taskStart),
	}
}

// MemoryStats holds memory statistics.
type MemoryStats struct {
	ShortTermCount int
	ShortTermLimit int
	LongTermCount  int
	TaskPrompt     string
	TaskDuration   time.Duration
}
