// Package tools provides ADK tool definitions for browser automation actions.
package tools

import (
	"context"
	"fmt"

	"github.com/anxuanzi/bua-go/browser"
	"github.com/anxuanzi/bua-go/dom"
)

// BrowserContext provides access to browser functionality for tools.
type BrowserContext struct {
	Browser *browser.Browser
}

// ClickParams holds parameters for the click action.
type ClickParams struct {
	// ElementIndex is the index of the element to click (from the element map).
	ElementIndex int `json:"element_index" description:"The index of the element to click, as shown in the element map"`
}

// ClickResult holds the result of a click action.
type ClickResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Click clicks on an element by its index in the element map.
// Use this action when you need to click on a button, link, or any interactive element.
// The element_index corresponds to the numbered elements shown in the annotated screenshot.
func Click(ctx context.Context, bc *BrowserContext, params ClickParams) (*ClickResult, error) {
	if bc.Browser == nil {
		return &ClickResult{Success: false, Message: "Browser not initialized"}, nil
	}

	err := bc.Browser.Click(ctx, params.ElementIndex)
	if err != nil {
		return &ClickResult{Success: false, Message: err.Error()}, nil
	}

	// Wait for page to stabilize after click
	bc.Browser.WaitForStable(ctx)

	return &ClickResult{
		Success: true,
		Message: fmt.Sprintf("Clicked element %d", params.ElementIndex),
	}, nil
}

// TypeParams holds parameters for the type action.
type TypeParams struct {
	// ElementIndex is the index of the input element to type into.
	ElementIndex int `json:"element_index" description:"The index of the input element to type into"`
	// Text is the text to type.
	Text string `json:"text" description:"The text to type into the element"`
}

// TypeResult holds the result of a type action.
type TypeResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Type types text into an input element.
// Use this action to enter text into input fields, text areas, or search boxes.
// First, the element will be clicked to focus it, then the text will be typed.
func Type(ctx context.Context, bc *BrowserContext, params TypeParams) (*TypeResult, error) {
	if bc.Browser == nil {
		return &TypeResult{Success: false, Message: "Browser not initialized"}, nil
	}

	err := bc.Browser.TypeInElement(ctx, params.ElementIndex, params.Text)
	if err != nil {
		return &TypeResult{Success: false, Message: err.Error()}, nil
	}

	return &TypeResult{
		Success: true,
		Message: fmt.Sprintf("Typed '%s' into element %d", params.Text, params.ElementIndex),
	}, nil
}

// ScrollParams holds parameters for the scroll action.
type ScrollParams struct {
	// Direction is the scroll direction: "up", "down", "left", "right".
	Direction string `json:"direction" description:"The scroll direction: up, down, left, or right"`
	// Amount is the scroll amount in pixels (default 500).
	Amount int `json:"amount,omitempty" description:"The scroll amount in pixels (default 500)"`
}

// ScrollResult holds the result of a scroll action.
type ScrollResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Scroll scrolls the page in the specified direction.
// Use this action to reveal more content that is not currently visible.
// Scrolling down is useful for infinite scroll pages or to see more items.
func Scroll(ctx context.Context, bc *BrowserContext, params ScrollParams) (*ScrollResult, error) {
	if bc.Browser == nil {
		return &ScrollResult{Success: false, Message: "Browser not initialized"}, nil
	}

	amount := params.Amount
	if amount == 0 {
		amount = 500
	}

	var deltaX, deltaY float64
	switch params.Direction {
	case "up":
		deltaY = -float64(amount)
	case "down":
		deltaY = float64(amount)
	case "left":
		deltaX = -float64(amount)
	case "right":
		deltaX = float64(amount)
	default:
		return &ScrollResult{Success: false, Message: "Invalid direction. Use: up, down, left, right"}, nil
	}

	err := bc.Browser.Scroll(ctx, deltaX, deltaY)
	if err != nil {
		return &ScrollResult{Success: false, Message: err.Error()}, nil
	}

	return &ScrollResult{
		Success: true,
		Message: fmt.Sprintf("Scrolled %s by %d pixels", params.Direction, amount),
	}, nil
}

// NavigateParams holds parameters for the navigate action.
type NavigateParams struct {
	// URL is the URL to navigate to.
	URL string `json:"url" description:"The URL to navigate to"`
}

// NavigateResult holds the result of a navigate action.
type NavigateResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
}

// Navigate navigates to a URL.
// Use this action to go to a specific webpage or website.
// The URL should be a complete URL including the protocol (https://).
func Navigate(ctx context.Context, bc *BrowserContext, params NavigateParams) (*NavigateResult, error) {
	if bc.Browser == nil {
		return &NavigateResult{Success: false, Message: "Browser not initialized"}, nil
	}

	err := bc.Browser.Navigate(ctx, params.URL)
	if err != nil {
		return &NavigateResult{Success: false, Message: err.Error()}, nil
	}

	return &NavigateResult{
		Success: true,
		Message: fmt.Sprintf("Navigated to %s", params.URL),
		URL:     bc.Browser.GetURL(),
		Title:   bc.Browser.GetTitle(),
	}, nil
}

// WaitParams holds parameters for the wait action.
type WaitParams struct {
	// Reason is the reason for waiting.
	Reason string `json:"reason" description:"The reason for waiting (for logging purposes)"`
}

// WaitResult holds the result of a wait action.
type WaitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Wait waits for the page to stabilize.
// Use this action after interactions that trigger page changes or loading.
// The browser will wait until no more DOM changes are detected.
func Wait(ctx context.Context, bc *BrowserContext, params WaitParams) (*WaitResult, error) {
	if bc.Browser == nil {
		return &WaitResult{Success: false, Message: "Browser not initialized"}, nil
	}

	err := bc.Browser.WaitForStable(ctx)
	if err != nil {
		return &WaitResult{Success: false, Message: err.Error()}, nil
	}

	return &WaitResult{
		Success: true,
		Message: fmt.Sprintf("Waited for page to stabilize: %s", params.Reason),
	}, nil
}

// ExtractParams holds parameters for the extract action.
type ExtractParams struct {
	// ElementIndex is the index of the element to extract data from.
	// If -1, extracts from the entire page.
	ElementIndex int `json:"element_index" description:"The index of the element to extract from (-1 for entire page)"`
	// Fields specifies what data to extract.
	Fields []string `json:"fields,omitempty" description:"Specific fields to extract (optional)"`
}

// ExtractResult holds the result of an extract action.
type ExtractResult struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

// Extract extracts data from an element or the page.
// Use this action to gather structured information from the page.
// Specify element_index=-1 to extract general page information.
func Extract(ctx context.Context, bc *BrowserContext, params ExtractParams) (*ExtractResult, error) {
	if bc.Browser == nil {
		return &ExtractResult{Success: false, Message: "Browser not initialized"}, nil
	}

	data := make(map[string]any)

	if params.ElementIndex < 0 {
		// Extract page-level information
		data["url"] = bc.Browser.GetURL()
		data["title"] = bc.Browser.GetTitle()

		// Get element map for additional info
		elements, err := bc.Browser.GetElementMap(ctx)
		if err == nil {
			data["element_count"] = elements.Count()
		}
	} else {
		// Extract from specific element
		elements, err := bc.Browser.GetElementMap(ctx)
		if err != nil {
			return &ExtractResult{Success: false, Message: err.Error()}, nil
		}

		el, ok := elements.ByIndex(params.ElementIndex)
		if !ok {
			return &ExtractResult{
				Success: false,
				Message: fmt.Sprintf("Element %d not found", params.ElementIndex),
			}, nil
		}

		data["tag"] = el.TagName
		data["text"] = el.Text
		if el.Href != "" {
			data["href"] = el.Href
		}
		if el.Value != "" {
			data["value"] = el.Value
		}
	}

	return &ExtractResult{
		Success: true,
		Message: "Data extracted successfully",
		Data:    data,
	}, nil
}

// HumanTakeoverParams holds parameters for the human takeover action.
type HumanTakeoverParams struct {
	// Reason explains why human intervention is needed.
	Reason string `json:"reason" description:"Explanation of why human intervention is needed"`
}

// HumanTakeoverResult holds the result of a human takeover request.
type HumanTakeoverResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

// RequestHumanTakeover requests human intervention.
// Use this action when you encounter:
// - Login pages requiring credentials
// - CAPTCHAs or bot detection
// - Two-factor authentication
// - Any situation where you cannot proceed automatically
func RequestHumanTakeover(ctx context.Context, bc *BrowserContext, params HumanTakeoverParams) (*HumanTakeoverResult, error) {
	return &HumanTakeoverResult{
		Success:   true,
		Message:   fmt.Sprintf("Human takeover requested: %s", params.Reason),
		Completed: false, // Will be updated when human completes the task
	}, nil
}

// DoneParams holds parameters for the done action.
type DoneParams struct {
	// Success indicates whether the task was completed successfully.
	Success bool `json:"success" description:"Whether the task was completed successfully"`
	// Summary summarizes what was accomplished.
	Summary string `json:"summary" description:"Summary of what was accomplished"`
	// Data contains any extracted data.
	Data map[string]any `json:"data,omitempty" description:"Any extracted or relevant data"`
}

// DoneResult holds the result of a done action.
type DoneResult struct {
	Success bool           `json:"success"`
	Summary string         `json:"summary"`
	Data    map[string]any `json:"data,omitempty"`
}

// Done signals that the task is complete.
// Use this action when you have:
// - Successfully completed the requested task
// - Determined that the task cannot be completed
// - Gathered all requested information
// Always provide a clear summary of what was accomplished.
func Done(ctx context.Context, bc *BrowserContext, params DoneParams) (*DoneResult, error) {
	return &DoneResult{
		Success: params.Success,
		Summary: params.Summary,
		Data:    params.Data,
	}, nil
}

// GetCurrentState returns the current page state for the LLM.
func GetCurrentState(ctx context.Context, bc *BrowserContext) (*PageState, error) {
	if bc.Browser == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	state := &PageState{
		URL:   bc.Browser.GetURL(),
		Title: bc.Browser.GetTitle(),
	}

	// Get element map
	elements, err := bc.Browser.GetElementMap(ctx)
	if err == nil {
		state.Elements = elements
		state.ElementSummary = elements.ToTokenString()
	}

	// Get accessibility tree
	tree, err := bc.Browser.GetAccessibilityTree(ctx)
	if err == nil {
		state.AccessibilityTree = tree
	}

	// Merge accessibility info with elements
	if state.Elements != nil && state.AccessibilityTree != nil {
		dom.MergeWithElementMap(state.Elements, state.AccessibilityTree)
	}

	return state, nil
}

// PageState represents the current state of the page.
type PageState struct {
	URL               string                 `json:"url"`
	Title             string                 `json:"title"`
	Elements          *dom.ElementMap        `json:"elements,omitempty"`
	ElementSummary    string                 `json:"element_summary,omitempty"`
	AccessibilityTree *dom.AccessibilityTree `json:"accessibility_tree,omitempty"`
}
