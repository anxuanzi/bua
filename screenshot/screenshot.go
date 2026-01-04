package screenshot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nfnt/resize"
)

// ErrBlankPage is returned when attempting to screenshot a blank page.
var ErrBlankPage = errors.New("page is blank (about:blank)")

// ErrEmptyScreenshot is returned when screenshot appears to be empty/white.
var ErrEmptyScreenshot = errors.New("screenshot appears to be empty or all white")

// Options configures screenshot capture.
type Options struct {
	// MaxWidth is the maximum width for the screenshot.
	// Images wider than this will be resized.
	MaxWidth int

	// Quality is the JPEG quality (1-100).
	// Only used when Format is JPEG.
	Quality int

	// Format is the output format (png or jpeg).
	Format string

	// FullPage captures the entire page, not just the viewport.
	FullPage bool

	// WaitForLoad waits for page load event before capturing.
	WaitForLoad bool

	// WaitForIdle waits for network idle before capturing.
	WaitForIdle bool

	// StabilityTimeout is how long to wait for page stability.
	// Default is 500ms if not set.
	StabilityTimeout time.Duration

	// SkipBlankPages returns ErrBlankPage instead of capturing blank pages.
	SkipBlankPages bool

	// ValidateContent checks if screenshot has actual content (not all white).
	ValidateContent bool
}

// DefaultOptions returns sensible defaults for LLM consumption.
func DefaultOptions() Options {
	return Options{
		MaxWidth:         1280,
		Quality:          80,
		Format:           "jpeg",
		FullPage:         false,
		WaitForLoad:      true,
		WaitForIdle:      false, // Can be slow, disabled by default
		StabilityTimeout: 500 * time.Millisecond,
		SkipBlankPages:   true,
		ValidateContent:  false, // Can be expensive, disabled by default
	}
}

// LLMOptions returns options optimized for LLM vision consumption.
// Includes page readiness checks and content validation.
func LLMOptions() Options {
	return Options{
		MaxWidth:         1280,
		Quality:          75,
		Format:           "jpeg",
		FullPage:         false,
		WaitForLoad:      true,
		WaitForIdle:      true,
		StabilityTimeout: 1 * time.Second,
		SkipBlankPages:   true,
		ValidateContent:  true,
	}
}

// Capture takes a screenshot of the page and returns compressed bytes.
// It implements proper page readiness checks following browser-use best practices.
func Capture(ctx context.Context, page *rod.Page, opts Options) ([]byte, error) {
	// Apply defaults
	if opts.MaxWidth == 0 {
		opts.MaxWidth = 1280
	}
	if opts.Quality == 0 {
		opts.Quality = 80
	}
	if opts.Format == "" {
		opts.Format = "jpeg"
	}
	if opts.StabilityTimeout == 0 {
		opts.StabilityTimeout = 500 * time.Millisecond
	}

	// Check for blank page if configured
	if opts.SkipBlankPages {
		if isBlankPage(page) {
			return nil, ErrBlankPage
		}
	}

	// Wait for page readiness
	if err := waitForPageReady(ctx, page, opts); err != nil {
		// Log warning but continue - page might still be usable
		// The error is non-fatal as we want to attempt screenshot anyway
	}

	// Configure screenshot options
	format := proto.PageCaptureScreenshotFormatJpeg
	if opts.Format == "png" {
		format = proto.PageCaptureScreenshotFormatPng
	}

	// Capture screenshot
	data, err := page.Screenshot(opts.FullPage, &proto.PageCaptureScreenshot{
		Format:  format,
		Quality: &opts.Quality,
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot capture failed: %w", err)
	}

	// Decode image for processing
	img, imgFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	// Validate content if configured
	if opts.ValidateContent {
		if isEmptyImage(img) {
			return nil, ErrEmptyScreenshot
		}
	}

	// Resize if needed
	bounds := img.Bounds()
	if bounds.Dx() > opts.MaxWidth {
		ratio := float64(opts.MaxWidth) / float64(bounds.Dx())
		newHeight := uint(float64(bounds.Dy()) * ratio)
		img = resize.Resize(uint(opts.MaxWidth), newHeight, img, resize.Lanczos3)
	}

	// Encode to output format
	var buf bytes.Buffer
	switch opts.Format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality})
	case "png":
		err = png.Encode(&buf, img)
	default:
		if imgFormat == "png" {
			err = png.Encode(&buf, img)
		} else {
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality})
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode screenshot: %w", err)
	}

	return buf.Bytes(), nil
}

// isBlankPage checks if the page is a blank page (about:blank or empty).
func isBlankPage(page *rod.Page) bool {
	info, err := page.Info()
	if err != nil {
		return false // Assume not blank if we can't get info
	}

	url := info.URL
	if url == "" || url == "about:blank" || strings.HasPrefix(url, "chrome://") {
		return true
	}

	return false
}

// waitForPageReady waits for the page to be ready for screenshot.
func waitForPageReady(ctx context.Context, page *rod.Page, opts Options) error {
	// Wait for page load event
	if opts.WaitForLoad {
		if err := page.WaitLoad(); err != nil {
			// Non-fatal, continue
		}
	}

	// Wait for network idle (no pending requests)
	if opts.WaitForIdle {
		if err := page.WaitIdle(opts.StabilityTimeout); err != nil {
			// Non-fatal, continue
		}
	}

	// Wait for DOM stability (no mutations)
	if opts.StabilityTimeout > 0 {
		if err := page.WaitStable(opts.StabilityTimeout); err != nil {
			// Non-fatal, continue
		}
	}

	return nil
}

// isEmptyImage checks if the image is mostly white/empty.
// Uses sampling to check a grid of pixels for efficiency.
func isEmptyImage(img image.Image) bool {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return true
	}

	// Sample grid of pixels (8x8 = 64 samples)
	sampleSize := 8
	whiteCount := 0
	totalSamples := 0

	// White threshold - pixels with R,G,B all above this are considered white
	const whiteThreshold = 250

	for i := 0; i < sampleSize; i++ {
		for j := 0; j < sampleSize; j++ {
			x := bounds.Min.X + (width * i / sampleSize)
			y := bounds.Min.Y + (height * j / sampleSize)

			c := img.At(x, y)
			r, g, b, _ := c.RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Check if pixel is near white
			if r8 >= whiteThreshold && g8 >= whiteThreshold && b8 >= whiteThreshold {
				whiteCount++
			}
			totalSamples++
		}
	}

	// If more than 95% of samples are white, consider it empty
	whiteRatio := float64(whiteCount) / float64(totalSamples)
	return whiteRatio > 0.95
}

// HasContent checks if the image has meaningful content (not mostly white).
func HasContent(img image.Image) bool {
	return !isEmptyImage(img)
}

// IsPageReady checks if the page is ready for screenshot capture.
func IsPageReady(page *rod.Page) bool {
	if isBlankPage(page) {
		return false
	}

	// Check if page has body content
	hasContent, err := page.Eval(`() => {
		const body = document.body;
		if (!body) return false;
		return body.innerHTML.trim().length > 0;
	}`)
	if err != nil {
		return false
	}

	return hasContent.Value.Bool()
}

// WaitUntilReady waits until the page is ready for screenshot, with timeout.
func WaitUntilReady(ctx context.Context, page *rod.Page, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for page to be ready")
		}

		if IsPageReady(page) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue checking
		}
	}
}

// GetPageInfo returns information about the current page state.
type PageInfo struct {
	URL        string
	Title      string
	IsBlank    bool
	HasContent bool
}

// GetPageInfo returns information about the current page state.
func GetPageInfo(page *rod.Page) (*PageInfo, error) {
	info, err := page.Info()
	if err != nil {
		return nil, err
	}

	isBlank := isBlankPage(page)
	hasContent := false

	if !isBlank {
		hasContent = IsPageReady(page)
	}

	return &PageInfo{
		URL:        info.URL,
		Title:      info.Title,
		IsBlank:    isBlank,
		HasContent: hasContent,
	}, nil
}

// CaptureElement takes a screenshot of a specific element.
func CaptureElement(ctx context.Context, page *rod.Page, selector string, opts Options) ([]byte, error) {
	if opts.Quality == 0 {
		opts.Quality = 80
	}
	if opts.Format == "" {
		opts.Format = "jpeg"
	}

	// Find the element
	el, err := page.Element(selector)
	if err != nil {
		return nil, fmt.Errorf("element not found: %w", err)
	}

	// Get element screenshot
	format := proto.PageCaptureScreenshotFormatJpeg
	if opts.Format == "png" {
		format = proto.PageCaptureScreenshotFormatPng
	}

	data, err := el.Screenshot(format, opts.Quality)
	if err != nil {
		return nil, fmt.Errorf("element screenshot failed: %w", err)
	}

	// Resize if needed
	if opts.MaxWidth > 0 {
		img, imgFormat, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return data, nil // Return original if decode fails
		}

		bounds := img.Bounds()
		if bounds.Dx() > opts.MaxWidth {
			ratio := float64(opts.MaxWidth) / float64(bounds.Dx())
			newHeight := uint(float64(bounds.Dy()) * ratio)
			img = resize.Resize(uint(opts.MaxWidth), newHeight, img, resize.Lanczos3)

			var buf bytes.Buffer
			if imgFormat == "png" || opts.Format == "png" {
				png.Encode(&buf, img)
			} else {
				jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality})
			}
			return buf.Bytes(), nil
		}
	}

	return data, nil
}

// CaptureViewport takes a screenshot of the current viewport only.
func CaptureViewport(ctx context.Context, page *rod.Page, opts Options) ([]byte, error) {
	opts.FullPage = false
	return Capture(ctx, page, opts)
}

// CaptureFullPage takes a screenshot of the entire page.
func CaptureFullPage(ctx context.Context, page *rod.Page, opts Options) ([]byte, error) {
	opts.FullPage = true
	return Capture(ctx, page, opts)
}

// ForLLM captures a screenshot optimized for LLM consumption.
// Uses JPEG format with reasonable compression for token efficiency.
// Includes page readiness checks and skips blank pages.
func ForLLM(ctx context.Context, page *rod.Page, maxWidth int) ([]byte, error) {
	opts := LLMOptions()
	if maxWidth > 0 {
		opts.MaxWidth = maxWidth
	}
	return Capture(ctx, page, opts)
}

// ForLLMSafe captures a screenshot for LLM consumption with full validation.
// Returns nil data (not error) if page is blank or screenshot is empty.
// This is useful for agent loops where blank screenshots should be skipped.
func ForLLMSafe(ctx context.Context, page *rod.Page, maxWidth int) ([]byte, error) {
	opts := LLMOptions()
	if maxWidth > 0 {
		opts.MaxWidth = maxWidth
	}

	data, err := Capture(ctx, page, opts)
	if err != nil {
		// Return nil data for expected blank/empty conditions
		if errors.Is(err, ErrBlankPage) || errors.Is(err, ErrEmptyScreenshot) {
			return nil, nil
		}
		return nil, err
	}

	return data, nil
}

// CaptureAfterAction captures a screenshot after an action has been performed.
// It waits for the page to stabilize after the action before capturing.
func CaptureAfterAction(ctx context.Context, page *rod.Page, maxWidth int) ([]byte, error) {
	opts := LLMOptions()
	if maxWidth > 0 {
		opts.MaxWidth = maxWidth
	}

	// Use longer stability timeout for post-action captures
	opts.StabilityTimeout = 1500 * time.Millisecond
	opts.WaitForIdle = true

	return Capture(ctx, page, opts)
}
