package bua

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anxuanzi/bua/agent"
	"github.com/anxuanzi/bua/browser"
)

// Agent is the main interface for browser automation with LLM.
type Agent struct {
	config  Config
	browser *browser.Browser
	agent   *agent.BrowserAgent
	started bool
	mu      sync.RWMutex
}

// New creates a new browser automation agent.
// Call Start() before using Run().
func New(cfg Config) (*Agent, error) {
	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Agent{
		config: cfg,
	}, nil
}

// Start launches the browser and initializes the agent.
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.started {
		return ErrAlreadyStarted
	}

	// Create browser configuration
	browserCfg := browser.Config{
		Headless:          a.config.Headless,
		ProfileDir:        a.config.ProfileDir,
		ProfileName:       a.config.ProfileName,
		ViewportWidth:     a.config.Viewport.Width,
		ViewportHeight:    a.config.Viewport.Height,
		ShowHighlight:     *a.config.ShowHighlight,
		HighlightDuration: time.Duration(a.config.HighlightDurationMs) * time.Millisecond,
		Debug:             a.config.Debug,
	}

	// Create browser
	b, err := browser.New(browserCfg)
	if err != nil {
		return fmt.Errorf("failed to create browser: %w", err)
	}

	// Start browser
	if err := b.Start(ctx); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}
	a.browser = b

	// Create browser agent
	agentCfg := agent.AgentConfig{
		APIKey:        a.config.APIKey,
		Model:         a.config.Model,
		MaxSteps:      a.config.MaxSteps,
		TextOnly:      a.config.TextOnly,
		MaxWidth:      a.config.ScreenshotMaxWidth,
		Debug:         a.config.Debug,
		ScreenshotDir: a.config.ScreenshotDir,
	}

	browserAgent, err := agent.NewBrowserAgent(ctx, agentCfg, b)
	if err != nil {
		b.Close()
		return fmt.Errorf("failed to create agent: %w", err)
	}
	a.agent = browserAgent

	a.started = true
	return nil
}

// Run executes a task described in natural language.
// Returns a Result containing the outcome and execution details.
func (a *Agent) Run(ctx context.Context, task string) (*Result, error) {
	a.mu.RLock()
	started := a.started
	a.mu.RUnlock()

	if !started {
		return nil, ErrNotStarted
	}

	// Execute the task
	agentResult, err := a.agent.Run(ctx, task)
	if err != nil {
		return nil, err
	}

	// Convert agent result to public Result type
	result := &Result{
		Success:         agentResult.Success,
		Data:            agentResult.Data,
		Error:           agentResult.Error,
		Duration:        agentResult.Duration,
		TokensUsed:      agentResult.TokensUsed,
		Steps:           make([]Step, len(agentResult.Steps)),
		ScreenshotPaths: agentResult.ScreenshotPaths,
	}

	for i, s := range agentResult.Steps {
		result.Steps[i] = Step{
			Number:         s.Number,
			Action:         s.Action,
			Target:         s.Target,
			Thinking:       s.Thinking,
			Evaluation:     s.Evaluation,
			NextGoal:       s.NextGoal,
			Memory:         s.Memory,
			Duration:       time.Duration(s.DurationMs) * time.Millisecond,
			ScreenshotPath: s.ScreenshotPath,
		}
	}

	return result, nil
}

// Navigate opens a URL in the browser.
// This is a convenience method for direct navigation without a task.
func (a *Agent) Navigate(ctx context.Context, url string) error {
	a.mu.RLock()
	started := a.started
	a.mu.RUnlock()

	if !started {
		return ErrNotStarted
	}

	return a.browser.Navigate(ctx, url)
}

// Close shuts down the browser and cleans up resources.
func (a *Agent) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.started {
		return nil
	}

	var errs []error

	if a.agent != nil {
		if err := a.agent.Close(); err != nil {
			errs = append(errs, err)
		}
		a.agent = nil
	}

	if a.browser != nil {
		if err := a.browser.Close(); err != nil {
			errs = append(errs, err)
		}
		a.browser = nil
	}

	a.started = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// GetURL returns the current page URL.
func (a *Agent) GetURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.browser == nil {
		return ""
	}
	return a.browser.GetURL()
}

// GetTitle returns the current page title.
func (a *Agent) GetTitle() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.browser == nil {
		return ""
	}
	return a.browser.GetTitle()
}

// IsStarted returns whether the agent has been started.
func (a *Agent) IsStarted() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.started
}

// NewTab opens a new browser tab.
func (a *Agent) NewTab(ctx context.Context, url string) (string, error) {
	a.mu.RLock()
	started := a.started
	a.mu.RUnlock()

	if !started {
		return "", ErrNotStarted
	}

	return a.browser.NewTab(ctx, url)
}

// SwitchTab switches to a different tab by ID.
func (a *Agent) SwitchTab(tabID string) error {
	a.mu.RLock()
	started := a.started
	a.mu.RUnlock()

	if !started {
		return ErrNotStarted
	}

	return a.browser.SwitchTab(tabID)
}

// CloseTab closes a tab by ID.
func (a *Agent) CloseTab(tabID string) error {
	a.mu.RLock()
	started := a.started
	a.mu.RUnlock()

	if !started {
		return ErrNotStarted
	}

	return a.browser.CloseTab(tabID)
}

// ListTabs returns information about all open tabs.
func (a *Agent) ListTabs() []TabInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.browser == nil {
		return nil
	}

	tabs := a.browser.ListTabs()
	result := make([]TabInfo, len(tabs))
	for i, t := range tabs {
		result[i] = TabInfo{
			ID:     t.ID,
			URL:    t.URL,
			Title:  t.Title,
			Active: t.Active,
		}
	}
	return result
}

// TabInfo contains information about a browser tab.
type TabInfo struct {
	ID     string
	URL    string
	Title  string
	Active bool
}

// WithContext returns a helper for chaining operations with context.
func (a *Agent) WithContext(ctx context.Context) *ContextualAgent {
	return &ContextualAgent{agent: a, ctx: ctx}
}

// ContextualAgent wraps Agent with a context for convenience methods.
type ContextualAgent struct {
	agent *Agent
	ctx   context.Context
}

// Run executes a task using the stored context.
func (ca *ContextualAgent) Run(task string) (*Result, error) {
	return ca.agent.Run(ca.ctx, task)
}

// Navigate opens a URL using the stored context.
func (ca *ContextualAgent) Navigate(url string) error {
	return ca.agent.Navigate(ca.ctx, url)
}

// NewTab opens a new tab using the stored context.
func (ca *ContextualAgent) NewTab(url string) (string, error) {
	return ca.agent.NewTab(ca.ctx, url)
}
