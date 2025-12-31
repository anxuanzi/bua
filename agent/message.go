package agent

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anxuanzi/bua/dom"
)

// MessageManager handles conversation state and message construction for the LLM.
type MessageManager struct {
	systemPrompt    string
	history         *AgentHistory
	sensitiveFilter *SensitiveDataFilter
	maxElements     int
	useVision       bool
}

// MessageManagerConfig configures the message manager.
type MessageManagerConfig struct {
	MaxHistoryItems int
	MaxElements     int
	UseVision       bool
}

// NewMessageManager creates a new message manager.
func NewMessageManager(cfg MessageManagerConfig) *MessageManager {
	maxHistory := cfg.MaxHistoryItems
	if maxHistory <= 0 {
		maxHistory = 20
	}

	maxElements := cfg.MaxElements
	if maxElements <= 0 {
		maxElements = 100
	}

	return &MessageManager{
		systemPrompt:    SystemPrompt(),
		history:         NewAgentHistory(maxHistory),
		sensitiveFilter: NewSensitiveDataFilter(),
		maxElements:     maxElements,
		useVision:       cfg.UseVision,
	}
}

// GetSystemPrompt returns the system prompt.
func (m *MessageManager) GetSystemPrompt() string {
	return m.systemPrompt
}

// SetTask sets the current task.
func (m *MessageManager) SetTask(task string) {
	m.history.SetTask(task)
}

// AddHistoryItem adds an item to the execution history.
func (m *MessageManager) AddHistoryItem(item HistoryItem) {
	m.history.AddItem(item)
}

// GetHistory returns the agent history.
func (m *MessageManager) GetHistory() *AgentHistory {
	return m.history
}

// BuildStateMessage builds the current state message for the LLM.
func (m *MessageManager) BuildStateMessage(elementMap *dom.ElementMap, lastActionResult string, screenshotIncluded bool) string {
	var sb strings.Builder

	// Add current page state
	if elementMap != nil {
		pageState := BuildPageStatePrompt(
			elementMap.PageURL,
			elementMap.PageTitle,
			elementMap.ToTokenStringLimited(m.maxElements),
			screenshotIncluded,
		)
		sb.WriteString(pageState)
		sb.WriteString("\n\n")
	}

	// Add history if we have previous steps
	if m.history.StepCount() > 0 {
		sb.WriteString(m.history.ToDescription())
		sb.WriteString("\n\n")
	}

	// Add last action result if provided
	if lastActionResult != "" {
		sb.WriteString("<last_action_result>\n")
		sb.WriteString(lastActionResult)
		sb.WriteString("\n</last_action_result>\n\n")
	}

	// Add context about consecutive failures if any
	failures := m.history.GetConsecutiveFailures()
	if failures >= 2 {
		sb.WriteString(fmt.Sprintf("<warning>You have had %d consecutive failed actions. Consider trying a different approach.</warning>\n\n", failures))
	}

	// Add step info
	sb.WriteString(fmt.Sprintf("<step_info>This is step %d</step_info>", m.history.StepCount()+1))

	return sb.String()
}

// BuildInitialTaskMessage builds the initial message with task description.
func (m *MessageManager) BuildInitialTaskMessage(task string, elementMap *dom.ElementMap) string {
	var sb strings.Builder

	// Add task
	sb.WriteString(BuildTaskPrompt(task))
	sb.WriteString("\n\n")

	// Add initial page state if available
	if elementMap != nil {
		pageState := BuildPageStatePrompt(
			elementMap.PageURL,
			elementMap.PageTitle,
			elementMap.ToTokenStringLimited(m.maxElements),
			false,
		)
		sb.WriteString(pageState)
	}

	return sb.String()
}

// BuildContinuationMessage builds a message for continuing after an action.
func (m *MessageManager) BuildContinuationMessage(elementMap *dom.ElementMap, actionName, actionResult string, success bool) string {
	var sb strings.Builder

	// Update the last history item with results
	m.history.UpdateLastItem(actionResult, success)

	// Build state message
	sb.WriteString(m.BuildStateMessage(elementMap, actionResult, false))
	sb.WriteString("\n\n")

	// Add continuation instruction
	sb.WriteString("<instruction>Continue with the task. Analyze the result and take the next action.</instruction>")

	return sb.String()
}

// BuildErrorRecoveryMessage builds a message for recovering from an error.
func (m *MessageManager) BuildErrorRecoveryMessage(elementMap *dom.ElementMap, errorMsg string) string {
	var sb strings.Builder

	// Add page state
	if elementMap != nil {
		pageState := BuildPageStatePrompt(
			elementMap.PageURL,
			elementMap.PageTitle,
			elementMap.ToTokenStringLimited(m.maxElements),
			false,
		)
		sb.WriteString(pageState)
		sb.WriteString("\n\n")
	}

	// Add error recovery prompt
	sb.WriteString(BuildErrorRecoveryPrompt(errorMsg))

	return sb.String()
}

// FilterSensitiveData filters sensitive data from a message.
func (m *MessageManager) FilterSensitiveData(message string) string {
	return m.sensitiveFilter.Filter(message)
}

// Clear resets the message manager state.
func (m *MessageManager) Clear() {
	m.history.Clear()
}

// SensitiveDataFilter filters sensitive data from messages.
type SensitiveDataFilter struct {
	patterns map[string]*regexp.Regexp
}

// NewSensitiveDataFilter creates a new sensitive data filter.
func NewSensitiveDataFilter() *SensitiveDataFilter {
	return &SensitiveDataFilter{
		patterns: map[string]*regexp.Regexp{
			"api_key":     regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?([a-zA-Z0-9_-]{20,})["']?`),
			"password":    regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?([^\s"']{4,})["']?`),
			"token":       regexp.MustCompile(`(?i)(token|bearer|auth)\s*[:=]\s*["']?([a-zA-Z0-9_.-]{20,})["']?`),
			"secret":      regexp.MustCompile(`(?i)(secret|private[_-]?key)\s*[:=]\s*["']?([a-zA-Z0-9_/+=]{16,})["']?`),
			"credit_card": regexp.MustCompile(`\b(?:\d{4}[- ]?){3}\d{4}\b`),
			"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		},
	}
}

// Filter replaces sensitive data with placeholders.
func (f *SensitiveDataFilter) Filter(text string) string {
	result := text

	for name, pattern := range f.patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return fmt.Sprintf("<secret type=\"%s\">[REDACTED]</secret>", name)
		})
	}

	return result
}

// AddPattern adds a custom sensitive data pattern.
func (f *SensitiveDataFilter) AddPattern(name, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	f.patterns[name] = re
	return nil
}
