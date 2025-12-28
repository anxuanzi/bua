// Package agent provides the LLM agent loop and token management.
package agent

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TokenCounter provides token counting and budget management.
type TokenCounter struct {
	maxTokens  int
	usedTokens int

	// Approximate token ratios for different content types
	// These are rough estimates based on typical tokenization
	charsPerToken    float64 // Average characters per token for text
	imageTokensBase  int     // Base tokens for an image
	imageTokensPerPx float64 // Additional tokens per 1000 pixels
}

// NewTokenCounter creates a new token counter with the given max tokens.
func NewTokenCounter(maxTokens int) *TokenCounter {
	return &TokenCounter{
		maxTokens:        maxTokens,
		charsPerToken:    4.0,    // GPT-style average
		imageTokensBase:  85,     // Base cost for an image
		imageTokensPerPx: 0.0017, // ~170 tokens per 100k pixels for high detail
	}
}

// EstimateTextTokens estimates the number of tokens in a text string.
func (tc *TokenCounter) EstimateTextTokens(text string) int {
	if text == "" {
		return 0
	}

	// Count characters (not bytes)
	charCount := utf8.RuneCountInString(text)

	// Rough estimate: divide by average chars per token
	tokens := float64(charCount) / tc.charsPerToken

	// Add overhead for special tokens, formatting
	tokens *= 1.1

	return int(tokens)
}

// EstimateImageTokens estimates tokens for an image based on dimensions.
func (tc *TokenCounter) EstimateImageTokens(width, height int) int {
	if width <= 0 || height <= 0 {
		return tc.imageTokensBase
	}

	pixels := width * height
	// For high detail images
	tokens := tc.imageTokensBase + int(float64(pixels)*tc.imageTokensPerPx/1000)

	return tokens
}

// Add adds tokens to the used count.
func (tc *TokenCounter) Add(tokens int) {
	tc.usedTokens += tokens
}

// AddText estimates and adds tokens for text.
func (tc *TokenCounter) AddText(text string) int {
	tokens := tc.EstimateTextTokens(text)
	tc.usedTokens += tokens
	return tokens
}

// AddImage estimates and adds tokens for an image.
func (tc *TokenCounter) AddImage(width, height int) int {
	tokens := tc.EstimateImageTokens(width, height)
	tc.usedTokens += tokens
	return tokens
}

// Reset resets the used tokens count.
func (tc *TokenCounter) Reset() {
	tc.usedTokens = 0
}

// Used returns the number of tokens used.
func (tc *TokenCounter) Used() int {
	return tc.usedTokens
}

// Available returns the number of tokens available.
func (tc *TokenCounter) Available() int {
	return tc.maxTokens - tc.usedTokens
}

// UsagePercent returns the percentage of tokens used.
func (tc *TokenCounter) UsagePercent() float64 {
	if tc.maxTokens == 0 {
		return 0
	}
	return float64(tc.usedTokens) / float64(tc.maxTokens) * 100
}

// NeedsCompaction returns true if token usage is high enough to warrant compaction.
func (tc *TokenCounter) NeedsCompaction() bool {
	return tc.UsagePercent() > 75
}

// CanFit returns true if the estimated tokens can fit in the remaining budget.
func (tc *TokenCounter) CanFit(estimatedTokens int) bool {
	return tc.usedTokens+estimatedTokens <= tc.maxTokens
}

// ConversationCompactor handles conversation compaction to reduce token usage.
type ConversationCompactor struct {
	counter         *TokenCounter
	compactionRatio float64
}

// NewConversationCompactor creates a new conversation compactor.
func NewConversationCompactor(counter *TokenCounter) *ConversationCompactor {
	return &ConversationCompactor{
		counter:         counter,
		compactionRatio: 0.3, // Aim to reduce to 30% of original size
	}
}

// CompactMessages compacts a list of messages to reduce token count.
func (cc *ConversationCompactor) CompactMessages(messages []Message) []Message {
	if len(messages) <= 3 {
		return messages
	}

	// Strategy: Keep first (system) and last few messages, summarize middle
	keepFirst := 1 // System message
	keepLast := 3  // Recent context

	if len(messages) <= keepFirst+keepLast {
		return messages
	}

	result := make([]Message, 0, keepFirst+1+keepLast)

	// Keep first messages
	result = append(result, messages[:keepFirst]...)

	// Summarize middle messages
	middleStart := keepFirst
	middleEnd := len(messages) - keepLast
	middle := messages[middleStart:middleEnd]

	summary := cc.summarizeMessages(middle)
	result = append(result, Message{
		Role:    "system",
		Content: fmt.Sprintf("[Previous conversation summary: %s]", summary),
	})

	// Keep last messages
	result = append(result, messages[len(messages)-keepLast:]...)

	return result
}

// summarizeMessages creates a summary of messages.
func (cc *ConversationCompactor) summarizeMessages(messages []Message) string {
	var sb strings.Builder

	actionCount := 0
	var pages []string
	pageSet := make(map[string]bool)

	for _, msg := range messages {
		// Track pages visited
		if strings.Contains(msg.Content, "Navigated to") || strings.Contains(msg.Content, "Page:") {
			// Extract page info
			parts := strings.Split(msg.Content, "\n")
			for _, part := range parts {
				if strings.HasPrefix(part, "Page:") || strings.HasPrefix(part, "URL:") {
					if !pageSet[part] {
						pageSet[part] = true
						pages = append(pages, part)
					}
				}
			}
		}

		// Count actions
		if msg.Role == "assistant" && (strings.Contains(msg.Content, "click") ||
			strings.Contains(msg.Content, "type") ||
			strings.Contains(msg.Content, "scroll")) {
			actionCount++
		}
	}

	sb.WriteString(fmt.Sprintf("%d actions taken. ", actionCount))

	if len(pages) > 0 {
		sb.WriteString("Pages visited: ")
		for i, page := range pages {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(truncateSummary(page, 50))
		}
		sb.WriteString(". ")
	}

	return sb.String()
}

func truncateSummary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Message represents a conversation message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant", "tool"
	Content string `json:"content"`
}

// TokenStats holds token usage statistics.
type TokenStats struct {
	MaxTokens       int
	UsedTokens      int
	AvailableTokens int
	UsagePercent    float64
	NeedsCompaction bool
}

// Stats returns current token statistics.
func (tc *TokenCounter) Stats() TokenStats {
	return TokenStats{
		MaxTokens:       tc.maxTokens,
		UsedTokens:      tc.usedTokens,
		AvailableTokens: tc.Available(),
		UsagePercent:    tc.UsagePercent(),
		NeedsCompaction: tc.NeedsCompaction(),
	}
}
