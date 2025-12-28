# Directory Structure

```
bua-go/
├── bua.go                    # Main public API
├── bua_test.go               # Main package tests
├── go.mod                    # Go 1.25, ADK v0.3.0
├── go.sum
├── agent/
│   ├── agent.go              # ADK BrowserAgent + tools
│   └── prompts.go            # System prompts
├── browser/
│   └── browser.go            # Rod wrapper, CDP ops
├── dom/
│   ├── accessibility.go      # Accessibility tree
│   ├── element.go            # ElementMap, BoundingBox
│   └── element_test.go
├── memory/
│   ├── memory.go             # Short/long-term memory
│   └── memory_test.go
├── screenshot/
│   └── screenshot.go         # Capture + annotation
├── tools/
│   └── tools.go              # Legacy tool definitions
└── examples/
    ├── simple/
    │   └── main.go           # Basic usage example
    └── scraping/
        └── main.go           # Web scraping example
```

## Key Files by Function

### Public API
- `bua.go` - Agent, Config, New(), Start(), Run(), Navigate(), Screenshot()

### LLM Integration
- `agent/agent.go` - BrowserAgent, ADK tools, Init()
- `agent/prompts.go` - SystemPrompt(), tool descriptions

### Browser Control
- `browser/browser.go` - Click, Type, Scroll, Navigate, GetElementMap

### Data Extraction
- `dom/element.go` - ElementMap, Element, BoundingBox
- `dom/accessibility.go` - AccessibilityTree, MergeWithElementMap

### Visual Processing
- `screenshot/screenshot.go` - Annotate(), Save()

### Memory System
- `memory/memory.go` - Manager, Observation, LongTermMemory
