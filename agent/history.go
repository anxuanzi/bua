package agent

import (
	"fmt"
	"strings"
	"time"
)

// HistoryItem represents a single step in the agent's execution history.
type HistoryItem struct {
	StepNumber    int       `json:"step_number"`
	Timestamp     time.Time `json:"timestamp"`
	Thinking      string    `json:"thinking,omitempty"`
	Evaluation    string    `json:"evaluation,omitempty"`
	Memory        string    `json:"memory,omitempty"`
	NextGoal      string    `json:"next_goal,omitempty"`
	ActionName    string    `json:"action_name"`
	ActionParams  string    `json:"action_params,omitempty"`
	ActionResult  string    `json:"action_result,omitempty"`
	ActionSuccess bool      `json:"action_success"`
	DurationMs    int64     `json:"duration_ms"`
}

// AgentHistory manages the execution history of an agent.
type AgentHistory struct {
	items           []HistoryItem
	maxItems        int
	currentMemory   string
	taskDescription string
}

// NewAgentHistory creates a new agent history manager.
func NewAgentHistory(maxItems int) *AgentHistory {
	if maxItems <= 0 {
		maxItems = 20 // Default to keeping 20 history items
	}
	return &AgentHistory{
		items:    make([]HistoryItem, 0),
		maxItems: maxItems,
	}
}

// SetTask sets the task description for context.
func (h *AgentHistory) SetTask(task string) {
	h.taskDescription = task
}

// AddItem adds a new history item.
func (h *AgentHistory) AddItem(item HistoryItem) {
	h.items = append(h.items, item)

	// Update current memory if provided
	if item.Memory != "" {
		h.currentMemory = item.Memory
	}
}

// GetItems returns all history items.
func (h *AgentHistory) GetItems() []HistoryItem {
	return h.items
}

// GetLastItem returns the most recent history item, or nil if empty.
func (h *AgentHistory) GetLastItem() *HistoryItem {
	if len(h.items) == 0 {
		return nil
	}
	return &h.items[len(h.items)-1]
}

// UpdateLastItem updates the result and success status of the last history item.
func (h *AgentHistory) UpdateLastItem(result string, success bool) {
	if len(h.items) == 0 {
		return
	}
	h.items[len(h.items)-1].ActionResult = result
	h.items[len(h.items)-1].ActionSuccess = success
}

// GetCurrentMemory returns the accumulated memory from history.
func (h *AgentHistory) GetCurrentMemory() string {
	return h.currentMemory
}

// StepCount returns the number of steps executed.
func (h *AgentHistory) StepCount() int {
	return len(h.items)
}

// Clear resets the history.
func (h *AgentHistory) Clear() {
	h.items = make([]HistoryItem, 0)
	h.currentMemory = ""
}

// ToDescription generates a text description of the history for the LLM.
// It implements truncation: keeps first item + most recent items within maxItems.
func (h *AgentHistory) ToDescription() string {
	if len(h.items) == 0 {
		return "No previous actions taken yet."
	}

	var sb strings.Builder
	sb.WriteString("<agent_history>\n")

	// Determine which items to include
	itemsToShow := h.items
	omittedCount := 0

	if len(h.items) > h.maxItems {
		// Keep first item + most recent (maxItems-1) items
		firstItem := h.items[0]
		recentItems := h.items[len(h.items)-(h.maxItems-1):]
		omittedCount = len(h.items) - h.maxItems

		itemsToShow = make([]HistoryItem, 0, h.maxItems)
		itemsToShow = append(itemsToShow, firstItem)
		itemsToShow = append(itemsToShow, recentItems...)
	}

	for i, item := range itemsToShow {
		// Add omission marker after first item if we truncated
		if i == 1 && omittedCount > 0 {
			sb.WriteString(fmt.Sprintf("\n... (%d steps omitted) ...\n\n", omittedCount))
		}

		sb.WriteString(fmt.Sprintf("<step number=\"%d\">\n", item.StepNumber))

		if item.ActionName != "" {
			status := "✓"
			if !item.ActionSuccess {
				status = "✗"
			}
			sb.WriteString(fmt.Sprintf("  <action status=\"%s\">%s</action>\n", status, item.ActionName))
		}

		if item.ActionResult != "" {
			// Truncate long results
			result := item.ActionResult
			if len(result) > 200 {
				result = result[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("  <result>%s</result>\n", result))
		}

		if item.Evaluation != "" {
			sb.WriteString(fmt.Sprintf("  <evaluation>%s</evaluation>\n", item.Evaluation))
		}

		if item.NextGoal != "" {
			sb.WriteString(fmt.Sprintf("  <goal>%s</goal>\n", item.NextGoal))
		}

		sb.WriteString("</step>\n")
	}

	// Add current memory if available
	if h.currentMemory != "" {
		sb.WriteString(fmt.Sprintf("\n<accumulated_memory>%s</accumulated_memory>\n", h.currentMemory))
	}

	sb.WriteString("</agent_history>")

	return sb.String()
}

// GetSuccessRate calculates the success rate of actions.
func (h *AgentHistory) GetSuccessRate() float64 {
	if len(h.items) == 0 {
		return 0
	}

	successCount := 0
	for _, item := range h.items {
		if item.ActionSuccess {
			successCount++
		}
	}

	return float64(successCount) / float64(len(h.items))
}

// GetConsecutiveFailures returns the number of consecutive failures at the end.
func (h *AgentHistory) GetConsecutiveFailures() int {
	failures := 0
	for i := len(h.items) - 1; i >= 0; i-- {
		if !h.items[i].ActionSuccess {
			failures++
		} else {
			break
		}
	}
	return failures
}
