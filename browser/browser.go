// Package browser provides the browser automation layer using go-rod.
package browser

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/google/uuid"

	"github.com/anxuanzi/bua-go/dom"
	"github.com/anxuanzi/bua-go/screenshot"
)

// Viewport defines browser viewport dimensions.
type Viewport struct {
	Width  int
	Height int
}

// Config holds browser configuration.
type Config struct {
	Viewport         *Viewport
	ScreenshotConfig *screenshot.Config
}

// TabInfo contains information about a browser tab.
type TabInfo struct {
	ID    string
	URL   string
	Title string
}

// Browser wraps a rod browser for controlled automation.
// Supports multi-tab management.
type Browser struct {
	rod      *rod.Browser
	config   Config
	screener *screenshot.Manager

	// Multi-tab support
	pages       map[string]*rod.Page // tabID -> page
	activeTabID string               // currently active tab

	// Deprecated: use pages map instead
	page *rod.Page

	mu sync.RWMutex
}

// New creates a new browser wrapper.
func New(rodBrowser *rod.Browser, cfg Config) *Browser {
	b := &Browser{
		rod:    rodBrowser,
		config: cfg,
		pages:  make(map[string]*rod.Page),
	}

	if cfg.ScreenshotConfig != nil {
		b.screener = screenshot.NewManager(cfg.ScreenshotConfig)
	}

	return b
}

// Navigate navigates to the specified URL.
// If no tab exists, creates a new one.
func (b *Browser) Navigate(ctx context.Context, url string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get current page (create if needed)
	page := b.getActivePageLocked()
	if page == nil {
		// Create first tab
		tabID, err := b.createTabLocked(url)
		if err != nil {
			return err
		}
		page = b.pages[tabID]
	} else {
		// Navigate existing page
		err := page.Navigate(url)
		if err != nil {
			return fmt.Errorf("failed to navigate: %w", err)
		}
	}

	// Wait for page to be ready
	err := page.WaitLoad()
	if err != nil {
		return fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Wait for page to stabilize (animations, lazy loading, etc.)
	page.MustWaitStable()

	return nil
}

// createTabLocked creates a new tab (must hold lock).
func (b *Browser) createTabLocked(url string) (string, error) {
	// Create a new page
	page, err := b.rod.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	// Set viewport
	if b.config.Viewport != nil {
		err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:             b.config.Viewport.Width,
			Height:            b.config.Viewport.Height,
			DeviceScaleFactor: 1.0,
			Mobile:            false,
		})
		if err != nil {
			return "", fmt.Errorf("failed to set viewport: %w", err)
		}
	}

	// Generate tab ID
	tabID := uuid.New().String()[:8]

	// Store tab
	b.pages[tabID] = page
	b.activeTabID = tabID

	// Also maintain backward compatibility
	b.page = page

	return tabID, nil
}

// getActivePageLocked returns the active page (must hold lock).
func (b *Browser) getActivePageLocked() *rod.Page {
	if b.activeTabID != "" {
		if page, ok := b.pages[b.activeTabID]; ok {
			return page
		}
	}
	// Fallback to legacy single page
	return b.page
}

// Screenshot takes a screenshot of the current page.
func (b *Browser) Screenshot(ctx context.Context) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	data, err := page.Screenshot(true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	return data, nil
}

// ScreenshotWithAnnotations takes an annotated screenshot with element indices.
func (b *Browser) ScreenshotWithAnnotations(ctx context.Context, elements *dom.ElementMap) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	// Take raw screenshot
	data, err := page.Screenshot(true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Annotate if we have elements and a screener
	if elements != nil && b.screener != nil {
		annotated, err := b.screener.Annotate(data, elements)
		if err != nil {
			return nil, fmt.Errorf("failed to annotate screenshot: %w", err)
		}
		return annotated, nil
	}

	return data, nil
}

// SaveScreenshot saves a screenshot to storage and returns the path.
func (b *Browser) SaveScreenshot(ctx context.Context, data []byte, name string) (string, error) {
	if b.screener == nil {
		return "", fmt.Errorf("screenshot manager not configured")
	}

	return b.screener.Save(data, name)
}

// GetElementMap extracts the element map from the current page.
func (b *Browser) GetElementMap(ctx context.Context) (*dom.ElementMap, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	return dom.ExtractElementMap(ctx, page)
}

// GetAccessibilityTree extracts the accessibility tree from the current page.
func (b *Browser) GetAccessibilityTree(ctx context.Context) (*dom.AccessibilityTree, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	return dom.ExtractAccessibilityTree(ctx, page)
}

// Click clicks on an element by its index in the element map.
func (b *Browser) Click(ctx context.Context, elementIndex int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Get the element map to find the element
	elements, err := dom.ExtractElementMap(ctx, page)
	if err != nil {
		return fmt.Errorf("failed to get element map: %w", err)
	}

	el, ok := elements.ByIndex(elementIndex)
	if !ok {
		return fmt.Errorf("element with index %d not found", elementIndex)
	}

	// Click at the center of the element using JavaScript
	centerX := el.BoundingBox.X + el.BoundingBox.Width/2
	centerY := el.BoundingBox.Y + el.BoundingBox.Height/2

	// Use CDP to click at coordinates
	err = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMouseMoved,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 0,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	err = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMousePressed,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 1,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to press mouse: %w", err)
	}

	err = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMouseReleased,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 1,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to release mouse: %w", err)
	}

	return nil
}

// ClickElement clicks on an element directly.
func (b *Browser) ClickElement(ctx context.Context, el *dom.Element) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Click at the center of the element using CDP
	centerX := el.BoundingBox.X + el.BoundingBox.Width/2
	centerY := el.BoundingBox.Y + el.BoundingBox.Height/2

	err := proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMouseMoved,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 0,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	err = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMousePressed,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 1,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to press mouse: %w", err)
	}

	err = proto.InputDispatchMouseEvent{
		Type:       proto.InputDispatchMouseEventTypeMouseReleased,
		X:          centerX,
		Y:          centerY,
		Button:     proto.InputMouseButtonLeft,
		ClickCount: 1,
	}.Call(page)
	if err != nil {
		return fmt.Errorf("failed to release mouse: %w", err)
	}

	return nil
}

// Type types text into the currently focused element.
func (b *Browser) Type(ctx context.Context, text string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Use InsertText for text input
	return page.InsertText(text)
}

// TypeInElement clicks on an element and types text into it.
func (b *Browser) TypeInElement(ctx context.Context, elementIndex int, text string) error {
	// First click to focus
	if err := b.Click(ctx, elementIndex); err != nil {
		return err
	}

	// Small delay to ensure focus
	b.mu.RLock()
	page := b.getActivePageLocked()
	b.mu.RUnlock()
	if page != nil {
		page.MustWaitStable()
	}

	// Type the text
	return b.Type(ctx, text)
}

// Scroll scrolls the page by the specified amount.
func (b *Browser) Scroll(ctx context.Context, deltaX, deltaY float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	return page.Mouse.Scroll(deltaX, deltaY, 1)
}

// ScrollToElement scrolls an element into view.
func (b *Browser) ScrollToElement(ctx context.Context, elementIndex int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	elements, err := dom.ExtractElementMap(ctx, page)
	if err != nil {
		return fmt.Errorf("failed to get element map: %w", err)
	}

	el, ok := elements.ByIndex(elementIndex)
	if !ok {
		return fmt.Errorf("element with index %d not found", elementIndex)
	}

	// Scroll the element into view using JavaScript
	_, err = page.Eval(fmt.Sprintf(
		`document.querySelector('[data-bua-index="%d"]')?.scrollIntoView({behavior: 'smooth', block: 'center'})`,
		el.Index,
	))
	if err != nil {
		// Fall back to coordinate-based scroll
		return page.Mouse.Scroll(0, el.BoundingBox.Y-300, 1)
	}

	return nil
}

// ScrollInElement scrolls within a specific scrollable element (e.g., modal, sidebar).
// This is useful for scrolling within containers that have their own scroll bars,
// like Instagram comment modals, chat windows, or dropdown lists.
func (b *Browser) ScrollInElement(ctx context.Context, elementIndex int, deltaX, deltaY float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	elements, err := dom.ExtractElementMap(ctx, page)
	if err != nil {
		return fmt.Errorf("failed to get element map: %w", err)
	}

	el, ok := elements.ByIndex(elementIndex)
	if !ok {
		return fmt.Errorf("element with index %d not found", elementIndex)
	}

	// Use JavaScript to scroll within the element
	// scrollBy is the most reliable way to scroll within a specific container
	_, err = page.Eval(fmt.Sprintf(
		`(function() {
			const el = document.querySelector('[data-bua-index="%d"]');
			if (!el) return false;
			el.scrollBy({top: %f, left: %f, behavior: 'smooth'});
			return true;
		})()`,
		el.Index, deltaY, deltaX,
	))
	if err != nil {
		return fmt.Errorf("failed to scroll in element: %w", err)
	}

	return nil
}

// WaitForNavigation waits for a navigation to complete.
func (b *Browser) WaitForNavigation(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	return page.WaitLoad()
}

// WaitForStable waits for the page to become stable (no more changes).
func (b *Browser) WaitForStable(ctx context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	return page.WaitStable(300)
}

// GetURL returns the current page URL.
func (b *Browser) GetURL() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
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
	b.mu.RLock()
	defer b.mu.RUnlock()

	page := b.getActivePageLocked()
	if page == nil {
		return ""
	}

	info, err := page.Info()
	if err != nil {
		return ""
	}
	return info.Title
}

// Page returns the underlying rod.Page for advanced operations.
// Deprecated: Use GetActiveTabID() and multi-tab methods instead.
func (b *Browser) Page() *rod.Page {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.getActivePageLocked()
}

// GetActiveTabID returns the ID of the currently active tab.
func (b *Browser) GetActiveTabID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.activeTabID
}

// intPtr returns a pointer to an int value.
func intPtr(v int) *int {
	return &v
}

// Close closes the browser and all tabs.
func (b *Browser) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Close all tabs
	for tabID, page := range b.pages {
		if page != nil {
			page.Close()
		}
		delete(b.pages, tabID)
	}
	b.activeTabID = ""

	// Legacy cleanup
	if b.page != nil {
		b.page.Close()
		b.page = nil
	}

	if b.rod != nil {
		err := b.rod.Close()
		b.rod = nil
		return err
	}

	return nil
}

// ========================================
// Multi-Tab Management Methods
// ========================================

// NewTab opens a new browser tab with the specified URL.
// Returns the tab ID for later reference.
func (b *Browser) NewTab(ctx context.Context, url string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	tabID, err := b.createTabLocked(url)
	if err != nil {
		return "", err
	}

	// Wait for page to load
	page := b.pages[tabID]
	if err := page.WaitLoad(); err != nil {
		return tabID, fmt.Errorf("page load failed: %w", err)
	}
	page.MustWaitStable()

	return tabID, nil
}

// SwitchTab switches to a different browser tab by its ID.
func (b *Browser) SwitchTab(ctx context.Context, tabID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page, ok := b.pages[tabID]
	if !ok {
		return fmt.Errorf("tab %s not found", tabID)
	}

	b.activeTabID = tabID
	b.page = page // maintain backward compatibility

	// Bring the tab to front
	page.MustActivate()

	return nil
}

// CloseTab closes a browser tab by its ID.
// Cannot close the last remaining tab.
func (b *Browser) CloseTab(ctx context.Context, tabID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page, ok := b.pages[tabID]
	if !ok {
		return fmt.Errorf("tab %s not found", tabID)
	}

	// Don't allow closing the last tab
	if len(b.pages) <= 1 {
		return fmt.Errorf("cannot close the last tab")
	}

	// Close the page
	page.Close()
	delete(b.pages, tabID)

	// If we closed the active tab, switch to another
	if b.activeTabID == tabID {
		for newTabID, newPage := range b.pages {
			b.activeTabID = newTabID
			b.page = newPage
			newPage.MustActivate()
			break
		}
	}

	return nil
}

// ListTabs returns information about all open tabs.
func (b *Browser) ListTabs(ctx context.Context) []TabInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var tabs []TabInfo
	for tabID, page := range b.pages {
		info, err := page.Info()
		if err != nil {
			continue
		}
		tabs = append(tabs, TabInfo{
			ID:    tabID,
			URL:   info.URL,
			Title: info.Title,
		})
	}
	return tabs
}
