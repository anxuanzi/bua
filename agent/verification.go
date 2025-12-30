// Package agent provides browser agent verification capabilities.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anxuanzi/bua-go/dom"
)

// VerificationConfig configures action verification behavior.
type VerificationConfig struct {
	// Enabled turns verification on/off.
	Enabled bool
	// MaxRetries is the number of retry attempts for failed verifications.
	MaxRetries int
	// RetryDelay is the wait time between retries.
	RetryDelay time.Duration
	// StabilizationTime is how long to wait after action before verifying.
	StabilizationTime time.Duration
}

// DefaultVerificationConfig returns sensible defaults.
func DefaultVerificationConfig() *VerificationConfig {
	return &VerificationConfig{
		Enabled:           true,
		MaxRetries:        2,
		RetryDelay:        500 * time.Millisecond,
		StabilizationTime: 300 * time.Millisecond,
	}
}

// ActionVerification represents the verification state for an action.
type ActionVerification struct {
	ActionType      string
	ActionTarget    int    // element index for click/type
	ExpectedChange  string // description of expected change
	PreState        *PageSnapshot
	PostState       *PageSnapshot
	Verified        bool
	VerificationMsg string
	Retries         int
}

// PageSnapshot captures page state for comparison.
type PageSnapshot struct {
	URL          string
	Title        string
	ElementCount int
	Elements     map[int]*dom.Element // key elements by index
	Timestamp    time.Time
}

// Verifier handles action verification logic.
type Verifier struct {
	config *VerificationConfig
	agent  *BrowserAgent
}

// NewVerifier creates a new Verifier instance.
func NewVerifier(agent *BrowserAgent, config *VerificationConfig) *Verifier {
	if config == nil {
		config = DefaultVerificationConfig()
	}
	return &Verifier{
		config: config,
		agent:  agent,
	}
}

// CaptureSnapshot captures the current page state.
func (v *Verifier) CaptureSnapshot(ctx context.Context) (*PageSnapshot, error) {
	if v.agent.browser == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	elements, err := v.agent.browser.GetElementMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get element map: %w", err)
	}

	snapshot := &PageSnapshot{
		URL:          v.agent.browser.GetURL(),
		Title:        v.agent.browser.GetTitle(),
		ElementCount: elements.Count(),
		Elements:     make(map[int]*dom.Element),
		Timestamp:    time.Now(),
	}

	// Store interactive elements for comparison
	for _, el := range elements.InteractiveElements() {
		snapshot.Elements[el.Index] = el
	}

	return snapshot, nil
}

// VerifyClick checks if a click action had the expected effect.
func (v *Verifier) VerifyClick(ctx context.Context, elementIndex int, pre, post *PageSnapshot) *ActionVerification {
	verification := &ActionVerification{
		ActionType:   "click",
		ActionTarget: elementIndex,
		PreState:     pre,
		PostState:    post,
	}

	// Check for state changes that indicate successful click
	changes := v.detectChanges(pre, post)

	if len(changes) > 0 {
		verification.Verified = true
		verification.VerificationMsg = fmt.Sprintf("Click verified: %s", strings.Join(changes, "; "))

		// Detect if a modal/popup likely opened (many new elements appeared)
		// This helps the agent know to use element_id or auto_detect for scrolling inside modals
		elementDiff := post.ElementCount - pre.ElementCount
		if elementDiff >= 20 {
			verification.VerificationMsg += fmt.Sprintf(". MODAL DETECTED: %d new elements appeared - to scroll this content, use scroll(auto_detect=true) or scroll(element_id=<container_index>)", elementDiff)
		}
	} else {
		// No changes detected - might be a problem
		verification.Verified = false
		verification.VerificationMsg = "No state changes detected after click"
	}

	return verification
}

// VerifyType checks if a type action had the expected effect.
func (v *Verifier) VerifyType(ctx context.Context, elementIndex int, text string, pre, post *PageSnapshot) *ActionVerification {
	verification := &ActionVerification{
		ActionType:     "type",
		ActionTarget:   elementIndex,
		ExpectedChange: fmt.Sprintf("text '%s' should appear", text),
		PreState:       pre,
		PostState:      post,
	}

	// Check if the target element's value changed
	preEl, preExists := pre.Elements[elementIndex]
	postEl, postExists := post.Elements[elementIndex]

	if postExists && preExists {
		if postEl.Value != preEl.Value && strings.Contains(postEl.Value, text) {
			verification.Verified = true
			verification.VerificationMsg = fmt.Sprintf("Type verified: value now contains '%s'", text)
		} else if postEl.Value == preEl.Value {
			verification.Verified = false
			verification.VerificationMsg = "Element value unchanged after typing"
		}
	} else if !postExists && preExists {
		// Element disappeared - might indicate form submission or DOM change
		verification.Verified = true
		verification.VerificationMsg = "Element no longer present (possible form submission)"
	} else {
		verification.Verified = false
		verification.VerificationMsg = "Unable to verify type action - element not trackable"
	}

	return verification
}

// VerifyNavigation checks if a navigation action succeeded.
func (v *Verifier) VerifyNavigation(ctx context.Context, targetURL string, pre, post *PageSnapshot) *ActionVerification {
	verification := &ActionVerification{
		ActionType:     "navigate",
		ExpectedChange: fmt.Sprintf("navigate to %s", targetURL),
		PreState:       pre,
		PostState:      post,
	}

	// Check URL change
	if post.URL != pre.URL {
		if strings.Contains(post.URL, targetURL) || targetURL == post.URL {
			verification.Verified = true
			verification.VerificationMsg = fmt.Sprintf("Navigation verified: now at %s", post.URL)
		} else {
			verification.Verified = true // URL changed, even if not exact match
			verification.VerificationMsg = fmt.Sprintf("Navigation resulted in: %s (expected: %s)", post.URL, targetURL)
		}
	} else {
		verification.Verified = false
		verification.VerificationMsg = "URL unchanged after navigation"
	}

	return verification
}

// VerifyScroll checks if a scroll action had effect.
func (v *Verifier) VerifyScroll(ctx context.Context, pre, post *PageSnapshot) *ActionVerification {
	verification := &ActionVerification{
		ActionType: "scroll",
		PreState:   pre,
		PostState:  post,
	}

	// Scroll verification is tricky - check if visible elements changed
	changes := v.detectChanges(pre, post)
	if len(changes) > 0 {
		verification.Verified = true
		verification.VerificationMsg = "Scroll resulted in viewport changes"
	} else {
		// Scrolling at edge of page might not change anything
		verification.Verified = true
		verification.VerificationMsg = "Scroll action completed (no visible changes)"
	}

	return verification
}

// detectChanges compares two snapshots and returns a list of detected changes.
func (v *Verifier) detectChanges(pre, post *PageSnapshot) []string {
	var changes []string

	// URL change
	if post.URL != pre.URL {
		changes = append(changes, fmt.Sprintf("URL: %s → %s", pre.URL, post.URL))
	}

	// Title change
	if post.Title != pre.Title {
		changes = append(changes, fmt.Sprintf("Title: %s → %s", pre.Title, post.Title))
	}

	// Element count change
	if post.ElementCount != pre.ElementCount {
		diff := post.ElementCount - pre.ElementCount
		if diff > 0 {
			changes = append(changes, fmt.Sprintf("+%d elements", diff))
		} else {
			changes = append(changes, fmt.Sprintf("%d elements", diff))
		}
	}

	// Check for specific element changes
	for idx, postEl := range post.Elements {
		if preEl, exists := pre.Elements[idx]; exists {
			// Compare element properties
			if postEl.Text != preEl.Text {
				changes = append(changes, fmt.Sprintf("Element %d text changed", idx))
			}
			if postEl.Value != preEl.Value {
				changes = append(changes, fmt.Sprintf("Element %d value changed", idx))
			}
		}
	}

	// Check for new elements
	for idx := range post.Elements {
		if _, exists := pre.Elements[idx]; !exists {
			changes = append(changes, fmt.Sprintf("New element %d appeared", idx))
		}
	}

	// Check for removed elements
	for idx := range pre.Elements {
		if _, exists := post.Elements[idx]; !exists {
			changes = append(changes, fmt.Sprintf("Element %d removed", idx))
		}
	}

	return changes
}

// ShouldRetry determines if the action should be retried based on verification.
func (v *Verifier) ShouldRetry(verification *ActionVerification) bool {
	if verification.Verified {
		return false
	}
	return verification.Retries < v.config.MaxRetries
}
