// Package browser provides the browser automation layer using go-rod.
package browser

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
)

// HighlightDebug controls debug logging for highlighting.
// Set to true to see when highlights are triggered.
var HighlightDebug = false

// Highlighter provides visual feedback for browser automation actions.
// Following Python browser-use's approach, it directly modifies element styles
// for more reliable visual feedback that works regardless of page layout.
type Highlighter struct {
	page    *rod.Page
	enabled bool
	delay   time.Duration // How long to show highlight before action
}

// NewHighlighter creates a new highlighter for the given page.
func NewHighlighter(page *rod.Page, enabled bool) *Highlighter {
	if HighlightDebug {
		log.Printf("[highlight] NewHighlighter created: enabled=%v, page=%v", enabled, page != nil)
	}
	return &Highlighter{
		page:    page,
		enabled: enabled,
		delay:   300 * time.Millisecond, // Default 300ms visual feedback
	}
}

// SetDelay sets how long the highlight is shown before action execution.
func (h *Highlighter) SetDelay(d time.Duration) {
	h.delay = d
}

// SetEnabled enables or disables highlighting.
func (h *Highlighter) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// injectStyles injects the CSS for highlight animations if not already present.
func (h *Highlighter) injectStyles() error {
	css := `
/* Flash animation for direct element highlighting (browser-use style) */
@keyframes bua-flash {
	0% { outline-color: #ff6b35; box-shadow: 0 0 0 4px rgba(255, 107, 53, 0.6); }
	50% { outline-color: #ff8c5a; box-shadow: 0 0 0 8px rgba(255, 107, 53, 0.3); }
	100% { outline-color: #ff6b35; box-shadow: 0 0 0 4px rgba(255, 107, 53, 0.6); }
}

.bua-flash-highlight {
	outline: 3px solid #ff6b35 !important;
	outline-offset: 2px !important;
	box-shadow: 0 0 0 4px rgba(255, 107, 53, 0.4) !important;
	animation: bua-flash 0.3s ease-in-out 2 !important;
	transition: none !important;
}

/* Coordinate-based click indicator */
.bua-click-indicator {
	position: fixed;
	pointer-events: none;
	z-index: 2147483647;
	width: 20px;
	height: 20px;
	border: 3px solid #ff6b35;
	border-radius: 50%;
	transform: translate(-50%, -50%);
	animation: bua-click-pulse 0.4s ease-out forwards;
}

@keyframes bua-click-pulse {
	0% {
		transform: translate(-50%, -50%) scale(0.5);
		opacity: 1;
		box-shadow: 0 0 0 0 rgba(255, 107, 53, 0.7);
	}
	100% {
		transform: translate(-50%, -50%) scale(2);
		opacity: 0;
		box-shadow: 0 0 0 10px rgba(255, 107, 53, 0);
	}
}

/* Crosshair for coordinate clicks */
.bua-crosshair {
	position: fixed;
	pointer-events: none;
	z-index: 2147483647;
	background: #ff6b35;
}
.bua-crosshair-h {
	width: 30px;
	height: 2px;
	transform: translateX(-50%);
}
.bua-crosshair-v {
	width: 2px;
	height: 30px;
	transform: translateY(-50%);
}

/* Action label */
.bua-action-label {
	position: fixed;
	pointer-events: none;
	z-index: 2147483647;
	background: #ff6b35;
	color: white;
	padding: 4px 8px;
	font-size: 12px;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
	font-weight: 600;
	border-radius: 4px;
	white-space: nowrap;
	box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

/* Scroll indicator */
.bua-scroll-indicator {
	position: fixed;
	pointer-events: none;
	z-index: 2147483647;
	background: rgba(255, 107, 53, 0.9);
	color: white;
	padding: 8px 16px;
	font-size: 16px;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
	font-weight: 600;
	border-radius: 8px;
	transform: translate(-50%, -50%);
	box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}
`
	js := fmt.Sprintf(`() => {
		if (document.getElementById('bua-highlight-styles')) return true;

		const style = document.createElement('style');
		style.id = 'bua-highlight-styles';
		style.textContent = %q;
		document.head.appendChild(style);
		return true;
	}`, css)

	_, err := h.page.Eval(js)
	return err
}

// FlashElementBySelector flashes an element by CSS selector.
// This is the browser-use style approach - direct element modification.
func (h *Highlighter) FlashElementBySelector(selector, label string) error {
	if !h.enabled || h.page == nil {
		return nil
	}

	if err := h.injectStyles(); err != nil {
		return err
	}

	js := fmt.Sprintf(`() => {
		const selector = %q;
		const label = %q;

		// Find the element
		const el = document.querySelector(selector);
		if (!el) {
			console.warn('bua-highlight: Element not found:', selector);
			return false;
		}

		// Remove any existing highlights
		document.querySelectorAll('.bua-action-label').forEach(e => e.remove());
		document.querySelectorAll('.bua-flash-highlight').forEach(e => {
			e.classList.remove('bua-flash-highlight');
		});

		// Add flash class to element
		el.classList.add('bua-flash-highlight');

		// Add label near the element
		if (label) {
			const rect = el.getBoundingClientRect();
			const labelEl = document.createElement('div');
			labelEl.className = 'bua-action-label';
			labelEl.textContent = label;
			labelEl.style.left = rect.left + 'px';
			labelEl.style.top = Math.max(0, rect.top - 28) + 'px';
			document.body.appendChild(labelEl);
		}

		return true;
	}`, selector, label)

	result, err := h.page.Eval(js)
	if err != nil {
		return fmt.Errorf("failed to flash element: %w", err)
	}

	// Check if element was found
	if result.Value.Bool() {
		time.Sleep(h.delay)
	}

	return nil
}

// FlashElementAtPoint flashes the element at specific viewport coordinates.
// Uses document.elementFromPoint to find and highlight the actual element.
func (h *Highlighter) FlashElementAtPoint(x, y float64, label string) error {
	if HighlightDebug {
		log.Printf("[highlight] FlashElementAtPoint called: x=%.0f, y=%.0f, label=%q, enabled=%v, page=%v",
			x, y, label, h.enabled, h.page != nil)
	}
	if !h.enabled || h.page == nil {
		if HighlightDebug {
			log.Printf("[highlight] FlashElementAtPoint SKIPPED: enabled=%v, page=%v", h.enabled, h.page != nil)
		}
		return nil
	}

	if err := h.injectStyles(); err != nil {
		return err
	}

	js := fmt.Sprintf(`() => {
		try {
			const x = %f;
			const y = %f;
			const label = %q;

			// Remove any existing highlights
			document.querySelectorAll('.bua-action-label, .bua-click-indicator, .bua-crosshair').forEach(e => e.remove());
			document.querySelectorAll('.bua-flash-highlight').forEach(e => {
				e.classList.remove('bua-flash-highlight');
			});

			// Find element at point
			const el = document.elementFromPoint(x, y);

			if (el && el !== document.body && el !== document.documentElement) {
				// Flash the actual element
				el.classList.add('bua-flash-highlight');

				// Add label
				if (label) {
					const rect = el.getBoundingClientRect();
					const labelEl = document.createElement('div');
					labelEl.className = 'bua-action-label';
					labelEl.textContent = label;
					labelEl.style.left = rect.left + 'px';
					labelEl.style.top = Math.max(0, rect.top - 28) + 'px';
					document.body.appendChild(labelEl);
				}
				return {success: true, mode: 'element', tag: el.tagName};
			}

			// Fallback: show click indicator at coordinates
			const indicator = document.createElement('div');
			indicator.className = 'bua-click-indicator';
			indicator.style.left = x + 'px';
			indicator.style.top = y + 'px';
			document.body.appendChild(indicator);

			// Add crosshairs
			const hLine = document.createElement('div');
			hLine.className = 'bua-crosshair bua-crosshair-h';
			hLine.style.left = x + 'px';
			hLine.style.top = y + 'px';
			document.body.appendChild(hLine);

			const vLine = document.createElement('div');
			vLine.className = 'bua-crosshair bua-crosshair-v';
			vLine.style.left = x + 'px';
			vLine.style.top = y + 'px';
			document.body.appendChild(vLine);

			// Add label
			if (label) {
				const labelEl = document.createElement('div');
				labelEl.className = 'bua-action-label';
				labelEl.textContent = label;
				labelEl.style.left = (x + 15) + 'px';
				labelEl.style.top = Math.max(0, y - 28) + 'px';
				document.body.appendChild(labelEl);
			}

			return {success: true, mode: 'crosshair'};
		} catch (e) {
			return {success: false, error: e.message};
		}
	}`, x, y, label)

	result, err := h.page.Eval(js)
	if err != nil {
		if HighlightDebug {
			log.Printf("[highlight] FlashElementAtPoint JS eval error: %v", err)
		}
		return fmt.Errorf("failed to flash element at point: %w", err)
	}

	if HighlightDebug {
		log.Printf("[highlight] FlashElementAtPoint result: %v", result.Value)
	}

	time.Sleep(h.delay)
	return nil
}

// HighlightElement shows animated highlight on an element using direct styling.
// x, y are viewport coordinates of the top-left corner.
// This method finds the element at the center point and flashes it directly.
func (h *Highlighter) HighlightElement(x, y, width, height float64, label string) error {
	if HighlightDebug {
		log.Printf("[highlight] HighlightElement called: enabled=%v, page=%v, coords=(%.0f, %.0f, %.0f, %.0f), label=%q",
			h.enabled, h.page != nil, x, y, width, height, label)
	}
	if !h.enabled || h.page == nil {
		if HighlightDebug {
			log.Printf("[highlight] HighlightElement SKIPPED: enabled=%v, page=%v", h.enabled, h.page != nil)
		}
		return nil
	}

	// Calculate center point and use FlashElementAtPoint for better reliability
	centerX := x + width/2
	centerY := y + height/2

	return h.FlashElementAtPoint(centerX, centerY, label)
}

// HighlightCoordinates shows a crosshair and expanding circle at the click position.
func (h *Highlighter) HighlightCoordinates(x, y float64, label string) error {
	if !h.enabled || h.page == nil {
		return nil
	}

	return h.FlashElementAtPoint(x, y, label)
}

// HighlightScroll shows a scroll indicator on the page or element.
func (h *Highlighter) HighlightScroll(x, y float64, direction string) error {
	if !h.enabled || h.page == nil {
		return nil
	}

	if err := h.injectStyles(); err != nil {
		return err
	}

	arrow := "↓"
	text := "Scrolling down"
	switch direction {
	case "up":
		arrow = "↑"
		text = "Scrolling up"
	case "left":
		arrow = "←"
		text = "Scrolling left"
	case "right":
		arrow = "→"
		text = "Scrolling right"
	}

	js := fmt.Sprintf(`() => {
		// Remove any existing indicators
		document.querySelectorAll('.bua-scroll-indicator').forEach(e => e.remove());

		const x = %f;
		const y = %f;
		const arrow = %q;
		const text = %q;

		const indicator = document.createElement('div');
		indicator.className = 'bua-scroll-indicator';
		indicator.textContent = arrow + ' ' + text;
		indicator.style.left = x + 'px';
		indicator.style.top = y + 'px';
		document.body.appendChild(indicator);

		return true;
	}`, x, y, arrow, text)

	_, err := h.page.Eval(js)
	if err != nil {
		return fmt.Errorf("failed to show scroll highlight: %w", err)
	}

	// Shorter delay for scroll
	time.Sleep(h.delay / 2)
	return nil
}

// HighlightType shows a typing indicator on an input element.
func (h *Highlighter) HighlightType(x, y, width, height float64, text string) error {
	if !h.enabled || h.page == nil {
		return nil
	}

	// Create label with typing indicator
	label := "⌨️ typing..."
	if len(text) > 20 {
		label = fmt.Sprintf("⌨️ %s...", text[:20])
	} else if len(text) > 0 {
		label = fmt.Sprintf("⌨️ %s", text)
	}

	return h.HighlightElement(x, y, width, height, label)
}

// RemoveHighlights removes all highlight elements from the page.
func (h *Highlighter) RemoveHighlights() error {
	if h.page == nil {
		return nil
	}

	_, err := h.page.Eval(`() => {
		// Remove all highlight overlays
		document.querySelectorAll('.bua-action-label, .bua-click-indicator, .bua-crosshair, .bua-scroll-indicator').forEach(el => el.remove());

		// Remove flash class from all elements
		document.querySelectorAll('.bua-flash-highlight').forEach(el => {
			el.classList.remove('bua-flash-highlight');
		});
		return true;
	}`)
	return err
}
