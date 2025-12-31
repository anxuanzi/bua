package agent

import (
	"fmt"

	"github.com/anxuanzi/bua/browser"
	"github.com/anxuanzi/bua/dom"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// BrowserToolkit holds browser context for tool execution.
type BrowserToolkit struct {
	browser    *browser.Browser
	elementMap *dom.ElementMap
	maxWidth   int
}

// NewBrowserToolkit creates a new browser toolkit.
func NewBrowserToolkit(b *browser.Browser, maxWidth int) *BrowserToolkit {
	return &BrowserToolkit{
		browser:  b,
		maxWidth: maxWidth,
	}
}

// RefreshElementMap updates the cached element map.
func (t *BrowserToolkit) RefreshElementMap() error {
	em, err := t.browser.GetElementMap(nil)
	if err != nil {
		return err
	}
	t.elementMap = em
	return nil
}

// GetElementMap returns the current element map.
func (t *BrowserToolkit) GetElementMap() *dom.ElementMap {
	return t.elementMap
}

// ---- Tool Argument Structs (ADK format with json + jsonschema tags) ----

// NavigateArgs is the input for the navigate tool.
type NavigateArgs struct {
	URL       string `json:"url" jsonschema:"The URL to navigate to"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why navigating to this URL"`
}

// NavigateResult is the output for the navigate tool.
type NavigateResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// ClickArgs is the input for the click tool.
type ClickArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to click"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why clicking this element"`
}

// ClickResult is the output for the click tool.
type ClickResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TypeTextArgs is the input for the type_text tool.
type TypeTextArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to type into"`
	Text         string `json:"text" jsonschema:"The text to type"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why typing this text"`
}

// TypeTextResult is the output for the type_text tool.
type TypeTextResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ClearAndTypeArgs is the input for the clear_and_type tool.
type ClearAndTypeArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to clear and type into"`
	Text         string `json:"text" jsonschema:"The text to type after clearing"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why clearing and typing"`
}

// ClearAndTypeResult is the output for the clear_and_type tool.
type ClearAndTypeResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ScrollArgs is the input for the scroll tool.
type ScrollArgs struct {
	Direction    string `json:"direction" jsonschema:"Scroll direction: up, down, left, right"`
	Amount       int    `json:"amount,omitzero" jsonschema:"Number of pixels to scroll (default 300)"`
	ElementIndex *int   `json:"element_index,omitempty" jsonschema:"Optional element index to scroll within"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why scrolling"`
}

// ScrollResult is the output for the scroll tool.
type ScrollResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendKeysArgs is the input for the send_keys tool.
type SendKeysArgs struct {
	Keys      string `json:"keys" jsonschema:"The keys to send (Enter, Escape, Tab, etc.)"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why sending these keys"`
}

// SendKeysResult is the output for the send_keys tool.
type SendKeysResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GoBackArgs is the input for the go_back tool.
type GoBackArgs struct {
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why going back"`
}

// GoBackResult is the output for the go_back tool.
type GoBackResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GoForwardArgs is the input for the go_forward tool.
type GoForwardArgs struct {
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why going forward"`
}

// GoForwardResult is the output for the go_forward tool.
type GoForwardResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HoverArgs is the input for the hover tool.
type HoverArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to hover over"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why hovering over this element"`
}

// HoverResult is the output for the hover tool.
type HoverResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DoubleClickArgs is the input for the double_click tool.
type DoubleClickArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to double-click"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why double-clicking this element"`
}

// DoubleClickResult is the output for the double_click tool.
type DoubleClickResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// FocusArgs is the input for the focus tool.
type FocusArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to focus"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why focusing this element"`
}

// FocusResult is the output for the focus tool.
type FocusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReloadArgs is the input for the reload tool.
type ReloadArgs struct {
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why reloading the page"`
}

// ReloadResult is the output for the reload tool.
type ReloadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ScrollToElementArgs is the input for the scroll_to_element tool.
type ScrollToElementArgs struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index of the element to scroll into view"`
	Reasoning    string `json:"reasoning,omitempty" jsonschema:"Why scrolling to this element"`
}

// ScrollToElementResult is the output for the scroll_to_element tool.
type ScrollToElementResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ExtractContentArgs is the input for the extract_content tool.
type ExtractContentArgs struct {
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why extracting content"`
}

// ExtractContentResult is the output for the extract_content tool.
type ExtractContentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"`
}

// ScreenshotArgs is the input for the screenshot tool.
type ScreenshotArgs struct {
	FullPage  bool   `json:"full_page,omitempty" jsonschema:"Whether to capture the full page or just the viewport"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why taking a screenshot"`
}

// ScreenshotResult is the output for the screenshot tool.
type ScreenshotResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Size    int    `json:"size,omitempty"`
}

// EvaluateJSArgs is the input for the evaluate_js tool.
type EvaluateJSArgs struct {
	Script    string `json:"script" jsonschema:"The JavaScript code to execute"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why running this JavaScript"`
}

// EvaluateJSResult is the output for the evaluate_js tool.
type EvaluateJSResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  string `json:"result,omitempty"`
}

// WaitArgs is the input for the wait tool.
type WaitArgs struct {
	DurationMs int    `json:"duration_ms,omitzero" jsonschema:"Number of milliseconds to wait (default 1000, max 10000)"`
	Reason     string `json:"reason,omitempty" jsonschema:"Why waiting"`
}

// WaitResult is the output for the wait tool.
type WaitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewTabArgs is the input for the new_tab tool.
type NewTabArgs struct {
	URL       string `json:"url,omitempty" jsonschema:"Optional URL to open in the new tab"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why opening a new tab"`
}

// NewTabResult is the output for the new_tab tool.
type NewTabResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TabID   string `json:"tab_id,omitempty"`
}

// SwitchTabArgs is the input for the switch_tab tool.
type SwitchTabArgs struct {
	TabID     string `json:"tab_id" jsonschema:"The ID of the tab to switch to"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why switching to this tab"`
}

// SwitchTabResult is the output for the switch_tab tool.
type SwitchTabResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CloseTabArgs is the input for the close_tab tool.
type CloseTabArgs struct {
	TabID     string `json:"tab_id" jsonschema:"The ID of the tab to close"`
	Reasoning string `json:"reasoning,omitempty" jsonschema:"Why closing this tab"`
}

// CloseTabResult is the output for the close_tab tool.
type CloseTabResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListTabsArgs is the input for the list_tabs tool (no args needed).
type ListTabsArgs struct{}

// ADKTabInfo represents information about a browser tab.
type ADKTabInfo struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Active bool   `json:"active"`
}

// ListTabsResult is the output for the list_tabs tool.
type ListTabsResult struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Tabs    []ADKTabInfo `json:"tabs"`
}

// GetPageStateArgs is the input for the get_page_state tool (no args needed).
type GetPageStateArgs struct{}

// GetPageStateResult is the output for the get_page_state tool.
type GetPageStateResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Elements string `json:"elements"`
	TabCount int    `json:"tab_count"`
}

// DoneArgs is the input for the done tool.
type DoneArgs struct {
	Success bool   `json:"success" jsonschema:"Whether the task was completed successfully"`
	Summary string `json:"summary" jsonschema:"Summary of what was accomplished"`
	Data    any    `json:"data,omitempty" jsonschema:"Any data to return from the task"`
}

// DoneResult is the output for the done tool.
type DoneResult struct {
	Success bool   `json:"success"`
	Summary string `json:"summary"`
	Data    any    `json:"data,omitempty"`
}

// ---- Tool Functions ----

// CreateNavigateTool creates the navigate function tool.
func (t *BrowserToolkit) CreateNavigateTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "navigate",
			Description: "Navigate the browser to a specified URL",
		},
		func(ctx tool.Context, args NavigateArgs) (NavigateResult, error) {
			if err := t.browser.Navigate(nil, args.URL); err != nil {
				return NavigateResult{Success: false, Message: fmt.Sprintf("Navigation failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return NavigateResult{Success: true, Message: fmt.Sprintf("Navigated to %s", args.URL), URL: args.URL}, nil
		},
	)
}

// CreateClickTool creates the click function tool.
func (t *BrowserToolkit) CreateClickTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "click",
			Description: "Click on an element by its index number",
		},
		func(ctx tool.Context, args ClickArgs) (ClickResult, error) {
			if t.elementMap == nil {
				return ClickResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.Click(nil, args.ElementIndex, t.elementMap); err != nil {
				return ClickResult{Success: false, Message: fmt.Sprintf("Click failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return ClickResult{Success: true, Message: fmt.Sprintf("Clicked element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateTypeTextTool creates the type_text function tool.
func (t *BrowserToolkit) CreateTypeTextTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "type_text",
			Description: "Type text into an input element by its index number",
		},
		func(ctx tool.Context, args TypeTextArgs) (TypeTextResult, error) {
			if t.elementMap == nil {
				return TypeTextResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.TypeText(nil, args.ElementIndex, args.Text, t.elementMap); err != nil {
				return TypeTextResult{Success: false, Message: fmt.Sprintf("Type failed: %v", err)}, nil
			}
			return TypeTextResult{Success: true, Message: fmt.Sprintf("Typed text into element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateClearAndTypeTool creates the clear_and_type function tool.
func (t *BrowserToolkit) CreateClearAndTypeTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "clear_and_type",
			Description: "Clear an input element and type new text into it",
		},
		func(ctx tool.Context, args ClearAndTypeArgs) (ClearAndTypeResult, error) {
			if t.elementMap == nil {
				return ClearAndTypeResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.ClearAndType(nil, args.ElementIndex, args.Text, t.elementMap); err != nil {
				return ClearAndTypeResult{Success: false, Message: fmt.Sprintf("Clear and type failed: %v", err)}, nil
			}
			return ClearAndTypeResult{Success: true, Message: fmt.Sprintf("Cleared and typed into element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateScrollTool creates the scroll function tool.
func (t *BrowserToolkit) CreateScrollTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "scroll",
			Description: "Scroll the page or a specific element in a direction",
		},
		func(ctx tool.Context, args ScrollArgs) (ScrollResult, error) {
			amount := float64(args.Amount)
			if amount == 0 {
				amount = 300
			}
			if err := t.browser.Scroll(nil, args.Direction, amount, args.ElementIndex, t.elementMap); err != nil {
				return ScrollResult{Success: false, Message: fmt.Sprintf("Scroll failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return ScrollResult{Success: true, Message: fmt.Sprintf("Scrolled %s by %.0f pixels", args.Direction, amount)}, nil
		},
	)
}

// CreateSendKeysTool creates the send_keys function tool.
func (t *BrowserToolkit) CreateSendKeysTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "send_keys",
			Description: "Send keyboard keys (Enter, Escape, Tab, ArrowUp, ArrowDown, etc.)",
		},
		func(ctx tool.Context, args SendKeysArgs) (SendKeysResult, error) {
			if err := t.browser.SendKeys(nil, args.Keys); err != nil {
				return SendKeysResult{Success: false, Message: fmt.Sprintf("Send keys failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return SendKeysResult{Success: true, Message: fmt.Sprintf("Sent keys: %s", args.Keys)}, nil
		},
	)
}

// CreateGoBackTool creates the go_back function tool.
func (t *BrowserToolkit) CreateGoBackTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "go_back",
			Description: "Navigate back in browser history",
		},
		func(ctx tool.Context, args GoBackArgs) (GoBackResult, error) {
			if err := t.browser.GoBack(nil); err != nil {
				return GoBackResult{Success: false, Message: fmt.Sprintf("Go back failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return GoBackResult{Success: true, Message: "Navigated back"}, nil
		},
	)
}

// CreateGoForwardTool creates the go_forward function tool.
func (t *BrowserToolkit) CreateGoForwardTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "go_forward",
			Description: "Navigate forward in browser history",
		},
		func(ctx tool.Context, args GoForwardArgs) (GoForwardResult, error) {
			if err := t.browser.GoForward(nil); err != nil {
				return GoForwardResult{Success: false, Message: fmt.Sprintf("Go forward failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return GoForwardResult{Success: true, Message: "Navigated forward"}, nil
		},
	)
}

// CreateHoverTool creates the hover function tool.
func (t *BrowserToolkit) CreateHoverTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "hover",
			Description: "Hover over an element by its index number to reveal tooltips or dropdowns",
		},
		func(ctx tool.Context, args HoverArgs) (HoverResult, error) {
			if t.elementMap == nil {
				return HoverResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.Hover(nil, args.ElementIndex, t.elementMap); err != nil {
				return HoverResult{Success: false, Message: fmt.Sprintf("Hover failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return HoverResult{Success: true, Message: fmt.Sprintf("Hovered over element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateDoubleClickTool creates the double_click function tool.
func (t *BrowserToolkit) CreateDoubleClickTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "double_click",
			Description: "Double-click on an element by its index number",
		},
		func(ctx tool.Context, args DoubleClickArgs) (DoubleClickResult, error) {
			if t.elementMap == nil {
				return DoubleClickResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.DoubleClick(nil, args.ElementIndex, t.elementMap); err != nil {
				return DoubleClickResult{Success: false, Message: fmt.Sprintf("Double-click failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return DoubleClickResult{Success: true, Message: fmt.Sprintf("Double-clicked element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateFocusTool creates the focus function tool.
func (t *BrowserToolkit) CreateFocusTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "focus",
			Description: "Focus on an element by its index number",
		},
		func(ctx tool.Context, args FocusArgs) (FocusResult, error) {
			if t.elementMap == nil {
				return FocusResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.Focus(nil, args.ElementIndex, t.elementMap); err != nil {
				return FocusResult{Success: false, Message: fmt.Sprintf("Focus failed: %v", err)}, nil
			}
			return FocusResult{Success: true, Message: fmt.Sprintf("Focused element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateReloadTool creates the reload function tool.
func (t *BrowserToolkit) CreateReloadTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "reload",
			Description: "Reload the current page",
		},
		func(ctx tool.Context, args ReloadArgs) (ReloadResult, error) {
			if err := t.browser.Reload(nil); err != nil {
				return ReloadResult{Success: false, Message: fmt.Sprintf("Reload failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return ReloadResult{Success: true, Message: "Page reloaded"}, nil
		},
	)
}

// CreateScrollToElementTool creates the scroll_to_element function tool.
func (t *BrowserToolkit) CreateScrollToElementTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "scroll_to_element",
			Description: "Scroll to make an element visible in the viewport",
		},
		func(ctx tool.Context, args ScrollToElementArgs) (ScrollToElementResult, error) {
			if t.elementMap == nil {
				return ScrollToElementResult{Success: false, Message: "No elements available. Call get_page_state first."}, nil
			}
			if err := t.browser.ScrollToElement(nil, args.ElementIndex, t.elementMap); err != nil {
				return ScrollToElementResult{Success: false, Message: fmt.Sprintf("Scroll to element failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return ScrollToElementResult{Success: true, Message: fmt.Sprintf("Scrolled to element [%d]", args.ElementIndex)}, nil
		},
	)
}

// CreateExtractContentTool creates the extract_content function tool.
func (t *BrowserToolkit) CreateExtractContentTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "extract_content",
			Description: "Extract the main text content from the current page",
		},
		func(ctx tool.Context, args ExtractContentArgs) (ExtractContentResult, error) {
			content, err := t.browser.ExtractContent(nil)
			if err != nil {
				return ExtractContentResult{Success: false, Message: fmt.Sprintf("Extract content failed: %v", err)}, nil
			}
			// Truncate if too long
			if len(content) > 10000 {
				content = content[:10000] + "... (truncated)"
			}
			return ExtractContentResult{Success: true, Message: "Content extracted", Content: content}, nil
		},
	)
}

// CreateScreenshotTool creates the screenshot function tool.
func (t *BrowserToolkit) CreateScreenshotTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "screenshot",
			Description: "Take a screenshot of the current page",
		},
		func(ctx tool.Context, args ScreenshotArgs) (ScreenshotResult, error) {
			data, err := t.browser.Screenshot(nil, args.FullPage)
			if err != nil {
				return ScreenshotResult{Success: false, Message: fmt.Sprintf("Screenshot failed: %v", err)}, nil
			}
			// Note: We just report success; actual image data would be handled by the agent
			return ScreenshotResult{Success: true, Message: "Screenshot captured", Size: len(data)}, nil
		},
	)
}

// CreateEvaluateJSTool creates the evaluate_js function tool.
func (t *BrowserToolkit) CreateEvaluateJSTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "evaluate_js",
			Description: "Execute JavaScript code on the page and return the result",
		},
		func(ctx tool.Context, args EvaluateJSArgs) (EvaluateJSResult, error) {
			result, err := t.browser.EvaluateJS(nil, args.Script)
			if err != nil {
				return EvaluateJSResult{Success: false, Message: fmt.Sprintf("JS evaluation failed: %v", err)}, nil
			}
			return EvaluateJSResult{Success: true, Message: "JavaScript executed", Result: result}, nil
		},
	)
}

// CreateWaitTool creates the wait function tool.
func (t *BrowserToolkit) CreateWaitTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "wait",
			Description: "Wait for page stability for a specified duration",
		},
		func(ctx tool.Context, args WaitArgs) (WaitResult, error) {
			durationMs := args.DurationMs
			if durationMs <= 0 {
				durationMs = 1000
			}
			if durationMs > 10000 {
				durationMs = 10000
			}
			// Use browser's wait stable
			t.browser.WaitStable(nil)
			t.RefreshElementMap()
			return WaitResult{Success: true, Message: fmt.Sprintf("Waited for %d ms", durationMs)}, nil
		},
	)
}

// CreateNewTabTool creates the new_tab function tool.
func (t *BrowserToolkit) CreateNewTabTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "new_tab",
			Description: "Open a new browser tab, optionally navigating to a URL",
		},
		func(ctx tool.Context, args NewTabArgs) (NewTabResult, error) {
			tabID, err := t.browser.NewTab(nil, args.URL)
			if err != nil {
				return NewTabResult{Success: false, Message: fmt.Sprintf("New tab failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return NewTabResult{Success: true, Message: fmt.Sprintf("Opened new tab: %s", tabID), TabID: tabID}, nil
		},
	)
}

// CreateSwitchTabTool creates the switch_tab function tool.
func (t *BrowserToolkit) CreateSwitchTabTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "switch_tab",
			Description: "Switch to a different browser tab by its ID",
		},
		func(ctx tool.Context, args SwitchTabArgs) (SwitchTabResult, error) {
			if err := t.browser.SwitchTab(args.TabID); err != nil {
				return SwitchTabResult{Success: false, Message: fmt.Sprintf("Switch tab failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return SwitchTabResult{Success: true, Message: fmt.Sprintf("Switched to tab: %s", args.TabID)}, nil
		},
	)
}

// CreateCloseTabTool creates the close_tab function tool.
func (t *BrowserToolkit) CreateCloseTabTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "close_tab",
			Description: "Close a browser tab by its ID",
		},
		func(ctx tool.Context, args CloseTabArgs) (CloseTabResult, error) {
			if err := t.browser.CloseTab(args.TabID); err != nil {
				return CloseTabResult{Success: false, Message: fmt.Sprintf("Close tab failed: %v", err)}, nil
			}
			t.RefreshElementMap()
			return CloseTabResult{Success: true, Message: fmt.Sprintf("Closed tab: %s", args.TabID)}, nil
		},
	)
}

// CreateListTabsTool creates the list_tabs function tool.
func (t *BrowserToolkit) CreateListTabsTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "list_tabs",
			Description: "List all open browser tabs",
		},
		func(ctx tool.Context, args ListTabsArgs) (ListTabsResult, error) {
			tabs := t.browser.ListTabs()
			tabInfos := make([]ADKTabInfo, len(tabs))
			for i, tab := range tabs {
				tabInfos[i] = ADKTabInfo{
					ID:     tab.ID,
					URL:    tab.URL,
					Title:  tab.Title,
					Active: tab.Active,
				}
			}
			return ListTabsResult{Success: true, Message: fmt.Sprintf("Found %d tabs", len(tabs)), Tabs: tabInfos}, nil
		},
	)
}

// CreateGetPageStateTool creates the get_page_state function tool.
func (t *BrowserToolkit) CreateGetPageStateTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "get_page_state",
			Description: "Get the current page state including URL, title, and interactive elements",
		},
		func(ctx tool.Context, args GetPageStateArgs) (GetPageStateResult, error) {
			if err := t.RefreshElementMap(); err != nil {
				return GetPageStateResult{Success: false, Message: fmt.Sprintf("Failed to get page state: %v", err)}, nil
			}

			elementsText := t.elementMap.ToTokenStringLimited(100)

			return GetPageStateResult{
				Success:  true,
				Message:  "Page state retrieved",
				URL:      t.elementMap.PageURL,
				Title:    t.elementMap.PageTitle,
				Elements: elementsText,
				TabCount: len(t.browser.ListTabs()),
			}, nil
		},
	)
}

// CreateDoneTool creates the done function tool.
func (t *BrowserToolkit) CreateDoneTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "done",
			Description: "Mark the task as complete with a summary of what was accomplished",
		},
		func(ctx tool.Context, args DoneArgs) (DoneResult, error) {
			return DoneResult{
				Success: args.Success,
				Summary: args.Summary,
				Data:    args.Data,
			}, nil
		},
	)
}

// CreateAllTools creates all browser automation tools.
func (t *BrowserToolkit) CreateAllTools() ([]tool.Tool, error) {
	tools := make([]tool.Tool, 0, 23)

	navigateTool, err := t.CreateNavigateTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create navigate tool: %w", err)
	}
	tools = append(tools, navigateTool)

	clickTool, err := t.CreateClickTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create click tool: %w", err)
	}
	tools = append(tools, clickTool)

	typeTextTool, err := t.CreateTypeTextTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create type_text tool: %w", err)
	}
	tools = append(tools, typeTextTool)

	clearAndTypeTool, err := t.CreateClearAndTypeTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create clear_and_type tool: %w", err)
	}
	tools = append(tools, clearAndTypeTool)

	scrollTool, err := t.CreateScrollTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create scroll tool: %w", err)
	}
	tools = append(tools, scrollTool)

	sendKeysTool, err := t.CreateSendKeysTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create send_keys tool: %w", err)
	}
	tools = append(tools, sendKeysTool)

	goBackTool, err := t.CreateGoBackTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create go_back tool: %w", err)
	}
	tools = append(tools, goBackTool)

	goForwardTool, err := t.CreateGoForwardTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create go_forward tool: %w", err)
	}
	tools = append(tools, goForwardTool)

	hoverTool, err := t.CreateHoverTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create hover tool: %w", err)
	}
	tools = append(tools, hoverTool)

	doubleClickTool, err := t.CreateDoubleClickTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create double_click tool: %w", err)
	}
	tools = append(tools, doubleClickTool)

	focusTool, err := t.CreateFocusTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create focus tool: %w", err)
	}
	tools = append(tools, focusTool)

	reloadTool, err := t.CreateReloadTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create reload tool: %w", err)
	}
	tools = append(tools, reloadTool)

	scrollToElementTool, err := t.CreateScrollToElementTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create scroll_to_element tool: %w", err)
	}
	tools = append(tools, scrollToElementTool)

	extractContentTool, err := t.CreateExtractContentTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create extract_content tool: %w", err)
	}
	tools = append(tools, extractContentTool)

	screenshotTool, err := t.CreateScreenshotTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create screenshot tool: %w", err)
	}
	tools = append(tools, screenshotTool)

	evaluateJSTool, err := t.CreateEvaluateJSTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluate_js tool: %w", err)
	}
	tools = append(tools, evaluateJSTool)

	waitTool, err := t.CreateWaitTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create wait tool: %w", err)
	}
	tools = append(tools, waitTool)

	newTabTool, err := t.CreateNewTabTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create new_tab tool: %w", err)
	}
	tools = append(tools, newTabTool)

	switchTabTool, err := t.CreateSwitchTabTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create switch_tab tool: %w", err)
	}
	tools = append(tools, switchTabTool)

	closeTabTool, err := t.CreateCloseTabTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create close_tab tool: %w", err)
	}
	tools = append(tools, closeTabTool)

	listTabsTool, err := t.CreateListTabsTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create list_tabs tool: %w", err)
	}
	tools = append(tools, listTabsTool)

	getPageStateTool, err := t.CreateGetPageStateTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get_page_state tool: %w", err)
	}
	tools = append(tools, getPageStateTool)

	doneTool, err := t.CreateDoneTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create done tool: %w", err)
	}
	tools = append(tools, doneTool)

	return tools, nil
}
