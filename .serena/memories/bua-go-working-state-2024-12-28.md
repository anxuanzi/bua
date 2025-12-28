# bua-go Working State - December 28, 2024

## Model Configuration
**IMPORTANT**: Always use `gemini-3-flash-preview` model
- Input token limit: 1,048,576
- Output token limit: 65,536
- Default set in `bua.go` line 144

## Features Working

### 1. Browser Automation
- ✅ Navigate to URLs
- ✅ Click elements by index
- ✅ Type text into inputs
- ✅ Scroll up/down
- ✅ Wait for page stability
- ✅ Extract data from pages
- ✅ Request human takeover (CAPTCHA handling)
- ✅ Done action with extracted data

### 2. Annotations System
- ✅ `ShowAnnotations bool` in `bua.Config`
- ✅ Visual overlays displayed before each action (click, type, scroll)
- ✅ Colored boxes with element indices in browser
- ✅ Auto-hide after action completes

### 3. Screenshot System
- ✅ Annotated screenshots saved for each step
- ✅ Format: `step_XXX_action_HHMMSS.png`
- ✅ Location: `~/.bua/screenshots/steps/`
- ✅ Shows element overlays in screenshots

### 4. Logging System (agent/logger.go)
- ✅ Structured output with emoji indicators
- ✅ Step counter with timestamps
- ✅ Box formatting for actions:
  - 🎯 STEP indicator
  - 🔧 Action name
  - 🎪 Target element
  - 💭 Reasoning
- ✅ Status indicators: ✅ success, ❌ failure, ⏳ waiting
- ✅ Page state logging: 📄 title, 🔗 URL, 🧩 element count

### 5. Examples
All examples use:
- godotenv for `.env` file loading
- `gemini-3-flash-preview` model
- `Debug: true` for verbose output
- `ShowAnnotations: true` for visual debugging

Location: `examples/simple/`, `examples/scraping/`, `examples/multipage/`

## Key Files

### Configuration
- `bua.go` - Main agent config, ShowAnnotations flag
- `agent/agent.go` - ADK integration, preAction/postAction hooks
- `agent/logger.go` - Emoji-based structured logging

### Browser
- `browser/browser.go` - go-rod wrapper
- `browser/annotation.go` - JavaScript injection for visual overlays

### DOM
- `dom/element.go` - Element mapping and extraction
- `dom/accessibility.go` - Accessibility tree parsing

## Viewport Configuration
- DesktopViewport: 1280x800 (default)
- LargeDesktopViewport: 1920x1080
- TabletViewport: 768x1024
- MobileViewport: 375x812

**Important**: Both window-size AND viewport must match for responsive sites to work correctly.

## Environment
API key stored in `.env` at project root:
```
GOOGLE_API_KEY=your-key-here
```

Examples load with: `godotenv.Load("../../.env")`

## Testing Results (Dec 28, 2024)

### Scraping Example
- ✅ Successfully extracted top 5 Hacker News stories
- ✅ Returned JSON with titles and URLs
- ✅ Task completed in ~9 seconds

### Simple Example
- ✅ Navigated to Google
- ✅ Typed search query
- ✅ Detected CAPTCHA and requested human takeover
- ✅ Intelligently switched to DuckDuckGo
- ✅ Completed search and clicked result
- ✅ Screenshots saved for each step

### Download Example (Dec 29, 2024)
- ✅ Downloaded Go logo from go.dev (go-logo-white.svg, 1472 bytes)
- ✅ Downloaded Rust logo from rust-lang.org (rust-logo-blk.svg, 2396 bytes)
- ✅ Files saved to ~/.bua/downloads/

### ADK Tool Example (Dec 29, 2024)
- ✅ BrowserTool integrated as ADK tool in external agent
- ✅ Research agent used browser_automation tool successfully
- ✅ Extracted Go tagline: "Build simple, secure, scalable systems with Go"

## New Features (Dec 29, 2024)

### 6. Download Capability (browser/download.go)
- ✅ `download_file` tool for programmatic downloads
- ✅ Direct HTTP downloads (use_page_auth=false)
- ✅ CDP resource download with page context (use_page_auth=true)
- ✅ Auto-generated filenames from URL
- ✅ Downloads saved to ~/.bua/downloads/

### 7. Dual-Use Architecture (export/adktool.go)
- ✅ BrowserTool wrapper for single browser instance
- ✅ MultiBrowserTool for parallel browser management
- ✅ SimpleBrowserTask convenience function
- ✅ Can be embedded in other ADK applications as a tool
- ✅ Actions: create, execute, close, list browsers
