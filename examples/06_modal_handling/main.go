// Package main demonstrates scrolling within modals and popups.
// This example shows how bua-go handles scrollable containers like:
// - Instagram comment modals
// - Chat windows
// - Dropdown lists with many options
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
	// Load .env file
	_ = godotenv.Load(".env")

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// Create agent
	agent, err := bua.New(bua.Config{
		APIKey:   apiKey,
		Model:    bua.ModelGemini3Flash,
		Headless: false,
		Debug:    true,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Starting browser...")
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	// Navigate to a page with a modal/popup that has scrollable content
	// This example uses a dialog example page
	fmt.Println("Navigating to dialog example page...")
	if err := agent.Navigate(ctx, "https://www.w3schools.com/howto/tryit.asp?filename=tryhow_css_modal"); err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// The agent will:
	// 1. Open the modal by clicking the button
	// 2. Recognize it's a scrollable modal container
	// 3. Use element_id scrolling to scroll within the modal (not the page)
	fmt.Println("\nRunning modal scroll task...")
	fmt.Println("The agent should:")
	fmt.Println("  1. Click to open the modal")
	fmt.Println("  2. Scroll within the modal content")
	fmt.Println("  3. Use element_id or auto_detect for modal scrolling")

	result, err := agent.Run(ctx, `
		In the right iframe:
		1. Click the "Open Modal" button
		2. When the modal opens, scroll down within the modal to see more content
		3. Then close the modal by clicking the X or outside the modal

		Note: When scrolling in a modal, you must use element_id scrolling or auto_detect=true.
	`)
	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	fmt.Printf("\nResult: success=%v\n", result.Success)
	fmt.Printf("Steps taken: %d\n", len(result.Steps))
	for i, step := range result.Steps {
		fmt.Printf("  %d. %s: %s\n", i+1, step.Action, step.Target)
		if step.Reasoning != "" {
			fmt.Printf("     Reasoning: %s\n", step.Reasoning)
		}
	}

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}

	fmt.Println("\n=== Modal Scrolling Notes ===")
	fmt.Println("When scrolling within modals/popups:")
	fmt.Println("  - Use scroll(direction, amount, element_id=<container_index>)")
	fmt.Println("  - Or use scroll(direction, amount, auto_detect=true)")
	fmt.Println("  - Without element_id or auto_detect, scroll() moves the main page")
	fmt.Println("  - The system prompt guides the agent on modal detection")
}
