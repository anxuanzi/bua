package dom

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// AccessibilityNode represents a node in the accessibility tree.
type AccessibilityNode struct {
	// NodeID is the unique identifier for this node.
	NodeID string `json:"nodeId"`

	// Role is the accessibility role (e.g., "button", "textbox").
	Role string `json:"role"`

	// Name is the accessible name.
	Name string `json:"name,omitempty"`

	// Description is the accessible description.
	Description string `json:"description,omitempty"`

	// Value is the current value.
	Value string `json:"value,omitempty"`

	// Properties contains additional accessibility properties.
	Properties map[string]any `json:"properties,omitempty"`

	// IsIgnored indicates if this node is ignored in the accessibility tree.
	IsIgnored bool `json:"isIgnored"`

	// Children are the child nodes.
	Children []*AccessibilityNode `json:"children,omitempty"`

	// BackendDOMNodeID links to the DOM node.
	BackendDOMNodeID int `json:"backendDomNodeId,omitempty"`
}

// AccessibilityTree represents the full accessibility tree.
type AccessibilityTree struct {
	Root    *AccessibilityNode `json:"root"`
	NodeMap map[string]*AccessibilityNode
}

// NewAccessibilityTree creates a new accessibility tree.
func NewAccessibilityTree() *AccessibilityTree {
	return &AccessibilityTree{
		NodeMap: make(map[string]*AccessibilityNode),
	}
}

// ToTokenString converts the accessibility tree to a token-efficient string for LLM context.
func (t *AccessibilityTree) ToTokenString() string {
	var sb strings.Builder
	sb.WriteString("Accessibility Tree:\n")
	if t.Root != nil {
		writeNode(&sb, t.Root, 0)
	}
	return sb.String()
}

func writeNode(sb *strings.Builder, node *AccessibilityNode, depth int) {
	if node.IsIgnored && len(node.Children) == 0 {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Write node info
	if !node.IsIgnored {
		sb.WriteString(fmt.Sprintf("%s- %s", indent, node.Role))
		if node.Name != "" {
			sb.WriteString(fmt.Sprintf(": %q", truncate(node.Name, 50)))
		}
		if node.Value != "" {
			sb.WriteString(fmt.Sprintf(" [value=%q]", truncate(node.Value, 30)))
		}
		sb.WriteString("\n")
	}

	// Write children
	for _, child := range node.Children {
		writeNode(sb, child, depth+1)
	}
}

// InteractiveNodes returns all interactive nodes in the tree.
func (t *AccessibilityTree) InteractiveNodes() []*AccessibilityNode {
	var result []*AccessibilityNode
	if t.Root != nil {
		collectInteractive(t.Root, &result)
	}
	return result
}

func collectInteractive(node *AccessibilityNode, result *[]*AccessibilityNode) {
	if !node.IsIgnored && isInteractiveRole(node.Role) {
		*result = append(*result, node)
	}
	for _, child := range node.Children {
		collectInteractive(child, result)
	}
}

func isInteractiveRole(role string) bool {
	interactiveRoles := map[string]bool{
		"button":     true,
		"link":       true,
		"textbox":    true,
		"checkbox":   true,
		"radio":      true,
		"combobox":   true,
		"listbox":    true,
		"menu":       true,
		"menuitem":   true,
		"tab":        true,
		"switch":     true,
		"slider":     true,
		"spinbutton": true,
		"searchbox":  true,
	}
	return interactiveRoles[role]
}

// CDPAXNode represents a node from the CDP Accessibility.getFullAXTree response.
type CDPAXNode struct {
	NodeID           string          `json:"nodeId"`
	Ignored          bool            `json:"ignored"`
	IgnoredReasons   []CDPAXProperty `json:"ignoredReasons,omitempty"`
	Role             *CDPAXValue     `json:"role,omitempty"`
	Name             *CDPAXValue     `json:"name,omitempty"`
	Description      *CDPAXValue     `json:"description,omitempty"`
	Value            *CDPAXValue     `json:"value,omitempty"`
	Properties       []CDPAXProperty `json:"properties,omitempty"`
	ChildIDs         []string        `json:"childIds,omitempty"`
	ParentID         string          `json:"parentId,omitempty"`
	BackendDOMNodeID int             `json:"backendDOMNodeId,omitempty"`
}

// CDPAXValue represents an accessibility value from CDP.
type CDPAXValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// CDPAXProperty represents an accessibility property from CDP.
type CDPAXProperty struct {
	Name  string      `json:"name"`
	Value *CDPAXValue `json:"value,omitempty"`
}

// CDPAXTreeResponse is the response from Accessibility.getFullAXTree.
type CDPAXTreeResponse struct {
	Nodes []CDPAXNode `json:"nodes"`
}

// ExtractAccessibilityTree extracts the accessibility tree from the page using CDP.
func ExtractAccessibilityTree(ctx context.Context, page *rod.Page) (*AccessibilityTree, error) {
	tree := NewAccessibilityTree()

	// Call CDP Accessibility.getFullAXTree directly via rod
	result, err := page.Call(ctx, "", "Accessibility.getFullAXTree", nil)
	if err != nil {
		// Fall back to JavaScript-based extraction
		return extractAccessibilityTreeJS(ctx, page)
	}

	// Parse the response
	var response CDPAXTreeResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return extractAccessibilityTreeJS(ctx, page)
	}

	if len(response.Nodes) == 0 {
		return tree, nil
	}

	// Build node map
	nodeMap := make(map[string]*AccessibilityNode)
	childMap := make(map[string][]string)

	for _, n := range response.Nodes {
		node := convertCDPNode(&n)
		nodeMap[n.NodeID] = node
		tree.NodeMap[n.NodeID] = node

		// Track parent-child relationships
		if n.ParentID != "" {
			childMap[n.ParentID] = append(childMap[n.ParentID], n.NodeID)
		}
	}

	// Build tree structure using childMap
	for parentID, childIDs := range childMap {
		if parent, ok := nodeMap[parentID]; ok {
			for _, childID := range childIDs {
				if child, ok := nodeMap[childID]; ok {
					parent.Children = append(parent.Children, child)
				}
			}
		}
	}

	// Find root (node without parent)
	for _, n := range response.Nodes {
		if n.ParentID == "" {
			tree.Root = nodeMap[n.NodeID]
			break
		}
	}

	return tree, nil
}

func convertCDPNode(n *CDPAXNode) *AccessibilityNode {
	node := &AccessibilityNode{
		NodeID:           n.NodeID,
		IsIgnored:        n.Ignored,
		BackendDOMNodeID: n.BackendDOMNodeID,
	}

	if n.Role != nil {
		if role, ok := n.Role.Value.(string); ok {
			node.Role = role
		}
	}

	if n.Name != nil {
		if name, ok := n.Name.Value.(string); ok {
			node.Name = name
		}
	}

	if n.Description != nil {
		if desc, ok := n.Description.Value.(string); ok {
			node.Description = desc
		}
	}

	if n.Value != nil {
		if val, ok := n.Value.Value.(string); ok {
			node.Value = val
		}
	}

	// Extract properties
	if len(n.Properties) > 0 {
		node.Properties = make(map[string]any)
		for _, prop := range n.Properties {
			if prop.Value != nil {
				node.Properties[prop.Name] = prop.Value.Value
			}
		}
	}

	return node
}

// extractAccessibilityTreeJS is a fallback that uses JavaScript to extract accessibility info.
func extractAccessibilityTreeJS(ctx context.Context, page *rod.Page) (*AccessibilityTree, error) {
	tree := NewAccessibilityTree()

	// JavaScript fallback for accessibility info
	js := `() => {
		function getAccessibleInfo(el, depth = 0) {
			if (depth > 10) return null;

			const role = el.getAttribute('role') || inferRole(el);
			const name = getAccessibleName(el);
			const value = el.value || '';

			const node = {
				role: role,
				name: name,
				value: value,
				children: []
			};

			// Get interactive children
			for (const child of el.children) {
				const childInfo = getAccessibleInfo(child, depth + 1);
				if (childInfo && (childInfo.role || childInfo.children.length > 0)) {
					node.children.push(childInfo);
				}
			}

			return node;
		}

		function inferRole(el) {
			const tag = el.tagName.toLowerCase();
			const roleMap = {
				'button': 'button',
				'a': 'link',
				'input': el.type === 'checkbox' ? 'checkbox' : el.type === 'radio' ? 'radio' : 'textbox',
				'select': 'combobox',
				'textarea': 'textbox',
				'nav': 'navigation',
				'main': 'main',
				'header': 'banner',
				'footer': 'contentinfo',
				'aside': 'complementary',
				'form': 'form'
			};
			return roleMap[tag] || '';
		}

		function getAccessibleName(el) {
			return el.getAttribute('aria-label') ||
				   el.getAttribute('aria-labelledby') && document.getElementById(el.getAttribute('aria-labelledby'))?.textContent ||
				   el.getAttribute('title') ||
				   el.getAttribute('alt') ||
				   (el.tagName.toLowerCase() === 'input' && el.labels?.[0]?.textContent) ||
				   el.innerText?.substring(0, 100) ||
				   '';
		}

		return getAccessibleInfo(document.body);
	}`

	result, err := page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("failed to extract accessibility tree: %w", err)
	}

	if result.Value.Nil() {
		return tree, nil
	}

	var rawNode map[string]any
	if err := result.Value.Unmarshal(&rawNode); err != nil {
		return nil, fmt.Errorf("failed to parse accessibility tree: %w", err)
	}

	tree.Root = parseJSNode(rawNode, "0")
	return tree, nil
}

func parseJSNode(raw map[string]any, id string) *AccessibilityNode {
	node := &AccessibilityNode{
		NodeID: id,
		Role:   getString(raw, "role"),
		Name:   getString(raw, "name"),
		Value:  getString(raw, "value"),
	}

	if children, ok := raw["children"].([]any); ok {
		for i, child := range children {
			if childMap, ok := child.(map[string]any); ok {
				childNode := parseJSNode(childMap, fmt.Sprintf("%s.%d", id, i))
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}

// MergeWithElementMap enriches an element map with accessibility information.
func MergeWithElementMap(elements *ElementMap, tree *AccessibilityTree) {
	if tree == nil || tree.Root == nil || elements == nil {
		return
	}

	// Create a map of interactive nodes by their approximate position
	interactiveNodes := tree.InteractiveNodes()

	// Try to match accessibility nodes to elements
	for _, el := range elements.Elements {
		// Find best matching accessibility node
		for _, node := range interactiveNodes {
			// Match by name similarity
			if el.Role == "" && node.Role != "" {
				el.Role = node.Role
			}
			if el.Name == "" && node.Name != "" && (el.Text == node.Name || el.AriaLabel == node.Name) {
				el.Name = node.Name
			}
		}
	}
}

// Ensure proto is used to avoid import error
var _ = proto.TargetCreateTarget{}
