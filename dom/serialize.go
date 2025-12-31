package dom

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SerializeOptions configures how elements are serialized.
type SerializeOptions struct {
	// MaxElements limits the number of elements to include.
	MaxElements int

	// IncludeBoundingBox includes position information.
	IncludeBoundingBox bool

	// IncludeSelector includes CSS selectors.
	IncludeSelector bool

	// Compact uses minimal whitespace.
	Compact bool
}

// DefaultSerializeOptions returns sensible defaults.
func DefaultSerializeOptions() SerializeOptions {
	return SerializeOptions{
		MaxElements:        100,
		IncludeBoundingBox: true,
		IncludeSelector:    false,
		Compact:            true,
	}
}

// ToTokenString serializes the element map for LLM consumption.
// Uses a compact format to minimize token usage.
func (m *ElementMap) ToTokenString(opts SerializeOptions) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Page: %s\n", m.PageTitle))
	sb.WriteString(fmt.Sprintf("URL: %s\n\n", m.PageURL))

	// Count elements
	count := len(m.Elements)
	if opts.MaxElements > 0 && count > opts.MaxElements {
		count = opts.MaxElements
	}

	sb.WriteString(fmt.Sprintf("Interactive Elements (%d):\n", count))

	for i, el := range m.Elements {
		if opts.MaxElements > 0 && i >= opts.MaxElements {
			sb.WriteString(fmt.Sprintf("... and %d more elements\n", len(m.Elements)-opts.MaxElements))
			break
		}

		line := formatElement(el, opts)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ToTokenStringLimited is a convenience method with a max elements limit.
func (m *ElementMap) ToTokenStringLimited(maxElements int) string {
	opts := DefaultSerializeOptions()
	opts.MaxElements = maxElements
	return m.ToTokenString(opts)
}

// formatElement formats a single element as a compact string.
func formatElement(el *Element, opts SerializeOptions) string {
	var parts []string

	// Index
	parts = append(parts, fmt.Sprintf("[%d]", el.Index))

	// Tag and type
	if el.Type != "" && el.TagName == "input" {
		parts = append(parts, fmt.Sprintf("%s[%s]", el.TagName, el.Type))
	} else {
		parts = append(parts, el.TagName)
	}

	// Role if different from tag
	if el.Role != "" && !isImplicitRole(el.TagName, el.Role) {
		parts = append(parts, fmt.Sprintf("role=%s", el.Role))
	}

	// Description (aria-label, name, placeholder, or text)
	desc := el.Description()
	if desc != "" && desc != el.TagName {
		// Quote and truncate description
		if len(desc) > 40 {
			desc = desc[:40] + "..."
		}
		parts = append(parts, fmt.Sprintf(`"%s"`, desc))
	}

	// Href for links (truncated)
	if el.Href != "" && el.TagName == "a" {
		href := el.Href
		if len(href) > 50 {
			href = href[:50] + "..."
		}
		parts = append(parts, fmt.Sprintf("href=%s", href))
	}

	// Value for inputs with content
	if el.Value != "" && (el.TagName == "input" || el.TagName == "textarea") {
		val := el.Value
		if len(val) > 30 {
			val = val[:30] + "..."
		}
		parts = append(parts, fmt.Sprintf("value=%q", val))
	}

	// Bounding box
	if opts.IncludeBoundingBox {
		parts = append(parts, fmt.Sprintf("(%.0f,%.0f)", el.BoundingBox.X, el.BoundingBox.Y))
	}

	// Disabled state
	if !el.IsEnabled {
		parts = append(parts, "[disabled]")
	}

	// Selector
	if opts.IncludeSelector && el.Selector != "" {
		parts = append(parts, fmt.Sprintf("sel=%q", el.Selector))
	}

	return strings.Join(parts, " ")
}

// isImplicitRole returns true if the role is implied by the tag.
func isImplicitRole(tag, role string) bool {
	implicitRoles := map[string]string{
		"a":        "link",
		"button":   "button",
		"input":    "textbox",
		"select":   "combobox",
		"textarea": "textbox",
	}
	return implicitRoles[tag] == role
}

// ToMarkdown serializes the element map as markdown for documentation.
func (m *ElementMap) ToMarkdown() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", m.PageTitle))
	sb.WriteString(fmt.Sprintf("**URL:** %s\n\n", m.PageURL))
	sb.WriteString("## Interactive Elements\n\n")
	sb.WriteString("| Index | Element | Description | Position |\n")
	sb.WriteString("|-------|---------|-------------|----------|\n")

	for _, el := range m.Elements {
		desc := el.Description()
		if len(desc) > 30 {
			desc = desc[:30] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | (%.0f,%.0f) |\n",
			el.Index,
			el.TagName,
			desc,
			el.BoundingBox.X,
			el.BoundingBox.Y,
		))
	}

	return sb.String()
}

// ToJSON serializes the element map as JSON.
func (m *ElementMap) ToJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type jsonMap struct {
		PageURL   string     `json:"pageUrl"`
		PageTitle string     `json:"pageTitle"`
		Elements  []*Element `json:"elements"`
	}

	data := jsonMap{
		PageURL:   m.PageURL,
		PageTitle: m.PageTitle,
		Elements:  m.Elements,
	}

	return jsonMarshal(data)
}

// jsonMarshal wraps json.Marshal.
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
