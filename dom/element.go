package dom

import "sync"

// BoundingBox represents an element's position and size on the page.
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Center returns the center point of the bounding box.
func (b BoundingBox) Center() (x, y float64) {
	return b.X + b.Width/2, b.Y + b.Height/2
}

// Contains checks if a point is within the bounding box.
func (b BoundingBox) Contains(x, y float64) bool {
	return x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height
}

// IsEmpty returns true if the bounding box has no area.
func (b BoundingBox) IsEmpty() bool {
	return b.Width <= 0 || b.Height <= 0
}

// Element represents an interactive page element.
type Element struct {
	// Index is the element's index for LLM reference (0-based).
	Index int `json:"index"`

	// TagName is the HTML tag name (lowercase).
	TagName string `json:"tagName"`

	// Role is the ARIA role or inferred role.
	Role string `json:"role,omitempty"`

	// Name is the accessible name of the element.
	Name string `json:"name,omitempty"`

	// Text is the visible text content (truncated).
	Text string `json:"text,omitempty"`

	// Type is the input type for input elements.
	Type string `json:"type,omitempty"`

	// Href is the link URL for anchor elements.
	Href string `json:"href,omitempty"`

	// Placeholder is the placeholder text for inputs.
	Placeholder string `json:"placeholder,omitempty"`

	// Value is the current value for form elements.
	Value string `json:"value,omitempty"`

	// AriaLabel is the aria-label attribute.
	AriaLabel string `json:"ariaLabel,omitempty"`

	// BoundingBox is the element's position and size.
	BoundingBox BoundingBox `json:"boundingBox"`

	// IsVisible indicates if the element is visible in the viewport.
	IsVisible bool `json:"isVisible"`

	// IsEnabled indicates if the element is not disabled.
	IsEnabled bool `json:"isEnabled"`

	// IsFocusable indicates if the element can receive focus.
	IsFocusable bool `json:"isFocusable"`

	// IsInteractive indicates if the element is interactive.
	IsInteractive bool `json:"isInteractive"`

	// Selector is a unique CSS selector for the element.
	Selector string `json:"selector,omitempty"`

	// BackendNodeID is the CDP backend node ID.
	BackendNodeID int `json:"backendNodeId,omitempty"`
}

// Description returns a human-readable description of the element.
func (e *Element) Description() string {
	if e.AriaLabel != "" {
		return e.AriaLabel
	}
	if e.Name != "" {
		return e.Name
	}
	if e.Placeholder != "" {
		return e.Placeholder
	}
	if e.Text != "" {
		if len(e.Text) > 50 {
			return e.Text[:50] + "..."
		}
		return e.Text
	}
	if e.Value != "" {
		return e.Value
	}
	return e.TagName
}

// ElementMap holds all interactive elements on a page.
type ElementMap struct {
	// Elements is the list of interactive elements.
	Elements []*Element

	// PageURL is the current page URL.
	PageURL string

	// PageTitle is the current page title.
	PageTitle string

	// indexMap provides O(1) lookup by index.
	indexMap map[int]*Element

	mu sync.RWMutex
}

// NewElementMap creates a new empty element map.
func NewElementMap() *ElementMap {
	return &ElementMap{
		Elements: make([]*Element, 0),
		indexMap: make(map[int]*Element),
	}
}

// Add adds an element to the map.
func (m *ElementMap) Add(el *Element) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Elements = append(m.Elements, el)
	m.indexMap[el.Index] = el
}

// Get returns an element by index.
func (m *ElementMap) Get(index int) (*Element, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	el, ok := m.indexMap[index]
	return el, ok
}

// Len returns the number of elements.
func (m *ElementMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Elements)
}

// Clear removes all elements.
func (m *ElementMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Elements = make([]*Element, 0)
	m.indexMap = make(map[int]*Element)
}

// FindBySelector returns the first element matching the selector.
func (m *ElementMap) FindBySelector(selector string) (*Element, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, el := range m.Elements {
		if el.Selector == selector {
			return el, true
		}
	}
	return nil, false
}

// FindByText returns elements containing the given text.
func (m *ElementMap) FindByText(text string) []*Element {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*Element
	for _, el := range m.Elements {
		if containsIgnoreCase(el.Text, text) ||
			containsIgnoreCase(el.AriaLabel, text) ||
			containsIgnoreCase(el.Name, text) {
			results = append(results, el)
		}
	}
	return results
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	// Simple ASCII lowercase comparison
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

// toLower converts ASCII characters to lowercase.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
