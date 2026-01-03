package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anxuanzi/bua"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Create agent configuration
	cfg := bua.Config{
		APIKey:        apiKey,
		Model:         "gemini-3-flash-preview", // Fast model for quick tasks
		Headless:      false,                    // Show browser window
		Preset:        bua.PresetBalanced,       // Good balance of quality and cost
		Debug:         true,                     // Enable debug logging
		ScreenshotDir: "./screenshots",
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
	fmt.Println("Starting browser...")
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// Run a simple task
	fmt.Println("\n--- Running task: Search for 'golang tutorials' on Google ---")

	result, err := agent.Run(ctx, "Go to duckduckgo and search for 'golang tutorials'")
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	// Print results
	fmt.Printf("\n--- Task Completed ---\n")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Steps taken: %d\n", len(result.Steps))

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}

	// Print step details
	fmt.Println("\n--- Steps ---")
	for _, step := range result.Steps {
		fmt.Printf("[%d] %s - %s\n", step.Number, step.Action, step.NextGoal)
		if step.Thinking != "" {
			fmt.Printf("    Thinking: %s\n", truncate(step.Thinking, 80))
		}
	}

	// Wait for user to see the result
	fmt.Println("\n--- Press Ctrl+C to close ---")
	select {}
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
