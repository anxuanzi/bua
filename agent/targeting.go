// Package agent provides enhanced element targeting capabilities.
package agent

import (
	"regexp"
	"sort"
	"strings"

	"github.com/anxuanzi/bua-go/dom"
)

// TargetingStrategy defines how to find elements.
type TargetingStrategy string

const (
	// StrategyExact requires exact text match.
	StrategyExact TargetingStrategy = "exact"
	// StrategyFuzzy uses fuzzy text matching.
	StrategyFuzzy TargetingStrategy = "fuzzy"
	// StrategySemantic uses role and purpose matching.
	StrategySemantic TargetingStrategy = "semantic"
	// StrategyPosition uses spatial positioning.
	StrategyPosition TargetingStrategy = "position"
)

// ElementQuery defines criteria for finding elements.
type ElementQuery struct {
	// Text to match (supports partial/fuzzy matching).
	Text string `json:"text,omitempty"`

	// TagName to match (e.g., "button", "input").
	TagName string `json:"tagName,omitempty"`

	// Role to match (e.g., "button", "textbox").
	Role string `json:"role,omitempty"`

	// Type for input elements (e.g., "submit", "text").
	Type string `json:"type,omitempty"`

	// Placeholder text to match.
	Placeholder string `json:"placeholder,omitempty"`

	// AriaLabel to match.
	AriaLabel string `json:"ariaLabel,omitempty"`

	// HrefContains matches link URLs.
	HrefContains string `json:"hrefContains,omitempty"`

	// NearElement finds elements near another element index.
	NearElement *int `json:"nearElement,omitempty"`

	// NearDistance is the max pixel distance for NearElement.
	NearDistance float64 `json:"nearDistance,omitempty"`

	// Strategy determines matching approach.
	Strategy TargetingStrategy `json:"strategy,omitempty"`

	// MustBeVisible requires the element to be visible.
	MustBeVisible bool `json:"mustBeVisible,omitempty"`

	// MustBeEnabled requires the element to be enabled.
	MustBeEnabled bool `json:"mustBeEnabled,omitempty"`

	// MaxResults limits the number of results.
	MaxResults int `json:"maxResults,omitempty"`
}

// ElementMatch represents a matching element with its score.
type ElementMatch struct {
	Element   *dom.Element `json:"element"`
	Score     float64      `json:"score"`
	MatchType string       `json:"matchType"`
	Reasons   []string     `json:"reasons"`
}

// ElementMatcher provides advanced element targeting.
type ElementMatcher struct {
	// wordSimilarityCache caches computed similarities.
	wordSimilarityCache map[string]float64
}

// NewElementMatcher creates a new element matcher.
func NewElementMatcher() *ElementMatcher {
	return &ElementMatcher{
		wordSimilarityCache: make(map[string]float64),
	}
}

// FindElements finds elements matching the query criteria.
func (m *ElementMatcher) FindElements(elements *dom.ElementMap, query ElementQuery) []ElementMatch {
	if elements == nil || len(elements.Elements) == 0 {
		return nil
	}

	// Set defaults
	if query.Strategy == "" {
		query.Strategy = StrategyFuzzy
	}
	if query.NearDistance == 0 {
		query.NearDistance = 200 // Default 200px radius
	}
	if query.MaxResults == 0 {
		query.MaxResults = 10
	}

	matches := make([]ElementMatch, 0)

	for _, el := range elements.Elements {
		// Apply visibility filter
		if query.MustBeVisible && !el.IsVisible {
			continue
		}
		if query.MustBeEnabled && !el.IsEnabled {
			continue
		}

		match := m.scoreElement(el, elements, query)
		if match.Score > 0 {
			matches = append(matches, match)
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Limit results
	if len(matches) > query.MaxResults {
		matches = matches[:query.MaxResults]
	}

	return matches
}

// FindBestMatch returns the single best matching element.
func (m *ElementMatcher) FindBestMatch(elements *dom.ElementMap, query ElementQuery) *ElementMatch {
	query.MaxResults = 1
	matches := m.FindElements(elements, query)
	if len(matches) == 0 {
		return nil
	}
	return &matches[0]
}

// scoreElement scores how well an element matches the query.
func (m *ElementMatcher) scoreElement(el *dom.Element, allElements *dom.ElementMap, query ElementQuery) ElementMatch {
	match := ElementMatch{
		Element: el,
		Score:   0,
		Reasons: make([]string, 0),
	}

	// Tag name matching
	if query.TagName != "" {
		if strings.EqualFold(el.TagName, query.TagName) {
			match.Score += 0.2
			match.Reasons = append(match.Reasons, "tag match")
		} else {
			return match // Wrong tag, skip
		}
	}

	// Role matching
	if query.Role != "" {
		if strings.EqualFold(el.Role, query.Role) {
			match.Score += 0.15
			match.Reasons = append(match.Reasons, "role match")
		}
	}

	// Type matching
	if query.Type != "" {
		if strings.EqualFold(el.Type, query.Type) {
			match.Score += 0.15
			match.Reasons = append(match.Reasons, "type match")
		}
	}

	// Text matching
	if query.Text != "" {
		textScore, textReason := m.matchText(el, query.Text, query.Strategy)
		if textScore > 0 {
			match.Score += textScore
			match.Reasons = append(match.Reasons, textReason)
			match.MatchType = string(query.Strategy)
		}
	}

	// Placeholder matching
	if query.Placeholder != "" {
		similarity := m.textSimilarity(el.Placeholder, query.Placeholder)
		if similarity > 0.5 {
			match.Score += similarity * 0.2
			match.Reasons = append(match.Reasons, "placeholder match")
		}
	}

	// AriaLabel matching
	if query.AriaLabel != "" {
		similarity := m.textSimilarity(el.AriaLabel, query.AriaLabel)
		if similarity > 0.5 {
			match.Score += similarity * 0.15
			match.Reasons = append(match.Reasons, "aria-label match")
		}
	}

	// Href matching
	if query.HrefContains != "" {
		if strings.Contains(strings.ToLower(el.Href), strings.ToLower(query.HrefContains)) {
			match.Score += 0.2
			match.Reasons = append(match.Reasons, "href match")
		}
	}

	// Proximity matching
	if query.NearElement != nil {
		refEl, ok := allElements.ByIndex(*query.NearElement)
		if ok {
			distance := m.elementDistance(el, refEl)
			if distance <= query.NearDistance {
				// Closer = higher score
				proximityScore := 1.0 - (distance / query.NearDistance)
				match.Score += proximityScore * 0.3
				match.Reasons = append(match.Reasons, "proximity match")
			}
		}
	}

	// Bonus for interactive elements
	if el.IsInteractive && el.IsEnabled {
		match.Score *= 1.1
	}

	// Clamp score
	if match.Score > 1.0 {
		match.Score = 1.0
	}

	return match
}

// matchText scores text matching with different strategies.
func (m *ElementMatcher) matchText(el *dom.Element, queryText string, strategy TargetingStrategy) (float64, string) {
	queryLower := strings.ToLower(strings.TrimSpace(queryText))

	// Collect all text sources from the element
	textSources := []struct {
		text   string
		weight float64
	}{
		{el.Text, 1.0},
		{el.Name, 0.8},
		{el.AriaLabel, 0.9},
		{el.Placeholder, 0.7},
		{el.Value, 0.5},
	}

	var bestScore float64
	var bestReason string

	for _, source := range textSources {
		if source.text == "" {
			continue
		}

		sourceLower := strings.ToLower(strings.TrimSpace(source.text))
		var score float64
		var reason string

		switch strategy {
		case StrategyExact:
			if sourceLower == queryLower {
				score = 0.5 * source.weight
				reason = "exact match"
			}

		case StrategyFuzzy:
			// Check for exact match first
			if sourceLower == queryLower {
				score = 0.5 * source.weight
				reason = "exact match"
			} else if strings.Contains(sourceLower, queryLower) {
				// Contains the query
				score = 0.4 * source.weight
				reason = "contains match"
			} else if strings.Contains(queryLower, sourceLower) {
				// Query contains the element text
				score = 0.3 * source.weight
				reason = "partial match"
			} else {
				// Fuzzy similarity
				similarity := m.textSimilarity(sourceLower, queryLower)
				if similarity > 0.5 {
					score = similarity * 0.4 * source.weight
					reason = "fuzzy match"
				}
			}

		case StrategySemantic:
			// Semantic matching: check for related words
			if m.semanticMatch(sourceLower, queryLower) {
				score = 0.35 * source.weight
				reason = "semantic match"
			} else if strings.Contains(sourceLower, queryLower) {
				score = 0.4 * source.weight
				reason = "contains match"
			}

		default:
			// Default to fuzzy
			if strings.Contains(sourceLower, queryLower) {
				score = 0.4 * source.weight
				reason = "contains match"
			}
		}

		if score > bestScore {
			bestScore = score
			bestReason = reason
		}
	}

	return bestScore, bestReason
}

// textSimilarity computes a simple similarity score between two strings.
func (m *ElementMatcher) textSimilarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}

	// Check cache
	cacheKey := a + "|" + b
	if cached, ok := m.wordSimilarityCache[cacheKey]; ok {
		return cached
	}

	// Normalize strings
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))

	if a == b {
		return 1.0
	}

	// Word overlap approach
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	// Count matching words
	matches := 0
	for _, wa := range wordsA {
		for _, wb := range wordsB {
			if wa == wb || strings.Contains(wa, wb) || strings.Contains(wb, wa) {
				matches++
				break
			}
		}
	}

	// Jaccard-like similarity
	similarity := float64(matches) / float64(len(wordsA)+len(wordsB)-matches)

	// Cache the result
	m.wordSimilarityCache[cacheKey] = similarity

	return similarity
}

// semanticMatch checks for semantic relationships between texts.
func (m *ElementMatcher) semanticMatch(elementText, queryText string) bool {
	// Common semantic groupings
	semanticGroups := map[string][]string{
		"submit": {"submit", "send", "go", "ok", "confirm", "save", "apply"},
		"cancel": {"cancel", "close", "dismiss", "back", "exit", "abort"},
		"login":  {"login", "sign in", "signin", "log in", "authenticate"},
		"logout": {"logout", "sign out", "signout", "log out"},
		"search": {"search", "find", "look", "query"},
		"next":   {"next", "continue", "forward", "proceed"},
		"prev":   {"previous", "back", "prior"},
		"add":    {"add", "create", "new", "plus", "+"},
		"delete": {"delete", "remove", "trash", "erase"},
		"edit":   {"edit", "modify", "change", "update"},
		"menu":   {"menu", "hamburger", "navigation", "nav"},
		"close":  {"close", "x", "dismiss", "hide"},
	}

	queryLower := strings.ToLower(queryText)
	elementLower := strings.ToLower(elementText)

	// Find which semantic group the query belongs to
	for _, synonyms := range semanticGroups {
		queryMatches := false
		elementMatches := false

		for _, synonym := range synonyms {
			if strings.Contains(queryLower, synonym) {
				queryMatches = true
			}
			if strings.Contains(elementLower, synonym) {
				elementMatches = true
			}
		}

		if queryMatches && elementMatches {
			return true
		}
	}

	return false
}

// elementDistance calculates the distance between two element centers.
func (m *ElementMatcher) elementDistance(a, b *dom.Element) float64 {
	// Get center points
	ax := a.BoundingBox.X + a.BoundingBox.Width/2
	ay := a.BoundingBox.Y + a.BoundingBox.Height/2
	bx := b.BoundingBox.X + b.BoundingBox.Width/2
	by := b.BoundingBox.Y + b.BoundingBox.Height/2

	// Euclidean distance
	dx := ax - bx
	dy := ay - by
	return sqrt(dx*dx + dy*dy)
}

// sqrt computes square root without importing math for this simple case.
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}
	// Newton's method
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// FindByText is a convenience method to find elements by text.
func (m *ElementMatcher) FindByText(elements *dom.ElementMap, text string) []ElementMatch {
	return m.FindElements(elements, ElementQuery{
		Text:          text,
		Strategy:      StrategyFuzzy,
		MustBeVisible: true,
	})
}

// FindByRole finds elements by their accessibility role.
func (m *ElementMatcher) FindByRole(elements *dom.ElementMap, role string) []ElementMatch {
	return m.FindElements(elements, ElementQuery{
		Role:          role,
		MustBeVisible: true,
	})
}

// FindButton finds button elements by text.
func (m *ElementMatcher) FindButton(elements *dom.ElementMap, text string) *ElementMatch {
	return m.FindBestMatch(elements, ElementQuery{
		Text:          text,
		Role:          "button",
		Strategy:      StrategyFuzzy,
		MustBeVisible: true,
		MustBeEnabled: true,
	})
}

// FindInput finds input elements by placeholder or label.
func (m *ElementMatcher) FindInput(elements *dom.ElementMap, labelOrPlaceholder string) *ElementMatch {
	return m.FindBestMatch(elements, ElementQuery{
		Text:          labelOrPlaceholder,
		Placeholder:   labelOrPlaceholder,
		TagName:       "input",
		MustBeVisible: true,
		MustBeEnabled: true,
	})
}

// FindLink finds links by text or href.
func (m *ElementMatcher) FindLink(elements *dom.ElementMap, textOrHref string) *ElementMatch {
	return m.FindBestMatch(elements, ElementQuery{
		Text:         textOrHref,
		HrefContains: textOrHref,
		TagName:      "a",
	})
}

// FindNear finds elements near a reference element.
func (m *ElementMatcher) FindNear(elements *dom.ElementMap, refIndex int, query ElementQuery) []ElementMatch {
	query.NearElement = &refIndex
	if query.NearDistance == 0 {
		query.NearDistance = 150
	}
	return m.FindElements(elements, query)
}

// ElementSelector generates a CSS selector for an element.
type ElementSelector struct{}

// NewElementSelector creates a new selector generator.
func NewElementSelector() *ElementSelector {
	return &ElementSelector{}
}

// GenerateSelector creates a unique CSS selector for an element.
func (s *ElementSelector) GenerateSelector(el *dom.Element) string {
	var parts []string

	// Start with tag name
	parts = append(parts, el.TagName)

	// Add ID if present (most specific)
	if el.Name != "" && isValidIdentifier(el.Name) {
		return el.TagName + "#" + el.Name
	}

	// Add role if present
	if el.Role != "" {
		parts = append(parts, "[role=\""+el.Role+"\"]")
	}

	// Add type for inputs
	if el.Type != "" && el.TagName == "input" {
		parts = append(parts, "[type=\""+el.Type+"\"]")
	}

	// Add aria-label if present
	if el.AriaLabel != "" {
		escapedLabel := strings.ReplaceAll(el.AriaLabel, "\"", "\\\"")
		parts = append(parts, "[aria-label=\""+escapedLabel+"\"]")
	}

	return strings.Join(parts, "")
}

// GenerateXPath creates an XPath expression for an element.
func (s *ElementSelector) GenerateXPath(el *dom.Element) string {
	var conditions []string

	conditions = append(conditions, "//"+el.TagName)

	if el.Text != "" && len(el.Text) < 50 {
		escapedText := strings.ReplaceAll(el.Text, "'", "\\'")
		conditions = append(conditions, "[contains(text(), '"+escapedText+"')]")
	}

	if el.Role != "" {
		conditions = append(conditions, "[@role='"+el.Role+"']")
	}

	return strings.Join(conditions, "")
}

// isValidIdentifier checks if a string is a valid CSS identifier.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	// Simple check: starts with letter or underscore, contains only alphanumeric, hyphens, underscores
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_-]*$`, s)
	return match
}

// ElementGroup represents a group of related elements.
type ElementGroup struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"` // form, menu, list, etc.
	Elements []*dom.Element `json:"elements"`
}

// ElementGrouper groups related elements together.
type ElementGrouper struct{}

// NewElementGrouper creates a new element grouper.
func NewElementGrouper() *ElementGrouper {
	return &ElementGrouper{}
}

// FindFormGroups identifies form field groups.
func (g *ElementGrouper) FindFormGroups(elements *dom.ElementMap) []ElementGroup {
	groups := make([]ElementGroup, 0)

	// Find inputs and their associated labels
	formElements := make([]*dom.Element, 0)
	for _, el := range elements.Elements {
		if el.TagName == "input" || el.TagName == "textarea" || el.TagName == "select" {
			formElements = append(formElements, el)
		}
	}

	if len(formElements) > 0 {
		groups = append(groups, ElementGroup{
			Name:     "Form Fields",
			Type:     "form",
			Elements: formElements,
		})
	}

	return groups
}

// FindNavigationGroups identifies navigation elements.
func (g *ElementGrouper) FindNavigationGroups(elements *dom.ElementMap) []ElementGroup {
	groups := make([]ElementGroup, 0)

	navElements := make([]*dom.Element, 0)
	for _, el := range elements.Elements {
		if el.Role == "navigation" || el.Role == "menu" ||
			el.TagName == "nav" || strings.Contains(strings.ToLower(el.AriaLabel), "nav") {
			navElements = append(navElements, el)
		}
	}

	if len(navElements) > 0 {
		groups = append(groups, ElementGroup{
			Name:     "Navigation",
			Type:     "navigation",
			Elements: navElements,
		})
	}

	return groups
}

// FindButtonGroups identifies button groups (e.g., action buttons).
func (g *ElementGrouper) FindButtonGroups(elements *dom.ElementMap) []ElementGroup {
	groups := make([]ElementGroup, 0)

	buttonElements := make([]*dom.Element, 0)
	for _, el := range elements.Elements {
		if el.Role == "button" || el.TagName == "button" ||
			(el.TagName == "input" && (el.Type == "submit" || el.Type == "button")) {
			buttonElements = append(buttonElements, el)
		}
	}

	if len(buttonElements) > 0 {
		groups = append(groups, ElementGroup{
			Name:     "Buttons",
			Type:     "buttons",
			Elements: buttonElements,
		})
	}

	return groups
}
