package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"

	"github.com/anxuanzi/bua/dom"
	screenshotpkg "github.com/anxuanzi/bua/screenshot"
)

// Navigate navigates the current page to a URL.
func (b *Browser) Navigate(ctx context.Context, url string) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Add human-like delay before navigation
	if b.config.Stealth.HumanLikeDelays {
		humanDelay(b.config.Stealth.MinDelay, b.config.Stealth.MaxDelay)
	}

	// Navigate to URL
	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page to load
	if err := page.WaitLoad(); err != nil {
		// Continue even if wait fails - page might be dynamic
	}

	// Wait for stability
	_ = ctx // Context available for future use
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// GoBack navigates back in history.
func (b *Browser) GoBack(ctx context.Context) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	if err := page.NavigateBack(); err != nil {
		return fmt.Errorf("go back failed: %w", err)
	}

	_ = ctx
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// GoForward navigates forward in history.
func (b *Browser) GoForward(ctx context.Context) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	if err := page.NavigateForward(); err != nil {
		return fmt.Errorf("go forward failed: %w", err)
	}

	_ = ctx
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// Reload reloads the current page.
func (b *Browser) Reload(ctx context.Context) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	if err := page.Reload(); err != nil {
		return fmt.Errorf("reload failed: %w", err)
	}

	_ = ctx
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// Click clicks on an element by index.
func (b *Browser) Click(ctx context.Context, elementIndex int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Show highlight if enabled
	if b.config.ShowHighlight {
		b.highlightElement(ctx, element)
	}

	// Add human-like delay before click
	if b.config.Stealth.HumanLikeDelays {
		humanDelay(b.config.Stealth.MinDelay, b.config.Stealth.MaxDelay)
	}

	// Get center coordinates with optional random offset for human-like behavior
	centerX, centerY := element.BoundingBox.Center()
	if b.config.Stealth.HumanLikeDelays {
		offsetX, offsetY := randomMouseOffset(3.0) // Max 3px offset
		centerX += offsetX
		centerY += offsetY
	}

	// Move mouse with human-like motion (linear interpolation)
	if err := page.Mouse.MoveLinear(proto.Point{X: centerX, Y: centerY}, 5); err != nil {
		// Fallback to direct move if linear fails
		if err := page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY}); err != nil {
			return fmt.Errorf("failed to move mouse: %w", err)
		}
	}

	// Small delay before click (like human reaction time)
	if b.config.Stealth.HumanLikeDelays {
		humanDelay(20, 50)
	}

	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click failed: %w", err)
	}

	// Wait for stability after click
	time.Sleep(100 * time.Millisecond)
	_ = ctx
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// ClickAt clicks at specific coordinates.
func (b *Browser) ClickAt(ctx context.Context, x, y float64) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Move mouse and click
	if err := page.Mouse.MoveTo(proto.Point{X: x, Y: y}); err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click failed: %w", err)
	}

	// Wait for stability after click
	time.Sleep(100 * time.Millisecond)
	_ = ctx
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	return nil
}

// DoubleClick double-clicks on an element by index.
func (b *Browser) DoubleClick(ctx context.Context, elementIndex int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Show highlight if enabled
	if b.config.ShowHighlight {
		b.highlightElement(ctx, element)
	}

	centerX, centerY := element.BoundingBox.Center()

	if err := page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY}); err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	// Double click
	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 2); err != nil {
		return fmt.Errorf("double click failed: %w", err)
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}

// TypeText types text into an element by index.
func (b *Browser) TypeText(ctx context.Context, elementIndex int, text string, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Show highlight if enabled
	if b.config.ShowHighlight {
		b.highlightElement(ctx, element)
	}

	// Add human-like delay before typing
	if b.config.Stealth.HumanLikeDelays {
		humanDelay(b.config.Stealth.MinDelay, b.config.Stealth.MaxDelay)
	}

	// Click to focus the element first
	centerX, centerY := element.BoundingBox.Center()
	if b.config.Stealth.HumanLikeDelays {
		offsetX, offsetY := randomMouseOffset(2.0)
		centerX += offsetX
		centerY += offsetY
	}

	if err := page.Mouse.MoveLinear(proto.Point{X: centerX, Y: centerY}, 5); err != nil {
		if err := page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY}); err != nil {
			return fmt.Errorf("failed to move mouse: %w", err)
		}
	}
	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click to focus failed: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Clear existing content
	if err := b.clearInput(page); err != nil {
		// Continue even if clear fails
	}

	// Type the text - use character-by-character for more human-like behavior
	if b.config.Stealth.HumanLikeDelays && len(text) < 100 {
		// Type character by character with small random delays
		for _, char := range text {
			if err := page.InsertText(string(char)); err != nil {
				return fmt.Errorf("type failed: %w", err)
			}
			humanDelay(30, 80) // Random delay between keystrokes
		}
	} else {
		// Fast insert for longer text
		if err := page.InsertText(text); err != nil {
			return fmt.Errorf("type failed: %w", err)
		}
	}

	return nil
}

// ClearAndType clears an input and types new text.
func (b *Browser) ClearAndType(ctx context.Context, elementIndex int, text string, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Show highlight if enabled
	if b.config.ShowHighlight {
		b.highlightElement(ctx, element)
	}

	// Click to focus
	centerX, centerY := element.BoundingBox.Center()
	if err := page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY}); err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}
	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click to focus failed: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Select all and delete
	if err := b.clearInput(page); err != nil {
		// Try triple-click to select all as fallback
		if err := page.Mouse.Click(proto.InputMouseButtonLeft, 3); err == nil {
			time.Sleep(50 * time.Millisecond)
			page.Keyboard.Type(input.Backspace)
		}
	}

	// Type new text using InsertText for string input
	if err := page.InsertText(text); err != nil {
		return fmt.Errorf("type failed: %w", err)
	}

	return nil
}

// clearInput clears the currently focused input.
func (b *Browser) clearInput(page *rod.Page) error {
	// Select all with Ctrl+A / Cmd+A
	if err := page.Keyboard.Press(input.ControlLeft); err != nil {
		return err
	}
	if err := page.Keyboard.Type(input.KeyA); err != nil {
		return err
	}
	if err := page.Keyboard.Release(input.ControlLeft); err != nil {
		return err
	}

	// Delete selected text
	if err := page.Keyboard.Type(input.Backspace); err != nil {
		return err
	}

	return nil
}

// SendKeys sends keyboard keys to the page.
func (b *Browser) SendKeys(ctx context.Context, keys string) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	// Parse and send keys
	// Support special keys like "Enter", "Escape", "Tab", etc.
	keyMap := map[string]input.Key{
		"Enter":      input.Enter,
		"Escape":     input.Escape,
		"Tab":        input.Tab,
		"Backspace":  input.Backspace,
		"Delete":     input.Delete,
		"ArrowUp":    input.ArrowUp,
		"ArrowDown":  input.ArrowDown,
		"ArrowLeft":  input.ArrowLeft,
		"ArrowRight": input.ArrowRight,
		"Home":       input.Home,
		"End":        input.End,
		"PageUp":     input.PageUp,
		"PageDown":   input.PageDown,
		"Space":      input.Space,
	}

	if key, ok := keyMap[keys]; ok {
		if err := page.Keyboard.Type(key); err != nil {
			return fmt.Errorf("send keys failed: %w", err)
		}
	} else {
		// Type as regular text using InsertText for string input
		if err := page.InsertText(keys); err != nil {
			return fmt.Errorf("send keys failed: %w", err)
		}
	}

	return nil
}

// Scroll scrolls the page or an element.
func (b *Browser) Scroll(ctx context.Context, direction string, amount float64, elementIndex *int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	var scrollX, scrollY float64
	switch direction {
	case "up":
		scrollY = -amount
	case "down":
		scrollY = amount
	case "left":
		scrollX = -amount
	case "right":
		scrollX = amount
	default:
		return fmt.Errorf("invalid scroll direction: %s", direction)
	}

	// If element index is specified, scroll within that element
	if elementIndex != nil && elementMap != nil {
		element, ok := elementMap.Get(*elementIndex)
		if !ok {
			return fmt.Errorf("element not found: index %d", *elementIndex)
		}

		// Show highlight if enabled
		if b.config.ShowHighlight {
			b.highlightElement(ctx, element)
		}

		// Scroll within element using JavaScript
		scrollJS := fmt.Sprintf(`(x, y) => {
			const el = document.elementFromPoint(%f, %f);
			if (el) {
				el.scrollBy(x, y);
				return true;
			}
			return false;
		}`, element.BoundingBox.X+10, element.BoundingBox.Y+10)

		_, err := page.Eval(scrollJS, scrollX, scrollY)
		if err != nil {
			return fmt.Errorf("scroll element failed: %w", err)
		}
	} else {
		// Scroll the page
		if err := page.Mouse.Scroll(scrollX, scrollY, 1); err != nil {
			return fmt.Errorf("scroll page failed: %w", err)
		}
	}

	// Wait for content to load after scroll
	time.Sleep(200 * time.Millisecond)

	return nil
}

// ScrollToElement scrolls an element into view.
func (b *Browser) ScrollToElement(ctx context.Context, elementIndex int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Use JavaScript to scroll element into view
	scrollJS := fmt.Sprintf(`() => {
		const el = document.elementFromPoint(%f, %f);
		if (el) {
			el.scrollIntoView({behavior: 'smooth', block: 'center'});
			return true;
		}
		return false;
	}`, element.BoundingBox.X+10, element.BoundingBox.Y+10)

	_, err := page.Eval(scrollJS)
	if err != nil {
		return fmt.Errorf("scroll to element failed: %w", err)
	}

	time.Sleep(300 * time.Millisecond)
	return nil
}

// highlightElement shows a visual highlight on an element.
func (b *Browser) highlightElement(ctx context.Context, element *dom.Element) {
	page := b.ActivePage()
	if page == nil {
		return
	}

	highlightJS := fmt.Sprintf(`() => {
		const overlay = document.createElement('div');
		overlay.id = 'bua-highlight';
		overlay.style.cssText = 'position:fixed;pointer-events:none;z-index:999999;' +
			'border:3px solid #ff6b6b;background:rgba(255,107,107,0.2);' +
			'left:%fpx;top:%fpx;width:%fpx;height:%fpx;transition:opacity 0.2s;';
		document.body.appendChild(overlay);
		setTimeout(() => {
			overlay.style.opacity = '0';
			setTimeout(() => overlay.remove(), 200);
		}, %d);
	}`,
		element.BoundingBox.X,
		element.BoundingBox.Y,
		element.BoundingBox.Width,
		element.BoundingBox.Height,
		b.config.HighlightDuration.Milliseconds()-200,
	)

	page.Eval(highlightJS)
	time.Sleep(b.config.HighlightDuration)
}

// Hover moves the mouse to hover over an element.
func (b *Browser) Hover(ctx context.Context, elementIndex int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	centerX, centerY := element.BoundingBox.Center()

	if err := page.Mouse.MoveLinear(proto.Point{X: centerX, Y: centerY}, 10); err != nil {
		return fmt.Errorf("hover failed: %w", err)
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}

// Focus focuses on an element.
func (b *Browser) Focus(ctx context.Context, elementIndex int, elementMap *dom.ElementMap) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	element, ok := elementMap.Get(elementIndex)
	if !ok {
		return fmt.Errorf("element not found: index %d", elementIndex)
	}

	// Click to focus
	centerX, centerY := element.BoundingBox.Center()
	if err := page.Mouse.MoveTo(proto.Point{X: centerX, Y: centerY}); err != nil {
		return fmt.Errorf("failed to move mouse: %w", err)
	}
	if err := page.Mouse.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click to focus failed: %w", err)
	}

	return nil
}

// Screenshot takes a screenshot of the current page.
// Uses the enhanced screenshot package with proper page readiness checks.
func (b *Browser) Screenshot(ctx context.Context, fullPage bool) ([]byte, error) {
	page := b.ActivePage()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	// Use the screenshot package with LLM-optimized options
	opts := screenshotpkg.LLMOptions()
	opts.FullPage = fullPage

	return screenshotpkg.Capture(ctx, page, opts)
}

// ScreenshotSafe takes a screenshot, returning nil (not error) for blank pages.
// This is useful for agent loops where blank screenshots should be skipped.
func (b *Browser) ScreenshotSafe(ctx context.Context, fullPage bool) ([]byte, error) {
	page := b.ActivePage()
	if page == nil {
		return nil, nil // No page, return nil safely
	}

	return screenshotpkg.ForLLMSafe(ctx, page, b.config.ViewportWidth)
}

// ScreenshotAfterAction captures a screenshot after an action completes.
// Waits for page stability before capturing.
func (b *Browser) ScreenshotAfterAction(ctx context.Context) ([]byte, error) {
	page := b.ActivePage()
	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	return screenshotpkg.CaptureAfterAction(ctx, page, b.config.ViewportWidth)
}

// IsPageReady checks if the current page is ready for screenshot capture.
func (b *Browser) IsPageReady() bool {
	page := b.ActivePage()
	if page == nil {
		return false
	}

	return screenshotpkg.IsPageReady(page)
}

// WaitForPageReady waits until the page is ready, with timeout.
func (b *Browser) WaitForPageReady(ctx context.Context, timeout time.Duration) error {
	page := b.ActivePage()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	return screenshotpkg.WaitUntilReady(ctx, page, timeout)
}

// ExtractContent extracts text content from the page.
func (b *Browser) ExtractContent(ctx context.Context) (string, error) {
	page := b.ActivePage()
	if page == nil {
		return "", fmt.Errorf("no active page")
	}

	// Extract main text content using JavaScript
	result, err := page.Eval(`() => {
		// Try to get main content area first
		const main = document.querySelector('main, article, [role="main"], .content, #content');
		if (main) {
			return main.innerText;
		}
		// Fallback to body
		return document.body.innerText;
	}`)
	if err != nil {
		return "", fmt.Errorf("content extraction failed: %w", err)
	}

	return result.Value.String(), nil
}

// EvaluateJS evaluates JavaScript code on the page.
func (b *Browser) EvaluateJS(ctx context.Context, script string) (string, error) {
	page := b.ActivePage()
	if page == nil {
		return "", fmt.Errorf("no active page")
	}

	// Wrap script in arrow function if not already
	wrappedScript := script
	if len(script) > 0 && script[0] != '(' {
		wrappedScript = fmt.Sprintf("() => { %s }", script)
	}

	result, err := page.Eval(wrappedScript)
	if err != nil {
		return "", fmt.Errorf("JS evaluation failed: %w", err)
	}

	return result.Value.String(), nil
}
