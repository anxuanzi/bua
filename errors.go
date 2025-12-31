package bua

import "errors"

// Common errors returned by the bua package.
var (
	// ErrMissingAPIKey is returned when Config.APIKey is not set.
	ErrMissingAPIKey = errors.New("bua: API key is required")

	// ErrNotStarted is returned when Run is called before Start.
	ErrNotStarted = errors.New("bua: agent not started, call Start() first")

	// ErrAlreadyStarted is returned when Start is called twice.
	ErrAlreadyStarted = errors.New("bua: agent already started")

	// ErrMaxStepsReached is returned when the agent exceeds MaxSteps.
	ErrMaxStepsReached = errors.New("bua: maximum steps reached without completing task")

	// ErrBrowserClosed is returned when the browser is unexpectedly closed.
	ErrBrowserClosed = errors.New("bua: browser was closed")

	// ErrElementNotFound is returned when an element index is invalid.
	ErrElementNotFound = errors.New("bua: element not found")

	// ErrElementNotVisible is returned when an element is not visible.
	ErrElementNotVisible = errors.New("bua: element is not visible")

	// ErrNavigationFailed is returned when page navigation fails.
	ErrNavigationFailed = errors.New("bua: navigation failed")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("bua: operation timed out")

	// ErrHumanTakeoverTimeout is returned when human intervention times out.
	ErrHumanTakeoverTimeout = errors.New("bua: human takeover timed out")
)
