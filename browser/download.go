// Package browser provides the browser automation layer using go-rod.
package browser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// DownloadConfig holds download configuration.
type DownloadConfig struct {
	// DownloadDir is the directory to save downloads.
	DownloadDir string
	// Timeout is the maximum time to wait for a download.
	Timeout time.Duration
	// AllowOverwrite allows overwriting existing files.
	AllowOverwrite bool
}

// DefaultDownloadConfig returns the default download configuration.
func DefaultDownloadConfig() *DownloadConfig {
	homeDir, _ := os.UserHomeDir()
	return &DownloadConfig{
		DownloadDir:    filepath.Join(homeDir, ".bua", "downloads"),
		Timeout:        5 * time.Minute,
		AllowOverwrite: true,
	}
}

// DownloadInfo contains information about a completed download.
type DownloadInfo struct {
	// URL is the source URL of the download.
	URL string
	// Filename is the name of the downloaded file.
	Filename string
	// FilePath is the full path to the downloaded file.
	FilePath string
	// Size is the size of the downloaded file in bytes.
	Size int64
	// MimeType is the MIME type of the downloaded file.
	MimeType string
	// Data contains the raw bytes (for small files or when requested).
	Data []byte
}

// downloadState tracks active downloads.
type downloadState struct {
	mu        sync.RWMutex
	downloads map[string]*DownloadInfo // GUID -> DownloadInfo
	completed chan string              // GUID of completed downloads
}

// EnableDownloads enables file downloads for the browser.
func (b *Browser) EnableDownloads(ctx context.Context, cfg *DownloadConfig) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if cfg == nil {
		cfg = DefaultDownloadConfig()
	}

	// Ensure download directory exists
	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Enable downloads at browser level
	err := proto.BrowserSetDownloadBehavior{
		Behavior:      proto.BrowserSetDownloadBehaviorBehaviorAllowAndName,
		DownloadPath:  cfg.DownloadDir,
		EventsEnabled: true,
	}.Call(b.rod)
	if err != nil {
		return fmt.Errorf("failed to set download behavior: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from a URL and returns the download info.
// This uses HTTP to fetch the file directly, which is more reliable for programmatic downloads.
func (b *Browser) DownloadFile(ctx context.Context, url string, cfg *DownloadConfig) (*DownloadInfo, error) {
	if cfg == nil {
		cfg = DefaultDownloadConfig()
	}

	// Ensure download directory exists
	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Determine filename
	filename := extractFilename(url, resp)

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Determine file path
	filePath := filepath.Join(cfg.DownloadDir, filename)

	// Check if file exists and handle accordingly
	if !cfg.AllowOverwrite {
		if _, err := os.Stat(filePath); err == nil {
			// File exists, generate unique name
			filePath = generateUniqueFilename(filePath)
			filename = filepath.Base(filePath)
		}
	}

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &DownloadInfo{
		URL:      url,
		Filename: filename,
		FilePath: filePath,
		Size:     int64(len(data)),
		MimeType: resp.Header.Get("Content-Type"),
		Data:     data,
	}, nil
}

// DownloadResource downloads a resource using the browser's CDP protocol.
// This method uses the page's context for authentication and cookies.
func (b *Browser) DownloadResource(ctx context.Context, url string, cfg *DownloadConfig) (*DownloadInfo, error) {
	b.mu.RLock()
	page := b.getActivePageLocked()
	b.mu.RUnlock()

	if page == nil {
		return nil, fmt.Errorf("no active page")
	}

	if cfg == nil {
		cfg = DefaultDownloadConfig()
	}

	// Ensure download directory exists
	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create download directory: %w", err)
	}

	// Use page.GetResource to fetch the resource with page context
	data, err := page.GetResource(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	// Determine filename from URL
	filename := extractFilenameFromURL(url)

	// Determine file path
	filePath := filepath.Join(cfg.DownloadDir, filename)

	// Check if file exists and handle accordingly
	if !cfg.AllowOverwrite {
		if _, err := os.Stat(filePath); err == nil {
			filePath = generateUniqueFilename(filePath)
			filename = filepath.Base(filePath)
		}
	}

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Try to detect MIME type from data
	mimeType := http.DetectContentType(data)

	return &DownloadInfo{
		URL:      url,
		Filename: filename,
		FilePath: filePath,
		Size:     int64(len(data)),
		MimeType: mimeType,
		Data:     data,
	}, nil
}

// Helper functions

// extractFilename extracts the filename from URL or Content-Disposition header.
func extractFilename(url string, resp *http.Response) string {
	// Try Content-Disposition header first
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if strings.Contains(cd, "filename=") {
			parts := strings.Split(cd, "filename=")
			if len(parts) > 1 {
				filename := strings.Trim(parts[1], `"' `)
				if filename != "" {
					return sanitizeDownloadFilename(filename)
				}
			}
		}
	}

	return extractFilenameFromURL(url)
}

// extractFilenameFromURL extracts the filename from a URL.
func extractFilenameFromURL(url string) string {
	// Extract path from URL
	parts := strings.Split(url, "?")
	path := parts[0]

	// Get the last segment
	segments := strings.Split(path, "/")
	filename := segments[len(segments)-1]

	if filename == "" {
		filename = "download"
	}

	return sanitizeDownloadFilename(filename)
}

// sanitizeDownloadFilename removes unsafe characters from a filename.
func sanitizeDownloadFilename(filename string) string {
	// Remove path separators and other unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	result := filename
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Limit length
	if len(result) > 255 {
		ext := filepath.Ext(result)
		name := result[:255-len(ext)]
		result = name + ext
	}

	return result
}

// generateUniqueFilename generates a unique filename by appending a number.
func generateUniqueFilename(filePath string) string {
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)

	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
