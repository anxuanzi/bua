package bua

import (
	"os"
	"path/filepath"
)

// Preset defines token/quality tradeoffs for different use cases.
type Preset string

const (
	// PresetFast uses text-only mode for lowest token usage.
	PresetFast Preset = "fast"

	// PresetEfficient uses low quality screenshots.
	PresetEfficient Preset = "efficient"

	// PresetBalanced is the default with good balance of quality and cost.
	PresetBalanced Preset = "balanced"

	// PresetQuality uses higher quality screenshots.
	PresetQuality Preset = "quality"

	// PresetMax uses maximum quality for complex pages.
	PresetMax Preset = "max"
)

// Viewport defines browser viewport dimensions.
type Viewport struct {
	Width  int
	Height int
}

// DefaultViewport returns the default viewport size.
func DefaultViewport() Viewport {
	return Viewport{Width: 1280, Height: 720}
}

// Config holds agent configuration.
type Config struct {
	// APIKey is the Gemini API key (required).
	APIKey string

	// Model is the Gemini model to use. Default: "gemini-2.5-flash".
	Model string

	// Headless runs the browser without a visible window. Default: false.
	Headless bool

	// Debug enables verbose logging. Default: false.
	Debug bool

	// ProfileName specifies a named browser profile for session persistence.
	// Empty string uses a temporary profile that is deleted on close.
	ProfileName string

	// ProfileDir is the directory to store browser profiles.
	// Default: ~/.bua/profiles
	ProfileDir string

	// Viewport sets the browser viewport dimensions.
	// Default: 1280x720
	Viewport *Viewport

	// MaxSteps is the maximum number of agent steps before giving up.
	// Default: 100
	MaxSteps int

	// Preset configures token/quality tradeoffs.
	// Default: PresetBalanced
	Preset Preset

	// MaxTokens is the maximum token budget for context.
	// Set automatically based on Preset if not specified.
	MaxTokens int

	// MaxElements is the maximum number of elements to include in state.
	// Set automatically based on Preset if not specified.
	MaxElements int

	// ScreenshotMaxWidth is the maximum width for screenshots.
	// Set automatically based on Preset if not specified.
	ScreenshotMaxWidth int

	// ScreenshotQuality is the JPEG quality (1-100) for screenshots.
	// Set automatically based on Preset if not specified.
	ScreenshotQuality int

	// TextOnly disables screenshots entirely for minimum token usage.
	// Set automatically based on Preset if not specified.
	TextOnly bool

	// ShowAnnotations displays element indices on the page during execution.
	// Useful for debugging. Default: false.
	ShowAnnotations bool

	// ShowHighlight highlights elements before actions.
	// Default: true when not headless.
	ShowHighlight *bool

	// HighlightDuration is how long to show action highlights.
	// Default: 300ms.
	HighlightDurationMs int

	// ScreenshotDir is the directory to save screenshots.
	// Default: system temp directory.
	ScreenshotDir string
}

// presetConfig defines the configuration for each preset.
type presetConfig struct {
	MaxTokens          int
	MaxElements        int
	ScreenshotMaxWidth int
	ScreenshotQuality  int
	TextOnly           bool
}

var presetConfigs = map[Preset]presetConfig{
	PresetFast: {
		MaxTokens:          8000,
		MaxElements:        30,
		ScreenshotMaxWidth: 0,
		ScreenshotQuality:  0,
		TextOnly:           true,
	},
	PresetEfficient: {
		MaxTokens:          16000,
		MaxElements:        50,
		ScreenshotMaxWidth: 800,
		ScreenshotQuality:  60,
		TextOnly:           false,
	},
	PresetBalanced: {
		MaxTokens:          32000,
		MaxElements:        100,
		ScreenshotMaxWidth: 1280,
		ScreenshotQuality:  75,
		TextOnly:           false,
	},
	PresetQuality: {
		MaxTokens:          64000,
		MaxElements:        200,
		ScreenshotMaxWidth: 1920,
		ScreenshotQuality:  85,
		TextOnly:           false,
	},
	PresetMax: {
		MaxTokens:          128000,
		MaxElements:        500,
		ScreenshotMaxWidth: 2560,
		ScreenshotQuality:  95,
		TextOnly:           false,
	},
}

// applyDefaults fills in default values for the config.
func (c *Config) applyDefaults() {
	if c.Model == "" {
		c.Model = "gemini-2.5-flash"
	}

	if c.ProfileDir == "" {
		home, _ := os.UserHomeDir()
		c.ProfileDir = filepath.Join(home, ".bua", "profiles")
	}

	if c.Viewport == nil {
		v := DefaultViewport()
		c.Viewport = &v
	}

	if c.MaxSteps == 0 {
		c.MaxSteps = 100
	}

	if c.Preset == "" {
		c.Preset = PresetBalanced
	}

	// Apply preset configuration
	preset, ok := presetConfigs[c.Preset]
	if !ok {
		preset = presetConfigs[PresetBalanced]
	}

	if c.MaxTokens == 0 {
		c.MaxTokens = preset.MaxTokens
	}
	if c.MaxElements == 0 {
		c.MaxElements = preset.MaxElements
	}
	if c.ScreenshotMaxWidth == 0 {
		c.ScreenshotMaxWidth = preset.ScreenshotMaxWidth
	}
	if c.ScreenshotQuality == 0 {
		c.ScreenshotQuality = preset.ScreenshotQuality
	}
	// TextOnly is only set from preset if not explicitly configured
	// We use the zero value check
	if !c.TextOnly && preset.TextOnly {
		c.TextOnly = preset.TextOnly
	}

	if c.ShowHighlight == nil {
		show := !c.Headless
		c.ShowHighlight = &show
	}

	if c.HighlightDurationMs == 0 {
		c.HighlightDurationMs = 300
	}

	if c.ScreenshotDir == "" {
		c.ScreenshotDir = os.TempDir()
	}
}

// validate checks that required configuration is provided.
func (c *Config) validate() error {
	if c.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}
