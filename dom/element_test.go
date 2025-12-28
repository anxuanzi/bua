package dom

import (
	"testing"
)

func TestNewElementMap(t *testing.T) {
	em := NewElementMap()
	if em == nil {
		t.Fatal("NewElementMap() returned nil")
	}
	if em.Elements == nil {
		t.Error("Elements slice is nil")
	}
	if em.indexMap == nil {
		t.Error("indexMap is nil")
	}
	if em.Count() != 0 {
		t.Errorf("Count() = %d, want 0", em.Count())
	}
}

func TestElementMap_Add(t *testing.T) {
	em := NewElementMap()

	el := &Element{
		Index:   0,
		TagName: "button",
		Text:    "Click me",
	}
	em.Add(el)

	if em.Count() != 1 {
		t.Errorf("Count() = %d, want 1", em.Count())
	}

	// Add more elements
	em.Add(&Element{Index: 1, TagName: "input"})
	em.Add(&Element{Index: 2, TagName: "a"})

	if em.Count() != 3 {
		t.Errorf("Count() = %d, want 3", em.Count())
	}
}

func TestElementMap_ByIndex(t *testing.T) {
	em := NewElementMap()
	em.Add(&Element{Index: 0, TagName: "button", Text: "Button 0"})
	em.Add(&Element{Index: 1, TagName: "input", Text: "Input 1"})
	em.Add(&Element{Index: 2, TagName: "a", Text: "Link 2"})

	// Test found case
	el, ok := em.ByIndex(1)
	if !ok {
		t.Error("ByIndex(1) returned false")
	}
	if el.TagName != "input" {
		t.Errorf("ByIndex(1).TagName = %s, want input", el.TagName)
	}

	// Test not found case
	_, ok = em.ByIndex(99)
	if ok {
		t.Error("ByIndex(99) should return false")
	}
}

func TestElementMap_InteractiveElements(t *testing.T) {
	em := NewElementMap()
	em.Add(&Element{Index: 0, TagName: "button", IsInteractive: true, IsVisible: true})
	em.Add(&Element{Index: 1, TagName: "div", IsInteractive: false, IsVisible: true})
	em.Add(&Element{Index: 2, TagName: "input", IsInteractive: true, IsVisible: false})
	em.Add(&Element{Index: 3, TagName: "a", IsInteractive: true, IsVisible: true})

	interactive := em.InteractiveElements()
	if len(interactive) != 2 {
		t.Errorf("InteractiveElements() returned %d elements, want 2", len(interactive))
	}
}

func TestElementMap_ToTokenString(t *testing.T) {
	em := NewElementMap()
	em.PageTitle = "Test Page"
	em.PageURL = "https://example.com"
	em.Add(&Element{
		Index:     0,
		TagName:   "button",
		Text:      "Click me",
		IsVisible: true,
	})
	em.Add(&Element{
		Index:     1,
		TagName:   "input",
		Type:      "text",
		IsVisible: true,
	})
	em.Add(&Element{
		Index:     2,
		TagName:   "a",
		Href:      "https://example.com/link",
		IsVisible: false, // Should be skipped
	})

	str := em.ToTokenString()
	if str == "" {
		t.Error("ToTokenString() returned empty string")
	}

	// Check that visible elements are included
	if !contains(str, "[0]") {
		t.Error("ToTokenString() missing element 0")
	}
	if !contains(str, "[1]") {
		t.Error("ToTokenString() missing element 1")
	}
}

func TestBoundingBox(t *testing.T) {
	bb := BoundingBox{
		X:      100,
		Y:      200,
		Width:  50,
		Height: 30,
	}

	if bb.X != 100 {
		t.Errorf("X = %f, want 100", bb.X)
	}
	if bb.Y != 200 {
		t.Errorf("Y = %f, want 200", bb.Y)
	}
	if bb.Width != 50 {
		t.Errorf("Width = %f, want 50", bb.Width)
	}
	if bb.Height != 30 {
		t.Errorf("Height = %f, want 30", bb.Height)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a long string", 10, "this is..."},
		{"", 10, ""},
		{"test", 4, "test"},
		{"testing", 5, "te..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
