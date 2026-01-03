package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anxuanzi/bua/browser"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// BrowserAgent is the main agent that controls browser automation via LLM using ADK.
type BrowserAgent struct {
	agent           agent.Agent
	runner          *runner.Runner
	sessionService  session.Service
	browser         *browser.Browser
	toolkit         *BrowserToolkit
	messageManager  *MessageManager
	maxSteps        int
	maxFailures     int
	debug           bool
	steps           []Step
	screenshotDir   string
	screenshotPaths []string
	useVision       bool
	maxWidth        int
}

// Step represents a single step in the agent's execution.
type Step struct {
	Number         int       `json:"number"`
	Action         string    `json:"action"`
	Target         string    `json:"target,omitempty"`
	Thinking       string    `json:"thinking,omitempty"`
	Evaluation     string    `json:"evaluation,omitempty"`
	Memory         string    `json:"memory,omitempty"`
	NextGoal       string    `json:"next_goal,omitempty"`
	Result         string    `json:"result,omitempty"`
	Success        bool      `json:"success"`
	Timestamp      time.Time `json:"timestamp"`
	DurationMs     int64     `json:"duration_ms"`
	ScreenshotPath string    `json:"screenshot_path,omitempty"`
}

// AgentConfig configures the browser agent.
type AgentConfig struct {
	APIKey          string
	Model           string
	MaxSteps        int
	MaxHistoryItems int
	MaxElements     int
	MaxFailures     int
	TextOnly        bool
	MaxWidth        int
	Debug           bool
	ScreenshotDir   string // Directory to save screenshots (empty = no saving)
}

// Result represents the outcome of an agent run.
type Result struct {
	Success         bool          `json:"success"`
	Data            any           `json:"data,omitempty"`
	Error           string        `json:"error,omitempty"`
	Steps           []Step        `json:"steps"`
	Duration        time.Duration `json:"duration"`
	TokensUsed      int           `json:"tokens_used,omitempty"`
	ScreenshotPaths []string      `json:"screenshot_paths,omitempty"`
}

// NewBrowserAgent creates a new browser agent using ADK.
func NewBrowserAgent(ctx context.Context, cfg AgentConfig, b *browser.Browser) (*BrowserAgent, error) {
	// Get API key from config or environment
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key required: set APIKey in config or GOOGLE_API_KEY environment variable")
	}

	// Set model with default
	modelName := cfg.Model
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}

	// Set max steps with default
	maxSteps := cfg.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 100
	}

	// Set max history items with default
	maxHistoryItems := cfg.MaxHistoryItems
	if maxHistoryItems <= 0 {
		maxHistoryItems = 20
	}

	// Set max elements with default
	maxElements := cfg.MaxElements
	if maxElements <= 0 {
		maxElements = 100
	}

	// Set max consecutive failures with default
	maxFailures := cfg.MaxFailures
	if maxFailures <= 0 {
		maxFailures = 5
	}

	// Set max width with default
	maxWidth := cfg.MaxWidth
	if maxWidth <= 0 {
		maxWidth = 1280
	}

	// Create Gemini model using ADK
	model, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// Create browser toolkit with tools
	toolkit := NewBrowserToolkit(b, maxWidth)
	tools, err := toolkit.CreateAllTools()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser tools: %w", err)
	}

	// Create message manager
	messageManager := NewMessageManager(MessageManagerConfig{
		MaxHistoryItems: maxHistoryItems,
		MaxElements:     maxElements,
		UseVision:       !cfg.TextOnly,
	})

	// Create LLM agent using ADK
	llmAgent, err := llmagent.New(llmagent.Config{
		Name:        "browser_agent",
		Model:       model,
		Description: "An expert web browser automation agent that helps users accomplish tasks by interacting with web pages.",
		Instruction: messageManager.GetSystemPrompt(),
		Tools:       tools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM agent: %w", err)
	}

	// Create in-memory session service using ADK
	sessionService := session.InMemoryService()

	// Create runner using ADK
	agentRunner, err := runner.New(runner.Config{
		AppName:        "bua-browser-agent",
		Agent:          llmAgent,
		SessionService: sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	// Create screenshot directory if specified
	screenshotDir := cfg.ScreenshotDir
	if screenshotDir != "" {
		if err := os.MkdirAll(screenshotDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create screenshot directory: %w", err)
		}
	}

	return &BrowserAgent{
		agent:           llmAgent,
		runner:          agentRunner,
		sessionService:  sessionService,
		browser:         b,
		toolkit:         toolkit,
		messageManager:  messageManager,
		maxSteps:        maxSteps,
		maxFailures:     maxFailures,
		debug:           cfg.Debug,
		steps:           make([]Step, 0),
		screenshotDir:   screenshotDir,
		screenshotPaths: make([]string, 0),
		useVision:       !cfg.TextOnly,
		maxWidth:        maxWidth,
	}, nil
}

// Run executes a task and returns the result.
func (a *BrowserAgent) Run(ctx context.Context, task string) (*Result, error) {
	startTime := time.Now()
	a.steps = make([]Step, 0)
	a.screenshotPaths = make([]string, 0)
	a.messageManager.Clear()
	a.messageManager.SetTask(task)

	// Get initial page state
	if err := a.toolkit.RefreshElementMap(); err != nil {
		// Continue even if initial state fails - page might be blank
		if a.debug {
			fmt.Printf("[Debug] Initial page state: %v\n", err)
		}
	}

	// Generate a unique session ID for this task
	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	userID := "user"

	// Create session before running
	_, err := a.sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "bua-browser-agent",
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Build the initial task message with page state
	taskMessage := a.messageManager.BuildInitialTaskMessage(task, a.toolkit.GetElementMap())

	// Filter sensitive data
	taskMessage = a.messageManager.FilterSensitiveData(taskMessage)

	// Create user message content (with optional screenshot)
	var userContent *genai.Content
	if a.useVision {
		screenshotData, _, err := a.captureAndSaveScreenshot(ctx, 0)
		if err == nil && len(screenshotData) > 0 {
			userContent = a.createMultimodalContent(taskMessage, screenshotData)
		} else {
			userContent = genai.NewContentFromText(taskMessage, "user")
		}
	} else {
		userContent = genai.NewContentFromText(taskMessage, "user")
	}

	// Run the agent using ADK runner
	turnNum := 0
	toolCallNum := 0
	taskComplete := false
	var lastResult *Result
	var lastActionName string
	var lastActionResult string
	var lastActionSuccess bool
	var lastScreenshotData []byte // Reuse screenshot for continuation message

	for toolCallNum < a.maxSteps && !taskComplete {
		turnNum++

		if a.debug {
			fmt.Printf("[Turn %d] Starting...\n", turnNum)
		}

		// Check for too many consecutive failures
		if a.messageManager.GetHistory().GetConsecutiveFailures() >= a.maxFailures {
			if a.debug {
				fmt.Printf("[Turn %d] Too many consecutive failures (%d), forcing completion\n", turnNum, a.maxFailures)
			}
			return &Result{
				Success:         false,
				Error:           fmt.Sprintf("Task aborted after %d consecutive failures", a.maxFailures),
				Steps:           a.steps,
				Duration:        time.Since(startTime),
				ScreenshotPaths: a.screenshotPaths,
			}, nil
		}

		// Capture screenshot at START of each turn (before action execution)
		// This follows browser-use pattern: model sees current state before deciding
		// The screenshot path is saved with the Step to record what the model saw
		var turnScreenshotPath string
		if a.useVision {
			_, path, err := a.captureAndSaveScreenshot(ctx, turnNum)
			if err == nil {
				turnScreenshotPath = path
			}
		}

		// Run the agent for one turn using iter.Seq2 pattern
		for event, err := range a.runner.Run(ctx, userID, sessionID, userContent, agent.RunConfig{}) {
			if err != nil {
				return nil, fmt.Errorf("agent error at turn %d: %w", turnNum, err)
			}

			if event == nil {
				continue
			}

			// Check for function calls (tool usage)
			if event.Content != nil {
				for _, part := range event.Content.Parts {
					// Check for function calls
					if part.FunctionCall != nil {
						toolCallNum++
						toolName := part.FunctionCall.Name
						toolArgs, _ := json.Marshal(part.FunctionCall.Args)
						callStart := time.Now()

						if a.debug {
							fmt.Printf("[Step %d] Tool call: %s\n", toolCallNum, toolName)
						}

						lastActionName = toolName
						lastActionSuccess = true // Will be updated by response

						// Record the step with the screenshot taken at start of this turn
						step := Step{
							Number:         toolCallNum,
							Action:         toolName,
							Target:         string(toolArgs),
							Timestamp:      callStart,
							DurationMs:     0, // Will be updated
							Success:        true,
							ScreenshotPath: turnScreenshotPath,
						}
						a.steps = append(a.steps, step)

						// Add to history
						historyItem := HistoryItem{
							StepNumber:    toolCallNum,
							Timestamp:     callStart,
							ActionName:    toolName,
							ActionParams:  string(toolArgs),
							ActionSuccess: true,
							DurationMs:    0,
						}
						a.messageManager.AddHistoryItem(historyItem)

						// Check if done tool was called
						if toolName == "done" {
							taskComplete = true
							var doneArgs DoneArgs
							if err := json.Unmarshal(toolArgs, &doneArgs); err == nil {
								lastResult = &Result{
									Success:         doneArgs.Success,
									Data:            doneArgs.Data,
									Steps:           a.steps,
									Duration:        time.Since(startTime),
									ScreenshotPaths: a.screenshotPaths,
								}
								if !doneArgs.Success {
									lastResult.Error = doneArgs.Summary
								}
							}
						}
					}

					// Check for function responses (tool results)
					if part.FunctionResponse != nil {
						if a.debug {
							fmt.Printf("[Step %d] Tool response: %s\n", toolCallNum, part.FunctionResponse.Name)
						}

						// Extract result for history
						resp := part.FunctionResponse.Response
						if resp != nil {
							resultBytes, _ := json.Marshal(resp)
							lastActionResult = string(resultBytes)

							// Check if action failed
							if success, exists := resp["success"]; exists {
								if successBool, ok := success.(bool); ok {
									lastActionSuccess = successBool
								}
							}
						}

						// Capture screenshot after tool execution for continuation message
						// Note: We DON'T update the step's ScreenshotPath - that was already set
						// to the pre-action screenshot (browser-use pattern: screenshot shows what
						// the model saw BEFORE deciding, not the result after)
						if a.useVision {
							data, _, err := a.captureAndSaveScreenshot(ctx, toolCallNum)
							if err == nil {
								lastScreenshotData = data // Store for continuation message
							}
						}
					}

					// Check for text content (agent reasoning)
					if part.Text != "" && a.debug {
						// Only show first 200 chars of reasoning
						text := part.Text
						if len(text) > 200 {
							text = text[:200] + "..."
						}
						fmt.Printf("[Turn %d] Agent: %s\n", turnNum, text)
					}
				}
			}

			// Check if this is the final response for this turn
			if event.IsFinalResponse() {
				break
			}
		}

		// If task is complete, break out of the loop
		if taskComplete {
			break
		}

		// Refresh page state for next iteration
		if err := a.toolkit.RefreshElementMap(); err != nil {
			if a.debug {
				fmt.Printf("[Turn %d] Failed to refresh page state: %v\n", turnNum, err)
			}
		}

		// Build continuation message with history and updated page state
		continuationMsg := a.messageManager.BuildContinuationMessage(
			a.toolkit.GetElementMap(),
			lastActionName,
			lastActionResult,
			lastActionSuccess,
		)

		// Filter sensitive data
		continuationMsg = a.messageManager.FilterSensitiveData(continuationMsg)

		// Create content with optional screenshot (reuse the last captured screenshot)
		if a.useVision && len(lastScreenshotData) > 0 {
			userContent = a.createMultimodalContent(continuationMsg, lastScreenshotData)
			lastScreenshotData = nil // Clear after use
		} else {
			userContent = genai.NewContentFromText(continuationMsg, "user")
		}
	}

	// Return result
	if lastResult != nil {
		return lastResult, nil
	}

	// Max steps reached without completion
	return &Result{
		Success:         false,
		Error:           fmt.Sprintf("Max steps (%d) reached without completion", a.maxSteps),
		Steps:           a.steps,
		Duration:        time.Since(startTime),
		ScreenshotPaths: a.screenshotPaths,
	}, nil
}

// GetSteps returns all executed steps.
func (a *BrowserAgent) GetSteps() []Step {
	return a.steps
}

// GetHistory returns the agent's execution history.
func (a *BrowserAgent) GetHistory() *AgentHistory {
	return a.messageManager.GetHistory()
}

// Close cleans up the agent resources.
func (a *BrowserAgent) Close() error {
	// Clean up any resources if needed
	return nil
}

// captureAndSaveScreenshot captures a screenshot and saves it to disk if configured.
// Returns the screenshot bytes and the saved path (empty if not saved).
func (a *BrowserAgent) captureAndSaveScreenshot(ctx context.Context, stepNum int) ([]byte, string, error) {
	// Capture screenshot
	data, err := a.browser.Screenshot(ctx, false)
	if err != nil {
		return nil, "", fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Save to disk if directory is configured
	var savedPath string
	if a.screenshotDir != "" {
		filename := fmt.Sprintf("step_%03d_%d.jpg", stepNum, time.Now().UnixMilli())
		savedPath = filepath.Join(a.screenshotDir, filename)
		if err := os.WriteFile(savedPath, data, 0644); err != nil {
			return data, "", fmt.Errorf("failed to save screenshot: %w", err)
		}
		a.screenshotPaths = append(a.screenshotPaths, savedPath)
	}

	return data, savedPath, nil
}

// GetScreenshotPaths returns all screenshot paths from this run.
func (a *BrowserAgent) GetScreenshotPaths() []string {
	return a.screenshotPaths
}

// createMultimodalContent creates a genai.Content with both text and image.
func (a *BrowserAgent) createMultimodalContent(text string, imageData []byte) *genai.Content {
	parts := []*genai.Part{
		{Text: text},
		{InlineData: &genai.Blob{
			Data:     imageData,
			MIMEType: "image/jpeg",
		}},
	}
	return &genai.Content{
		Parts: parts,
		Role:  "user",
	}
}
