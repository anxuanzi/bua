<h1 align="center">BUA - Browser Use Agent for Go</h1>

<p align="center">
  <strong>🤖 Make websites accessible for AI agents. Automate the web with natural language.</strong>
</p>

<p align="center">
  <a href="#installation">Installation</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#features">Features</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#examples">Examples</a> •
  <a href="#architecture">Architecture</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"/>
  <img src="https://img.shields.io/badge/LLM-Gemini-4285F4?style=flat&logo=google" alt="Gemini"/>
</p>

---

## Why BUA?

Traditional browser automation is **fragile**. CSS selectors break. XPaths change. Every website update means rewriting
scripts.

**BUA changes the game.** Instead of writing brittle selectors, describe what you want in plain English:

```go
result, _ := agent.Run(ctx, "Go to Amazon and find the best-rated wireless headphones under $100")
```

The AI agent **sees** the page, **understands** your intent, and **adapts** to any layout. No selectors. No maintenance.
Just results.

---

## ✨ What Makes BUA Special

| Feature                  | Traditional Automation  | BUA                  |
|--------------------------|-------------------------|----------------------|
| **Selector Maintenance** | Constant updates needed | Zero maintenance     |
| **Dynamic Content**      | Complex waits & retries | AI understands state |
| **Multi-step Workflows** | Hundreds of lines       | One sentence         |
| **Layout Changes**       | Scripts break           | Adapts automatically |
| **New Sites**            | Write new selectors     | Works immediately    |

### 🎯 Perfect For

- **Web Scraping** - Extract data from any site without writing parsers
- **Form Automation** - Fill applications, registrations, checkout flows
- **E2E Testing** - Test user journeys with natural language
- **Data Entry** - Automate repetitive web-based tasks
- **Research** - Gather information across multiple sources
- **Monitoring** - Track prices, inventory, content changes

---

## 🚀 Installation

```bash
go get github.com/anxuanzi/bua
```

### Prerequisites

- **Go 1.25+**
- **Chrome/Chromium** installed on your system
- **Gemini API Key** from [Google AI Studio](https://aistudio.google.com/apikey)

---

## ⚡ Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anxuanzi/bua"
)

func main() {
	// Create agent with your Gemini API key
	agent, err := bua.New(bua.Config{
		APIKey:   os.Getenv("GEMINI_API_KEY"),
		Headless: false, // Watch the magic happen
		Debug:    true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer agent.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start the browser
	if err := agent.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// Run a task with natural language
	result, err := agent.Run(ctx,
		"Go to Hacker News and find the top 3 stories about AI")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Success: %v\n", result.Success)
	fmt.Printf("📊 Steps taken: %d\n", len(result.Steps))
	fmt.Printf("⏱️  Duration: %v\n", result.Duration)
}
```

**That's it.** The agent navigates to Hacker News, scans the stories, identifies AI-related content, and returns the
results.

---

## 🛠️ Features

### 🧠 Intelligent Navigation

BUA doesn't just click buttons—it **understands** web pages:

```go
// The agent figures out HOW to accomplish the task
agent.Run(ctx, "Find flights from NYC to London for next weekend, sort by price")

// Multi-step workflows handled automatically
agent.Run(ctx, "Log into my account, go to settings, and change my timezone to PST")
```

### 👁️ Vision-Enabled

Screenshots are analyzed by the LLM for visual understanding:

```go
cfg := bua.Config{
Preset: bua.PresetQuality, // High-res screenshots
// Vision is enabled by default
}
```

### 🥷 Stealth Mode

Built-in anti-detection measures help avoid bot blocking:

- Navigator property spoofing
- WebGL fingerprint masking
- Plugin emulation
- Human-like mouse movements
- Random action delays

### 📸 Screenshot Annotations

Visual debugging with element indices overlaid on screenshots:

```go
cfg := bua.Config{
ShowAnnotations: true, // See what the AI sees
ScreenshotDir:   "./debug",
}
```

<p align="center">
  <img src="https://raw.githubusercontent.com/anxuanzi/bua/main/assets/annotated-screenshot.png" alt="Annotated Screenshot" width="600"/>
</p>

### 🎛️ Flexible Presets

Optimize for speed, cost, or quality:

| Preset            | Tokens | Screenshot       | Best For                  |
|-------------------|--------|------------------|---------------------------|
| `PresetFast`      | 8K     | None (text-only) | Simple tasks, lowest cost |
| `PresetEfficient` | 16K    | 800px @ 60%      | Balanced cost/capability  |
| `PresetBalanced`  | 32K    | 1280px @ 75%     | **Default** - most tasks  |
| `PresetQuality`   | 64K    | 1920px @ 85%     | Complex visual tasks      |
| `PresetMax`       | 128K   | 2560px @ 95%     | Maximum accuracy          |

### 🔐 Sensitive Data Protection

Automatic redaction of sensitive information in logs:

```go
// API keys, passwords, SSNs, credit cards are automatically masked
// <secret type="api_key">[REDACTED]</secret>
```

### 🗂️ Tab Management

Handle complex multi-tab workflows:

```go
// Open comparison shopping tabs
tab1, _ := agent.NewTab(ctx, "https://amazon.com")
tab2, _ := agent.NewTab(ctx, "https://ebay.com")

agent.SwitchTab(tab1)
agent.Run(ctx, "Search for 'mechanical keyboard'")

agent.SwitchTab(tab2)
agent.Run(ctx, "Search for 'mechanical keyboard' and compare prices")
```

### 💾 Session Persistence

Save and restore browser sessions:

```go
cfg := bua.Config{
ProfileName: "my-shopping-session",
ProfileDir:  "~/.bua/profiles",
// Cookies, localStorage, login state preserved
}
```

---

## ⚙️ Configuration

### Full Configuration Options

```go
cfg := bua.Config{
// Required
APIKey: "your-gemini-api-key",

// LLM Settings
Model: "gemini-2.5-flash", // or "gemini-2.0-flash", etc.

// Browser Settings
Headless:    false,        // true for background operation
ProfileName: "persistent", // empty = temporary profile
ProfileDir:  "~/.bua/profiles",
Viewport:    &bua.Viewport{Width: 1920, Height: 1080},

// Agent Behavior
MaxSteps:    100, // Max actions before giving up
Preset:      bua.PresetBalanced,

// Screenshot Settings
ScreenshotDir:      "./screenshots",
ScreenshotMaxWidth: 1280,
ScreenshotQuality:  75,
TextOnly:           false, // true disables screenshots
ShowAnnotations:    false, // true shows element indices

// Visual Feedback
ShowHighlight:       boolPtr(true),
HighlightDurationMs: 300,

// Debugging
Debug: true,
}
```

### Environment Variables

```bash
export GEMINI_API_KEY="your-api-key-here"
```

---

## 📖 Examples

### Example 1: Web Research

```go
result, _ := agent.Run(ctx, `
    Go to Wikipedia and find information about the Go programming language.
    Extract the release date, original author, and main features.
`)

fmt.Println(result.Data) // Extracted information
```

### Example 2: E-commerce Automation

```go
result, _ := agent.Run(ctx, `
    Go to Amazon, search for "USB-C hub", filter by 4+ stars,
    and find the cheapest option with Prime shipping.
`)

for _, step := range result.Steps {
fmt.Printf("[%d] %s: %s\n", step.Number, step.Action, step.NextGoal)
}
```

### Example 3: Form Filling

```go
result, _ := agent.Run(ctx, `
    Go to the contact form at example.com/contact.
    Fill in:
    - Name: John Doe
    - Email: john@example.com
    - Message: I'm interested in your services
    Then submit the form.
`)
```

### Example 4: Multi-Page Workflow

```go
// Navigate first
agent.Navigate(ctx, "https://github.com/login")

// Then automate
result, _ := agent.Run(ctx, `
    Log in with username 'myuser' and password from the password field.
    After logging in, go to my repositories and find the most starred one.
`)
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application                          │
├─────────────────────────────────────────────────────────────┤
│                   BUA Public API                             │
│         bua.New() → Start() → Run() → Close()               │
├─────────────────────────────────────────────────────────────┤
│                   Agent Layer                                │
│    ┌─────────────┐  ┌──────────────┐  ┌────────────────┐   │
│    │ BrowserAgent│  │ ADK Toolkit  │  │ Message Builder│   │
│    │  (LLM Loop) │  │ (20+ Tools)  │  │  (History)     │   │
│    └─────────────┘  └──────────────┘  └────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                   Browser Layer                              │
│    ┌─────────────┐  ┌──────────────┐  ┌────────────────┐   │
│    │   Browser   │  │    Page      │  │   Stealth      │   │
│    │ (Lifecycle) │  │ (Actions)    │  │   (Evasion)    │   │
│    └─────────────┘  └──────────────┘  └────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                   Support Layer                              │
│    ┌─────────────┐  ┌──────────────┐                        │
│    │     DOM     │  │  Screenshot  │                        │
│    │ (Extraction)│  │ (Annotation) │                        │
│    └─────────────┘  └──────────────┘                        │
├─────────────────────────────────────────────────────────────┤
│              go-rod (Chrome DevTools Protocol)              │
├─────────────────────────────────────────────────────────────┤
│                   Chrome / Chromium                          │
└─────────────────────────────────────────────────────────────┘
```

### Agent Loop

```
Task: "Search for Go tutorials"
           │
           ▼
┌──────────────────────┐
│   Get Page State     │◄─────────────────────┐
│   (DOM + Screenshot) │                      │
└──────────┬───────────┘                      │
           │                                  │
           ▼                                  │
┌──────────────────────┐                      │
│   LLM Reasoning      │                      │
│   (Gemini + Tools)   │                      │
└──────────┬───────────┘                      │
           │                                  │
           ▼                                  │
┌──────────────────────┐                      │
│   Execute Action     │                      │
│   (click, type, etc) │                      │
└──────────┬───────────┘                      │
           │                                  │
           ▼                                  │
    ┌──────────────┐         ┌───────────────┐
    │ Task Done?   │───No───►│ Update State  │
    └──────┬───────┘         └───────┬───────┘
           │Yes                      │
           ▼                         │
    ┌──────────────┐                 │
    │ Return Result│                 │
    └──────────────┘                 │
                                     └────────┘
```

---

## 🔧 Available Tools

The agent has access to 20+ browser automation tools:

| Category        | Tools                                                                    |
|-----------------|--------------------------------------------------------------------------|
| **Navigation**  | `navigate`, `go_back`, `go_forward`, `reload`                            |
| **Interaction** | `click`, `type_text`, `clear_and_type`, `hover`, `double_click`, `focus` |
| **Scrolling**   | `scroll`, `scroll_to_element`                                            |
| **Keyboard**    | `send_keys` (Enter, Tab, Escape, etc.)                                   |
| **Observation** | `get_page_state`, `screenshot`, `extract_content`                        |
| **JavaScript**  | `evaluate_js`                                                            |
| **Tabs**        | `new_tab`, `switch_tab`, `close_tab`, `list_tabs`                        |
| **Completion**  | `done`                                                                   |

---

## 📊 Comparison with Browser-Use (Python)

BUA is inspired by the popular [browser-use](https://github.com/browser-use/browser-use) Python library. Here's how they
compare:

| Aspect             | Browser-Use (Python)           | BUA (Go)                             |
|--------------------|--------------------------------|--------------------------------------|
| **Language**       | Python 3.11+                   | Go 1.25+                             |
| **LLM Support**    | OpenAI, Claude, Gemini, Ollama | Gemini (via ADK), other models soon. |
| **Browser Engine** | Playwright                     | go-rod (CDP direct)                  |
| **Performance**    | Good                           | Excellent (compiled, no runtime)     |
| **Deployment**     | Python environment             | Single binary                        |
| **Memory**         | Higher (Python + Node.js)      | Lower (native Go)                    |
| **Concurrency**    | asyncio                        | Native goroutines                    |
| **Anti-Detection** | ✅                              | ✅                                    |
| **Vision Support** | ✅                              | ✅                                    |
| **Custom Tools**   | ✅                              | ✅                                    |

### Why Choose BUA?

- **🚀 Performance**: Go's compiled nature means faster startup and lower memory
- **📦 Simple Deployment**: Single binary, no Python/Node.js dependencies
- **⚡ Concurrency**: Native goroutines for parallel operations
- **🔒 Type Safety**: Catch errors at compile time
- **🏢 Enterprise Ready**: Common choice for production services

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [browser-use](https://github.com/browser-use/browser-use) - The original Python inspiration
- [go-rod](https://github.com/go-rod/rod) - Excellent Go CDP implementation
- [Google ADK](https://google.github.io/adk-docs/) - Agent Development Kit for Go
- [Google Gemini](https://ai.google.dev/) - Powerful multimodal LLM

---

<p align="center">
  <strong>Built with ❤️ for the Go community</strong>
</p>

<p align="center">
  <a href="https://github.com/anxuanzi/bua/issues">Report Bug</a> •
  <a href="https://github.com/anxuanzi/bua/issues">Request Feature</a>
</p>
