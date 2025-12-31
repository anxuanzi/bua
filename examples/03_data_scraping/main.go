// Package main demonstrates web scraping with bua-go.
package main

import (
	"context"
	"encoding/json"
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

	// Create agent - simplified config, only set what you need
	cfg := bua.Config{
		APIKey:      apiKey,
		ProfileName: "scraping",
		Debug:       true,
		// For scraping, PresetFast (text-only) is often sufficient and faster:
		// Preset: bua.PresetFast,
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

	// Example: Scrape Hacker News top stories - short autonomous prompt!
	fmt.Println("📰 Scraping Hacker News top stories...")
	result, err := agent.Run(ctx, `Go to https://news.ycombinator.com and extract the top 5 stories with their title, url, points, and number of comments.`)
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	// Print result
	fmt.Println()
	fmt.Printf("✅ Task completed: success=%v\n", result.Success)

	if result.Data != nil {
		// Pretty print the extracted data
		data, _ := json.MarshalIndent(result.Data, "", "  ")
		fmt.Printf("📊 Extracted data:\n%s\n", data)
	}

	// Print steps for debugging
	fmt.Printf("\n📝 Steps taken: %d\n", len(result.Steps))
	for i, step := range result.Steps {
		fmt.Printf("  %d. %s: %s\n", i+1, step.Action, step.Reasoning)
	}

	if result.Error != "" {
		fmt.Printf("❌ Error: %s\n", result.Error)
	}
}
