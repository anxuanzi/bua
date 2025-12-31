// Highlight Test - Direct test of highlighting without LLM
//
// This test bypasses the LLM to directly test the highlighting mechanism.
// Run with: go run main.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anxuanzi/bua-go/browser"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	// Enable highlight debug logging
	browser.HighlightDebug = true

	fmt.Println("=== Highlight Test ===")
	fmt.Println("This test directly verifies the highlighting mechanism works.")
	fmt.Println()

	// Launch browser
	fmt.Println("1. Launching browser...")
	u := launcher.New().
		Headless(false).
		Devtools(false).
		MustLaunch()

	rodBrowser := rod.New().ControlURL(u).MustConnect()
	defer rodBrowser.MustClose()

	// Create page
	fmt.Println("2. Creating page...")
	page := rodBrowser.MustPage("")

	// Navigate to a simple page
	fmt.Println("3. Navigating to example.com...")
	page.MustNavigate("https://example.com")
	page.MustWaitStable()

	// Create highlighter
	fmt.Println("4. Creating highlighter (enabled=true)...")
	highlighter := browser.NewHighlighter(page, true)
	highlighter.SetDelay(1 * time.Second) // Longer delay for visibility

	// Test 1: Flash element by CSS selector
	fmt.Println("\n=== Test 1: FlashElementBySelector ===")
	fmt.Println("Looking for 'h1' element...")
	err := highlighter.FlashElementBySelector("h1", "Test 1: h1")
	if err != nil {
		log.Printf("ERROR: FlashElementBySelector failed: %v", err)
	} else {
		fmt.Println("SUCCESS: FlashElementBySelector completed")
	}
	time.Sleep(500 * time.Millisecond)

	// Test 2: Flash at coordinates
	fmt.Println("\n=== Test 2: FlashElementAtPoint ===")
	fmt.Println("Flashing element at coordinates (200, 200)...")
	err = highlighter.FlashElementAtPoint(200, 200, "Test 2: Point")
	if err != nil {
		log.Printf("ERROR: FlashElementAtPoint failed: %v", err)
	} else {
		fmt.Println("SUCCESS: FlashElementAtPoint completed")
	}
	time.Sleep(500 * time.Millisecond)

	// Test 3: Highlight element with bounding box
	fmt.Println("\n=== Test 3: HighlightElement ===")
	fmt.Println("Highlighting element at (100, 100, 200, 50)...")
	err = highlighter.HighlightElement(100, 100, 200, 50, "Test 3: Box")
	if err != nil {
		log.Printf("ERROR: HighlightElement failed: %v", err)
	} else {
		fmt.Println("SUCCESS: HighlightElement completed")
	}
	time.Sleep(500 * time.Millisecond)

	// Test 4: Highlight scroll
	fmt.Println("\n=== Test 4: HighlightScroll ===")
	fmt.Println("Showing scroll indicator...")
	err = highlighter.HighlightScroll(400, 300, "down")
	if err != nil {
		log.Printf("ERROR: HighlightScroll failed: %v", err)
	} else {
		fmt.Println("SUCCESS: HighlightScroll completed")
	}

	// Keep browser open
	fmt.Println("\n=== Keeping browser open for 5 seconds ===")
	fmt.Println("You should have seen orange highlights on the page.")
	time.Sleep(5 * time.Second)

	// Cleanup
	fmt.Println("Removing highlights...")
	highlighter.RemoveHighlights()

	fmt.Println("\n=== Test Complete ===")
}
