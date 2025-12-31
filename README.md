# BrowserUse Agent - Golang

> Browser automation powered by AI — just tell it what to do in plain English.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## What is bua-go?

bua-go is a Go library that lets you automate web browsers using natural language. Instead of writing complex selectors
and scripts, just describe what you want:

```go
agent.Run(ctx, "Search for 'golang' and click the first result")
```

The AI sees the page (screenshots + DOM), decides what to click/type/scroll, and executes the actions for you.

## Features

| Feature                      | Description                                        |
|------------------------------|----------------------------------------------------|
| **Natural Language**         | Describe tasks in plain English                    |
| **Vision + DOM**             | AI sees both screenshots and page structure        |
| **Google ADK**               | Powered by Gemini via Agent Development Kit        |
| **Presets**                  | Optimize for speed, cost, or quality               |
| **TextOnly Mode**            | Skip screenshots for fastest operation             |
| **Enhanced DOM**             | CDP-based extraction with paint order filtering    |
| **Action Highlighting**      | Orange visual feedback showing clicks/typing       |
| **Session Memory**           | Remembers cookies, logins, and patterns            |
| **Headless Mode**            | Run invisibly in the background                    |
| **Viewport Presets**         | Desktop, tablet, and mobile sizes                  |
| **Visual Annotations**       | See what the AI sees with colored element overlays |
| **File Downloads**           | Download files with authentication support         |
| **Dual-Use Architecture**    | Use as library OR as tool in other ADK agents      |

## Installation

```bash
go get github.com/anxuanzi/bua-go
```

**Requirements:**

- Go 1.25+
- [Google API Key](https://aistudio.google.com/) (for Gemini)
- Chrome/Chromium (auto-managed)

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
	// Only APIKey is required - everything else has sensible defaults!
	agent, err := bua.New(bua.Config{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer agent.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	agent.Start(ctx)
	agent.Navigate(ctx, "https://www.google.com")

	result, _ := agent.Run(ctx, "Search for 'Go programming' and click the first result")
	fmt.Printf("Done! Steps: %d, Tokens: %d\n", len(result.Steps), result.TokensUsed)
}
```

## Web Scraping Example

```go
// Use PresetFast for text-only, fastest operation
agent, _ := bua.New(bua.Config{
	APIKey: os.Getenv("GOOGLE_API_KEY"),
	Preset: bua.PresetFast, // No screenshots, ~5-15K tokens/page
})

result, _ := agent.Run(ctx, `
    Go to news.ycombinator.com and extract the top 5 story titles and URLs.
    Return as JSON array with 'title' and 'url' fields.
`)

data, _ := json.MarshalIndent(result.Data, "", "  ")
fmt.Println(string(data))
```

## Configuration

### Simple Config (Recommended)

Only `APIKey` is required. Everything else has sensible defaults:

```go
// Minimal - just works!
bua.New(bua.Config{APIKey: apiKey})

// With preset for specific needs
bua.New(bua.Config{
    APIKey: apiKey,
    Preset: bua.PresetFast, // Text-only, fastest
})

// For debugging
bua.New(bua.Config{
    APIKey: apiKey,
    Debug:  true,
})
```

### Presets

Use presets to control screenshot quality and token usage:

| Preset | Screenshots | Tokens/Page | Best For |
|--------|-------------|-------------|----------|
| `PresetFast` | None | ~5-15K | Text extraction, form filling, scraping |
| `PresetEfficient` | 640px, q50 | ~15-25K | High-volume automation, cost savings |
| `PresetBalanced` | 800px, q60 | ~25-40K | Most tasks (default) |
| `PresetQuality` | 1024px, q75 | ~40-60K | Complex UIs, visual verification |
| `PresetMax` | 1280px, q85 | ~60-100K | Debugging, maximum accuracy |

### Full Configuration

```go
bua.Config{
    // Required
    APIKey: "your-api-key",

    // Commonly used (all optional with good defaults)
    Model:    bua.ModelGemini3Flash,  // Default
    Headless: false,                  // Show browser window
    Debug:    true,                   // Verbose logging
    Preset:   bua.PresetBalanced,     // Token/quality tradeoff

    // Browser options
    Viewport:    bua.DesktopViewport, // Or TabletViewport, MobileViewport
    ProfileName: "my-session",        // Persist cookies/logins

    // Visual debugging
    ShowAnnotations: true,            // Draw element indices on page
}
```

### Model Constants

All Gemini 2.x/3.x models have 1M token context window.

| Constant | Model ID | Description |
|----------|----------|-------------|
| `bua.ModelGemini3Pro` | gemini-3-pro-preview | Latest, best multimodal |
| `bua.ModelGemini3Flash` | gemini-3-flash-preview | Fast, 1M context |
| `bua.ModelGemini25Pro` | gemini-2.5-pro | Stable, production-ready |
| `bua.ModelGemini25Flash` | gemini-2.5-flash | Fast & efficient |
| `bua.ModelGemini25FlashLite` | gemini-2.5-flash-lite | Most cost-effective |
| `bua.ModelGemini20Flash` | gemini-2.0-flash | Previous gen, stable |

### Advanced: Manual Token Configuration

> **Note:** Most users should use `Preset` instead. These fields are for advanced use cases only.

```go
bua.Config{
    APIKey: apiKey,
    Preset: bua.PresetBalanced, // Start with a preset, then override if needed

    // Override specific settings (rarely needed)
    MaxElements:        150,  // Max elements sent to LLM
    ScreenshotMaxWidth: 800,  // Screenshot width in pixels
    ScreenshotQuality:  60,   // JPEG quality 1-100
}
```

### Enhanced DOM Extraction

Enable CDP-based enhanced DOM extraction for better element detection and token efficiency:

```go
bua.Config{
    APIKey:      apiKey,
    EnhancedDOM: true, // Enable advanced element extraction
}
```

**Features enabled with `EnhancedDOM: true`:**

| Feature | Description |
|---------|-------------|
| **Paint Order Filtering** | Removes visually occluded elements (hidden behind other elements) |
| **Containment Filtering** | Removes redundant child elements inside interactive parents |
| **Cursor-based Detection** | Detects clickable elements via `cursor: pointer` style |
| **New Element Markers** | Marks elements that appeared since last snapshot with `*[id]` |
| **Backend Node ID Tracking** | Stable element references across page updates |
| **Accessibility Integration** | Merges accessibility tree properties for better semantics |

**When to use:**
- Complex pages with overlapping elements (modals, dropdowns)
- Pages with many decorative/non-interactive elements
- Scenarios requiring change detection between page states
- Applications needing more accurate interactive element detection

**Trade-offs:**
- Slightly higher extraction overhead (parallel CDP calls)
- Falls back to standard extraction if CDP calls fail

### Viewport Sizes

| Preset | Size |
|--------|------|
| `bua.DesktopViewport` | 1280×800 |
| `bua.LargeDesktopViewport` | 1920×1080 |
| `bua.TabletViewport` | 768×1024 |
| `bua.MobileViewport` | 375×812 |

## Available Actions

The AI can perform these actions (10 tools):

| Action | What it does |
|--------|--------------|
| `click` | Click on elements |
| `type_text` | Type into inputs |
| `scroll` | Scroll page or specific container (modals, sidebars) |
| `navigate` | Go to a URL |
| `wait` | Wait for page to load |
| `extract` | Pull data from page |
| `get_page_state` | Get current URL, title, elements (optional screenshot) |
| `download_file` | Download files (with auth support) |
| `request_human_takeover` | Ask for human help (CAPTCHA, etc.) |
| `done` | Complete the task |

## How It Works

```
You: "Click the login button"
         ↓
    📸 Screenshot + 🌳 DOM Tree (or TextOnly: DOM only)
         ↓
    🤖 AI analyzes the page
         ↓
    🎯 Finds "Login" button at index [3]
         ↓
    🖱️ Clicks element [3]
         ↓
    ✅ Reports success
```

The AI uses a **hybrid approach**:

- **Vision**: Sees the page layout via compressed screenshots (unless TextOnly)
- **DOM**: Understands element structure and semantics via element map

## Browser Profiles

Keep your sessions alive across runs:

```go
cfg := bua.Config{
    ProfileName: "my-account", // Saves to ~/.bua/profiles/my-account/
}
```

Persists: cookies, localStorage, auth tokens, IndexedDB

## Debugging

```go
cfg := bua.Config{
    Debug:           true, // See what the AI is thinking
    ShowAnnotations: true, // Visual element overlays (requires Headless: false)
}
```

Screenshots are saved to `~/.bua/screenshots/steps/` with annotations showing element indices.

### Action Highlighting

When running with `Headless: false`, you'll see visual feedback for each action:

- **Orange outlines** around clicked elements
- **"typing..."** labels on text inputs
- **Scroll indicators** showing direction
- **Crosshairs** for coordinate-based clicks

Configure highlighting:

```go
cfg := bua.Config{
    Headless:       false,                  // Required to see highlights
    ShowHighlights: &showHighlights,        // Explicit on/off (defaults to !Headless)
    HighlightDelay: 500 * time.Millisecond, // How long highlights show
}
```

## Dual-Use Architecture

bua-go can be used as a **tool within other ADK agents**:

```go
import "github.com/anxuanzi/bua-go/export"

// Create browser tool
browserTool := export.NewBrowserTool(&export.BrowserToolConfig{
    APIKey: apiKey,
    Model:  "gemini-3-flash-preview",
})
defer browserTool.Close()

// Get ADK tool and add to your agent
adkTool, _ := browserTool.Tool()

myAgent, _ := llmagent.New(llmagent.Config{
    Name:  "my_agent",
    Model: model,
    Tools: []tool.Tool{adkTool, otherTools...},
})
```

### Multi-Browser Support

For parallel browser operations:

```go
multiBrowser := export.NewMultiBrowserTool(&export.MultiBrowserToolConfig{
    BrowserToolConfig:     export.DefaultBrowserToolConfig(),
    MaxConcurrentBrowsers: 3,
})
// Actions: create, execute, close, list
```

## Downloads

Files are downloaded to `~/.bua/downloads/`:

```go
result, _ := agent.Run(ctx, `
    Go to example.com/files and download the PDF report.
    Use the download_file tool with the file URL.
`)
```

## Rate Limiting

The agent automatically handles Gemini API rate limits (429 errors):
- Parses retry delay from error response
- Waits and retries automatically
- Add delays between tasks for high-volume operations

## Token Counting

bua-go includes an accurate tokenizer using Google's API:

```go
// Count tokens for budget management
count := agent.CountTokens(ctx, "Your text here")
fmt.Printf("This text uses %d tokens\n", count)

// Token usage is also tracked in results
result, _ := agent.Run(ctx, "do something")
fmt.Printf("Task used %d tokens\n", result.TokensUsed)
```

The tokenizer:
- Uses Google's `CountTokens` API for accuracy
- Caches results to reduce API calls
- Falls back to estimation if API is unavailable

## API Reference

```go
// Create agent
agent, err := bua.New(cfg)

// Start browser
agent.Start(ctx)

// Navigate
agent.Navigate(ctx, "https://example.com")

// Run natural language task
result, err := agent.Run(ctx, "fill out the contact form")

// Result contains:
// - result.Success     bool
// - result.Data        any (extracted data)
// - result.Steps       []Step (action history)
// - result.TokensUsed  int
// - result.Duration    time.Duration
// - result.Confidence  *TaskConfidence

// Take screenshot
screenshot, err := agent.Screenshot(ctx)

// Get element map
elements, err := agent.GetElementMap(ctx)

// Show/hide annotations
agent.ShowAnnotations(ctx, nil)
agent.HideAnnotations(ctx)

// Count tokens (uses Google's tokenizer for accuracy)
tokenCount := agent.CountTokens(ctx, "your text here")

// Cleanup
agent.Close()
```

## Testing

```bash
go test ./...        # Run tests
go test -v ./...     # Verbose output
```

## Examples

The `examples/` directory contains working demonstrations:

| Example | Description |
|---------|-------------|
| `01_quick_start` | Minimal setup, Google search and click |
| `02_google_search` | Search with preset configuration |
| `03_data_scraping` | Scrape Hacker News headlines |
| `04_file_download` | Download files with auth support |
| `05_multi_page` | Navigate across multiple sites |
| `06_modal_handling` | Scroll within modals and popups |
| `07_adk_embedding` | Use as a tool in other ADK agents |
| `08_deep_research` | Multi-site research with data extraction |
| `09_instagram_research` | Multi-tab content analysis |
| `10_highlight_demo` | Visual action highlighting |

Run any example:

```bash
cd examples/01_quick_start
go run main.go
```

## Contributing

1. Fork the repo
2. Make your changes
3. Run `go fmt ./...` and `go vet ./...`
4. Submit a PR

## License

MIT License — see [LICENSE](LICENSE)

## Credits

- [go-rod/rod](https://github.com/go-rod/rod) — Browser automation
- [Google ADK](https://google.golang.org/adk) — Agent Development Kit
- [fogleman/gg](https://github.com/fogleman/gg) — Screenshot annotation
