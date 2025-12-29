package agent

import (
	"testing"
	"time"

	"github.com/anxuanzi/bua-go/dom"
)

func TestDefaultVerificationConfig(t *testing.T) {
	cfg := DefaultVerificationConfig()

	if cfg == nil {
		t.Fatal("DefaultVerificationConfig() returned nil")
	}
	if !cfg.Enabled {
		t.Error("Enabled should be true by default")
	}
	if cfg.MaxRetries != 2 {
		t.Errorf("MaxRetries = %d, want 2", cfg.MaxRetries)
	}
	if cfg.RetryDelay != 500*time.Millisecond {
		t.Errorf("RetryDelay = %v, want 500ms", cfg.RetryDelay)
	}
	if cfg.StabilizationTime != 300*time.Millisecond {
		t.Errorf("StabilizationTime = %v, want 300ms", cfg.StabilizationTime)
	}
}

func TestNewVerifier(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		v := NewVerifier(nil, nil)
		if v == nil {
			t.Fatal("NewVerifier() returned nil")
		}
		if v.config == nil {
			t.Error("config should not be nil")
		}
		if !v.config.Enabled {
			t.Error("config should use defaults")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &VerificationConfig{
			Enabled:    false,
			MaxRetries: 5,
		}
		v := NewVerifier(nil, cfg)
		if v.config.Enabled {
			t.Error("Enabled should be false")
		}
		if v.config.MaxRetries != 5 {
			t.Errorf("MaxRetries = %d, want 5", v.config.MaxRetries)
		}
	})
}

func TestActionVerification(t *testing.T) {
	av := &ActionVerification{
		ActionType:      "click",
		ActionTarget:    5,
		ExpectedChange:  "button should trigger navigation",
		Verified:        true,
		VerificationMsg: "URL changed",
		Retries:         1,
	}

	if av.ActionType != "click" {
		t.Errorf("ActionType = %q, want 'click'", av.ActionType)
	}
	if av.ActionTarget != 5 {
		t.Errorf("ActionTarget = %d, want 5", av.ActionTarget)
	}
	if !av.Verified {
		t.Error("Verified should be true")
	}
	if av.Retries != 1 {
		t.Errorf("Retries = %d, want 1", av.Retries)
	}
}

func TestPageSnapshot(t *testing.T) {
	snapshot := &PageSnapshot{
		URL:          "https://example.com",
		Title:        "Example",
		ElementCount: 10,
		Elements:     make(map[int]*dom.Element),
		Timestamp:    time.Now(),
	}

	if snapshot.URL != "https://example.com" {
		t.Errorf("URL = %q", snapshot.URL)
	}
	if snapshot.Title != "Example" {
		t.Errorf("Title = %q", snapshot.Title)
	}
	if snapshot.ElementCount != 10 {
		t.Errorf("ElementCount = %d", snapshot.ElementCount)
	}
}

func TestVerifier_detectChanges(t *testing.T) {
	v := NewVerifier(nil, nil)

	t.Run("URL change", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://b.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		changes := v.detectChanges(pre, post)
		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changes))
		}
	})

	t.Run("title change", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "B", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		changes := v.detectChanges(pre, post)
		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changes))
		}
	})

	t.Run("element count change", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 10, Elements: make(map[int]*dom.Element)}

		changes := v.detectChanges(pre, post)
		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changes))
		}
	})

	t.Run("element text change", func(t *testing.T) {
		pre := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Text: "Hello"}},
		}
		post := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Text: "World"}},
		}

		changes := v.detectChanges(pre, post)
		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d: %v", len(changes), changes)
		}
	})

	t.Run("no changes", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		changes := v.detectChanges(pre, post)
		if len(changes) != 0 {
			t.Errorf("Expected 0 changes, got %d", len(changes))
		}
	})

	t.Run("new element", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 6,
			Elements: map[int]*dom.Element{99: {Index: 99, TagName: "div"}},
		}

		changes := v.detectChanges(pre, post)
		if len(changes) != 2 { // count change + new element
			t.Errorf("Expected 2 changes, got %d: %v", len(changes), changes)
		}
	})

	t.Run("removed element", func(t *testing.T) {
		pre := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{99: {Index: 99, TagName: "div"}},
		}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 4, Elements: make(map[int]*dom.Element)}

		changes := v.detectChanges(pre, post)
		if len(changes) != 2 { // count change + removed element
			t.Errorf("Expected 2 changes, got %d: %v", len(changes), changes)
		}
	})
}

func TestVerifier_VerifyClick(t *testing.T) {
	v := NewVerifier(nil, nil)

	t.Run("verified with changes", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://b.com", Title: "B", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyClick(nil, 1, pre, post)
		if !result.Verified {
			t.Error("Should be verified when changes detected")
		}
		if result.ActionType != "click" {
			t.Errorf("ActionType = %q, want 'click'", result.ActionType)
		}
	})

	t.Run("not verified without changes", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyClick(nil, 1, pre, post)
		if result.Verified {
			t.Error("Should not be verified when no changes detected")
		}
	})
}

func TestVerifier_VerifyType(t *testing.T) {
	v := NewVerifier(nil, nil)

	t.Run("verified when value contains typed text", func(t *testing.T) {
		pre := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Value: ""}},
		}
		post := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Value: "hello world"}},
		}

		result := v.VerifyType(nil, 1, "hello", pre, post)
		if !result.Verified {
			t.Errorf("Should be verified: %s", result.VerificationMsg)
		}
	})

	t.Run("not verified when value unchanged", func(t *testing.T) {
		pre := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Value: "test"}},
		}
		post := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Value: "test"}},
		}

		result := v.VerifyType(nil, 1, "new text", pre, post)
		if result.Verified {
			t.Error("Should not be verified when value unchanged")
		}
	})

	t.Run("verified when element disappears (form submit)", func(t *testing.T) {
		pre := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 5,
			Elements: map[int]*dom.Element{1: {Index: 1, Value: ""}},
		}
		post := &PageSnapshot{
			URL: "https://a.com", Title: "A", ElementCount: 4,
			Elements: make(map[int]*dom.Element),
		}

		result := v.VerifyType(nil, 1, "text", pre, post)
		if !result.Verified {
			t.Errorf("Should be verified when element disappears: %s", result.VerificationMsg)
		}
	})
}

func TestVerifier_VerifyNavigation(t *testing.T) {
	v := NewVerifier(nil, nil)

	t.Run("verified when URL matches target", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://example.com/page", Title: "B", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyNavigation(nil, "https://example.com/page", pre, post)
		if !result.Verified {
			t.Errorf("Should be verified: %s", result.VerificationMsg)
		}
	})

	t.Run("verified when URL changed to different target", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://other.com", Title: "B", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyNavigation(nil, "https://example.com", pre, post)
		if !result.Verified {
			t.Error("Should still be verified when URL changed, even if different target")
		}
	})

	t.Run("not verified when URL unchanged", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyNavigation(nil, "https://example.com", pre, post)
		if result.Verified {
			t.Error("Should not be verified when URL unchanged")
		}
	})
}

func TestVerifier_VerifyScroll(t *testing.T) {
	v := NewVerifier(nil, nil)

	t.Run("verified with viewport changes", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 8, Elements: make(map[int]*dom.Element)}

		result := v.VerifyScroll(nil, pre, post)
		if !result.Verified {
			t.Errorf("Should be verified: %s", result.VerificationMsg)
		}
	})

	t.Run("verified even without changes (scroll at edge)", func(t *testing.T) {
		pre := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}
		post := &PageSnapshot{URL: "https://a.com", Title: "A", ElementCount: 5, Elements: make(map[int]*dom.Element)}

		result := v.VerifyScroll(nil, pre, post)
		if !result.Verified {
			t.Error("Scroll should always be verified (might be at edge of page)")
		}
	})
}

func TestVerifier_ShouldRetry(t *testing.T) {
	v := NewVerifier(nil, &VerificationConfig{MaxRetries: 2})

	t.Run("should not retry if verified", func(t *testing.T) {
		av := &ActionVerification{Verified: true, Retries: 0}
		if v.ShouldRetry(av) {
			t.Error("Should not retry verified action")
		}
	})

	t.Run("should retry if not verified and retries available", func(t *testing.T) {
		av := &ActionVerification{Verified: false, Retries: 0}
		if !v.ShouldRetry(av) {
			t.Error("Should retry unverified action with retries available")
		}
	})

	t.Run("should retry until max", func(t *testing.T) {
		av := &ActionVerification{Verified: false, Retries: 1}
		if !v.ShouldRetry(av) {
			t.Error("Should retry when retries < maxRetries")
		}

		av.Retries = 2
		if v.ShouldRetry(av) {
			t.Error("Should not retry when retries >= maxRetries")
		}
	})
}

// Benchmarks

func BenchmarkDetectChanges(b *testing.B) {
	v := NewVerifier(nil, nil)
	pre := &PageSnapshot{
		URL: "https://a.com", Title: "A", ElementCount: 100,
		Elements: make(map[int]*dom.Element),
	}
	post := &PageSnapshot{
		URL: "https://b.com", Title: "B", ElementCount: 105,
		Elements: make(map[int]*dom.Element),
	}

	// Add some elements
	for i := 0; i < 50; i++ {
		pre.Elements[i] = &dom.Element{Index: i, TagName: "div", Text: "text"}
		post.Elements[i] = &dom.Element{Index: i, TagName: "div", Text: "changed"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.detectChanges(pre, post)
	}
}
