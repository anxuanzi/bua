package agent

import (
	"strings"
	"unicode"
)

// TokenCounter provides token estimation for text content.
// This is a simple approximation - actual token counts depend on the specific tokenizer.
type TokenCounter struct {
	// Average characters per token (approximation for most LLMs)
	charsPerToken float64
}

// NewTokenCounter creates a new token counter with default settings.
func NewTokenCounter() *TokenCounter {
	return &TokenCounter{
		charsPerToken: 4.0, // Reasonable approximation for English text
	}
}

// EstimateTokens estimates the number of tokens in a text string.
// This uses a simple character-based heuristic:
// - ~4 characters per token for English text
// - Adjustments for whitespace, punctuation, and special characters
func (tc *TokenCounter) EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Count different character types for better estimation
	var alphaCount, digitCount, whitespaceCount, punctCount, otherCount int

	for _, r := range text {
		switch {
		case unicode.IsLetter(r):
			alphaCount++
		case unicode.IsDigit(r):
			digitCount++
		case unicode.IsSpace(r):
			whitespaceCount++
		case unicode.IsPunct(r):
			punctCount++
		default:
			otherCount++
		}
	}

	// Words typically become 1-2 tokens
	wordCount := len(strings.Fields(text))

	// Punctuation often becomes separate tokens
	// Digits in groups typically share tokens
	// Adjust estimation based on content type

	// Base estimate from character count
	charEstimate := float64(len(text)) / tc.charsPerToken

	// Adjust based on word count (minimum of 1 token per word for short words)
	wordEstimate := float64(wordCount) * 1.3

	// Punctuation tends to create extra tokens
	punctAdjust := float64(punctCount) * 0.5

	// Take the higher of char or word estimate, add punctuation adjustment
	estimate := charEstimate
	if wordEstimate > estimate {
		estimate = wordEstimate
	}
	estimate += punctAdjust

	return int(estimate + 0.5) // Round to nearest integer
}

// EstimateFromElements estimates tokens for element descriptions.
// Elements have a predictable format so we can be more precise.
func (tc *TokenCounter) EstimateFromElements(elementCount int, avgTextLength int) int {
	// Each element has: index, tag, attributes, text
	// Format: "[0] button text='Click me' (100,200)"
	// Average ~15-25 tokens per element depending on content
	baseTokens := 15
	textTokens := avgTextLength / 4 // ~4 chars per token for text

	return elementCount * (baseTokens + textTokens)
}

// TruncateToTokenLimit truncates text to fit within a token limit.
// Returns the truncated text and whether truncation occurred.
func (tc *TokenCounter) TruncateToTokenLimit(text string, maxTokens int) (string, bool) {
	currentTokens := tc.EstimateTokens(text)
	if currentTokens <= maxTokens {
		return text, false
	}

	// Estimate character limit
	targetChars := int(float64(maxTokens) * tc.charsPerToken * 0.9) // 90% to be safe

	if targetChars >= len(text) {
		return text, false
	}

	// Truncate at word boundary if possible
	truncated := text[:targetChars]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > targetChars/2 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "...", true
}

// TokenBudget helps manage token allocation across message components.
type TokenBudget struct {
	Total     int
	System    int
	History   int
	PageState int
	Reserved  int
}

// NewTokenBudget creates a budget for a given model's context window.
func NewTokenBudget(contextWindow int) *TokenBudget {
	// Reserve some tokens for response and safety margin
	reserved := contextWindow / 10 // 10% reserved

	available := contextWindow - reserved

	return &TokenBudget{
		Total:     contextWindow,
		System:    available / 4, // 25% for system prompt
		History:   available / 4, // 25% for history
		PageState: available / 2, // 50% for page state and task
		Reserved:  reserved,
	}
}

// Available returns tokens available after subtracting used tokens.
func (b *TokenBudget) Available(used int) int {
	return b.Total - b.Reserved - used
}

// Common context window sizes for reference
const (
	ContextGeminiFlash = 1000000 // Gemini 2.0 Flash: 1M tokens
	ContextGeminiPro   = 2000000 // Gemini 1.5 Pro: 2M tokens
	ContextGPT4        = 128000  // GPT-4 Turbo: 128K tokens
	ContextClaude      = 200000  // Claude 3: 200K tokens
)
