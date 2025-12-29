package agent

import (
	"testing"

	"github.com/anxuanzi/bua-go/dom"
)

func createTestElementMap() *dom.ElementMap {
	em := dom.NewElementMap()
	em.PageURL = "https://example.com"
	em.PageTitle = "Test Page"

	em.Add(&dom.Element{
		Index:         0,
		TagName:       "button",
		Role:          "button",
		Text:          "Submit Form",
		IsVisible:     true,
		IsEnabled:     true,
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 100, Y: 100, Width: 80, Height: 30},
	})

	em.Add(&dom.Element{
		Index:         1,
		TagName:       "button",
		Role:          "button",
		Text:          "Cancel",
		IsVisible:     true,
		IsEnabled:     true,
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 200, Y: 100, Width: 80, Height: 30},
	})

	em.Add(&dom.Element{
		Index:         2,
		TagName:       "input",
		Type:          "text",
		Placeholder:   "Enter your email",
		IsVisible:     true,
		IsEnabled:     true,
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 100, Y: 50, Width: 200, Height: 30},
	})

	em.Add(&dom.Element{
		Index:         3,
		TagName:       "a",
		Text:          "Sign In",
		Href:          "https://example.com/login",
		IsVisible:     true,
		IsEnabled:     true,
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 300, Y: 10, Width: 60, Height: 20},
	})

	em.Add(&dom.Element{
		Index:         4,
		TagName:       "button",
		Role:          "button",
		Text:          "Login Now",
		IsVisible:     true,
		IsEnabled:     false, // Disabled
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 400, Y: 100, Width: 80, Height: 30},
	})

	em.Add(&dom.Element{
		Index:         5,
		TagName:       "div",
		Text:          "Hidden element",
		IsVisible:     false, // Not visible
		IsEnabled:     true,
		IsInteractive: true,
		BoundingBox:   dom.BoundingBox{X: 0, Y: 0, Width: 0, Height: 0},
	})

	return em
}

func TestNewElementMatcher(t *testing.T) {
	m := NewElementMatcher()
	if m == nil {
		t.Fatal("NewElementMatcher returned nil")
	}
	if m.wordSimilarityCache == nil {
		t.Error("wordSimilarityCache should be initialized")
	}
}

func TestFindElements_ExactMatch(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		Text:     "Submit Form",
		Strategy: StrategyExact,
	})

	if len(matches) == 0 {
		t.Fatal("Expected at least one match")
	}

	if matches[0].Element.Text != "Submit Form" {
		t.Errorf("Expected 'Submit Form', got %q", matches[0].Element.Text)
	}
}

func TestFindElements_FuzzyMatch(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		Text:     "submit",
		Strategy: StrategyFuzzy,
	})

	if len(matches) == 0 {
		t.Fatal("Expected at least one match for fuzzy 'submit'")
	}

	// Should find "Submit Form" button
	found := false
	for _, match := range matches {
		if match.Element.Text == "Submit Form" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Submit Form' with fuzzy match")
	}
}

func TestFindElements_TagFilter(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		TagName: "button",
	})

	for _, match := range matches {
		if match.Element.TagName != "button" {
			t.Errorf("Expected only buttons, got %q", match.Element.TagName)
		}
	}
}

func TestFindElements_RoleFilter(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		Role: "button",
	})

	for _, match := range matches {
		if match.Element.Role != "button" {
			t.Errorf("Expected role 'button', got %q", match.Element.Role)
		}
	}
}

func TestFindElements_MustBeVisible(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		MustBeVisible: true,
	})

	for _, match := range matches {
		if !match.Element.IsVisible {
			t.Error("Expected only visible elements")
		}
	}
}

func TestFindElements_MustBeEnabled(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindElements(elements, ElementQuery{
		Text:          "Login",
		MustBeEnabled: true,
	})

	for _, match := range matches {
		if !match.Element.IsEnabled {
			t.Error("Expected only enabled elements")
		}
	}
}

func TestFindBestMatch(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	match := m.FindBestMatch(elements, ElementQuery{
		Text:     "Cancel",
		Strategy: StrategyExact,
	})

	if match == nil {
		t.Fatal("Expected a match")
	}
	if match.Element.Text != "Cancel" {
		t.Errorf("Expected 'Cancel', got %q", match.Element.Text)
	}
}

func TestFindBestMatch_NoMatch(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	match := m.FindBestMatch(elements, ElementQuery{
		Text:     "NonexistentText",
		Strategy: StrategyExact,
	})

	if match != nil {
		t.Error("Expected no match")
	}
}

func TestFindByText(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindByText(elements, "Sign In")

	if len(matches) == 0 {
		t.Fatal("Expected at least one match")
	}

	found := false
	for _, match := range matches {
		if match.Element.Text == "Sign In" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'Sign In' link")
	}
}

func TestFindByRole(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	matches := m.FindByRole(elements, "button")

	if len(matches) == 0 {
		t.Fatal("Expected button matches")
	}

	for _, match := range matches {
		if match.Element.Role != "button" {
			t.Errorf("Expected role 'button', got %q", match.Element.Role)
		}
	}
}

func TestFindButton(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	match := m.FindButton(elements, "Submit")

	if match == nil {
		t.Fatal("Expected a match")
	}
	if match.Element.TagName != "button" && match.Element.Role != "button" {
		t.Error("Expected a button element")
	}
}

func TestFindInput(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	match := m.FindInput(elements, "email")

	if match == nil {
		t.Fatal("Expected a match")
	}
	if match.Element.TagName != "input" {
		t.Errorf("Expected input, got %q", match.Element.TagName)
	}
}

func TestFindLink(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	match := m.FindLink(elements, "login")

	if match == nil {
		t.Fatal("Expected a match")
	}
	if match.Element.TagName != "a" {
		t.Errorf("Expected anchor, got %q", match.Element.TagName)
	}
}

func TestFindNear(t *testing.T) {
	m := NewElementMatcher()
	elements := createTestElementMap()

	// Find elements near the email input (index 2)
	matches := m.FindNear(elements, 2, ElementQuery{
		TagName:       "button",
		MustBeVisible: true,
	})

	if len(matches) == 0 {
		t.Fatal("Expected nearby buttons")
	}

	// The Submit and Cancel buttons should be close to the email input
	foundNearby := false
	for _, match := range matches {
		if match.Element.Text == "Submit Form" || match.Element.Text == "Cancel" {
			foundNearby = true
			break
		}
	}
	if !foundNearby {
		t.Error("Expected to find nearby buttons")
	}
}

func TestTextSimilarity(t *testing.T) {
	m := NewElementMatcher()

	tests := []struct {
		a, b     string
		minScore float64
	}{
		{"submit", "submit", 1.0},
		{"Submit Form", "submit", 0.3},
		{"", "test", 0.0},
		{"login button", "sign in button", 0.2},
	}

	for _, tt := range tests {
		score := m.textSimilarity(tt.a, tt.b)
		if score < tt.minScore {
			t.Errorf("textSimilarity(%q, %q) = %f, expected >= %f", tt.a, tt.b, score, tt.minScore)
		}
	}
}

func TestSemanticMatch(t *testing.T) {
	m := NewElementMatcher()

	tests := []struct {
		element, query string
		expected       bool
	}{
		{"Submit", "send", true},
		{"Sign In", "login", true},
		{"Cancel", "close", true},
		{"Search", "find", true},
		{"Random Text", "login", false},
	}

	for _, tt := range tests {
		result := m.semanticMatch(tt.element, tt.query)
		if result != tt.expected {
			t.Errorf("semanticMatch(%q, %q) = %v, expected %v", tt.element, tt.query, result, tt.expected)
		}
	}
}

func TestElementDistance(t *testing.T) {
	m := NewElementMatcher()

	el1 := &dom.Element{
		BoundingBox: dom.BoundingBox{X: 0, Y: 0, Width: 100, Height: 100},
	}
	el2 := &dom.Element{
		BoundingBox: dom.BoundingBox{X: 100, Y: 0, Width: 100, Height: 100},
	}

	distance := m.elementDistance(el1, el2)

	// Centers are at (50,50) and (150,50), distance should be 100
	if distance < 99 || distance > 101 {
		t.Errorf("elementDistance = %f, expected ~100", distance)
	}
}

func TestElementSelector_GenerateSelector(t *testing.T) {
	s := NewElementSelector()

	tests := []struct {
		element  *dom.Element
		expected string
	}{
		{
			&dom.Element{TagName: "button", Name: "submit_btn"},
			"button#submit_btn",
		},
		{
			&dom.Element{TagName: "button", Role: "button"},
			"button[role=\"button\"]",
		},
		{
			&dom.Element{TagName: "input", Type: "text"},
			"input[type=\"text\"]",
		},
	}

	for _, tt := range tests {
		result := s.GenerateSelector(tt.element)
		if result != tt.expected {
			t.Errorf("GenerateSelector() = %q, expected %q", result, tt.expected)
		}
	}
}

func TestElementSelector_GenerateXPath(t *testing.T) {
	s := NewElementSelector()

	el := &dom.Element{
		TagName: "button",
		Text:    "Submit",
		Role:    "button",
	}

	xpath := s.GenerateXPath(el)

	if xpath == "" {
		t.Error("Expected non-empty XPath")
	}
	if xpath[:8] != "//button" {
		t.Errorf("XPath should start with //button, got %q", xpath)
	}
}

func TestElementGrouper_FindFormGroups(t *testing.T) {
	g := NewElementGrouper()
	elements := createTestElementMap()

	groups := g.FindFormGroups(elements)

	if len(groups) == 0 {
		t.Fatal("Expected form groups")
	}

	formGroup := groups[0]
	if formGroup.Type != "form" {
		t.Errorf("Expected type 'form', got %q", formGroup.Type)
	}

	// Should contain the email input
	hasInput := false
	for _, el := range formGroup.Elements {
		if el.TagName == "input" {
			hasInput = true
			break
		}
	}
	if !hasInput {
		t.Error("Form group should contain input element")
	}
}

func TestElementGrouper_FindButtonGroups(t *testing.T) {
	g := NewElementGrouper()
	elements := createTestElementMap()

	groups := g.FindButtonGroups(elements)

	if len(groups) == 0 {
		t.Fatal("Expected button groups")
	}

	buttonGroup := groups[0]
	if buttonGroup.Type != "buttons" {
		t.Errorf("Expected type 'buttons', got %q", buttonGroup.Type)
	}

	if len(buttonGroup.Elements) < 2 {
		t.Error("Expected multiple buttons in group")
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"submit_btn", true},
		{"myButton", true},
		{"btn-primary", true},
		{"123abc", false},
		{"-invalid", false},
		{"", false},
		{"has space", false},
	}

	for _, tt := range tests {
		result := isValidIdentifier(tt.input)
		if result != tt.expected {
			t.Errorf("isValidIdentifier(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},
		{100, 10},
		{-1, 0},
	}

	for _, tt := range tests {
		result := sqrt(tt.input)
		diff := result - tt.expected
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.0001 {
			t.Errorf("sqrt(%f) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

// Benchmarks

func BenchmarkFindElements_Fuzzy(b *testing.B) {
	m := NewElementMatcher()
	elements := createTestElementMap()
	query := ElementQuery{
		Text:     "Submit",
		Strategy: StrategyFuzzy,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.FindElements(elements, query)
	}
}

func BenchmarkTextSimilarity(b *testing.B) {
	m := NewElementMatcher()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.textSimilarity("Submit Form Button", "submit button")
	}
}
