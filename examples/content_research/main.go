// Package main demonstrates comprehensive browser automation capabilities using bua-go.
//
// This example showcases:
// - Multi-tab browser management (open, switch, close tabs)
// - Persistent findings storage (save and search data across 100+ steps)
// - Complex multi-site navigation and interaction
// - Structured data extraction with JSON schemas
// - Conditional logic and error handling
// - Professional prompt engineering patterns
//
// Use Case: Instagram Content Research & Trend Analysis
// The agent analyzes multiple hashtags to discover trending content patterns,
// high-engagement posts, and emerging content creators in a specific niche.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/anxuanzi/bua-go"
)

func main() {
	// Load .env file from project root
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Note: Could not load .env file: %v", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}

	// ═══════════════════════════════════════════════════════════════════════
	// AGENT CONFIGURATION
	// ═══════════════════════════════════════════════════════════════════════
	cfg := bua.Config{
		APIKey:          apiKey,
		Model:           "gemini-3-pro-preview", // Fast model for complex tasks
		ProfileName:     "instagram",            // Separate browser profile
		Headless:        false,                  // Show browser for observation
		Viewport:        bua.DesktopViewport,    // 1920x1080 desktop view
		Debug:           true,                   // Enable detailed logging
		ShowAnnotations: true,                   // Visual element annotations
	}

	// Create the automation agent
	agent, err := bua.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	// Extended timeout for complex research tasks
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	// ═══════════════════════════════════════════════════════════════════════
	// BROWSER INITIALIZATION
	// ═══════════════════════════════════════════════════════════════════════
	printHeader("BUA-GO CONTENT RESEARCH DEMO", "Instagram Trend Analysis & Content Discovery")

	fmt.Println("🚀 Starting browser automation agent...")
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
	fmt.Println("✅ Browser started successfully")
	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════
	// PHASE 1: MULTI-TAB HASHTAG EXPLORATION
	// Demonstrates: new_tab, switch_tab, list_tabs tools
	// ═══════════════════════════════════════════════════════════════════════
	printPhase(1, "Multi-Tab Hashtag Exploration",
		"Opening multiple hashtag pages simultaneously to compare trends")

	phase1Prompt := `
## OBJECTIVE
Open multiple Instagram hashtag explore pages in separate tabs to analyze content trends in the photography niche.

## TOOLS TO USE
- Use 'new_tab' to open each hashtag in a separate browser tab
- Use 'list_tabs' to verify all tabs are open
- Use 'switch_tab' to navigate between them

## EXECUTION STEPS

1. **Open First Tab - Main Hashtag**
   Navigate to: https://www.instagram.com/explore/tags/photography/
   Wait for the page to load completely.

2. **Open Second Tab - Related Hashtag**
   Use 'new_tab' to open: https://www.instagram.com/explore/tags/streetphotography/
   This opens in a NEW tab while keeping the first tab active.

3. **Open Third Tab - Niche Hashtag**
   Use 'new_tab' to open: https://www.instagram.com/explore/tags/mobilephotography/

4. **Verify Tab Setup**
   Use 'list_tabs' to confirm all 3 tabs are open.
   Report the tab IDs and URLs.

5. **Quick Analysis Per Tab**
   For each tab (use 'switch_tab' to navigate):
   - Count approximately how many "Top posts" are visible
   - Note the general content style (portraits, landscapes, urban, etc.)
   - Check if posts appear recent (look for time indicators)

## OUTPUT FORMAT
{
  "tabs_opened": [
    {"tab_id": "<id>", "hashtag": "photography", "url": "<url>"},
    {"tab_id": "<id>", "hashtag": "streetphotography", "url": "<url>"},
    {"tab_id": "<id>", "hashtag": "mobilephotography", "url": "<url>"}
  ],
  "quick_analysis": {
    "photography": {"top_posts_visible": <n>, "dominant_style": "<style>"},
    "streetphotography": {"top_posts_visible": <n>, "dominant_style": "<style>"},
    "mobilephotography": {"top_posts_visible": <n>, "dominant_style": "<style>"}
  },
  "multi_tab_success": true
}

## CONSTRAINTS
- Do NOT log in to Instagram (use public explore pages)
- If a page requires login, note it and continue
- Keep all tabs open for the next phase
`

	result1, err := agent.Run(ctx, phase1Prompt)
	if err != nil {
		log.Printf("Phase 1 error: %v", err)
	}
	printResult("Multi-Tab Setup", result1)

	// ═══════════════════════════════════════════════════════════════════════
	// PHASE 2: CONTENT DISCOVERY WITH PERSISTENT STORAGE
	// Demonstrates: save_finding tool for data persistence
	// ═══════════════════════════════════════════════════════════════════════
	printPhase(2, "Content Discovery with Persistent Storage",
		"Analyzing posts and saving interesting content creators")

	phase2Prompt := `
## OBJECTIVE
Discover and save interesting content from the photography hashtag. Use the 'save_finding' tool to persist data that survives across all future interactions.

## CRITICAL TOOL USAGE
Use 'save_finding' to store each discovery. This is PERSISTENT MEMORY - saved data can be retrieved later even after 100+ steps!

Categories to use:
- "trending_post" - For high-engagement posts
- "content_creator" - For interesting accounts
- "content_pattern" - For recurring themes/styles

## EXECUTION STEPS

1. **Switch to #photography tab**
   Use 'switch_tab' to go to the photography hashtag page.

2. **Analyze Top Posts Section**
   Look at the top posts grid (usually 9 posts at the top).
   For each visually interesting post:
   - Click to open the post modal/detail view
   - Extract: post type, visible engagement (likes/comments if shown), content description
   - Note any visible username

3. **Save 3 Trending Posts**
   For each interesting post found, use 'save_finding':

   Example call:
   save_finding(
     category: "trending_post",
     title: "Urban night photography",
     details: "High contrast cityscape, ~5000 likes visible, posted by @example_user"
   )

4. **Identify Content Patterns**
   After reviewing posts, identify 2-3 recurring patterns:
   - Common editing styles
   - Popular subjects
   - Engagement patterns

   Save each pattern with 'save_finding' using category "content_pattern"

5. **Save Notable Creator**
   If you find an account with interesting content, save it:
   save_finding(
     category: "content_creator",
     title: "@username",
     details: "Specializes in X, consistent style, appears to have high engagement"
   )

## OUTPUT FORMAT
{
  "posts_analyzed": <number>,
  "findings_saved": {
    "trending_posts": <count>,
    "content_patterns": <count>,
    "content_creators": <count>
  },
  "key_insight": "<most interesting discovery>",
  "hashtag_health": "<active/moderate/slow based on post freshness>"
}

## CONSTRAINTS
- Save at least 5 findings total across all categories
- Do NOT follow/like/comment on any posts
- If login popup appears, dismiss it or work around it
- Close post modals before moving to next post
`

	result2, err := agent.Run(ctx, phase2Prompt)
	if err != nil {
		log.Printf("Phase 2 error: %v", err)
	}
	printResult("Content Discovery", result2)

	// ═══════════════════════════════════════════════════════════════════════
	// PHASE 3: CROSS-TAB COMPARISON ANALYSIS
	// Demonstrates: switching between tabs, comparative analysis
	// ═══════════════════════════════════════════════════════════════════════
	printPhase(3, "Cross-Tab Comparison Analysis",
		"Comparing content across different hashtags")

	phase3Prompt := `
## OBJECTIVE
Compare content characteristics across the three hashtag tabs to identify niche-specific trends.

## EXECUTION STEPS

1. **Street Photography Analysis**
   - Switch to the #streetphotography tab
   - Scroll through the top posts section
   - Identify 2-3 distinctive characteristics of this niche
   - Save 2 findings (trending_post or content_pattern)

2. **Mobile Photography Analysis**
   - Switch to the #mobilephotography tab
   - Scroll through the top posts section
   - Note how content differs from DSLR/professional photography
   - Save 2 findings

3. **Comparative Analysis**
   Based on all three hashtags, analyze:
   - Which hashtag appears most active? (freshest posts)
   - Which has higher average engagement? (if visible)
   - What's unique about each niche?

4. **Save Comparison Insight**
   Use save_finding with category "content_pattern":
   save_finding(
     category: "content_pattern",
     title: "Cross-hashtag comparison",
     details: "<your comparative analysis>"
   )

## OUTPUT FORMAT
{
  "street_photography": {
    "distinctive_traits": ["<trait1>", "<trait2>"],
    "activity_level": "<high/medium/low>",
    "findings_saved": <count>
  },
  "mobile_photography": {
    "distinctive_traits": ["<trait1>", "<trait2>"],
    "activity_level": "<high/medium/low>",
    "findings_saved": <count>
  },
  "comparison": {
    "most_active_hashtag": "<hashtag>",
    "most_engagement": "<hashtag or 'unable to determine'>",
    "key_differences": ["<diff1>", "<diff2>"]
  }
}

## CONSTRAINTS
- Use switch_tab to move between tabs (don't navigate away)
- Save at least 4 new findings in this phase
- Keep observations factual based on visible content
`

	result3, err := agent.Run(ctx, phase3Prompt)
	if err != nil {
		log.Printf("Phase 3 error: %v", err)
	}
	printResult("Cross-Tab Comparison", result3)

	// ═══════════════════════════════════════════════════════════════════════
	// PHASE 4: MEMORY RECALL & SYNTHESIS
	// Demonstrates: search_findings tool to recall saved data
	// ═══════════════════════════════════════════════════════════════════════
	printPhase(4, "Memory Recall & Research Synthesis",
		"Retrieving all saved findings and creating comprehensive report")

	phase4Prompt := `
## OBJECTIVE
Use the 'search_findings' tool to recall ALL data saved throughout this research session, then synthesize into a comprehensive report.

## CRITICAL: MEMORY RECALL
The 'search_findings' tool retrieves data you saved earlier. This demonstrates that findings persist across all steps!

## EXECUTION STEPS

1. **Recall All Trending Posts**
   search_findings(category: "trending_post")
   List all posts you saved earlier.

2. **Recall All Content Patterns**
   search_findings(category: "content_pattern")
   List all patterns identified.

3. **Recall All Content Creators**
   search_findings(category: "content_creator")
   List all creators saved.

4. **Full Search (No Filter)**
   search_findings(query: "")
   This returns EVERYTHING saved. Count total findings.

5. **Synthesize Research Report**
   Based on ALL recalled findings, create a comprehensive summary:
   - Total findings across categories
   - Key themes and patterns
   - Actionable insights for content strategy
   - Recommended hashtags for different content types

## OUTPUT FORMAT
{
  "memory_recall": {
    "trending_posts_found": <count>,
    "content_patterns_found": <count>,
    "content_creators_found": <count>,
    "total_findings": <count>
  },
  "research_synthesis": {
    "top_content_themes": ["<theme1>", "<theme2>", "<theme3>"],
    "engagement_patterns": "<summary of what drives engagement>",
    "niche_recommendations": {
      "photography": "<recommendation>",
      "streetphotography": "<recommendation>",
      "mobilephotography": "<recommendation>"
    },
    "content_strategy_insight": "<key takeaway for content creators>"
  },
  "findings_detail": [
    {"category": "<cat>", "title": "<title>", "summary": "<brief>"}
  ]
}

## SUCCESS CRITERIA
- Must recall findings from Phase 2 and Phase 3 (proves persistence!)
- Total findings should be 9+ (5 from Phase 2, 4+ from Phase 3)
- Synthesis should reference specific saved findings
`

	result4, err := agent.Run(ctx, phase4Prompt)
	if err != nil {
		log.Printf("Phase 4 error: %v", err)
	}
	printResult("Memory Recall & Synthesis", result4)

	// ═══════════════════════════════════════════════════════════════════════
	// PHASE 5: TAB CLEANUP & FINAL REPORT
	// Demonstrates: close_tab, final synthesis
	// ═══════════════════════════════════════════════════════════════════════
	printPhase(5, "Cleanup & Final Report",
		"Closing extra tabs and generating final research report")

	phase5Prompt := `
## OBJECTIVE
Clean up browser tabs and generate the final research summary.

## EXECUTION STEPS

1. **List Current Tabs**
   Use 'list_tabs' to see all open tabs.

2. **Close Extra Tabs**
   Use 'close_tab' to close all but ONE tab.
   Note: Cannot close the last remaining tab.

3. **Final Tab Check**
   Use 'list_tabs' to confirm only one tab remains.

4. **Generate Final Report**
   Based on everything discovered in this session:

   Create a comprehensive "Content Research Report" with:
   - Executive Summary (2-3 sentences)
   - Hashtags Analyzed
   - Key Findings (bullet points)
   - Content Opportunities (what's missing/underserved)
   - Recommended Actions

## OUTPUT FORMAT
{
  "tab_cleanup": {
    "tabs_before": <count>,
    "tabs_closed": <count>,
    "tabs_remaining": 1
  },
  "final_report": {
    "executive_summary": "<2-3 sentence overview>",
    "hashtags_analyzed": ["photography", "streetphotography", "mobilephotography"],
    "key_findings": [
      "<finding1>",
      "<finding2>",
      "<finding3>"
    ],
    "content_opportunities": [
      "<opportunity1>",
      "<opportunity2>"
    ],
    "recommended_actions": [
      "<action1>",
      "<action2>",
      "<action3>"
    ]
  },
  "session_stats": {
    "total_findings_saved": <count>,
    "tabs_used": 3,
    "phases_completed": 5
  }
}
`

	result5, err := agent.Run(ctx, phase5Prompt)
	if err != nil {
		log.Printf("Phase 5 error: %v", err)
	}
	printResult("Final Report", result5)

	// ═══════════════════════════════════════════════════════════════════════
	// SESSION SUMMARY
	// ═══════════════════════════════════════════════════════════════════════
	fmt.Println()
	printHeader("RESEARCH SESSION COMPLETE", "")
	fmt.Println()
	fmt.Println("  ✅ Phase 1: Multi-Tab Setup         - Opened 3 hashtag tabs")
	fmt.Println("  ✅ Phase 2: Content Discovery       - Saved findings to memory")
	fmt.Println("  ✅ Phase 3: Cross-Tab Comparison    - Analyzed multiple niches")
	fmt.Println("  ✅ Phase 4: Memory Recall           - Retrieved all saved data")
	fmt.Println("  ✅ Phase 5: Cleanup & Report        - Generated final insights")
	fmt.Println()
	fmt.Println("  📊 Capabilities Demonstrated:")
	fmt.Println("     • Multi-tab browser management (new_tab, switch_tab, close_tab)")
	fmt.Println("     • Persistent findings storage (save_finding persists 100+ steps)")
	fmt.Println("     • Memory recall across phases (search_findings)")
	fmt.Println("     • Complex navigation and scrolling")
	fmt.Println("     • Structured data extraction with JSON schemas")
	fmt.Println("     • Conditional logic and error handling")
	fmt.Println()
	fmt.Println("  🔗 This example shows bua-go as a GENERAL-PURPOSE automation toolkit")
	fmt.Println("     suitable for research, testing, data extraction, and more!")
	fmt.Println()
}

// ═══════════════════════════════════════════════════════════════════════════
// HELPER FUNCTIONS
// ═══════════════════════════════════════════════════════════════════════════

func printHeader(title, subtitle string) {
	width := 70
	fmt.Println()
	fmt.Println("╔" + strings.Repeat("═", width-2) + "╗")
	fmt.Printf("║%s║\n", centerText(title, width-2))
	if subtitle != "" {
		fmt.Printf("║%s║\n", centerText(subtitle, width-2))
	}
	fmt.Println("╚" + strings.Repeat("═", width-2) + "╝")
	fmt.Println()
}

func printPhase(num int, title, description string) {
	fmt.Println()
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  📌 PHASE %d: %s\n", num, title)
	fmt.Printf("  %s\n", description)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func printResult(title string, result *bua.Result) {
	fmt.Printf("\n┌─ %s ", title)
	for i := 0; i < 55-len(title); i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	if result == nil {
		fmt.Println("│ ❌ Result is nil")
		fmt.Println("└" + strings.Repeat("─", 60) + "┘")
		return
	}

	fmt.Printf("│ Status: %s\n", statusIcon(result.Success))

	if result.Data != nil {
		data, err := json.MarshalIndent(result.Data, "│ ", "  ")
		if err == nil {
			// Truncate if too long for display
			dataStr := string(data)
			if len(dataStr) > 2000 {
				dataStr = dataStr[:2000] + "\n│   ... (truncated)"
			}
			fmt.Printf("│ Data:\n%s\n", dataStr)
		}
	}

	if result.Error != "" {
		fmt.Printf("│ ⚠️  Error: %s\n", result.Error)
	}

	fmt.Println("└" + strings.Repeat("─", 60) + "┘")
}

func statusIcon(success bool) string {
	if success {
		return "✅ Success"
	}
	return "❌ Failed"
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	leftPad := (width - len(text)) / 2
	rightPad := width - len(text) - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}
