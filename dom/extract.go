package dom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

// extractionJS is the JavaScript code injected to extract interactive elements.
// IMPORTANT: Must use arrow function syntax for rod.Eval()
const extractionJS = `() => {
    const elements = [];
    let index = 0;

    // Selectors for interactive elements
    const interactiveSelectors = [
        'a[href]',
        'button',
        'input:not([type="hidden"])',
        'select',
        'textarea',
        '[role="button"]',
        '[role="link"]',
        '[role="textbox"]',
        '[role="checkbox"]',
        '[role="radio"]',
        '[role="menuitem"]',
        '[role="tab"]',
        '[role="switch"]',
        '[role="combobox"]',
        '[onclick]',
        '[tabindex]:not([tabindex="-1"])',
        'summary',
        'details',
        'label[for]'
    ];

    const allElements = document.querySelectorAll(interactiveSelectors.join(','));
    const viewportHeight = window.innerHeight;
    const viewportWidth = window.innerWidth;

    for (const node of allElements) {
        const rect = node.getBoundingClientRect();

        // Skip elements with no size
        if (rect.width <= 0 || rect.height <= 0) continue;

        // Skip elements completely outside viewport (with buffer)
        const buffer = 100;
        if (rect.bottom < -buffer || rect.top > viewportHeight + buffer) continue;
        if (rect.right < -buffer || rect.left > viewportWidth + buffer) continue;

        // Check computed styles
        const style = window.getComputedStyle(node);
        if (style.display === 'none') continue;
        if (style.visibility === 'hidden') continue;
        if (parseFloat(style.opacity) < 0.1) continue;
        if (style.pointerEvents === 'none') continue;

        // Get text content (truncated)
        let text = '';
        if (node.tagName === 'INPUT' || node.tagName === 'TEXTAREA') {
            text = node.value || '';
        } else {
            text = (node.textContent || '').trim();
        }
        if (text.length > 100) {
            text = text.slice(0, 100) + '...';
        }

        // Build unique selector
        let selector = '';
        if (node.id) {
            selector = '#' + CSS.escape(node.id);
        } else if (node.className && typeof node.className === 'string') {
            const classes = node.className.trim().split(/\s+/).slice(0, 2);
            if (classes.length > 0 && classes[0]) {
                selector = node.tagName.toLowerCase() + '.' + classes.map(c => CSS.escape(c)).join('.');
            }
        }
        if (!selector) {
            selector = node.tagName.toLowerCase();
            const parent = node.parentElement;
            if (parent) {
                const siblings = Array.from(parent.children).filter(c => c.tagName === node.tagName);
                if (siblings.length > 1) {
                    const idx = siblings.indexOf(node) + 1;
                    selector += ':nth-of-type(' + idx + ')';
                }
            }
        }

        // Determine role
        let role = node.getAttribute('role') || '';
        if (!role) {
            const tagRoles = {
                'A': 'link',
                'BUTTON': 'button',
                'INPUT': node.type === 'checkbox' ? 'checkbox' :
                         node.type === 'radio' ? 'radio' :
                         node.type === 'submit' ? 'button' : 'textbox',
                'SELECT': 'combobox',
                'TEXTAREA': 'textbox'
            };
            role = tagRoles[node.tagName] || '';
        }

        elements.push({
            index: index,
            tagName: node.tagName.toLowerCase(),
            role: role,
            name: node.getAttribute('aria-label') || node.getAttribute('name') || '',
            text: text,
            type: node.type || '',
            href: node.href || '',
            placeholder: node.placeholder || '',
            value: node.value || '',
            ariaLabel: node.getAttribute('aria-label') || '',
            boundingBox: {
                x: rect.x,
                y: rect.y,
                width: rect.width,
                height: rect.height
            },
            isVisible: true,
            isEnabled: !node.disabled,
            isFocusable: node.tabIndex >= 0,
            isInteractive: true,
            selector: selector
        });

        index++;
    }

    return {
        elements: elements,
        pageUrl: window.location.href,
        pageTitle: document.title
    };
}`

// extractionResult is the structure returned by the extraction JavaScript.
type extractionResult struct {
	Elements  []*Element `json:"elements"`
	PageURL   string     `json:"pageUrl"`
	PageTitle string     `json:"pageTitle"`
}

// Extractor handles DOM element extraction from a page.
type Extractor struct {
	maxElements int
}

// NewExtractor creates a new DOM extractor.
func NewExtractor(maxElements int) *Extractor {
	if maxElements <= 0 {
		maxElements = 100
	}
	return &Extractor{maxElements: maxElements}
}

// Extract extracts interactive elements from the page.
func (e *Extractor) Extract(ctx context.Context, page *rod.Page) (*ElementMap, error) {
	// Wait for page to be ready (500ms stability window)
	_ = ctx // Context available for future use
	if err := page.WaitStable(500 * time.Millisecond); err != nil {
		// Continue even if wait fails - page might be dynamic
	}

	// Execute extraction JavaScript
	result, err := page.Eval(extractionJS)
	if err != nil {
		return nil, fmt.Errorf("dom extraction failed: %w", err)
	}

	// Parse the result
	var data extractionResult
	jsonBytes, err := result.Value.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extraction result: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	// Build element map with limit
	elementMap := NewElementMap()
	elementMap.PageURL = data.PageURL
	elementMap.PageTitle = data.PageTitle

	for i, el := range data.Elements {
		if i >= e.maxElements {
			break
		}
		elementMap.Add(el)
	}

	return elementMap, nil
}

// ExtractElementMap is a convenience function for extracting elements.
func ExtractElementMap(ctx context.Context, page *rod.Page, maxElements int) (*ElementMap, error) {
	extractor := NewExtractor(maxElements)
	return extractor.Extract(ctx, page)
}
