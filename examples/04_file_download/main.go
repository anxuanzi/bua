// Package main demonstrates the download capability of bua-go.
//
// This example showcases:
// - Downloading files using the download_file tool
// - Both authenticated and unauthenticated downloads
// - Downloading from dynamically discovered URLs
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/anxuanzi/bua-go"
)

func main() {
	// Load .env file from project root
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Create agent configuration
	cfg := bua.Config{
		APIKey:          apiKey,
		Model:           "gemini-3-flash-preview",
		ProfileName:     "download-example",
		Headless:        false, // Show browser for demonstration
		Viewport:        bua.DesktopViewport,
		Debug:           true,
		ShowAnnotations: true,
	}

	// Create the agent
	agent, err := bua.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start the browser
	fmt.Println("🚀 Starting browser...")
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║            📥 DOWNLOAD CAPABILITY DEMONSTRATION 📥            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════
	// TASK 1: Download a PDF from a public website
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("📥 TASK 1: Download a public PDF file")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	downloadPrompt := `
OBJECTIVE: Download the Go programming language logo from the official Go website.

STEPS:
1. Navigate to https://go.dev/
2. Look for a logo or branding image on the page
3. When you find an image URL (usually ending in .svg, .png, or .jpg), use the download_file tool to download it
   - Use: download_file with url=<image_url>
4. Report the download result

If you can't find a downloadable image on the main page:
1. Try downloading any available resource like a favicon
2. Or navigate to go.dev/images/go-logo-white.svg and download that

OUTPUT: Confirm the filename and file path of the downloaded file.
`

	result, err := agent.Run(ctx, downloadPrompt)
	if err != nil {
		log.Printf("Task 1 error: %v", err)
	}
	printResult("Download Task", result)

	// ═══════════════════════════════════════════════════════════════════
	// TASK 2: Find and download from a dynamic page
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println()
	fmt.Println("📥 TASK 2: Download from a content discovery task")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	dynamicDownloadPrompt := `
OBJECTIVE: Find and download the Rust logo from the official website.

STEPS:
1. Navigate to https://www.rust-lang.org/
2. Look for the Rust logo or any downloadable asset
3. Find the logo image URL (check the page source or images)
4. Download the logo using the download_file tool with use_page_auth=false
5. Report what was downloaded

If you find a logo, download it. If not, try downloading:
- The favicon from https://www.rust-lang.org/favicon.ico

OUTPUT: Report the downloaded file details (name, size, type).
`

	result2, err := agent.Run(ctx, dynamicDownloadPrompt)
	if err != nil {
		log.Printf("Task 2 error: %v", err)
	}
	printResult("Dynamic Download", result2)

	// ═══════════════════════════════════════════════════════════════════
	// FINAL SUMMARY
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                📊 DOWNLOAD DEMO COMPLETE 📊                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Demonstrated features:")
	fmt.Println("   - download_file tool for programmatic downloads")
	fmt.Println("   - Direct HTTP downloads (use_page_auth=false)")
	fmt.Println("   - Dynamic URL discovery and download")
	fmt.Println()
	fmt.Println("📁 Downloads saved to: ~/.bua/downloads/")
	fmt.Println()
}

func printResult(title string, result *bua.Result) {
	fmt.Printf("\n┌─ %s ", title)
	for i := 0; i < 50-len(title); i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	if result == nil {
		fmt.Println("│ ❌ Result is nil")
		fmt.Println("└" + "────────────────────────────────────────────────────────────┘")
		return
	}

	if result.Success {
		fmt.Println("│ ✅ Success")
	} else {
		fmt.Println("│ ❌ Failed")
	}

	if result.Data != nil {
		fmt.Printf("│ Data: %v\n", result.Data)
	}

	if result.Error != "" {
		fmt.Printf("│ ⚠️  Error: %s\n", result.Error)
	}

	fmt.Println("└" + "────────────────────────────────────────────────────────────┘")
}
