# bua-go

A Go library for browser automation powered by LLMs (Large Language Models). Uses a **vision + DOM hybrid approach** combining annotated screenshots with parsed HTML and accessibility trees to navigate websites and extract data based on natural language prompts.

## Features

- **Vision + DOM Hybrid**: Combines annotated screenshots with DOM structure for robust element identification
- **Natural Language Tasks**: Describe what you want to do in plain English
- **Gemini Integration**: Powered by Google's Gemini 2.5 Flash/Pro models
- **Session Persistence**: Browser profiles maintain cookies, localStorage, and authentication state
- **Memory System**: Short-term (in-task) and long-term (cross-session) memory for learning patterns
- **Headless Support**: Run with or without a visible browser window
- **Element Annotation**: Visual bounding boxes with indices matching the element map

## Installation

```bash
go get github.com/anxuanzi/bua-go
```

## Requirements

- Go 1.23 or later
- Google Gemini API key (get one at [Google AI Studio](https://aistudio.google.com/))
- Chrome/Chromium browser (automatically managed by rod)

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/anxuanzi/bua-go"
)

func main() {
    // Get API key from environment
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        log.Fatal("GEMINI_API_KEY environment variable is required")
    }

    // Create agent configuration
    cfg := bua.Config{
        APIKey:      apiKey,
        Model:       "gemini-2.5-flash",
        ProfileName: "my-session",
        Headless:    false,
        Viewport:    bua.DesktopViewport,
        Debug:       true,
    }

    // Create the agent
    agent, err := bua.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }
    defer agent.Close()

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Start the browser
    if err := agent.Start(ctx); err != nil {
        log.Fatalf("Failed to start agent: %v", err)
    }

    // Navigate to a website
    if err := agent.Navigate(ctx, "https://www.google.com"); err != nil {
        log.Fatalf("Failed to navigate: %v", err)
    }

    // Run a task with natural language
    result, err := agent.Run(ctx, "Search for 'Go programming language' and click on the first result")
    if err != nil {
        log.Fatalf("Task failed: %v", err)
    }

    // Print result
    fmt.Printf("Task completed: success=%v\n", result.Success)
    fmt.Printf("Steps taken: %d\n", len(result.Steps))
}
```

## Configuration Options

```go
type Config struct {
    // APIKey is the Google Gemini API key (required)
    APIKey string

    // Model specifies which Gemini model to use
    // Options: "gemini-2.5-flash" (default), "gemini-2.5-pro"
    Model string

    // ProfileName enables session persistence (cookies, localStorage, etc.)
    // Leave empty for incognito mode
    ProfileName string

    // Headless runs the browser without a visible window
    Headless bool

    // Viewport sets the browser window size
    // Use bua.DesktopViewport, bua.TabletViewport, or bua.MobileViewport
    Viewport *Viewport

    // MaxTokens limits the LLM response length (default: 4096)
    MaxTokens int

    // MaxSteps limits the number of actions per task (default: 50)
    MaxSteps int

    // Debug enables verbose logging
    Debug bool
}
```

## Viewport Presets

```go
bua.DesktopViewport  // 1920x1080
bua.TabletViewport   // 768x1024
bua.MobileViewport   // 375x812
```

## Web Scraping Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/anxuanzi/bua-go"
)

func main() {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        log.Fatal("GEMINI_API_KEY required")
    }

    cfg := bua.Config{
        APIKey:   apiKey,
        Headless: true, // Run in background
    }

    agent, err := bua.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer agent.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := agent.Start(ctx); err != nil {
        log.Fatal(err)
    }

    // Scrape Hacker News top stories
    result, err := agent.Run(ctx, `
        Navigate to news.ycombinator.com and extract the titles and URLs
        of the top 5 stories. Return the data as a JSON array with objects
        containing 'title' and 'url' fields.
    `)
    if err != nil {
        log.Fatal(err)
    }

    if result.Data != nil {
        data, _ := json.MarshalIndent(result.Data, "", "  ")
        fmt.Printf("Extracted data:\n%s\n", data)
    }
}
```

## Architecture

### Vision + DOM Hybrid Approach

bua-go uses a dual approach for understanding web pages:

1. **Vision (Screenshots)**: Annotated screenshots with bounding boxes and element indices help the LLM understand visual layout and identify interactive elements.

2. **DOM (Element Map)**: Parsed HTML combined with the Chrome Accessibility Tree provides semantic information about elements, their roles, states, and text content.

This hybrid approach is more robust than either method alone:
- Vision helps with visually-positioned elements and understanding layout context
- DOM provides reliable selectors and semantic information
- Cross-referencing both catches edge cases missed by either approach

### Element Annotation

Interactive elements are annotated with:
- Visual bounding boxes on screenshots
- Unique indices (e.g., `[0]`, `[1]`, `[2]`)
- Element type and properties in the token map

```
[0] button "Sign In"
[1] input[type=text] placeholder="Search..."
[2] a "Learn More" href="/about"
```

### Memory System

**Short-term Memory**: Observations within the current task
- Page states, actions taken, extracted data
- Automatically compacted when exceeding limits

**Long-term Memory**: Persistent patterns across sessions
- Successful action patterns for specific sites
- Login flows, navigation patterns, extraction strategies
- Saved to disk for cross-session learning

## API Reference

### Agent Methods

```go
// Create a new agent
agent, err := bua.New(cfg)

// Start the browser
err := agent.Start(ctx)

// Navigate to a URL
err := agent.Navigate(ctx, "https://example.com")

// Run a natural language task
result, err := agent.Run(ctx, "click the login button")

// Take a screenshot
screenshot, err := agent.Screenshot(ctx)

// Get the current page state
state, err := agent.GetState(ctx)

// Close the browser
agent.Close()
```

### Result Structure

```go
type Result struct {
    Success bool        // Whether the task completed successfully
    Steps   []Step      // Actions taken during the task
    Data    interface{} // Extracted data (for extraction tasks)
    Error   string      // Error message if failed
}

type Step struct {
    Action    string // e.g., "click", "type", "navigate"
    Target    string // Element or URL targeted
    Reasoning string // LLM's reasoning for this action
    Success   bool   // Whether this step succeeded
}
```

## Available Actions

The LLM can use these actions to interact with pages:

| Action | Description | Parameters |
|--------|-------------|------------|
| `click` | Click an element | `index` (element index) |
| `type` | Type text into an input | `index`, `text` |
| `scroll` | Scroll the page | `direction` (up/down/left/right), `amount` |
| `navigate` | Go to a URL | `url` |
| `wait` | Wait for content | `seconds` or `selector` |
| `extract` | Extract data | `data` (structured extraction) |
| `request_human_takeover` | Ask for human help | `reason` |
| `done` | Mark task complete | `success`, `message` |

## Browser Profiles

Profiles persist browser state across sessions:

```go
cfg := bua.Config{
    ProfileName: "my-account", // Creates ~/.bua/profiles/my-account/
}
```

Profile data includes:
- Cookies
- localStorage and sessionStorage
- IndexedDB data
- Authentication tokens

## Debugging

Enable debug mode for verbose output:

```go
cfg := bua.Config{
    Debug: true,
}
```

This logs:
- LLM prompts and responses
- Actions taken and their results
- Screenshot and element map details
- Memory operations

## Error Handling

```go
result, err := agent.Run(ctx, "complete the checkout")
if err != nil {
    // Task execution error (timeout, browser crash, etc.)
    log.Printf("Execution error: %v", err)
    return
}

if !result.Success {
    // Task failed (couldn't complete the requested action)
    log.Printf("Task failed: %s", result.Error)
    for _, step := range result.Steps {
        if !step.Success {
            log.Printf("Failed step: %s - %s", step.Action, step.Reasoning)
        }
    }
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go conventions (`go fmt`, `go vet`)
2. Tests pass (`go test ./...`)
3. New features include tests
4. Documentation is updated

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [go-rod/rod](https://github.com/go-rod/rod) - Browser automation
- [fogleman/gg](https://github.com/fogleman/gg) - Image annotation
- [Google Gemini](https://ai.google.dev/) - LLM provider
