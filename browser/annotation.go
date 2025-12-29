// Package browser provides the browser automation layer using go-rod.
package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anxuanzi/bua-go/dom"
)

// AnnotationConfig holds configuration for element annotations.
type AnnotationConfig struct {
	// ShowIndex displays the element index number
	ShowIndex bool
	// ShowType displays the element type (button, input, link, etc.)
	ShowType bool
	// ShowBoundingBox draws a border around elements
	ShowBoundingBox bool
	// Opacity of the overlay (0.0 - 1.0)
	Opacity float64
}

// DefaultAnnotationConfig returns the default annotation configuration.
func DefaultAnnotationConfig() *AnnotationConfig {
	return &AnnotationConfig{
		ShowIndex:       true,
		ShowType:        true,
		ShowBoundingBox: true,
		Opacity:         0.8,
	}
}

// annotationCSS returns the CSS for element annotations.
func annotationCSS(opacity float64) string {
	return fmt.Sprintf(`
		.bua-annotation-overlay {
			position: fixed;
			pointer-events: none;
			z-index: 2147483647;
			top: 0;
			left: 0;
			width: 100%%;
			height: 100%%;
		}
		.bua-element-box {
			position: absolute;
			border: 2px solid;
			box-sizing: border-box;
			pointer-events: none;
		}
		.bua-element-label {
			position: absolute;
			font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', monospace;
			font-size: 10px;
			font-weight: bold;
			padding: 2px 4px;
			border-radius: 3px;
			white-space: nowrap;
			opacity: %.2f;
			pointer-events: none;
		}
		/* Element type colors */
		.bua-type-button { border-color: #e74c3c; }
		.bua-type-button .bua-element-label { background: #e74c3c; color: white; }

		.bua-type-link { border-color: #3498db; }
		.bua-type-link .bua-element-label { background: #3498db; color: white; }

		.bua-type-input { border-color: #2ecc71; }
		.bua-type-input .bua-element-label { background: #2ecc71; color: white; }

		.bua-type-select { border-color: #9b59b6; }
		.bua-type-select .bua-element-label { background: #9b59b6; color: white; }

		.bua-type-textarea { border-color: #1abc9c; }
		.bua-type-textarea .bua-element-label { background: #1abc9c; color: white; }

		.bua-type-image { border-color: #f39c12; }
		.bua-type-image .bua-element-label { background: #f39c12; color: white; }

		.bua-type-other { border-color: #95a5a6; }
		.bua-type-other .bua-element-label { background: #95a5a6; color: white; }
	`, opacity)
}

// getElementTypeClass returns the CSS class for an element type.
func getElementTypeClass(tagName string, el *dom.Element) string {
	switch tagName {
	case "button":
		return "bua-type-button"
	case "a":
		return "bua-type-link"
	case "input":
		if el != nil {
			switch el.Type {
			case "submit", "button":
				return "bua-type-button"
			default:
				return "bua-type-input"
			}
		}
		return "bua-type-input"
	case "select":
		return "bua-type-select"
	case "textarea":
		return "bua-type-textarea"
	case "img":
		return "bua-type-image"
	default:
		// Check if it's clickable (has onclick or role=button)
		if el != nil && el.Role == "button" {
			return "bua-type-button"
		}
		return "bua-type-other"
	}
}

// elementAnnotation holds data for a single element annotation (used for batched JS).
type elementAnnotation struct {
	TypeClass string  `json:"t"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	W         float64 `json:"w"`
	H         float64 `json:"h"`
	Label     string  `json:"l"`
}

// buildElementAnnotations creates annotation data for all elements.
func buildElementAnnotations(elements []*dom.Element, cfg *AnnotationConfig) []elementAnnotation {
	var annotations []elementAnnotation
	for _, el := range elements {
		if el.BoundingBox.Width <= 0 || el.BoundingBox.Height <= 0 {
			continue
		}

		typeClass := getElementTypeClass(el.TagName, el)

		labelText := ""
		if cfg.ShowIndex {
			labelText = fmt.Sprintf("%d", el.Index)
		}
		if cfg.ShowType && el.TagName != "" {
			if labelText != "" {
				labelText += " "
			}
			labelText += el.TagName
		}

		annotations = append(annotations, elementAnnotation{
			TypeClass: typeClass,
			X:         el.BoundingBox.X,
			Y:         el.BoundingBox.Y,
			W:         el.BoundingBox.Width,
			H:         el.BoundingBox.Height,
			Label:     labelText,
		})
	}
	return annotations
}

// ShowAnnotations draws annotation overlays on all detected elements.
// Uses batched JavaScript for performance (single CDP call instead of O(n) calls).
func (b *Browser) ShowAnnotations(ctx context.Context, elements *dom.ElementMap, cfg *AnnotationConfig) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	if cfg == nil {
		cfg = DefaultAnnotationConfig()
	}

	// Build annotation data for all elements
	annotations := buildElementAnnotations(elements.InteractiveElements(), cfg)

	// Marshal to JSON for passing to JavaScript
	annotationsJSON, err := json.Marshal(annotations)
	if err != nil {
		return fmt.Errorf("failed to marshal annotations: %w", err)
	}

	css := annotationCSS(cfg.Opacity)

	// Single batched JavaScript call that:
	// 1. Clears existing annotations
	// 2. Injects CSS
	// 3. Creates container
	// 4. Adds all element boxes in one go
	batchedJS := fmt.Sprintf(`(annotationsData) => {
		// Clear existing annotations
		const existing = document.getElementById('bua-annotation-container');
		if (existing) existing.remove();

		// Inject or update CSS
		let style = document.getElementById('bua-annotation-style');
		if (!style) {
			style = document.createElement('style');
			style.id = 'bua-annotation-style';
			document.head.appendChild(style);
		}
		style.textContent = %q;

		// Create overlay container
		const container = document.createElement('div');
		container.id = 'bua-annotation-container';
		container.className = 'bua-annotation-overlay';

		// Add all element boxes in a single loop (no DOM access per iteration)
		const fragment = document.createDocumentFragment();
		for (const el of annotationsData) {
			const box = document.createElement('div');
			box.className = 'bua-element-box ' + el.t;
			box.style.cssText = 'left:' + el.x + 'px;top:' + el.y + 'px;width:' + el.w + 'px;height:' + el.h + 'px';

			const label = document.createElement('div');
			label.className = 'bua-element-label';
			label.textContent = el.l;
			label.style.cssText = 'left:0;top:-18px';

			box.appendChild(label);
			fragment.appendChild(box);
		}

		container.appendChild(fragment);
		document.body.appendChild(container);

		// Wait for browser paint to complete (prevents white background in screenshots)
		return new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
	}`, css)

	_, err = page.Eval(batchedJS, annotationsJSON)
	if err != nil {
		return fmt.Errorf("failed to create annotations: %w", err)
	}

	return nil
}

// HideAnnotations removes all annotation overlays from the page.
func (b *Browser) HideAnnotations(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	page := b.getActivePageLocked()
	if page == nil {
		return fmt.Errorf("no active page")
	}

	_, err := page.Eval(`() => {
		const container = document.getElementById('bua-annotation-container');
		if (container) container.remove();
		const style = document.getElementById('bua-annotation-style');
		if (style) style.remove();
	}`)
	if err != nil {
		return fmt.Errorf("failed to remove annotations: %w", err)
	}

	return nil
}

// ToggleAnnotations shows or hides annotations based on current state.
func (b *Browser) ToggleAnnotations(ctx context.Context, elements *dom.ElementMap, cfg *AnnotationConfig) (bool, error) {
	b.mu.RLock()
	page := b.getActivePageLocked()
	if page == nil {
		b.mu.RUnlock()
		return false, fmt.Errorf("no active page")
	}
	b.mu.RUnlock()

	// Check if annotations currently exist
	result, err := page.Eval(`() => {
		return document.getElementById('bua-annotation-container') !== null;
	}`)
	if err != nil {
		return false, fmt.Errorf("failed to check annotation state: %w", err)
	}

	hasAnnotations := result.Value.Bool()

	if hasAnnotations {
		err = b.HideAnnotations(ctx)
		return false, err
	}

	err = b.ShowAnnotations(ctx, elements, cfg)
	return true, err
}
