// Package screenshot provides screenshot capture, annotation, and storage functionality.
package screenshot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/anxuanzi/bua-go/dom"
	"github.com/fogleman/gg"
)

// Config holds screenshot configuration.
type Config struct {
	// Enabled indicates whether screenshots should be captured.
	Enabled bool

	// Annotate indicates whether screenshots should be annotated with element indices.
	Annotate bool

	// StorageDir is the directory where screenshots are stored.
	StorageDir string

	// MaxScreenshots is the maximum number of screenshots to keep.
	// Set to 0 for unlimited. Oldest screenshots are deleted first.
	MaxScreenshots int

	// ImageFormat is the format for saved screenshots ("png" or "jpeg").
	ImageFormat string

	// Quality is the JPEG quality (1-100). Only used if ImageFormat is "jpeg".
	Quality int

	// AnnotationStyle configures how annotations are drawn.
	AnnotationStyle *AnnotationStyle
}

// AnnotationStyle configures the visual style of element annotations.
type AnnotationStyle struct {
	// BoxColor is the color of bounding boxes.
	BoxColor color.Color

	// LabelColor is the background color of labels.
	LabelColor color.Color

	// TextColor is the color of label text.
	TextColor color.Color

	// BoxWidth is the width of bounding box lines.
	BoxWidth float64

	// FontSize is the size of label text.
	FontSize float64

	// ShowIndex determines whether to show element indices.
	ShowIndex bool

	// ShowRole determines whether to show element roles.
	ShowRole bool
}

// DefaultAnnotationStyle returns the default annotation style.
func DefaultAnnotationStyle() *AnnotationStyle {
	return &AnnotationStyle{
		BoxColor:   color.RGBA{255, 107, 107, 200}, // Coral red
		LabelColor: color.RGBA{255, 107, 107, 230},
		TextColor:  color.White,
		BoxWidth:   2,
		FontSize:   12,
		ShowIndex:  true,
		ShowRole:   false,
	}
}

// Manager handles screenshot operations.
type Manager struct {
	config *Config
	mu     sync.Mutex
}

// NewManager creates a new screenshot manager.
func NewManager(cfg *Config) *Manager {
	if cfg.ImageFormat == "" {
		cfg.ImageFormat = "png"
	}
	if cfg.Quality == 0 {
		cfg.Quality = 90
	}
	if cfg.AnnotationStyle == nil {
		cfg.AnnotationStyle = DefaultAnnotationStyle()
	}
	if cfg.StorageDir != "" {
		os.MkdirAll(cfg.StorageDir, 0755)
	}

	return &Manager{
		config: cfg,
	}
}

// Annotate adds element annotations to a screenshot.
// Each interactive element gets a numbered bounding box.
func (m *Manager) Annotate(screenshotData []byte, elements *dom.ElementMap) ([]byte, error) {
	if elements == nil || len(elements.Elements) == 0 {
		return screenshotData, nil
	}

	// Decode the screenshot
	img, _, err := image.Decode(bytes.NewReader(screenshotData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	// Create drawing context
	dc := gg.NewContextForImage(img)

	style := m.config.AnnotationStyle
	if style == nil {
		style = DefaultAnnotationStyle()
	}

	// Draw annotations for each element
	for _, el := range elements.Elements {
		if !el.IsVisible {
			continue
		}

		bb := el.BoundingBox
		if bb.Width <= 0 || bb.Height <= 0 {
			continue
		}

		// Draw bounding box
		dc.SetColor(style.BoxColor)
		dc.SetLineWidth(style.BoxWidth)
		dc.DrawRectangle(bb.X, bb.Y, bb.Width, bb.Height)
		dc.Stroke()

		// Draw a numbered circle at the top-left corner
		circleRadius := 10.0
		circleX := bb.X + circleRadius
		circleY := bb.Y - circleRadius
		if circleY < circleRadius {
			circleY = bb.Y + circleRadius
		}

		// Draw circle background
		dc.SetColor(style.LabelColor)
		dc.DrawCircle(circleX, circleY, circleRadius)
		dc.Fill()

		// Draw circle border
		dc.SetColor(style.BoxColor)
		dc.SetLineWidth(1)
		dc.DrawCircle(circleX, circleY, circleRadius)
		dc.Stroke()

		// Draw the index number using simple digit rendering
		// Note: For proper text rendering, a font file would need to be loaded
		// This draws the index as a series of small shapes representing the number
		drawNumber(dc, el.Index, circleX, circleY, style.TextColor)
	}

	// Encode result
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode annotated screenshot: %w", err)
	}

	return buf.Bytes(), nil
}

// drawNumber draws a simple representation of a number at the given position.
// This is a basic implementation that draws dots for each digit.
func drawNumber(dc *gg.Context, n int, x, y float64, c color.Color) {
	dc.SetColor(c)

	// Draw a small filled rectangle as a placeholder for the number
	// For a real implementation, you would load a font file
	label := fmt.Sprintf("%d", n)

	// Simple approach: draw small rectangles/dots for each character position
	startX := x - float64(len(label))*2
	for i := range label {
		// Draw a small vertical bar for each digit position
		dx := startX + float64(i)*4
		dc.DrawRectangle(dx, y-3, 2, 6)
		dc.Fill()
	}
}

// Save saves a screenshot to storage and returns the file path.
func (m *Manager) Save(data []byte, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.StorageDir == "" {
		return "", fmt.Errorf("storage directory not configured")
	}

	// Ensure directory exists
	if err := os.MkdirAll(m.config.StorageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate filename
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.%s", timestamp, sanitizeFilename(name), m.config.ImageFormat)
	filepath := filepath.Join(m.config.StorageDir, filename)

	// Write file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	// Cleanup old screenshots if needed
	if m.config.MaxScreenshots > 0 {
		m.cleanup()
	}

	return filepath, nil
}

// cleanup removes old screenshots to stay within the MaxScreenshots limit.
func (m *Manager) cleanup() {
	files, err := os.ReadDir(m.config.StorageDir)
	if err != nil {
		return
	}

	// Filter to screenshot files
	var screenshots []os.DirEntry
	for _, f := range files {
		if !f.IsDir() && isScreenshotFile(f.Name()) {
			screenshots = append(screenshots, f)
		}
	}

	// Sort by name (which includes timestamp)
	sort.Slice(screenshots, func(i, j int) bool {
		return screenshots[i].Name() < screenshots[j].Name()
	})

	// Remove excess files
	excess := len(screenshots) - m.config.MaxScreenshots
	for i := 0; i < excess; i++ {
		os.Remove(filepath.Join(m.config.StorageDir, screenshots[i].Name()))
	}
}

func isScreenshotFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func sanitizeFilename(name string) string {
	// Remove or replace invalid characters
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c == ' ' {
			result = append(result, '_')
		}
	}
	if len(result) == 0 {
		return "screenshot"
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return string(result)
}

// List returns a list of saved screenshot paths, sorted by date (newest first).
func (m *Manager) List() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.StorageDir == "" {
		return nil, nil
	}

	files, err := os.ReadDir(m.config.StorageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var paths []string
	for _, f := range files {
		if !f.IsDir() && isScreenshotFile(f.Name()) {
			paths = append(paths, filepath.Join(m.config.StorageDir, f.Name()))
		}
	}

	// Sort newest first (reverse order since filenames contain timestamps)
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))

	return paths, nil
}

// Clear removes all saved screenshots.
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.StorageDir == "" {
		return nil
	}

	files, err := os.ReadDir(m.config.StorageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, f := range files {
		if !f.IsDir() && isScreenshotFile(f.Name()) {
			os.Remove(filepath.Join(m.config.StorageDir, f.Name()))
		}
	}

	return nil
}
