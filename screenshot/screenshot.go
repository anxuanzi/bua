package screenshot

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nfnt/resize"
)

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
}

// DefaultOptions returns sensible defaults for LLM consumption.
func DefaultOptions() Options {
	return Options{
		MaxWidth: 1280,
		Quality:  80,
		Format:   "jpeg",
		FullPage: false,
	}
}

// Capture takes a screenshot of the page and returns compressed bytes.
func Capture(ctx context.Context, page *rod.Page, opts Options) ([]byte, error) {
	if opts.MaxWidth == 0 {
		opts.MaxWidth = 1280
	}
	if opts.Quality == 0 {
		opts.Quality = 80
	}
	if opts.Format == "" {
		opts.Format = "jpeg"
	}

	// Wait for page stability (500ms stability window)
	_ = ctx // Context available for future use
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails
	}

	// Configure screenshot options
	format := proto.PageCaptureScreenshotFormatJpeg
	if opts.Format == "png" {
		format = proto.PageCaptureScreenshotFormatPng
	}

	var screenshotOpts []any
	if opts.FullPage {
		screenshotOpts = append(screenshotOpts, true)
	}

	// Capture screenshot
	var data []byte
	var err error

	if opts.FullPage {
		data, err = page.Screenshot(opts.FullPage, &proto.PageCaptureScreenshot{
			Format:  format,
			Quality: &opts.Quality,
		})
	} else {
		data, err = page.Screenshot(false, &proto.PageCaptureScreenshot{
			Format:  format,
			Quality: &opts.Quality,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("screenshot capture failed: %w", err)
	}

	// Decode and resize if needed
	img, imgFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	// Check if resizing is needed
	bounds := img.Bounds()
	if bounds.Dx() > opts.MaxWidth {
		// Calculate new height maintaining aspect ratio
		ratio := float64(opts.MaxWidth) / float64(bounds.Dx())
		newHeight := uint(float64(bounds.Dy()) * ratio)

		// Resize using Lanczos3 for quality
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
		// Use original format if unrecognized
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
func ForLLM(ctx context.Context, page *rod.Page, maxWidth int) ([]byte, error) {
	opts := Options{
		MaxWidth: maxWidth,
		Quality:  75, // Good balance of quality and size
		Format:   "jpeg",
		FullPage: false,
	}
	return Capture(ctx, page, opts)
}
