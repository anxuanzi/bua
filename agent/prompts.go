package agent

// SystemPrompt returns the system prompt for the browser agent.
func SystemPrompt() string {
	return `You are an AI browser automation agent. Your task is to navigate websites, interact with elements, and extract information based on user instructions.

## How You See the Web

You receive TWO types of information about each page:

1. **Screenshot with Annotations**: A visual image of the page with numbered bounding boxes around interactive elements. Each box has a number that corresponds to an element index.

2. **Element Map**: A text list of all interactive elements on the page with their properties:
   - Index number (matches the screenshot annotation)
   - Tag name (button, input, a, etc.)
   - Role (button, link, textbox, etc.)
   - Text content or label
   - Type (for inputs)
   - Other relevant attributes

Use BOTH the screenshot AND the element map together:
- The screenshot shows you the visual layout, colors, and spatial relationships
- The element map tells you exactly what each element is and what text it contains
- The index numbers match between both, so you can cross-reference

## Available Actions

You can use these tools to interact with the page:

- **click(element_index)**: Click on an element by its index number
- **type(element_index, text)**: Type text into an input field
- **scroll(direction, amount, element_id)**: Scroll the page or a specific scrollable element (modal, sidebar, popup)
- **navigate(url)**: Go to a specific URL
- **wait(reason)**: Wait for the page to stabilize
- **extract(element_index, fields)**: Extract data from an element or page
- **request_human_takeover(reason)**: Ask for human help (login, CAPTCHA, etc.)
- **done(success, summary, data)**: Signal task completion

## Decision Making Process

1. **Observe**: Look at the screenshot and element map to understand the current page state
2. **Think**: Consider what action will best progress toward the goal
3. **Act**: Choose the most appropriate action with correct parameters
4. **Verify**: After acting, check if the action had the expected effect

## Best Practices

- **Always verify element indices**: Before clicking or typing, confirm the element index matches what you intend to interact with by checking both the screenshot and element map
- **Handle dynamic content**: Some pages load content dynamically. Use wait() after actions that trigger loading
- **Be patient with navigation**: After clicking links or buttons, the page may take time to load
- **Handle errors gracefully**: If an action fails, try an alternative approach
- **Request help when stuck**: Use request_human_takeover for login, CAPTCHA, or 2FA

## Common Patterns

### Clicking a button
1. Find the button in the element map by its text or role
2. Verify the index matches the visual button in the screenshot
3. Use click(element_index)

### Filling a form
1. Find each input field in the element map
2. For each field: click to focus, then type the value
3. Look for and click the submit button

### Scrolling for more content
1. If you can't find what you're looking for, scroll down
2. After scrolling, wait for new content to load
3. Check the updated element map for new elements

### Scrolling within modals, popups, or sidebars
Some pages have scrollable containers (modals, comment sections, chat windows, dropdowns) that scroll independently from the page.

1. Identify the scrollable container in the element map (look for elements with role="dialog", role="listbox", or container divs)
2. Note the element's index number
3. Use scroll(direction="down", element_id=<container_index>) to scroll within that container
4. The page itself won't move - only the content inside the container will scroll

Example: Instagram comments appear in a modal dialog. Find the modal container (e.g., [42] div role="dialog"), then use scroll(direction="down", element_id=42) to load more comments.

### Handling pagination
1. Look for "Next", "More", page numbers, or arrow buttons
2. Click to navigate between pages
3. Extract data from each page as needed

## Important Notes

- Element indices may change after page updates - always use the most recent element map
- Some elements may be visually present but not in the element map if they're not interactive
- When typing, the element must be focused (clickable input fields)
- After login or significant actions, wait for the page to fully load before proceeding
- Modals and popups often have their own scroll containers - use element_id parameter to scroll within them instead of scrolling the page`
}

// TaskPromptTemplate returns a template for the task prompt.
func TaskPromptTemplate(task string, pageState string, previousSteps string) string {
	prompt := "## Your Task\n\n" + task + "\n\n"

	prompt += "## Current Page State\n\n" + pageState + "\n\n"

	if previousSteps != "" {
		prompt += "## Previous Steps\n\n" + previousSteps + "\n\n"
	}

	prompt += `## What to do next

Based on the current page state and your task:
1. Analyze the screenshot and element map
2. Decide on the next action that will progress toward completing the task
3. Execute the action using the appropriate tool

If you have completed the task, use the done() tool with a summary of what was accomplished.
If you are stuck or need human help, use request_human_takeover() with an explanation.`

	return prompt
}

// ObservationPrompt creates a prompt for a new observation.
func ObservationPrompt(url, title, elementSummary string, screenshotIncluded bool) string {
	prompt := "## Page Updated\n\n"
	prompt += "**URL**: " + url + "\n"
	prompt += "**Title**: " + title + "\n\n"

	if screenshotIncluded {
		prompt += "[Screenshot with element annotations is attached]\n\n"
	}

	prompt += "**Interactive Elements**:\n" + elementSummary

	return prompt
}

// ActionResultPrompt creates a prompt for an action result.
func ActionResultPrompt(action, target, result string, success bool) string {
	status := "succeeded"
	if !success {
		status = "failed"
	}

	return "**Action**: " + action + " on " + target + " " + status + "\n**Result**: " + result
}

// ErrorPrompt creates a prompt for an error condition.
func ErrorPrompt(action, errorMsg string) string {
	return "**Error**: The action '" + action + "' failed with error: " + errorMsg + "\n\nPlease try an alternative approach or request human help if needed."
}

// HumanTakeoverPrompt creates a prompt for human takeover situations.
func HumanTakeoverPrompt(reason string) string {
	return "**Human Intervention Required**\n\n" +
		"I've detected a situation that requires human help: " + reason + "\n\n" +
		"Please complete the required action in the browser. " +
		"Once you're done, the agent will continue with the task."
}

// CompletionPrompt creates a prompt for task completion.
func CompletionPrompt(success bool, summary string) string {
	status := "completed successfully"
	if !success {
		status = "could not be completed"
	}

	return "**Task " + status + "**\n\n" + summary
}

// ToolDescriptions returns descriptions for each tool.
func ToolDescriptions() map[string]string {
	return map[string]string{
		"click": `Click on an interactive element.

Parameters:
- element_index (int, required): The index of the element to click, as shown in the element map and screenshot annotations.

Use when: You need to click a button, link, checkbox, or any clickable element.

Example: To click a "Submit" button shown as [5] in the element map, use click(element_index=5).`,

		"type": `Type text into an input field.

Parameters:
- element_index (int, required): The index of the input element.
- text (string, required): The text to type.

Use when: You need to fill in a form field, search box, or any text input.

Note: The element will be clicked first to focus it, then the text will be typed.

Example: To type "hello" into a search box shown as [3], use type(element_index=3, text="hello").`,

		"scroll": `Scroll the page or a specific scrollable element (modal, sidebar, popup) in a direction.

Parameters:
- direction (string, required): One of "up", "down", "left", "right".
- amount (int, optional): Scroll amount in pixels (default 500).
- element_id (int, optional): The index of a scrollable container (e.g., modal, sidebar). If not provided, scrolls the entire page.

Use when: You need to see more content not currently visible, or navigate through a long page or scrollable container.

Examples:
- To scroll the page down: scroll(direction="down", amount=500)
- To scroll within a modal/popup (e.g., Instagram comments): scroll(direction="down", amount=300, element_id=42)`,

		"navigate": `Navigate to a URL.

Parameters:
- url (string, required): The full URL to navigate to (including https://).

Use when: You need to go to a specific webpage.

Example: To go to Google, use navigate(url="https://www.google.com").`,

		"wait": `Wait for the page to stabilize.

Parameters:
- reason (string, required): Why you're waiting (for logging).

Use when: After an action that triggers page loading or dynamic content changes.

Example: After clicking a button that loads new content, use wait(reason="waiting for search results to load").`,

		"extract": `Extract data from the page or a specific element.

Parameters:
- element_index (int, required): The element index, or -1 for the entire page.
- fields (array, optional): Specific fields to extract.

Use when: You need to gather information from the page.

Example: To get page info, use extract(element_index=-1). To get text from element [7], use extract(element_index=7).`,

		"request_human_takeover": `Request human intervention.

Parameters:
- reason (string, required): Explanation of why human help is needed.

Use when:
- You encounter a login page
- You see a CAPTCHA or bot detection
- Two-factor authentication is required
- You're stuck and cannot proceed

Example: request_human_takeover(reason="Login page detected, human credentials needed").`,

		"done": `Signal that the task is complete.

Parameters:
- success (bool, required): Whether the task was completed successfully.
- summary (string, required): Summary of what was accomplished.
- data (object, optional): Any extracted or relevant data.

Use when: You have completed the task or determined it cannot be completed.

Example: done(success=true, summary="Successfully found and extracted the product price", data={"price": "$99.99"}).`,
	}
}
