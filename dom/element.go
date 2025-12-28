// Package dom provides DOM extraction and element mapping functionality.
package dom

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
)

// BoundingBox represents the bounding box of an element.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Element represents a single element in the element map.
type Element struct {
	// Index is the unique identifier for this element (used in annotations).
	Index int `json:"index"`

	// TagName is the HTML tag name (e.g., "button", "input", "a").
	TagName string `json:"tagName"`

	// Role is the accessibility role (e.g., "button", "link", "textbox").
	Role string `json:"role,omitempty"`

	// Name is the accessible name of the element.
	Name string `json:"name,omitempty"`

	// Text is the visible text content of the element.
	Text string `json:"text,omitempty"`

	// Type is the input type for input elements.
	Type string `json:"type,omitempty"`

	// Href is the link URL for anchor elements.
	Href string `json:"href,omitempty"`

	// Placeholder is the placeholder text for input elements.
	Placeholder string `json:"placeholder,omitempty"`

	// Value is the current value for input elements.
	Value string `json:"value,omitempty"`

	// AriaLabel is the aria-label attribute.
	AriaLabel string `json:"ariaLabel,omitempty"`

	// BoundingBox is the element's position and size on the page.
	BoundingBox BoundingBox `json:"boundingBox"`

	// IsVisible indicates whether the element is visible.
	IsVisible bool `json:"isVisible"`

	// IsEnabled indicates whether the element is enabled (not disabled).
	IsEnabled bool `json:"isEnabled"`

	// IsFocusable indicates whether the element can receive focus.
	IsFocusable bool `json:"isFocusable"`

	// IsInteractive indicates whether the element is interactive.
	IsInteractive bool `json:"isInteractive"`

	// Selector is a CSS selector that can be used to find this element.
	Selector string `json:"selector,omitempty"`

	// BackendNodeID is the Chrome DevTools Protocol backend node ID.
	BackendNodeID int `json:"backendNodeId,omitempty"`
}

// ElementMap holds all interactive elements on a page.
type ElementMap struct {
	Elements  []*Element       `json:"elements"`
	indexMap  map[int]*Element // Quick lookup by index
	PageURL   string           `json:"pageUrl"`
	PageTitle string           `json:"pageTitle"`
}

// NewElementMap creates a new element map.
func NewElementMap() *ElementMap {
	return &ElementMap{
		Elements: make([]*Element, 0),
		indexMap: make(map[int]*Element),
	}
}

// Add adds an element to the map.
func (m *ElementMap) Add(el *Element) {
	m.Elements = append(m.Elements, el)
	m.indexMap[el.Index] = el
}

// ByIndex returns an element by its index.
func (m *ElementMap) ByIndex(index int) (*Element, bool) {
	el, ok := m.indexMap[index]
	return el, ok
}

// Count returns the number of elements in the map.
func (m *ElementMap) Count() int {
	return len(m.Elements)
}

// InteractiveElements returns only interactive elements.
func (m *ElementMap) InteractiveElements() []*Element {
	result := make([]*Element, 0)
	for _, el := range m.Elements {
		if el.IsInteractive && el.IsVisible {
			result = append(result, el)
		}
	}
	return result
}

// ToTokenString converts the element map to a token-efficient string for LLM context.
func (m *ElementMap) ToTokenString() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Page: %s\n", m.PageTitle))
	sb.WriteString(fmt.Sprintf("URL: %s\n", m.PageURL))
	sb.WriteString(fmt.Sprintf("Elements (%d):\n", len(m.Elements)))

	for _, el := range m.Elements {
		if !el.IsVisible {
			continue
		}

		// Format: [index] tag role "text" (type=value, href=url)
		sb.WriteString(fmt.Sprintf("[%d] %s", el.Index, el.TagName))

		if el.Role != "" && el.Role != el.TagName {
			sb.WriteString(fmt.Sprintf(" role=%s", el.Role))
		}

		if el.Name != "" {
			sb.WriteString(fmt.Sprintf(" name=%q", truncate(el.Name, 50)))
		} else if el.Text != "" {
			sb.WriteString(fmt.Sprintf(" %q", truncate(el.Text, 50)))
		} else if el.AriaLabel != "" {
			sb.WriteString(fmt.Sprintf(" aria=%q", truncate(el.AriaLabel, 50)))
		} else if el.Placeholder != "" {
			sb.WriteString(fmt.Sprintf(" placeholder=%q", truncate(el.Placeholder, 50)))
		}

		if el.Type != "" {
			sb.WriteString(fmt.Sprintf(" type=%s", el.Type))
		}

		if el.Href != "" {
			sb.WriteString(fmt.Sprintf(" href=%q", truncate(el.Href, 80)))
		}

		if !el.IsEnabled {
			sb.WriteString(" [disabled]")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ExtractElementMap extracts interactive elements from the page.
func ExtractElementMap(ctx context.Context, page *rod.Page) (*ElementMap, error) {
	elementMap := NewElementMap()

	// Get page info
	info, err := page.Info()
	if err == nil {
		elementMap.PageURL = info.URL
		elementMap.PageTitle = info.Title
	}

	// JavaScript to extract interactive elements
	js := `() => {
		const interactiveTags = ['a', 'button', 'input', 'select', 'textarea', 'label'];
		const interactiveRoles = ['button', 'link', 'textbox', 'checkbox', 'radio', 'combobox', 'listbox', 'menuitem', 'tab', 'switch'];

		const elements = [];
		let index = 0;

		// Helper to check if element is visible
		function isVisible(el) {
			const rect = el.getBoundingClientRect();
			const style = window.getComputedStyle(el);
			return rect.width > 0 &&
				   rect.height > 0 &&
				   style.visibility !== 'hidden' &&
				   style.display !== 'none' &&
				   parseFloat(style.opacity) > 0;
		}

		// Helper to check if element is interactive
		function isInteractive(el) {
			const tag = el.tagName.toLowerCase();
			const role = el.getAttribute('role');
			const tabIndex = el.getAttribute('tabindex');

			if (interactiveTags.includes(tag)) return true;
			if (role && interactiveRoles.includes(role)) return true;
			if (tabIndex !== null && tabIndex !== '-1') return true;
			if (el.onclick || el.getAttribute('onclick')) return true;
			if (el.classList.contains('clickable') || el.classList.contains('btn')) return true;

			return false;
		}

		// Find all potentially interactive elements
		const allElements = document.querySelectorAll('a, button, input, select, textarea, [role], [tabindex], [onclick], .clickable, .btn');

		for (const el of allElements) {
			if (!isInteractive(el)) continue;

			const visible = isVisible(el);
			if (!visible) continue;

			const rect = el.getBoundingClientRect();
			const tag = el.tagName.toLowerCase();

			// Extract text content
			let text = '';
			if (tag === 'input' || tag === 'textarea') {
				text = el.value || '';
			} else {
				text = el.innerText || el.textContent || '';
			}
			text = text.trim().substring(0, 200);

			elements.push({
				index: index++,
				tagName: tag,
				role: el.getAttribute('role') || '',
				name: el.getAttribute('name') || '',
				text: text,
				type: el.type || '',
				href: el.href || '',
				placeholder: el.placeholder || '',
				value: el.value || '',
				ariaLabel: el.getAttribute('aria-label') || '',
				boundingBox: {
					x: rect.x,
					y: rect.y,
					width: rect.width,
					height: rect.height
				},
				isVisible: visible,
				isEnabled: !el.disabled,
				isFocusable: el.tabIndex >= 0,
				isInteractive: true
			});

			// Limit to prevent huge element maps
			if (index >= 500) break;
		}

		return elements;
	}`

	result, err := page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("failed to extract elements: %w", err)
	}

	// Parse the result
	var rawElements []map[string]any
	if err := result.Value.Unmarshal(&rawElements); err != nil {
		return nil, fmt.Errorf("failed to parse elements: %w", err)
	}

	for _, raw := range rawElements {
		el := &Element{
			Index:         int(raw["index"].(float64)),
			TagName:       getString(raw, "tagName"),
			Role:          getString(raw, "role"),
			Name:          getString(raw, "name"),
			Text:          getString(raw, "text"),
			Type:          getString(raw, "type"),
			Href:          getString(raw, "href"),
			Placeholder:   getString(raw, "placeholder"),
			Value:         getString(raw, "value"),
			AriaLabel:     getString(raw, "ariaLabel"),
			IsVisible:     getBool(raw, "isVisible"),
			IsEnabled:     getBool(raw, "isEnabled"),
			IsFocusable:   getBool(raw, "isFocusable"),
			IsInteractive: getBool(raw, "isInteractive"),
		}

		if bb, ok := raw["boundingBox"].(map[string]any); ok {
			el.BoundingBox = BoundingBox{
				X:      getFloat(bb, "x"),
				Y:      getFloat(bb, "y"),
				Width:  getFloat(bb, "width"),
				Height: getFloat(bb, "height"),
			}
		}

		elementMap.Add(el)
	}

	return elementMap, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// ExtractElementBySelector finds a specific element by CSS selector.
func ExtractElementBySelector(ctx context.Context, page *rod.Page, selector string) (*Element, error) {
	js := fmt.Sprintf(`() => {
		const el = document.querySelector(%q);
		if (!el) return null;

		const rect = el.getBoundingClientRect();
		const style = window.getComputedStyle(el);
		const tag = el.tagName.toLowerCase();

		let text = '';
		if (tag === 'input' || tag === 'textarea') {
			text = el.value || '';
		} else {
			text = el.innerText || el.textContent || '';
		}

		return {
			tagName: tag,
			role: el.getAttribute('role') || '',
			name: el.getAttribute('name') || '',
			text: text.trim().substring(0, 200),
			type: el.type || '',
			href: el.href || '',
			placeholder: el.placeholder || '',
			value: el.value || '',
			ariaLabel: el.getAttribute('aria-label') || '',
			boundingBox: {
				x: rect.x,
				y: rect.y,
				width: rect.width,
				height: rect.height
			},
			isVisible: rect.width > 0 && rect.height > 0 && style.visibility !== 'hidden' && style.display !== 'none',
			isEnabled: !el.disabled,
			isFocusable: el.tabIndex >= 0,
			isInteractive: true
		};
	}`, selector)

	result, err := page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("failed to find element: %w", err)
	}

	if result.Value.Nil() {
		return nil, fmt.Errorf("element not found: %s", selector)
	}

	var raw map[string]any
	if err := result.Value.Unmarshal(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse element: %w", err)
	}

	el := &Element{
		Index:         0,
		TagName:       getString(raw, "tagName"),
		Role:          getString(raw, "role"),
		Name:          getString(raw, "name"),
		Text:          getString(raw, "text"),
		Type:          getString(raw, "type"),
		Href:          getString(raw, "href"),
		Placeholder:   getString(raw, "placeholder"),
		Value:         getString(raw, "value"),
		AriaLabel:     getString(raw, "ariaLabel"),
		Selector:      selector,
		IsVisible:     getBool(raw, "isVisible"),
		IsEnabled:     getBool(raw, "isEnabled"),
		IsFocusable:   getBool(raw, "isFocusable"),
		IsInteractive: getBool(raw, "isInteractive"),
	}

	if bb, ok := raw["boundingBox"].(map[string]any); ok {
		el.BoundingBox = BoundingBox{
			X:      getFloat(bb, "x"),
			Y:      getFloat(bb, "y"),
			Width:  getFloat(bb, "width"),
			Height: getFloat(bb, "height"),
		}
	}

	return el, nil
}
