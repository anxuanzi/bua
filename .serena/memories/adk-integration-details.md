# ADK Integration Details

## Migration Summary (Completed)
Successfully migrated from raw Gemini API to Google ADK framework.

## Key Changes Made

### go.mod
- Go version: 1.25
- Added: `google.golang.org/adk v0.3.0`

### bua.go
- Added imports:
  - `adkagent "google.golang.org/adk/agent"` (aliased to avoid conflict)
  - `"google.golang.org/adk/runner"`
  - `"google.golang.org/adk/session"`
- Agent struct now includes:
  - `browserAgent *agent.BrowserAgent`
  - `runner *runner.Runner`
- Start() creates ADK runner with session.InMemoryService()
- Run() uses `r.Run(ctx, "user", sessionID, userMessage, adkagent.RunConfig{})`

### agent/agent.go
- BrowserAgent.Init() creates:
  - Gemini model via `gemini.NewModel()`
  - ADK agent via `llmagent.New()`
- createBrowserTools() defines 9 tools using `functiontool.New[TArgs, TResults]()`
- All handlers return `(Output, error)` per ADK requirement

## ADK Function Tool Signature
```go
type Func[TArgs, TResults any] func(tool.Context, TArgs) (TResults, error)
```

## Important: Handler Return Type
Handlers MUST return `(TResults, error)`, not just `TResults`.
Example:
```go
clickHandler := func(ctx tool.Context, input ClickInput) (ClickOutput, error) {
    return ClickOutput{Success: true, Message: "..."}, nil
}
```

## Import Alias Pattern
Local `agent` package conflicts with ADK's `agent` package.
Solution: `adkagent "google.golang.org/adk/agent"`

## Dual-Use Architecture (export/adktool.go)

### BrowserTool - Single Browser Instance
```go
browserTool := export.NewBrowserTool(&export.BrowserToolConfig{
    APIKey:          apiKey,
    Model:           "gemini-3-flash-preview",
    Headless:        false,
    ShowAnnotations: true,
})
defer browserTool.Close()

adkTool, _ := browserTool.Tool()  // Returns tool.Tool
```

### MultiBrowserTool - Parallel Browsers
```go
multiBrowserTool := export.NewMultiBrowserTool(&export.MultiBrowserToolConfig{
    BrowserToolConfig:     export.DefaultBrowserToolConfig(),
    MaxConcurrentBrowsers: 3,
})
defer multiBrowserTool.Close()

adkTool, _ := multiBrowserTool.Tool()  // Returns tool.Tool
// Actions: create, execute, close, list
```

### Integration with Other ADK Agents
```go
researchAgent, _ := llmagent.New(llmagent.Config{
    Name:        "research_agent",
    Model:       model,
    Tools:       []tool.Tool{browserTool},  // Add bua-go as tool
    Instruction: "Use browser_automation for web tasks...",
})
```

## Download Capability (browser/download.go)

### download_file Tool
```go
type DownloadFileInput struct {
    URL         string `json:"url"`
    Filename    string `json:"filename,omitempty"`
    UsePageAuth bool   `json:"use_page_auth,omitempty"`
    Reasoning   string `json:"reasoning"`
}
```

### Download Methods
- `DownloadFile()` - Direct HTTP download
- `DownloadResource()` - CDP download with page auth context
- Downloads saved to `~/.bua/downloads/`
