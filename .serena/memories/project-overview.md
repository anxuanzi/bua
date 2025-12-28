# bua-go Project Overview

## Repository
- GitHub: https://github.com/anxuanzi/bua-go
- Module: `github.com/anxuanzi/bua-go`
- Go Version: 1.25

## Description
Go library for browser automation powered by LLMs via **Google ADK (Agent Development Kit)**. Uses vision + DOM hybrid approach combining screenshots with parsed HTML and accessibility trees.

## Architecture - ADK Integration
The project uses Google ADK for Go (`google.golang.org/adk`) as the agent framework:
- `llmagent.New()` - Creates LLM agents with tools
- `functiontool.New[TArgs, TResults]()` - Defines browser action tools
- `runner.New()` + `runner.Run()` - Programmatic agent execution
- `session.InMemoryService()` - Session management
- `gemini.NewModel()` - Gemini model creation

## Package Structure
```
bua.go           - Main public API (Agent, Config, Run)
                   Uses adkagent alias to avoid conflict with local agent package
agent/
  agent.go       - ADK BrowserAgent with llmagent.New + functiontool.New
  prompts.go     - System prompts for browser automation
browser/         - Rod browser wrapper + CDP operations
dom/             - DOM extraction + accessibility tree + element map
memory/          - Short-term + long-term memory system  
screenshot/      - Screenshot capture + element annotation (fogleman/gg)
tools/           - Legacy tool definitions (kept for reference)
examples/        - Usage examples (simple, scraping)
```

## Key Dependencies
- `google.golang.org/adk v0.3.0` - Agent Development Kit
- `google.golang.org/genai` - Gemini API client
- `github.com/go-rod/rod` - Browser automation
- `github.com/fogleman/gg` - Image annotation

## ADK Tool Pattern
All browser tools follow ADK's functiontool signature:
```go
handler := func(ctx tool.Context, input InputType) (OutputType, error) {
    // implementation
    return OutputType{...}, nil
}
tool, err := functiontool.New(functiontool.Config{
    Name: "tool_name",
    Description: "...",
}, handler)
```

## Browser Tools (9 tools defined in agent/agent.go)
- click - Click element by index
- type_text - Type into input fields
- scroll - Scroll page up/down
- navigate - Go to URL
- wait - Wait for page stability
- extract - Extract data from elements
- get_page_state - Get URL, title, element map
- request_human_takeover - Request human help
- done - Signal task completion

## Environment
- GOOGLE_API_KEY - Required for Gemini API access

## Testing
- All tests pass (13 tests across 4 packages)
- Main test files: bua_test.go, dom/element_test.go, memory/memory_test.go