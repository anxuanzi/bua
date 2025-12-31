// Package main demonstrates basic usage of the bua-go library.
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

	// Create agent - only APIKey is required, everything else has sensible defaults!
	//
	// Simplest usage:
	//   agent, _ := bua.New(bua.Config{APIKey: apiKey})
	//
	// With preset for different needs:
	//   Preset: bua.PresetFast      // Text-only, fastest, lowest tokens
	//   Preset: bua.PresetEfficient // Small screenshots, cost savings
	//   Preset: bua.PresetBalanced  // Default, good for most tasks
	//   Preset: bua.PresetQuality   // Higher quality for complex UIs
	//   Preset: bua.PresetMax       // Maximum quality for debugging
	//
	cfg := bua.Config{
		APIKey:      apiKey,
		Debug:       true, // Enable debug logging
		ProfileName: "simple",
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

	// Navigate to a website
	fmt.Println("🌐 Navigating to Google...")
	if err := agent.Navigate(ctx, "https://www.google.com"); err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Run a task with natural language - short autonomous prompt!
	fmt.Println("🔍 Running search task...")
	result, err := agent.Run(ctx, `Search for "Go programming language" and click on the official Go website (go.dev).`)
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	// Print result
	fmt.Println()
	fmt.Printf("✅ Task completed: success=%v\n", result.Success)
	fmt.Printf("📝 Steps taken: %d\n", len(result.Steps))
	for i, step := range result.Steps {
		fmt.Printf("  %d. %s: %s (%s)\n", i+1, step.Action, step.Target, step.Reasoning)
	}

	if result.Error != "" {
		fmt.Printf("❌ Error: %s\n", result.Error)
	}

	fmt.Println("✨ Done!")
}
