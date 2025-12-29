// Package agent provides structured data extraction capabilities.
package agent

import (
	"fmt"
	"regexp"
	"strings"
)

// ExtractionSchema defines the structure for data extraction.
type ExtractionSchema struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      []ExtractionField `json:"fields"`
	Validators  []SchemaValidator `json:"-"` // Custom validators
}

// ExtractionField defines a single field to extract.
type ExtractionField struct {
	Name        string           `json:"name"`
	Type        FieldType        `json:"type"`
	Required    bool             `json:"required"`
	Description string           `json:"description"`
	Pattern     string           `json:"pattern,omitempty"` // Regex pattern for validation
	Default     any              `json:"default,omitempty"`
	Validators  []FieldValidator `json:"-"` // Custom field validators
}

// FieldType represents the type of extracted field.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeEmail   FieldType = "email"
	FieldTypeURL     FieldType = "url"
	FieldTypePhone   FieldType = "phone"
	FieldTypeDate    FieldType = "date"
	FieldTypeList    FieldType = "list"
	FieldTypeObject  FieldType = "object"
)

// FieldValidator is a function that validates a field value.
type FieldValidator func(value any) error

// SchemaValidator validates the entire extracted data.
type SchemaValidator func(data map[string]any) error

// ExtractionResult contains the result of a structured extraction.
type ExtractionResult struct {
	Schema     string         `json:"schema"`
	Data       map[string]any `json:"data"`
	Confidence float64        `json:"confidence"`
	Valid      bool           `json:"valid"`
	Errors     []string       `json:"errors,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

// Extractor handles structured data extraction.
type Extractor struct {
	schemas map[string]*ExtractionSchema
}

// NewExtractor creates a new Extractor with built-in schemas.
func NewExtractor() *Extractor {
	e := &Extractor{
		schemas: make(map[string]*ExtractionSchema),
	}
	e.registerBuiltInSchemas()
	return e
}

// RegisterSchema adds a custom schema.
func (e *Extractor) RegisterSchema(schema *ExtractionSchema) {
	e.schemas[schema.Name] = schema
}

// GetSchema returns a schema by name.
func (e *Extractor) GetSchema(name string) (*ExtractionSchema, bool) {
	schema, ok := e.schemas[name]
	return schema, ok
}

// ListSchemas returns all registered schema names.
func (e *Extractor) ListSchemas() []string {
	names := make([]string, 0, len(e.schemas))
	for name := range e.schemas {
		names = append(names, name)
	}
	return names
}

// Validate checks if the data matches the schema.
func (e *Extractor) Validate(schemaName string, data map[string]any) *ExtractionResult {
	result := &ExtractionResult{
		Schema:     schemaName,
		Data:       data,
		Confidence: 1.0,
		Valid:      true,
	}

	schema, ok := e.schemas[schemaName]
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("unknown schema: %s", schemaName))
		return result
	}

	// Validate each field
	for _, field := range schema.Fields {
		value, exists := data[field.Name]

		// Check required fields
		if field.Required && !exists {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field.Name))
			result.Confidence *= 0.5
			continue
		}

		if !exists {
			// Use default if available
			if field.Default != nil {
				data[field.Name] = field.Default
			}
			continue
		}

		// Validate field type
		if err := e.validateFieldType(field, value); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("field %s: %v", field.Name, err))
			result.Confidence *= 0.7
		}

		// Validate pattern if specified
		if field.Pattern != "" {
			if err := e.validatePattern(field.Pattern, value); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("field %s pattern mismatch: %v", field.Name, err))
				result.Confidence *= 0.9
			}
		}

		// Run custom field validators
		for _, validator := range field.Validators {
			if err := validator(value); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("field %s validation: %v", field.Name, err))
				result.Confidence *= 0.95
			}
		}
	}

	// Run schema-level validators
	for _, validator := range schema.Validators {
		if err := validator(data); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("schema validation: %v", err))
			result.Confidence *= 0.9
		}
	}

	return result
}

// validateFieldType checks if the value matches the expected type.
func (e *Extractor) validateFieldType(field ExtractionField, value any) error {
	switch field.Type {
	case FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case FieldTypeNumber:
		switch value.(type) {
		case int, int64, float64, float32:
			return nil
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case FieldTypeEmail:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for email, got %T", value)
		}
		if !isValidEmail(s) {
			return fmt.Errorf("invalid email format: %s", s)
		}
	case FieldTypeURL:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for URL, got %T", value)
		}
		if !isValidURL(s) {
			return fmt.Errorf("invalid URL format: %s", s)
		}
	case FieldTypePhone:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for phone, got %T", value)
		}
		if !isValidPhone(s) {
			return fmt.Errorf("invalid phone format: %s", s)
		}
	case FieldTypeList:
		if _, ok := value.([]any); !ok {
			// Also accept typed slices
			switch value.(type) {
			case []string, []int, []float64, []map[string]any:
				return nil
			}
			return fmt.Errorf("expected list, got %T", value)
		}
	case FieldTypeObject:
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// validatePattern checks if the value matches the regex pattern.
func (e *Extractor) validatePattern(pattern string, value any) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("pattern validation requires string value")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}

	if !re.MatchString(s) {
		return fmt.Errorf("value does not match pattern")
	}
	return nil
}

// registerBuiltInSchemas adds common extraction schemas.
func (e *Extractor) registerBuiltInSchemas() {
	// Contact schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "contact",
		Description: "Contact information extraction",
		Fields: []ExtractionField{
			{Name: "name", Type: FieldTypeString, Required: true, Description: "Full name"},
			{Name: "email", Type: FieldTypeEmail, Required: false, Description: "Email address"},
			{Name: "phone", Type: FieldTypePhone, Required: false, Description: "Phone number"},
			{Name: "company", Type: FieldTypeString, Required: false, Description: "Company name"},
			{Name: "title", Type: FieldTypeString, Required: false, Description: "Job title"},
			{Name: "location", Type: FieldTypeString, Required: false, Description: "Location/Address"},
		},
	})

	// Product schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "product",
		Description: "Product information extraction",
		Fields: []ExtractionField{
			{Name: "name", Type: FieldTypeString, Required: true, Description: "Product name"},
			{Name: "price", Type: FieldTypeString, Required: false, Description: "Price (with currency)"},
			{Name: "description", Type: FieldTypeString, Required: false, Description: "Product description"},
			{Name: "url", Type: FieldTypeURL, Required: false, Description: "Product URL"},
			{Name: "image_url", Type: FieldTypeURL, Required: false, Description: "Product image URL"},
			{Name: "category", Type: FieldTypeString, Required: false, Description: "Product category"},
			{Name: "rating", Type: FieldTypeString, Required: false, Description: "Product rating"},
			{Name: "availability", Type: FieldTypeString, Required: false, Description: "Availability status"},
		},
	})

	// Article schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "article",
		Description: "Article/content extraction",
		Fields: []ExtractionField{
			{Name: "title", Type: FieldTypeString, Required: true, Description: "Article title"},
			{Name: "author", Type: FieldTypeString, Required: false, Description: "Author name"},
			{Name: "date", Type: FieldTypeString, Required: false, Description: "Publication date"},
			{Name: "content", Type: FieldTypeString, Required: false, Description: "Article content/summary"},
			{Name: "url", Type: FieldTypeURL, Required: false, Description: "Article URL"},
			{Name: "tags", Type: FieldTypeList, Required: false, Description: "Article tags/categories"},
		},
	})

	// Business listing schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "business",
		Description: "Business listing extraction",
		Fields: []ExtractionField{
			{Name: "name", Type: FieldTypeString, Required: true, Description: "Business name"},
			{Name: "address", Type: FieldTypeString, Required: false, Description: "Street address"},
			{Name: "city", Type: FieldTypeString, Required: false, Description: "City"},
			{Name: "phone", Type: FieldTypePhone, Required: false, Description: "Phone number"},
			{Name: "website", Type: FieldTypeURL, Required: false, Description: "Website URL"},
			{Name: "rating", Type: FieldTypeString, Required: false, Description: "Rating/reviews"},
			{Name: "hours", Type: FieldTypeString, Required: false, Description: "Business hours"},
			{Name: "category", Type: FieldTypeString, Required: false, Description: "Business category"},
		},
	})

	// Social media profile schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "social_profile",
		Description: "Social media profile extraction",
		Fields: []ExtractionField{
			{Name: "username", Type: FieldTypeString, Required: true, Description: "Username/handle"},
			{Name: "display_name", Type: FieldTypeString, Required: false, Description: "Display name"},
			{Name: "bio", Type: FieldTypeString, Required: false, Description: "Profile bio/description"},
			{Name: "followers", Type: FieldTypeString, Required: false, Description: "Follower count"},
			{Name: "following", Type: FieldTypeString, Required: false, Description: "Following count"},
			{Name: "posts", Type: FieldTypeString, Required: false, Description: "Post count"},
			{Name: "profile_url", Type: FieldTypeURL, Required: false, Description: "Profile URL"},
			{Name: "avatar_url", Type: FieldTypeURL, Required: false, Description: "Profile image URL"},
		},
	})

	// Job listing schema
	e.RegisterSchema(&ExtractionSchema{
		Name:        "job_listing",
		Description: "Job listing extraction",
		Fields: []ExtractionField{
			{Name: "title", Type: FieldTypeString, Required: true, Description: "Job title"},
			{Name: "company", Type: FieldTypeString, Required: true, Description: "Company name"},
			{Name: "location", Type: FieldTypeString, Required: false, Description: "Job location"},
			{Name: "salary", Type: FieldTypeString, Required: false, Description: "Salary range"},
			{Name: "description", Type: FieldTypeString, Required: false, Description: "Job description"},
			{Name: "requirements", Type: FieldTypeList, Required: false, Description: "Requirements list"},
			{Name: "job_type", Type: FieldTypeString, Required: false, Description: "Full-time/Part-time/Contract"},
			{Name: "posted_date", Type: FieldTypeString, Required: false, Description: "Date posted"},
			{Name: "apply_url", Type: FieldTypeURL, Required: false, Description: "Application URL"},
		},
	})

	// Generic key-value extraction
	e.RegisterSchema(&ExtractionSchema{
		Name:        "generic",
		Description: "Generic key-value extraction",
		Fields: []ExtractionField{
			{Name: "key", Type: FieldTypeString, Required: true, Description: "Field name/key"},
			{Name: "value", Type: FieldTypeString, Required: true, Description: "Field value"},
			{Name: "source", Type: FieldTypeString, Required: false, Description: "Source element/location"},
		},
	})
}

// Validation helpers

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var phoneRegex = regexp.MustCompile(`^[\d\s\-\+\(\)\.]{7,}$`)
var urlRegex = regexp.MustCompile(`^(https?://)?[\w\-]+(\.[\w\-]+)+[/#?]?.*$`)

func isValidEmail(s string) bool {
	return emailRegex.MatchString(strings.TrimSpace(s))
}

func isValidURL(s string) bool {
	return urlRegex.MatchString(strings.TrimSpace(s))
}

func isValidPhone(s string) bool {
	// Remove common separators for validation
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
	// Phone should have at least 7 digits
	return len(cleaned) >= 7 && phoneRegex.MatchString(s)
}

// CalculateExtractionConfidence calculates overall confidence for extracted data.
func CalculateExtractionConfidence(result *ExtractionResult, totalFields, filledFields int) float64 {
	if totalFields == 0 {
		return 0.0
	}

	// Base confidence from field coverage
	coverage := float64(filledFields) / float64(totalFields)

	// Adjust based on validation results
	confidence := coverage * result.Confidence

	// Penalty for errors
	confidence -= float64(len(result.Errors)) * 0.1

	// Minor penalty for warnings
	confidence -= float64(len(result.Warnings)) * 0.02

	// Clamp to valid range
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}
