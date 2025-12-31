// Example 01: Quick Start - Basic browser automation with highlighting
//
// This example demonstrates the simplest way to use bua-go.
// It opens a browser, navigates to a page, and performs a simple task.
//
// Run with: go run main.go
// Make sure GOOGLE_API_KEY is set in your environment.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anxuanzi/bua-go"
	"github.com/anxuanzi/bua-go/browser"
)

func main() {
	// Enable highlight debugging to see when highlights are triggered
	browser.HighlightDebug = true

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Create agent with minimal config
	// Headless: false means browser window is visible
	// This also enables highlighting by default
	agent, err := bua.New(bua.Config{
		APIKey:   apiKey,
		Headless: false, // Show browser window
		Debug:    true,  // Enable debug logging
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Start the browser
	fmt.Println("Starting browser...")
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Failed to start browser: %v", err)
	}

	// Navigate to Google
	fmt.Println("Navigating to Google...")
	if err := agent.Navigate(ctx, "https://www.google.com"); err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Wait a moment to see the page
	time.Sleep(2 * time.Second)

	// Run a simple task
	fmt.Println("Running task: Search for 'golang'...")
	result, err := agent.Run(ctx, "Type 'golang' in the search box and press Enter")
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	// Print results
	fmt.Println("\n=== Task Results ===")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Steps: %d\n", len(result.Steps))
	fmt.Printf("Tokens used: %d\n", result.TokensUsed)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Data != nil {
		fmt.Printf("Extracted data: %v\n", result.Data)
	}

	// Keep browser open for a moment to see the result
	fmt.Println("\nKeeping browser open for 5 seconds...")
	time.Sleep(5 * time.Second)

	fmt.Println("Done!")
}
