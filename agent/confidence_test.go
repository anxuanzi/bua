package agent

import (
	"math"
	"testing"

	"github.com/anxuanzi/bua-go/dom"
)

func TestGetLevel(t *testing.T) {
	tests := []struct {
		value    float64
		expected ConfidenceLevel
	}{
		{1.0, ConfidenceVeryHigh},
		{0.95, ConfidenceVeryHigh},
		{0.9, ConfidenceVeryHigh},
		{0.85, ConfidenceHigh},
		{0.7, ConfidenceHigh},
		{0.65, ConfidenceMedium},
		{0.5, ConfidenceMedium},
		{0.45, ConfidenceLow},
		{0.3, ConfidenceLow},
		{0.25, ConfidenceVeryLow},
		{0.0, ConfidenceVeryLow},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := GetLevel(tt.value)
			if result != tt.expected {
				t.Errorf("GetLevel(%f) = %s, want %s", tt.value, result, tt.expected)
			}
		})
	}
}

func TestNewConfidenceScore(t *testing.T) {
	t.Run("normal value", func(t *testing.T) {
		score := NewConfidenceScore(0.8, "test explanation")
		if score.Value != 0.8 {
			t.Errorf("Value = %f, want 0.8", score.Value)
		}
		if score.Level != ConfidenceHigh {
			t.Errorf("Level = %s, want high", score.Level)
		}
		if score.Explanation != "test explanation" {
			t.Errorf("Explanation = %s", score.Explanation)
		}
	})

	t.Run("clamp low", func(t *testing.T) {
		score := NewConfidenceScore(-0.5, "negative value")
		if score.Value != 0.0 {
			t.Errorf("Value = %f, want 0.0", score.Value)
		}
	})

	t.Run("clamp high", func(t *testing.T) {
		score := NewConfidenceScore(1.5, "over value")
		if score.Value != 1.0 {
			t.Errorf("Value = %f, want 1.0", score.Value)
		}
	})
}

func TestConfidenceScore_AddFactor(t *testing.T) {
	score := NewConfidenceScore(0.5, "test")
	score.AddFactor("factor1", 0.5, 0.8, "reason1")
	score.AddFactor("factor2", 0.5, 0.6, "reason2")

	if len(score.Factors) != 2 {
		t.Errorf("Expected 2 factors, got %d", len(score.Factors))
	}
	if score.Factors[0].Name != "factor1" {
		t.Errorf("Factor name = %s", score.Factors[0].Name)
	}
}

func TestConfidenceScore_RecalculateFromFactors(t *testing.T) {
	t.Run("weighted average", func(t *testing.T) {
		score := NewConfidenceScore(0.0, "test")
		score.AddFactor("f1", 1.0, 0.8, "r1") // weight 1, score 0.8
		score.AddFactor("f2", 1.0, 0.6, "r2") // weight 1, score 0.6
		score.RecalculateFromFactors()

		// Expected: (1.0*0.8 + 1.0*0.6) / (1.0 + 1.0) = 0.7
		expected := 0.7
		if score.Value != expected {
			t.Errorf("Value = %f, want %f", score.Value, expected)
		}
		if score.Level != ConfidenceHigh {
			t.Errorf("Level = %s, want high", score.Level)
		}
	})

	t.Run("unequal weights", func(t *testing.T) {
		score := NewConfidenceScore(0.0, "test")
		score.AddFactor("f1", 3.0, 1.0, "r1") // weight 3, score 1.0
		score.AddFactor("f2", 1.0, 0.0, "r2") // weight 1, score 0.0
		score.RecalculateFromFactors()

		// Expected: (3.0*1.0 + 1.0*0.0) / (3.0 + 1.0) = 0.75
		expected := 0.75
		if score.Value != expected {
			t.Errorf("Value = %f, want %f", score.Value, expected)
		}
	})

	t.Run("empty factors", func(t *testing.T) {
		score := NewConfidenceScore(0.5, "test")
		score.RecalculateFromFactors()
		if score.Value != 0.5 {
			t.Errorf("Value should remain 0.5 with no factors")
		}
	})
}

func TestNewConfidenceTracker(t *testing.T) {
	tracker := NewConfidenceTracker()
	if tracker == nil {
		t.Fatal("NewConfidenceTracker() returned nil")
	}
	if tracker.confidenceDecay != 0.02 {
		t.Errorf("confidenceDecay = %f, want 0.02", tracker.confidenceDecay)
	}
	if len(tracker.actionHistory) != 0 {
		t.Error("actionHistory should be empty")
	}
}

func TestCalculateElementConfidence(t *testing.T) {
	t.Run("button element", func(t *testing.T) {
		element := &dom.Element{
			Index:   1,
			TagName: "button",
			Text:    "Click Me",
			Role:    "button",
			BoundingBox: dom.BoundingBox{
				X:      100,
				Y:      200,
				Width:  80,
				Height: 30,
			},
		}

		ec := CalculateElementConfidence(element, nil, "click")
		if ec.OverallScore == nil {
			t.Fatal("OverallScore is nil")
		}
		if ec.OverallScore.Value < 0.7 {
			t.Errorf("Button confidence too low: %f", ec.OverallScore.Value)
		}
		if ec.TypeMatch < 0.9 {
			t.Errorf("TypeMatch too low for button: %f", ec.TypeMatch)
		}
	})

	t.Run("input element for type", func(t *testing.T) {
		element := &dom.Element{
			Index:       2,
			TagName:     "input",
			Placeholder: "Enter email",
			BoundingBox: dom.BoundingBox{
				X:      100,
				Y:      200,
				Width:  200,
				Height: 30,
			},
		}

		ec := CalculateElementConfidence(element, nil, "type")
		if ec.TypeMatch < 0.9 {
			t.Errorf("TypeMatch too low for input: %f", ec.TypeMatch)
		}
	})

	t.Run("element without bounding box", func(t *testing.T) {
		element := &dom.Element{
			Index:   3,
			TagName: "div",
			Text:    "Some text",
		}

		ec := CalculateElementConfidence(element, nil, "click")
		// Should still work but with lower visibility score
		if ec.VisualMatch > 0.7 {
			t.Errorf("VisualMatch should be lower without bounding box: %f", ec.VisualMatch)
		}
	})

	t.Run("uniqueness with similar elements", func(t *testing.T) {
		element := &dom.Element{
			Index:   1,
			TagName: "button",
			Text:    "Submit",
		}

		elements := &dom.ElementMap{}
		// This is a simplified test - actual ElementMap would need proper initialization

		ec := CalculateElementConfidence(element, elements, "click")
		// Should have some uniqueness score
		if ec.UniquenessScore <= 0 {
			t.Error("UniquenessScore should be positive")
		}
	})
}

func TestConfidenceTracker_CalculateActionConfidence(t *testing.T) {
	tracker := NewConfidenceTracker()

	t.Run("verified action no retries", func(t *testing.T) {
		ac := tracker.CalculateActionConfidence("click", true, 0, nil, 0.5)
		if !ac.Verified {
			t.Error("Verified should be true")
		}
		if ac.OverallScore.Value < 0.7 {
			t.Errorf("Verified action should have high confidence: %f", ac.OverallScore.Value)
		}
	})

	t.Run("unverified action", func(t *testing.T) {
		ac := tracker.CalculateActionConfidence("click", false, 0, nil, 0.1)
		if ac.Verified {
			t.Error("Verified should be false")
		}
		if ac.OverallScore.Value > 0.6 {
			t.Errorf("Unverified action should have lower confidence: %f", ac.OverallScore.Value)
		}
	})

	t.Run("action with retries", func(t *testing.T) {
		acNoRetry := tracker.CalculateActionConfidence("click", true, 0, nil, 0.5)
		acWithRetry := tracker.CalculateActionConfidence("click", true, 2, nil, 0.5)

		if acWithRetry.OverallScore.Value >= acNoRetry.OverallScore.Value {
			t.Error("Action with retries should have lower confidence")
		}
	})

	t.Run("action with element confidence", func(t *testing.T) {
		elemConf := &ElementConfidence{
			OverallScore: NewConfidenceScore(0.9, "high confidence element"),
		}
		ac := tracker.CalculateActionConfidence("click", true, 0, elemConf, 0.5)

		if ac.TargetConfidence == nil {
			t.Error("TargetConfidence should be set")
		}
		if ac.TargetConfidence.Value != 0.9 {
			t.Errorf("TargetConfidence.Value = %f, want 0.9", ac.TargetConfidence.Value)
		}
	})
}

func TestConfidenceTracker_RecordAction(t *testing.T) {
	tracker := NewConfidenceTracker()

	ac := &ActionConfidence{
		ActionType:   "click",
		Verified:     true,
		OverallScore: NewConfidenceScore(0.8, "test"),
	}

	tracker.RecordAction(ac)

	if len(tracker.actionHistory) != 1 {
		t.Errorf("Expected 1 action in history, got %d", len(tracker.actionHistory))
	}
}

func TestConfidenceTracker_GetTaskConfidence(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		tracker := NewConfidenceTracker()
		tc := tracker.GetTaskConfidence()

		if tc.TotalSteps != 0 {
			t.Errorf("TotalSteps = %d, want 0", tc.TotalSteps)
		}
		if tc.OverallScore.Value != 0.0 {
			t.Errorf("OverallScore = %f, want 0.0", tc.OverallScore.Value)
		}
	})

	t.Run("with actions", func(t *testing.T) {
		tracker := NewConfidenceTracker()

		// Add some actions
		tracker.RecordAction(&ActionConfidence{
			ActionType:   "click",
			Verified:     true,
			OverallScore: NewConfidenceScore(0.9, "high"),
		})
		tracker.RecordAction(&ActionConfidence{
			ActionType:   "type",
			Verified:     true,
			OverallScore: NewConfidenceScore(0.8, "good"),
		})
		tracker.RecordAction(&ActionConfidence{
			ActionType:   "click",
			Verified:     false,
			OverallScore: NewConfidenceScore(0.4, "low"),
		})

		tc := tracker.GetTaskConfidence()

		if tc.TotalSteps != 3 {
			t.Errorf("TotalSteps = %d, want 3", tc.TotalSteps)
		}
		if tc.FailedSteps != 1 {
			t.Errorf("FailedSteps = %d, want 1", tc.FailedSteps)
		}

		// Average should be (0.9 + 0.8 + 0.4) / 3 = 0.7
		expectedAvg := 0.7
		if math.Abs(tc.AverageConfidence-expectedAvg) > 0.0001 {
			t.Errorf("AverageConfidence = %f, want %f", tc.AverageConfidence, expectedAvg)
		}

		if tc.MinConfidence != 0.4 {
			t.Errorf("MinConfidence = %f, want 0.4", tc.MinConfidence)
		}

		// Success rate = 2/3
		expectedRate := 2.0 / 3.0
		if tc.SuccessRate != expectedRate {
			t.Errorf("SuccessRate = %f, want %f", tc.SuccessRate, expectedRate)
		}
	})

	t.Run("all verified", func(t *testing.T) {
		tracker := NewConfidenceTracker()

		tracker.RecordAction(&ActionConfidence{
			ActionType:   "navigate",
			Verified:     true,
			OverallScore: NewConfidenceScore(1.0, "perfect"),
		})
		tracker.RecordAction(&ActionConfidence{
			ActionType:   "click",
			Verified:     true,
			OverallScore: NewConfidenceScore(0.9, "high"),
		})

		tc := tracker.GetTaskConfidence()

		if tc.SuccessRate != 1.0 {
			t.Errorf("SuccessRate = %f, want 1.0", tc.SuccessRate)
		}
		if tc.FailedSteps != 0 {
			t.Errorf("FailedSteps = %d, want 0", tc.FailedSteps)
		}
	})
}

func TestCalculatePageStateDiff(t *testing.T) {
	t.Run("nil snapshots", func(t *testing.T) {
		diff := CalculatePageStateDiff(nil, nil)
		if diff != 0.5 {
			t.Errorf("Diff = %f, want 0.5 for nil snapshots", diff)
		}
	})

	t.Run("same state", func(t *testing.T) {
		pre := &PageSnapshot{
			URL:          "https://example.com",
			Title:        "Example",
			ElementCount: 10,
		}
		post := &PageSnapshot{
			URL:          "https://example.com",
			Title:        "Example",
			ElementCount: 10,
		}

		diff := CalculatePageStateDiff(pre, post)
		if diff != 0.0 {
			t.Errorf("Diff = %f, want 0.0 for identical states", diff)
		}
	})

	t.Run("url change", func(t *testing.T) {
		pre := &PageSnapshot{
			URL:          "https://example.com",
			Title:        "Example",
			ElementCount: 10,
		}
		post := &PageSnapshot{
			URL:          "https://other.com",
			Title:        "Example",
			ElementCount: 10,
		}

		diff := CalculatePageStateDiff(pre, post)
		if diff < 0.3 {
			t.Errorf("Diff = %f, should be higher for URL change", diff)
		}
	})

	t.Run("element count change", func(t *testing.T) {
		pre := &PageSnapshot{
			URL:          "https://example.com",
			Title:        "Example",
			ElementCount: 10,
		}
		post := &PageSnapshot{
			URL:          "https://example.com",
			Title:        "Example",
			ElementCount: 20,
		}

		diff := CalculatePageStateDiff(pre, post)
		if diff <= 0.0 {
			t.Errorf("Diff = %f, should be > 0 for element count change", diff)
		}
	})
}

func TestAdjustConfidenceForContext(t *testing.T) {
	t.Run("normal step", func(t *testing.T) {
		adjusted := AdjustConfidenceForContext(0.8, 5, 10, 0)
		if adjusted != 0.8 {
			t.Errorf("Adjusted = %f, want 0.8 for normal step", adjusted)
		}
	})

	t.Run("many steps decay", func(t *testing.T) {
		// Use totalSteps=50 so progress (50%) doesn't trigger completion bonus
		adjusted := AdjustConfidenceForContext(0.8, 25, 50, 0)
		if adjusted >= 0.8 {
			t.Errorf("Adjusted = %f, should be less than 0.8 for many steps", adjusted)
		}
	})

	t.Run("consecutive failures", func(t *testing.T) {
		adjusted := AdjustConfidenceForContext(0.8, 5, 10, 2)
		if adjusted >= 0.8 {
			t.Errorf("Adjusted = %f, should be less than 0.8 with failures", adjusted)
		}
	})

	t.Run("near completion bonus", func(t *testing.T) {
		adjusted := AdjustConfidenceForContext(0.8, 9, 10, 0)
		if adjusted <= 0.8 {
			t.Errorf("Adjusted = %f, should have completion bonus", adjusted)
		}
	})

	t.Run("clamp to valid range", func(t *testing.T) {
		// Try to go below 0
		adjusted := AdjustConfidenceForContext(0.1, 30, 30, 10)
		if adjusted < 0.0 || adjusted > 1.0 {
			t.Errorf("Adjusted = %f, should be clamped to [0, 1]", adjusted)
		}
	})
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a longer string", 10, "this is..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// Benchmarks

func BenchmarkCalculateElementConfidence(b *testing.B) {
	element := &dom.Element{
		Index:   1,
		TagName: "button",
		Text:    "Click Me",
		Role:    "button",
		BoundingBox: dom.BoundingBox{
			X:      100,
			Y:      200,
			Width:  80,
			Height: 30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateElementConfidence(element, nil, "click")
	}
}

func BenchmarkGetTaskConfidence(b *testing.B) {
	tracker := NewConfidenceTracker()

	// Add some actions
	for i := 0; i < 20; i++ {
		tracker.RecordAction(&ActionConfidence{
			ActionType:   "click",
			Verified:     i%3 != 0,
			OverallScore: NewConfidenceScore(0.7+float64(i%3)*0.1, "test"),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.GetTaskConfidence()
	}
}
