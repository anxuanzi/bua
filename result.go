package bua

import "time"

// Result represents the outcome of a task execution.
type Result struct {
	// Success indicates whether the task completed successfully.
	Success bool

	// Data contains the extracted data or task output.
	// The type depends on what the agent was asked to do.
	Data any

	// Error contains the error message if Success is false.
	Error string

	// Steps contains the sequence of actions taken during execution.
	Steps []Step

	// Duration is the total execution time.
	Duration time.Duration

	// TokensUsed is the approximate number of tokens consumed.
	TokensUsed int

	// ScreenshotPaths contains paths to saved screenshots.
	ScreenshotPaths []string
}

// Step represents a single action in the execution sequence.
type Step struct {
	// Number is the step index (1-based).
	Number int

	// Action is the tool that was called (e.g., "click", "type_text").
	Action string

	// Target describes what the action was performed on.
	Target string

	// Thinking contains the agent's reasoning for this step.
	Thinking string

	// Evaluation is the agent's assessment of the previous action.
	Evaluation string

	// NextGoal describes what the agent planned to do.
	NextGoal string

	// Memory contains what the agent chose to remember.
	Memory string

	// URL is the page URL at this step.
	URL string

	// Title is the page title at this step.
	Title string

	// ScreenshotPath is the path to the screenshot for this step.
	ScreenshotPath string

	// Duration is how long this step took.
	Duration time.Duration

	// Error contains any error that occurred during this step.
	Error string
}
