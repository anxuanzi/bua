// Package agent provides the ADK-based browser automation agent.
package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"

	"github.com/anxuanzi/bua-go/browser"
	"github.com/anxuanzi/bua-go/dom"
	"github.com/anxuanzi/bua-go/memory"
)

// Config holds agent configuration.
type Config struct {
	// APIKey is the Gemini API key.
	APIKey string

	// Model is the model ID to use.
	Model string

	// MaxIterations is the maximum number of agent loop iterations.
	MaxIterations int

	// MaxTokens is the maximum context window size.
	MaxTokens int

	// Debug enables verbose logging.
	Debug bool
}

// BrowserAgent wraps an ADK agent with browser automation capabilities.
type BrowserAgent struct {
	config   Config
	browser  *browser.Browser
	memory   *memory.Manager
	adkAgent agent.Agent
}

// New creates a new browser agent.
func New(cfg Config, b *browser.Browser, m *memory.Manager) *BrowserAgent {
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 50
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 128000
	}
	if cfg.Model == "" {
		cfg.Model = "gemini-2.5-flash"
	}

	return &BrowserAgent{
		config:  cfg,
		browser: b,
		memory:  m,
	}
}

// Init initializes the ADK agent with browser tools.
func (a *BrowserAgent) Init(ctx context.Context) error {
	// Get API key
	apiKey := a.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	// Create Gemini model
	model, err := gemini.NewModel(ctx, a.config.Model, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// Create browser tools
	tools, err := a.createBrowserTools()
	if err != nil {
		return fmt.Errorf("failed to create browser tools: %w", err)
	}

	// Create ADK agent
	adkAgent, err := llmagent.New(llmagent.Config{
		Name:        "browser_automation_agent",
		Model:       model,
		Description: "A browser automation agent that can navigate websites, interact with elements, and extract data.",
		Instruction: SystemPrompt(),
		Tools:       tools,
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature:     genai.Ptr[float32](0.2),
			MaxOutputTokens: 4096,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create ADK agent: %w", err)
	}
	a.adkAgent = adkAgent

	return nil
}

// createBrowserTools creates the function tools for browser automation.
func (a *BrowserAgent) createBrowserTools() ([]tool.Tool, error) {
	var tools []tool.Tool

	// Click tool
	clickHandler := func(ctx tool.Context, input ClickInput) (ClickOutput, error) {
		if a.browser == nil {
			return ClickOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		err := a.browser.Click(context.Background(), input.ElementIndex)
		if err != nil {
			return ClickOutput{Success: false, Message: err.Error()}, nil
		}
		a.browser.WaitForStable(context.Background())
		return ClickOutput{
			Success: true,
			Message: fmt.Sprintf("Clicked element %d", input.ElementIndex),
		}, nil
	}
	clickTool, err := functiontool.New(
		functiontool.Config{
			Name:        "click",
			Description: "Click on an element by its index number shown in the annotated screenshot and element map.",
		},
		clickHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create click tool: %w", err)
	}
	tools = append(tools, clickTool)

	// Type tool
	typeHandler := func(ctx tool.Context, input TypeInput) (TypeOutput, error) {
		if a.browser == nil {
			return TypeOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		err := a.browser.TypeInElement(context.Background(), input.ElementIndex, input.Text)
		if err != nil {
			return TypeOutput{Success: false, Message: err.Error()}, nil
		}
		return TypeOutput{
			Success: true,
			Message: fmt.Sprintf("Typed '%s' into element %d", input.Text, input.ElementIndex),
		}, nil
	}
	typeTool, err := functiontool.New(
		functiontool.Config{
			Name:        "type_text",
			Description: "Type text into an input field. First clicks the element to focus it, then types the text.",
		},
		typeHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create type tool: %w", err)
	}
	tools = append(tools, typeTool)

	// Scroll tool
	scrollHandler := func(ctx tool.Context, input ScrollInput) (ScrollOutput, error) {
		if a.browser == nil {
			return ScrollOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		amount := input.Amount
		if amount == 0 {
			amount = 500
		}
		var deltaY float64
		switch input.Direction {
		case "up":
			deltaY = -float64(amount)
		case "down":
			deltaY = float64(amount)
		default:
			return ScrollOutput{Success: false, Message: "Invalid direction. Use: up or down"}, nil
		}
		err := a.browser.Scroll(context.Background(), 0, deltaY)
		if err != nil {
			return ScrollOutput{Success: false, Message: err.Error()}, nil
		}
		return ScrollOutput{
			Success: true,
			Message: fmt.Sprintf("Scrolled %s by %d pixels", input.Direction, amount),
		}, nil
	}
	scrollTool, err := functiontool.New(
		functiontool.Config{
			Name:        "scroll",
			Description: "Scroll the page in a direction (up or down) to reveal more content.",
		},
		scrollHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scroll tool: %w", err)
	}
	tools = append(tools, scrollTool)

	// Navigate tool
	navigateHandler := func(ctx tool.Context, input NavigateInput) (NavigateOutput, error) {
		if a.browser == nil {
			return NavigateOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		err := a.browser.Navigate(context.Background(), input.URL)
		if err != nil {
			return NavigateOutput{Success: false, Message: err.Error()}, nil
		}
		return NavigateOutput{
			Success: true,
			Message: fmt.Sprintf("Navigated to %s", input.URL),
			URL:     a.browser.GetURL(),
			Title:   a.browser.GetTitle(),
		}, nil
	}
	navigateTool, err := functiontool.New(
		functiontool.Config{
			Name:        "navigate",
			Description: "Navigate to a specific URL.",
		},
		navigateHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create navigate tool: %w", err)
	}
	tools = append(tools, navigateTool)

	// Wait tool
	waitHandler := func(ctx tool.Context, input WaitInput) (WaitOutput, error) {
		if a.browser == nil {
			return WaitOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		err := a.browser.WaitForStable(context.Background())
		if err != nil {
			return WaitOutput{Success: false, Message: err.Error()}, nil
		}
		return WaitOutput{
			Success: true,
			Message: fmt.Sprintf("Waited for page to stabilize: %s", input.Reason),
		}, nil
	}
	waitTool, err := functiontool.New(
		functiontool.Config{
			Name:        "wait",
			Description: "Wait for the page to stabilize after an action or for dynamic content to load.",
		},
		waitHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wait tool: %w", err)
	}
	tools = append(tools, waitTool)

	// Extract tool
	extractHandler := func(ctx tool.Context, input ExtractInput) (ExtractOutput, error) {
		if a.browser == nil {
			return ExtractOutput{Success: false, Message: "Browser not initialized"}, nil
		}
		data := make(map[string]any)
		if input.ElementIndex < 0 {
			data["url"] = a.browser.GetURL()
			data["title"] = a.browser.GetTitle()
			elements, err := a.browser.GetElementMap(context.Background())
			if err == nil {
				data["element_count"] = elements.Count()
			}
		} else {
			elements, err := a.browser.GetElementMap(context.Background())
			if err != nil {
				return ExtractOutput{Success: false, Message: err.Error()}, nil
			}
			el, ok := elements.ByIndex(input.ElementIndex)
			if !ok {
				return ExtractOutput{
					Success: false,
					Message: fmt.Sprintf("Element %d not found", input.ElementIndex),
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
		return ExtractOutput{
			Success: true,
			Message: "Data extracted successfully",
			Data:    data,
		}, nil
	}
	extractTool, err := functiontool.New(
		functiontool.Config{
			Name:        "extract",
			Description: "Extract data from an element or the page. Use element_index=-1 to extract general page information.",
		},
		extractHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create extract tool: %w", err)
	}
	tools = append(tools, extractTool)

	// Get page state tool
	getPageStateHandler := func(ctx tool.Context, input GetPageStateInput) (GetPageStateOutput, error) {
		if a.browser == nil {
			return GetPageStateOutput{Success: false, Error: "Browser not initialized"}, nil
		}
		bgCtx := context.Background()
		output := GetPageStateOutput{
			Success: true,
			URL:     a.browser.GetURL(),
			Title:   a.browser.GetTitle(),
		}
		elements, err := a.browser.GetElementMap(bgCtx)
		if err != nil {
			output.Error = fmt.Sprintf("Failed to get element map: %v", err)
			return output, nil
		}
		output.ElementMap = elements.ToTokenString()
		if a.memory != nil {
			a.memory.AddObservation(&memory.Observation{
				Timestamp:    time.Now(),
				URL:          output.URL,
				Title:        output.Title,
				ElementCount: elements.Count(),
			})
		}
		return output, nil
	}
	pageStateTool, err := functiontool.New(
		functiontool.Config{
			Name:        "get_page_state",
			Description: "Get the current page state including URL, title, and interactive elements. Call this to see what's on the page.",
		},
		getPageStateHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create page state tool: %w", err)
	}
	tools = append(tools, pageStateTool)

	// Request human takeover tool
	humanTakeoverHandler := func(ctx tool.Context, input HumanTakeoverInput) (HumanTakeoverOutput, error) {
		return HumanTakeoverOutput{
			Success:   true,
			Message:   fmt.Sprintf("Human takeover requested: %s. Please complete the action and confirm.", input.Reason),
			Completed: false,
		}, nil
	}
	humanTool, err := functiontool.New(
		functiontool.Config{
			Name:        "request_human_takeover",
			Description: "Request a human to take over for tasks like login, CAPTCHA, or other actions requiring human intervention.",
		},
		humanTakeoverHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create human takeover tool: %w", err)
	}
	tools = append(tools, humanTool)

	// Done tool
	doneHandler := func(ctx tool.Context, input DoneInput) (DoneOutput, error) {
		return DoneOutput{
			Success: input.Success,
			Summary: input.Summary,
		}, nil
	}
	doneTool, err := functiontool.New(
		functiontool.Config{
			Name:        "done",
			Description: "Indicate that the task is complete. Set success=true if the task was accomplished, false otherwise.",
		},
		doneHandler,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create done tool: %w", err)
	}
	tools = append(tools, doneTool)

	return tools, nil
}

// Tool input/output types

type ClickInput struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index number of the element to click (shown in the element map)"`
	Reasoning    string `json:"reasoning" jsonschema:"Brief explanation of why you're clicking this element"`
}

type ClickOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type TypeInput struct {
	ElementIndex int    `json:"element_index" jsonschema:"The index number of the input element"`
	Text         string `json:"text" jsonschema:"The text to type into the element"`
	Reasoning    string `json:"reasoning" jsonschema:"Brief explanation of why you're typing this text"`
}

type TypeOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ScrollInput struct {
	Direction string `json:"direction" jsonschema:"Direction to scroll: up or down"`
	Amount    int    `json:"amount" jsonschema:"Amount to scroll in pixels (default 500)"`
	Reasoning string `json:"reasoning" jsonschema:"Brief explanation of why you're scrolling"`
}

type ScrollOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type NavigateInput struct {
	URL       string `json:"url" jsonschema:"The URL to navigate to"`
	Reasoning string `json:"reasoning" jsonschema:"Brief explanation of why you're navigating to this URL"`
}

type NavigateOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
}

type WaitInput struct {
	Reason string `json:"reason" jsonschema:"What you're waiting for"`
}

type WaitOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ExtractInput struct {
	ElementIndex  int    `json:"element_index" jsonschema:"The index of the element to extract from (-1 for entire page)"`
	WhatToExtract string `json:"what_to_extract" jsonschema:"Description of what data to extract"`
}

type ExtractOutput struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

type GetPageStateInput struct {
	IncludeScreenshot bool `json:"include_screenshot" jsonschema:"Whether to include the annotated screenshot (default true)"`
}

type GetPageStateOutput struct {
	Success    bool   `json:"success"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	ElementMap string `json:"element_map"`
	Screenshot string `json:"screenshot,omitempty"`
	Error      string `json:"error,omitempty"`
}

type HumanTakeoverInput struct {
	Reason string `json:"reason" jsonschema:"Why human intervention is needed"`
}

type HumanTakeoverOutput struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

type DoneInput struct {
	Success       bool   `json:"success" jsonschema:"Whether the task was completed successfully"`
	Summary       string `json:"summary" jsonschema:"Summary of what was accomplished"`
	ExtractedData string `json:"extracted_data,omitempty" jsonschema:"Any data that was extracted during the task (as JSON)"`
}

type DoneOutput struct {
	Success bool   `json:"success"`
	Summary string `json:"summary"`
}

// GetADKAgent returns the underlying ADK agent for advanced use cases.
func (a *BrowserAgent) GetADKAgent() agent.Agent {
	return a.adkAgent
}

// GetBrowser returns the browser instance.
func (a *BrowserAgent) GetBrowser() *browser.Browser {
	return a.browser
}

// GetMemory returns the memory manager.
func (a *BrowserAgent) GetMemory() *memory.Manager {
	return a.memory
}

// Result represents the result of a task execution.
type Result struct {
	Success         bool
	Data            map[string]any
	Error           string
	Steps           []Step
	TokensUsed      int
	ScreenshotPaths []string
}

// Step represents a single step in the execution.
type Step struct {
	Action         string
	Target         string
	Reasoning      string
	URL            string
	Title          string
	ScreenshotPath string
}

// PageState represents the current state of the page.
type PageState struct {
	URL           string
	Title         string
	Elements      *dom.ElementMap
	Screenshot    []byte
	ScreenshotB64 string
}
