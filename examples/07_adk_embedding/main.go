// Package main demonstrates using bua-go as a tool within other ADK applications.
//
// This example showcases the Dual-Use Architecture:
// - Using BrowserTool to add browser automation capabilities to any ADK agent
// - Using MultiBrowserTool for parallel browser operations
// - Integrating browser automation with other ADK tools
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	"github.com/anxuanzi/bua-go/export"
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

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║      🔧 BUA-GO AS ADK TOOL DEMONSTRATION 🔧                  ║")
	fmt.Println("║      Dual-Use Architecture: Browser Tool in ADK Agent        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// ═══════════════════════════════════════════════════════════════════
	// DEMO 1: Simple Browser Tool Usage
	// ═══════════════════════════════════════════════════════════════════
	fmt.Println("🔧 DEMO 1: Browser Tool as part of a Research Agent")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	if err := runResearchAgentDemo(ctx, apiKey); err != nil {
		log.Printf("Demo 1 error: %v", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                📊 ADK TOOL DEMO COMPLETE 📊                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Demonstrated features:")
	fmt.Println("   - BrowserTool: Single browser instance for automation tasks")
	fmt.Println("   - Integration with ADK agents as a callable tool")
	fmt.Println("   - Browser automation within larger LLM-driven workflows")
	fmt.Println()
}

// runResearchAgentDemo demonstrates using the browser tool as part of a research agent
func runResearchAgentDemo(ctx context.Context, apiKey string) error {
	// Create the browser tool with configuration
	browserToolConfig := &export.BrowserToolConfig{
		APIKey:          apiKey,
		Model:           "gemini-3-flash-preview",
		Headless:        false, // Show browser for demonstration
		Viewport:        nil,   // Use default
		ShowAnnotations: true,
		Debug:           true,
	}

	browserTool := export.NewBrowserTool(browserToolConfig)
	defer browserTool.Close()

	// Get the ADK tool
	buaTool, err := browserTool.Tool()
	if err != nil {
		return fmt.Errorf("failed to create browser tool: %w", err)
	}

	// Create a Gemini model for our research agent
	model, err := gemini.NewModel(ctx, "gemini-3-flash-preview", &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// Create a research agent that can use the browser tool
	researchAgent, err := llmagent.New(llmagent.Config{
		Name:        "research_agent",
		Description: "A research agent that can browse the web to gather information",
		Model:       model,
		Instruction: `You are a research assistant that helps gather information from the web.
You have access to a browser_automation tool that can:
- Navigate to websites
- Extract information from web pages
- Click on elements and interact with pages
- Fill forms and submit data

When asked to research something, use the browser_automation tool to visit relevant websites and extract the needed information.

For the browser_automation tool:
- task: Describe what you want to do (e.g., "Go to example.com and extract the main heading")
- start_url: Optional URL to start at
- max_steps: Maximum steps (default 30)
- keep_browser: Set to true to keep browser open for follow-up tasks

Always summarize the findings after using the browser.`,
		Tools: []tool.Tool{buaTool},
		GenerateContentConfig: &genai.GenerateContentConfig{
			Temperature:     genai.Ptr[float32](0.3),
			MaxOutputTokens: 8192,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create research agent: %w", err)
	}

	// Create session service
	sessionService := session.InMemoryService()

	// Create runner
	r, err := runner.New(runner.Config{
		Agent:          researchAgent,
		AppName:        "research-demo",
		SessionService: sessionService,
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Create a session
	createResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "research-demo",
		UserID:  "demo_user",
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Run the research task
	fmt.Println("📋 Asking research agent to gather information...")
	fmt.Println()

	userMessage := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: "Please use the browser to visit go.dev and tell me what the main tagline or value proposition of Go is."},
		},
	}

	fmt.Println("User: Please use the browser to visit go.dev and tell me what the main tagline or value proposition of Go is.")
	fmt.Println()

	var response string
	for event, err := range r.Run(ctx, "demo_user", createResp.Session.ID(), userMessage, agent.RunConfig{}) {
		if err != nil {
			return fmt.Errorf("agent error: %w", err)
		}

		if event != nil && event.Content != nil {
			for _, part := range event.Content.Parts {
				if part != nil && part.Text != "" && !event.Partial {
					response = part.Text
				}
			}
		}
	}

	fmt.Println("Agent Response:")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	fmt.Println(response)
	fmt.Println("─────────────────────────────────────────────────────────────────")

	return nil
}
