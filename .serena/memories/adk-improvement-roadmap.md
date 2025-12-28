# BUA-GO ADK Optimization Roadmap

## Analysis Date: 2024-12-28

## Current ADK Usage (Correct)
- `session.InMemoryService()` ✅
- `session.Create()` ✅
- `runner.New()/Run()` ✅
- `llmagent.New()` ✅
- `functiontool.New()` ✅
- `gemini.NewModel()` ✅

## Gaps Identified

### 🔴 Priority 1: Memory Service
- **Gap**: Custom `memory/memory.go` (390 LOC) instead of ADK's `memory.InMemoryService()`
- **Fix**: Replace with ADK memory service in runner config
- **Benefit**: Semantic search, automatic consolidation, less maintenance

### 🔴 Priority 2: ADK Callbacks
- **Gap**: Manual `preAction()`/`postAction()` in agent/agent.go:120-185
- **Fix**: Use `BeforeToolCallback`/`AfterToolCallback` in llmagent.Config
- **Benefit**: Proper ADK integration, automatic context propagation

### 🔴 Priority 3: Artifact Service
- **Gap**: Custom file-based screenshot storage
- **Fix**: Use `artifact.InMemoryService()` with `ctx.SaveArtifact()`
- **Benefit**: Unified storage for screenshots AND downloads

### 🟡 Priority 4: Session State
- **Gap**: Not using `session.State` for persistence
- **Fix**: Store current_url, nav_history, element_count in state
- **Benefit**: Cross-turn persistence, resumable sessions

### 🟡 Priority 5: Multi-Tab Architecture
- **Gap**: Single `page *rod.Page` in browser/browser.go:30
- **Fix**: Change to `pages map[string]*rod.Page`, add 4 new tools
- **New Tools**: new_tab, switch_tab, close_tab, list_tabs

### 🟡 Priority 6: Download Capability
- **Gap**: No download handling
- **Fix**: CDP BrowserSetDownloadBehavior + artifact storage
- **New Tool**: download_file

### 🟢 Priority 7: Dual-Use Architecture
- **Gap**: Library only, not usable as ADK tool
- **Fix**: Add `AgentTool` wrapper in export/adktool.go
- **Benefit**: Can be used in other ADK applications

## Implementation Phases

### Phase 1: ADK Component Adoption (P1-P3)
- Replace custom memory with ADK memory service
- Implement ADK callbacks
- Use artifacts for screenshots

### Phase 2: Feature Expansion (P4-P6)
- Add session state usage
- Implement multi-tab support
- Add download capability

### Phase 3: Dual-Use Export (P7)
- Create AgentTool wrapper
- Document integration patterns

## Key Code Locations
- bua.go:241 - Custom memory initialization
- bua.go:277-286 - Runner config (add MemoryService, ArtifactService)
- agent/agent.go:99-112 - llmagent config (add callbacks)
- agent/agent.go:120-185 - Manual pre/post actions (replace with callbacks)
- browser/browser.go:30 - Single page (change to map for multi-tab)
