// Package main provides end-to-end tests for the browser agent.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/anxuanzi/bua-go"
)

// TestCase represents a single E2E test
type TestCase struct {
	Name        string
	URL         string
	Task        string
	Validate    func(result *bua.Result) error
	MaxDuration time.Duration
}

func main() {
	// Load environment
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// All test cases
	allTests := []TestCase{
		{
			Name: "Navigation Test",
			URL:  "https://example.com",
			Task: "Look at the page and tell me what is the main heading text. Then call done() with success=true and the heading text in the summary.",
			Validate: func(r *bua.Result) error {
				if !r.Success {
					return fmt.Errorf("task failed: %s", r.Error)
				}
				return nil
			},
			MaxDuration: 2 * time.Minute,
		},
		{
			Name: "Click Test",
			URL:  "https://example.com",
			Task: "Click on the 'More information...' link and wait for the new page to load. Then call done() with success=true.",
			Validate: func(r *bua.Result) error {
				if !r.Success {
					return fmt.Errorf("task failed: %s", r.Error)
				}
				// Should have at least one click action
				hasClick := false
				for _, step := range r.Steps {
					if strings.Contains(strings.ToLower(step.Action), "click") {
						hasClick = true
						break
					}
				}
				if !hasClick {
					return fmt.Errorf("expected click action, got none")
				}
				return nil
			},
			MaxDuration: 2 * time.Minute,
		},
		{
			Name: "Extraction Test",
			URL:  "https://news.ycombinator.com",
			Task: "Extract the titles of the top 3 stories on this page. Return them as a list in the summary and call done() with success=true.",
			Validate: func(r *bua.Result) error {
				if !r.Success {
					return fmt.Errorf("task failed: %s", r.Error)
				}
				return nil
			},
			MaxDuration: 3 * time.Minute,
		},
		{
			Name: "Search and Type Test",
			URL:  "https://www.google.com",
			Task: "Type 'golang testing' into the search box and press Enter to search. Wait for results to load, then call done() with success=true.",
			Validate: func(r *bua.Result) error {
				if !r.Success {
					return fmt.Errorf("task failed: %s", r.Error)
				}
				// Should have a type action
				hasType := false
				for _, step := range r.Steps {
					if strings.Contains(strings.ToLower(step.Action), "type") {
						hasType = true
						break
					}
				}
				if !hasType {
					return fmt.Errorf("expected type action, got none")
				}
				return nil
			},
			MaxDuration: 3 * time.Minute,
		},
		{
			Name: "Multi-Step Task",
			URL:  "https://en.wikipedia.org/wiki/Go_(programming_language)",
			Task: "Scroll down to see more content. Then find and click on the 'History' section link in the table of contents. Finally call done() with success=true.",
			Validate: func(r *bua.Result) error {
				if !r.Success {
					return fmt.Errorf("task failed: %s", r.Error)
				}
				// Should have both scroll and click actions
				hasScroll := false
				hasClick := false
				for _, step := range r.Steps {
					if strings.Contains(strings.ToLower(step.Action), "scroll") {
						hasScroll = true
					}
					if strings.Contains(strings.ToLower(step.Action), "click") {
						hasClick = true
					}
				}
				if !hasScroll {
					return fmt.Errorf("expected scroll action, got none")
				}
				if !hasClick {
					return fmt.Errorf("expected click action, got none")
				}
				return nil
			},
			MaxDuration: 4 * time.Minute,
		},
	}

	// Run specific test if TEST_NAME env is set, otherwise run all
	tests := allTests
	if testName := os.Getenv("TEST_NAME"); testName != "" {
		for _, tc := range allTests {
			if tc.Name == testName {
				tests = []TestCase{tc}
				break
			}
		}
	}

	// Run tests
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("  BUA-GO End-to-End Test Suite")
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println()

	passed := 0
	failed := 0
	var failures []string

	for i, tc := range tests {
		fmt.Printf("[%d/%d] Running: %s\n", i+1, len(tests), tc.Name)
		fmt.Printf("      URL: %s\n", tc.URL)
		fmt.Printf("      Task: %s\n", truncate(tc.Task, 60))

		err := runTest(apiKey, tc)
		if err != nil {
			fmt.Printf("      ❌ FAILED: %v\n", err)
			failed++
			failures = append(failures, fmt.Sprintf("%s: %v", tc.Name, err))
		} else {
			fmt.Printf("      ✅ PASSED\n")
			passed++
		}
		fmt.Println()

		// Small delay between tests
		time.Sleep(2 * time.Second)
	}

	// Summary
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Printf("  Results: %d passed, %d failed, %d total\n", passed, failed, len(tests))
	fmt.Println("=" + strings.Repeat("=", 59))

	if failed > 0 {
		fmt.Println("\nFailures:")
		for _, f := range failures {
			fmt.Printf("  - %s\n", f)
		}
		os.Exit(1)
	}

	fmt.Println("\n✨ All tests passed!")
}

func runTest(apiKey string, tc TestCase) error {
	// Create agent
	cfg := bua.Config{
		APIKey:          apiKey,
		Model:           "gemini-2.0-flash",
		ProfileName:     "e2e_test",
		Headless:        true,
		Viewport:        bua.DesktopViewport,
		Debug:           false,
		ShowAnnotations: false,
	}

	agent, err := bua.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	defer agent.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), tc.MaxDuration)
	defer cancel()

	// Start browser
	if err := agent.Start(ctx); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Navigate to URL
	if err := agent.Navigate(ctx, tc.URL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Small delay for page load
	time.Sleep(1 * time.Second)

	// Run the task
	result, err := agent.Run(ctx, tc.Task)
	if err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}

	// Print steps for debugging
	fmt.Printf("      Steps (%d):\n", len(result.Steps))
	for i, step := range result.Steps {
		fmt.Printf("        %d. %s: %s\n", i+1, step.Action, truncate(step.Target, 40))
	}

	// Check confidence if available
	if result.Confidence != nil {
		fmt.Printf("      Confidence: %.2f (avg), %.2f (min), %.0f%% success rate\n",
			result.Confidence.AverageConfidence,
			result.Confidence.MinConfidence,
			result.Confidence.SuccessRate*100)
	}

	// Validate result
	if tc.Validate != nil {
		return tc.Validate(result)
	}

	return nil
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
