// Package agent provides confidence scoring capabilities for browser automation.
package agent

import (
	"math"
	"time"

	"github.com/anxuanzi/bua-go/dom"
)

// ConfidenceLevel represents categorical confidence levels.
type ConfidenceLevel string

const (
	ConfidenceVeryHigh ConfidenceLevel = "very_high" // 0.9-1.0
	ConfidenceHigh     ConfidenceLevel = "high"      // 0.7-0.9
	ConfidenceMedium   ConfidenceLevel = "medium"    // 0.5-0.7
	ConfidenceLow      ConfidenceLevel = "low"       // 0.3-0.5
	ConfidenceVeryLow  ConfidenceLevel = "very_low"  // 0.0-0.3
)

// ConfidenceScore represents a confidence value with metadata.
type ConfidenceScore struct {
	Value       float64            `json:"value"`       // 0.0 to 1.0
	Level       ConfidenceLevel    `json:"level"`       // categorical level
	Factors     []ConfidenceFactor `json:"factors"`     // contributing factors
	Explanation string             `json:"explanation"` // human-readable explanation
}

// ConfidenceFactor represents a factor contributing to confidence.
type ConfidenceFactor struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"` // how much this factor contributes
	Score  float64 `json:"score"`  // the factor's individual score
	Reason string  `json:"reason"` // why this score was given
}

// ActionConfidence tracks confidence for a specific action.
type ActionConfidence struct {
	ActionType       string           `json:"action_type"`
	PreConfidence    float64          `json:"pre_confidence"`  // before action
	PostConfidence   float64          `json:"post_confidence"` // after action
	Verified         bool             `json:"verified"`
	RetryCount       int              `json:"retry_count"`
	TargetConfidence *ConfidenceScore `json:"target_confidence,omitempty"` // element targeting
	OverallScore     *ConfidenceScore `json:"overall_score"`
}

// ElementConfidence scores how confident we are in element targeting.
type ElementConfidence struct {
	ElementIndex    int              `json:"element_index"`
	OverallScore    *ConfidenceScore `json:"overall_score"`
	VisualMatch     float64          `json:"visual_match"`     // 0-1, how well it visually matches
	TextMatch       float64          `json:"text_match"`       // 0-1, text/label matching
	PositionMatch   float64          `json:"position_match"`   // 0-1, position reasonableness
	TypeMatch       float64          `json:"type_match"`       // 0-1, element type appropriateness
	UniquenessScore float64          `json:"uniqueness_score"` // 0-1, how unique is this element
}

// TaskConfidence aggregates confidence across all steps.
type TaskConfidence struct {
	OverallScore      *ConfidenceScore   `json:"overall_score"`
	StepConfidences   []ActionConfidence `json:"step_confidences"`
	SuccessRate       float64            `json:"success_rate"`       // verified steps / total steps
	AverageConfidence float64            `json:"average_confidence"` // mean of all step confidences
	MinConfidence     float64            `json:"min_confidence"`     // lowest confidence step
	FailedSteps       int                `json:"failed_steps"`       // count of failed steps
	TotalSteps        int                `json:"total_steps"`
}

// ConfidenceTracker tracks confidence throughout task execution.
type ConfidenceTracker struct {
	taskStart       time.Time
	actionHistory   []ActionConfidence
	currentTask     *TaskConfidence
	confidenceDecay float64 // how much confidence decays over time/steps
}

// NewConfidenceTracker creates a new confidence tracker.
func NewConfidenceTracker() *ConfidenceTracker {
	return &ConfidenceTracker{
		taskStart:       time.Now(),
		actionHistory:   make([]ActionConfidence, 0),
		confidenceDecay: 0.02, // 2% decay per step by default
	}
}

// GetLevel returns the categorical confidence level for a value.
func GetLevel(value float64) ConfidenceLevel {
	switch {
	case value >= 0.9:
		return ConfidenceVeryHigh
	case value >= 0.7:
		return ConfidenceHigh
	case value >= 0.5:
		return ConfidenceMedium
	case value >= 0.3:
		return ConfidenceLow
	default:
		return ConfidenceVeryLow
	}
}

// NewConfidenceScore creates a confidence score from a value.
func NewConfidenceScore(value float64, explanation string) *ConfidenceScore {
	// Clamp value to valid range
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	return &ConfidenceScore{
		Value:       value,
		Level:       GetLevel(value),
		Factors:     make([]ConfidenceFactor, 0),
		Explanation: explanation,
	}
}

// AddFactor adds a contributing factor to the confidence score.
func (c *ConfidenceScore) AddFactor(name string, weight, score float64, reason string) {
	c.Factors = append(c.Factors, ConfidenceFactor{
		Name:   name,
		Weight: weight,
		Score:  score,
		Reason: reason,
	})
}

// RecalculateFromFactors recalculates the value from weighted factors.
func (c *ConfidenceScore) RecalculateFromFactors() {
	if len(c.Factors) == 0 {
		return
	}

	var totalWeight float64
	var weightedSum float64

	for _, f := range c.Factors {
		totalWeight += f.Weight
		weightedSum += f.Weight * f.Score
	}

	if totalWeight > 0 {
		c.Value = weightedSum / totalWeight
		c.Level = GetLevel(c.Value)
	}
}

// CalculateElementConfidence calculates confidence for targeting a specific element.
func CalculateElementConfidence(element *dom.Element, allElements *dom.ElementMap, expectedType string) *ElementConfidence {
	ec := &ElementConfidence{
		ElementIndex: element.Index,
	}

	score := NewConfidenceScore(1.0, "Element targeting confidence")

	// Type match - is this the right type of element?
	typeScore := calculateTypeMatch(element, expectedType)
	ec.TypeMatch = typeScore
	score.AddFactor("type_match", 0.3, typeScore, getTypeMatchReason(element, expectedType))

	// Text match - does the element have meaningful text?
	textScore := calculateTextPresence(element)
	ec.TextMatch = textScore
	score.AddFactor("text_presence", 0.2, textScore, getTextPresenceReason(element))

	// Visibility/interactivity score
	visibilityScore := calculateVisibilityScore(element)
	ec.VisualMatch = visibilityScore
	score.AddFactor("visibility", 0.2, visibilityScore, getVisibilityReason(element))

	// Uniqueness score - how unique is this element?
	uniquenessScore := calculateUniquenessScore(element, allElements)
	ec.UniquenessScore = uniquenessScore
	score.AddFactor("uniqueness", 0.2, uniquenessScore, getUniquenessReason(uniquenessScore))

	// Position reasonableness
	positionScore := calculatePositionScore(element)
	ec.PositionMatch = positionScore
	score.AddFactor("position", 0.1, positionScore, "Position within viewport")

	score.RecalculateFromFactors()
	ec.OverallScore = score

	return ec
}

// hasBoundingBox checks if the element has a valid bounding box.
func hasBoundingBox(element *dom.Element) bool {
	return element.BoundingBox.Width > 0 || element.BoundingBox.Height > 0
}

// calculateTypeMatch scores how well the element type matches expectations.
func calculateTypeMatch(element *dom.Element, expectedType string) float64 {
	if expectedType == "" {
		return 0.8 // no specific type expected
	}

	tagName := element.TagName

	// Define type expectations
	typeMap := map[string][]string{
		"click":    {"button", "a", "input", "div", "span", "img", "li"},
		"type":     {"input", "textarea"},
		"select":   {"select", "input"},
		"checkbox": {"input"},
		"link":     {"a"},
		"button":   {"button", "input"},
	}

	expectedTags, ok := typeMap[expectedType]
	if !ok {
		return 0.7 // unknown expected type
	}

	for i, tag := range expectedTags {
		if tagName == tag {
			// First in list is most preferred
			return 1.0 - (float64(i) * 0.1)
		}
	}

	return 0.3 // element type doesn't match expectations
}

// getTypeMatchReason returns a reason for the type match score.
func getTypeMatchReason(element *dom.Element, expectedType string) string {
	if expectedType == "" {
		return "No specific element type expected"
	}
	return "Element tag: " + element.TagName + " for action: " + expectedType
}

// calculateTextPresence scores elements based on text content.
func calculateTextPresence(element *dom.Element) float64 {
	if element.Text != "" && len(element.Text) > 0 {
		// More text generally means better identifiability
		textLen := len(element.Text)
		if textLen > 50 {
			return 1.0
		}
		if textLen > 20 {
			return 0.9
		}
		if textLen > 5 {
			return 0.8
		}
		return 0.6
	}

	// Check for aria-label or other attributes
	if element.AriaLabel != "" {
		return 0.8
	}
	if element.Placeholder != "" {
		return 0.7
	}

	return 0.4 // no text makes element harder to verify
}

// getTextPresenceReason returns a reason for text presence score.
func getTextPresenceReason(element *dom.Element) string {
	if element.Text != "" {
		return "Has text content: \"" + truncateString(element.Text, 30) + "\""
	}
	if element.AriaLabel != "" {
		return "Has aria-label: " + element.AriaLabel
	}
	if element.Placeholder != "" {
		return "Has placeholder: " + element.Placeholder
	}
	return "No text or labels found"
}

// calculateVisibilityScore scores element visibility and interactivity.
func calculateVisibilityScore(element *dom.Element) float64 {
	score := 0.5 // base score

	// Check if element has reasonable dimensions
	if hasBoundingBox(element) {
		width := element.BoundingBox.Width
		height := element.BoundingBox.Height

		// Element should be visible size
		if width > 10 && height > 10 {
			score += 0.2
		}
		if width > 50 && height > 20 {
			score += 0.1
		}
		// Penalty for very large elements (might be containers)
		if width > 800 || height > 600 {
			score -= 0.2
		}
	}

	// Bonus for interactive roles
	if element.Role != "" {
		interactiveRoles := map[string]bool{
			"button": true, "link": true, "textbox": true,
			"checkbox": true, "radio": true, "menuitem": true,
		}
		if interactiveRoles[element.Role] {
			score += 0.2
		}
	}

	return math.Min(1.0, math.Max(0.0, score))
}

// getVisibilityReason returns a reason for visibility score.
func getVisibilityReason(element *dom.Element) string {
	if hasBoundingBox(element) {
		return "Element has bounding box: " +
			formatDimensions(element.BoundingBox.Width, element.BoundingBox.Height)
	}
	return "No bounding box information"
}

// calculateUniquenessScore scores how unique an element is.
func calculateUniquenessScore(element *dom.Element, allElements *dom.ElementMap) float64 {
	if allElements == nil {
		return 0.7 // can't determine uniqueness
	}

	// Count similar elements
	similarCount := 0
	for _, el := range allElements.Elements {
		if el.Index == element.Index {
			continue
		}
		if el.TagName == element.TagName && el.Text == element.Text {
			similarCount++
		}
	}

	if similarCount == 0 {
		return 1.0 // unique element
	}
	if similarCount == 1 {
		return 0.8
	}
	if similarCount < 5 {
		return 0.6
	}
	return 0.3 // many similar elements
}

// getUniquenessReason returns a reason for uniqueness score.
func getUniquenessReason(score float64) string {
	switch {
	case score >= 0.9:
		return "Element is unique on page"
	case score >= 0.7:
		return "Few similar elements exist"
	case score >= 0.5:
		return "Several similar elements exist"
	default:
		return "Many similar elements exist"
	}
}

// calculatePositionScore scores element position reasonableness.
func calculatePositionScore(element *dom.Element) float64 {
	if !hasBoundingBox(element) {
		return 0.5 // can't determine position
	}

	x := element.BoundingBox.X
	y := element.BoundingBox.Y

	// Penalize elements at extreme positions
	score := 1.0
	if x < 0 || y < 0 {
		score -= 0.3 // off-screen
	}
	if x > 2000 || y > 5000 {
		score -= 0.2 // very far down/right
	}

	return math.Max(0.0, score)
}

// CalculateActionConfidence calculates confidence for an action.
func (t *ConfidenceTracker) CalculateActionConfidence(
	actionType string,
	verified bool,
	retryCount int,
	elementConf *ElementConfidence,
	prePostDiff float64, // 0-1, how much the page changed
) *ActionConfidence {
	ac := &ActionConfidence{
		ActionType: actionType,
		Verified:   verified,
		RetryCount: retryCount,
	}

	score := NewConfidenceScore(1.0, "Action confidence")

	// Verification status is most important
	if verified {
		score.AddFactor("verification", 0.4, 1.0, "Action verified successfully")
	} else {
		score.AddFactor("verification", 0.4, 0.3, "Action not verified")
	}

	// Retry count affects confidence
	retryScore := 1.0 - (float64(retryCount) * 0.2)
	if retryScore < 0.2 {
		retryScore = 0.2
	}
	score.AddFactor("retries", 0.2, retryScore, formatRetryReason(retryCount))

	// Element targeting confidence
	if elementConf != nil {
		ac.TargetConfidence = elementConf.OverallScore
		score.AddFactor("targeting", 0.2, elementConf.OverallScore.Value,
			"Element targeting confidence")
	} else {
		score.AddFactor("targeting", 0.2, 0.7, "No element targeting data")
	}

	// Page change detection
	changeScore := calculateChangeScore(actionType, prePostDiff)
	score.AddFactor("page_change", 0.2, changeScore,
		formatChangeReason(actionType, prePostDiff))

	score.RecalculateFromFactors()
	ac.OverallScore = score
	ac.PreConfidence = score.Value
	ac.PostConfidence = score.Value

	return ac
}

// formatRetryReason formats the retry reason.
func formatRetryReason(retryCount int) string {
	switch retryCount {
	case 0:
		return "No retries needed"
	case 1:
		return "1 retry was needed"
	default:
		return "Multiple retries needed (" + string(rune('0'+retryCount)) + ")"
	}
}

// calculateChangeScore scores page changes based on action type.
func calculateChangeScore(actionType string, prePostDiff float64) float64 {
	// Expected change levels by action type
	expectedChanges := map[string]float64{
		"navigate": 0.8, // expect big change
		"click":    0.5, // moderate change expected
		"type":     0.3, // small change expected
		"scroll":   0.2, // minor change expected
		"wait":     0.0, // no change expected
	}

	expected, ok := expectedChanges[actionType]
	if !ok {
		expected = 0.5
	}

	// Score based on how well actual matches expected
	diff := math.Abs(prePostDiff - expected)
	return 1.0 - diff
}

// formatChangeReason formats the page change reason.
func formatChangeReason(actionType string, prePostDiff float64) string {
	changeLevel := "minimal"
	if prePostDiff > 0.7 {
		changeLevel = "significant"
	} else if prePostDiff > 0.3 {
		changeLevel = "moderate"
	}
	return changeLevel + " page change after " + actionType
}

// RecordAction records an action's confidence.
func (t *ConfidenceTracker) RecordAction(ac *ActionConfidence) {
	t.actionHistory = append(t.actionHistory, *ac)
}

// GetTaskConfidence calculates overall task confidence.
func (t *ConfidenceTracker) GetTaskConfidence() *TaskConfidence {
	tc := &TaskConfidence{
		StepConfidences: t.actionHistory,
		TotalSteps:      len(t.actionHistory),
	}

	if len(t.actionHistory) == 0 {
		tc.OverallScore = NewConfidenceScore(0.0, "No actions recorded")
		return tc
	}

	var totalConf float64
	var minConf float64 = 1.0
	var verifiedCount int

	for _, ac := range t.actionHistory {
		conf := ac.OverallScore.Value
		totalConf += conf
		if conf < minConf {
			minConf = conf
		}
		if ac.Verified {
			verifiedCount++
		} else {
			tc.FailedSteps++
		}
	}

	tc.AverageConfidence = totalConf / float64(len(t.actionHistory))
	tc.MinConfidence = minConf
	tc.SuccessRate = float64(verifiedCount) / float64(len(t.actionHistory))

	// Calculate overall score from multiple factors
	overallScore := NewConfidenceScore(tc.AverageConfidence, "Overall task confidence")
	overallScore.AddFactor("average_step_confidence", 0.4, tc.AverageConfidence,
		"Average confidence across all steps")
	overallScore.AddFactor("success_rate", 0.4, tc.SuccessRate,
		formatSuccessRate(verifiedCount, len(t.actionHistory)))
	overallScore.AddFactor("minimum_confidence", 0.2, tc.MinConfidence,
		"Lowest step confidence")

	overallScore.RecalculateFromFactors()
	tc.OverallScore = overallScore

	return tc
}

// formatSuccessRate formats the success rate reason.
func formatSuccessRate(verified, total int) string {
	return formatInt(verified) + " of " + formatInt(total) + " actions verified"
}

// formatInt formats an integer as a string.
func formatInt(n int) string {
	return string(rune('0' + n%10))
}

// formatDimensions formats width/height dimensions.
func formatDimensions(width, height float64) string {
	return formatFloat(width) + "x" + formatFloat(height)
}

// formatFloat formats a float for display.
func formatFloat(f float64) string {
	return string([]byte{
		byte('0' + int(f/100)%10),
		byte('0' + int(f/10)%10),
		byte('0' + int(f)%10),
	})
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// CalculatePageStateDiff calculates the difference between two page states.
func CalculatePageStateDiff(pre, post *PageSnapshot) float64 {
	if pre == nil || post == nil {
		return 0.5 // can't determine
	}

	var diff float64
	var factors int

	// URL change
	if pre.URL != post.URL {
		diff += 1.0
	}
	factors++

	// Title change
	if pre.Title != post.Title {
		diff += 0.5
	}
	factors++

	// Element count change
	countDiff := math.Abs(float64(post.ElementCount-pre.ElementCount)) / math.Max(float64(pre.ElementCount), 1.0)
	diff += math.Min(1.0, countDiff)
	factors++

	return diff / float64(factors)
}

// AdjustConfidenceForContext adjusts confidence based on task context.
func AdjustConfidenceForContext(baseConf float64, stepNumber int, totalSteps int, consecutiveFailures int) float64 {
	conf := baseConf

	// Slight decay over many steps (fatigue factor)
	if stepNumber > 10 {
		conf *= 0.99
	}
	if stepNumber > 20 {
		conf *= 0.98
	}

	// Penalty for consecutive failures
	conf *= math.Pow(0.9, float64(consecutiveFailures))

	// Bonus for being near task completion
	if totalSteps > 0 {
		progress := float64(stepNumber) / float64(totalSteps)
		if progress > 0.8 {
			conf *= 1.05 // slight boost near completion
		}
	}

	// Clamp to valid range
	return math.Max(0.0, math.Min(1.0, conf))
}
