package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"github.com/anxuanzi/bua/dom"
)

// Config holds browser configuration.
type Config struct {
	// Headless runs the browser without a visible window.
	Headless bool

	// ProfileDir is the directory for browser profiles.
	ProfileDir string

	// ProfileName is the name of the profile to use.
	// Empty string uses a temporary profile.
	ProfileName string

	// Viewport is the browser viewport size.
	ViewportWidth  int
	ViewportHeight int

	// ShowHighlight shows visual feedback for actions.
	ShowHighlight bool

	// HighlightDuration is how long to show highlights.
	HighlightDuration time.Duration

	// Debug enables verbose logging.
	Debug bool
}

// DefaultConfig returns a default browser configuration.
func DefaultConfig() Config {
	return Config{
		Headless:          false,
		ViewportWidth:     1280,
		ViewportHeight:    720,
		ShowHighlight:     true,
		HighlightDuration: 300 * time.Millisecond,
	}
}

// TabInfo contains information about an open tab.
type TabInfo struct {
	ID     string
	URL    string
	Title  string
	Active bool
}

// Browser wraps rod.Browser with enhanced functionality.
type Browser struct {
	config   Config
	rod      *rod.Browser
	launcher *launcher.Launcher

	// Tab management
	pages       map[string]*rod.Page
	activeTabID string

	// DOM extraction
	extractor *dom.Extractor

	// Temporary profile path for cleanup
	tempProfilePath string

	mu sync.RWMutex
}

// New creates a new browser instance.
func New(cfg Config) (*Browser, error) {
	b := &Browser{
		config: cfg,
		pages:  make(map[string]*rod.Page),
	}

	// Set default values
	if cfg.ViewportWidth == 0 {
		b.config.ViewportWidth = 1280
	}
	if cfg.ViewportHeight == 0 {
		b.config.ViewportHeight = 720
	}
	if cfg.HighlightDuration == 0 {
		b.config.HighlightDuration = 300 * time.Millisecond
	}

	return b, nil
}

// Start launches the browser.
func (b *Browser) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.rod != nil {
		return fmt.Errorf("browser already started")
	}

	// Configure launcher
	l := launcher.New()

	if b.config.Headless {
		l = l.Headless(true)
	} else {
		l = l.Headless(false)
	}

	// Configure profile
	if b.config.ProfileName != "" {
		// Use named profile
		profilePath := filepath.Join(b.config.ProfileDir, b.config.ProfileName)
		if err := os.MkdirAll(profilePath, 0755); err != nil {
			return fmt.Errorf("failed to create profile directory: %w", err)
		}
		l = l.UserDataDir(profilePath)
	} else {
		// Use temporary profile
		tempDir, err := os.MkdirTemp("", "bua-browser-*")
		if err != nil {
			return fmt.Errorf("failed to create temp profile: %w", err)
		}
		b.tempProfilePath = tempDir
		l = l.UserDataDir(tempDir)
	}

	// Additional Chrome flags
	l = l.Set("disable-background-networking").
		Set("disable-backgrounding-occluded-windows").
		Set("disable-breakpad").
		Set("disable-client-side-phishing-detection").
		Set("disable-default-apps").
		Set("disable-extensions").
		Set("disable-hang-monitor").
		Set("disable-popup-blocking").
		Set("disable-prompt-on-repost").
		Set("disable-sync").
		Set("disable-translate").
		Set("metrics-recording-only").
		Set("no-first-run").
		Set("safebrowsing-disable-auto-update")

	// Launch browser
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	b.launcher = l

	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	b.rod = browser

	// Create initial page
	page, err := b.rod.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return fmt.Errorf("failed to create initial page: %w", err)
	}

	// Set viewport
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  b.config.ViewportWidth,
		Height: b.config.ViewportHeight,
	}); err != nil {
		return fmt.Errorf("failed to set viewport: %w", err)
	}

	// Register initial tab
	tabID := generateTabID()
	b.pages[tabID] = page
	b.activeTabID = tabID

	// Create extractor
	b.extractor = dom.NewExtractor(100)

	return nil
}

// Close shuts down the browser and cleans up resources.
func (b *Browser) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var errs []error

	// Close all pages
	for _, page := range b.pages {
		if err := page.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	b.pages = make(map[string]*rod.Page)

	// Close browser
	if b.rod != nil {
		if err := b.rod.Close(); err != nil {
			errs = append(errs, err)
		}
		b.rod = nil
	}

	// Clean up temporary profile
	if b.tempProfilePath != "" {
		if err := os.RemoveAll(b.tempProfilePath); err != nil {
			errs = append(errs, err)
		}
		b.tempProfilePath = ""
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// ActivePage returns the currently active page.
func (b *Browser) ActivePage() *rod.Page {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.pages[b.activeTabID]
}

// GetURL returns the current page URL.
func (b *Browser) GetURL() string {
	page := b.ActivePage()
	if page == nil {
		return ""
	}
	info, err := page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

// GetTitle returns the current page title.
func (b *Browser) GetTitle() string {
	page := b.ActivePage()
	if page == nil {
		return ""
	}
	info, err := page.Info()
	if err != nil {
		return ""
	}
	return info.Title
}

// ListTabs returns information about all open tabs.
func (b *Browser) ListTabs() []TabInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tabs := make([]TabInfo, 0, len(b.pages))
	for id, page := range b.pages {
		info, err := page.Info()
		if err != nil {
			continue
		}
		tabs = append(tabs, TabInfo{
			ID:     id,
			URL:    info.URL,
			Title:  info.Title,
			Active: id == b.activeTabID,
		})
	}
	return tabs
}

// NewTab creates a new tab and optionally navigates to a URL.
func (b *Browser) NewTab(ctx context.Context, url string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.rod == nil {
		return "", fmt.Errorf("browser not started")
	}

	targetURL := "about:blank"
	if url != "" {
		targetURL = url
	}

	page, err := b.rod.Page(proto.TargetCreateTarget{URL: targetURL})
	if err != nil {
		return "", fmt.Errorf("failed to create new tab: %w", err)
	}

	// Set viewport
	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  b.config.ViewportWidth,
		Height: b.config.ViewportHeight,
	}); err != nil {
		return "", fmt.Errorf("failed to set viewport: %w", err)
	}

	if url != "" {
		_ = page.WaitStable(500 * time.Millisecond)
	}

	tabID := generateTabID()
	b.pages[tabID] = page
	b.activeTabID = tabID

	return tabID, nil
}

// SwitchTab switches to a tab by ID.
func (b *Browser) SwitchTab(tabID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page, ok := b.pages[tabID]
	if !ok {
		return fmt.Errorf("tab not found: %s", tabID)
	}

	// Bring tab to front
	if _, err := page.Activate(); err != nil {
		return fmt.Errorf("failed to activate tab: %w", err)
	}

	b.activeTabID = tabID
	return nil
}

// CloseTab closes a tab by ID.
func (b *Browser) CloseTab(tabID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page, ok := b.pages[tabID]
	if !ok {
		return fmt.Errorf("tab not found: %s", tabID)
	}

	// Can't close the last tab
	if len(b.pages) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	if err := page.Close(); err != nil {
		return fmt.Errorf("failed to close tab: %w", err)
	}

	delete(b.pages, tabID)

	// Switch to another tab if we closed the active one
	if b.activeTabID == tabID {
		for id := range b.pages {
			b.activeTabID = id
			break
		}
	}

	return nil
}

// GetElementMap extracts interactive elements from the current page.
func (b *Browser) GetElementMap(ctx context.Context) (*dom.ElementMap, error) {
	page := b.ActivePage()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	return b.extractor.Extract(ctx, page)
}

// SetMaxElements sets the maximum number of elements to extract.
func (b *Browser) SetMaxElements(max int) {
	b.extractor = dom.NewExtractor(max)
}

// WaitStable waits for the page to become stable.
func (b *Browser) WaitStable(ctx context.Context) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}
	_ = ctx // Context available for future use
	return page.WaitStable(500 * time.Millisecond)
}

// generateTabID creates a unique 4-character tab ID.
func generateTabID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 4)
	for i := range id {
		id[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(id)
}
