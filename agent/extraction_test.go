package agent

import (
	"testing"
)

func TestNewExtractor(t *testing.T) {
	e := NewExtractor()
	if e == nil {
		t.Fatal("NewExtractor() returned nil")
	}
	if e.schemas == nil {
		t.Error("schemas map should be initialized")
	}

	// Check built-in schemas are registered
	builtIn := []string{"contact", "product", "article", "business", "social_profile", "job_listing", "generic"}
	for _, name := range builtIn {
		if _, ok := e.GetSchema(name); !ok {
			t.Errorf("built-in schema %q not found", name)
		}
	}
}

func TestExtractor_RegisterSchema(t *testing.T) {
	e := NewExtractor()

	customSchema := &ExtractionSchema{
		Name:        "custom",
		Description: "Custom test schema",
		Fields: []ExtractionField{
			{Name: "field1", Type: FieldTypeString, Required: true},
		},
	}

	e.RegisterSchema(customSchema)

	schema, ok := e.GetSchema("custom")
	if !ok {
		t.Fatal("custom schema not found after registration")
	}
	if schema.Name != "custom" {
		t.Errorf("schema.Name = %q, want 'custom'", schema.Name)
	}
}

func TestExtractor_ListSchemas(t *testing.T) {
	e := NewExtractor()
	schemas := e.ListSchemas()

	if len(schemas) < 7 {
		t.Errorf("Expected at least 7 built-in schemas, got %d", len(schemas))
	}

	// Check that built-in schemas are present
	schemaSet := make(map[string]bool)
	for _, s := range schemas {
		schemaSet[s] = true
	}

	if !schemaSet["contact"] {
		t.Error("contact schema should be in list")
	}
	if !schemaSet["product"] {
		t.Error("product schema should be in list")
	}
}

func TestExtractor_Validate_Contact(t *testing.T) {
	e := NewExtractor()

	t.Run("valid contact", func(t *testing.T) {
		data := map[string]any{
			"name":    "John Doe",
			"email":   "john@example.com",
			"phone":   "+1-555-123-4567",
			"company": "Acme Inc",
		}

		result := e.Validate("contact", data)
		if !result.Valid {
			t.Errorf("Expected valid, got errors: %v", result.Errors)
		}
		if result.Confidence < 0.5 {
			t.Errorf("Confidence too low: %f", result.Confidence)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		data := map[string]any{
			"email": "john@example.com",
		}

		result := e.Validate("contact", data)
		if result.Valid {
			t.Error("Expected invalid due to missing required 'name' field")
		}
		if len(result.Errors) == 0 {
			t.Error("Expected errors for missing required field")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		data := map[string]any{
			"name":  "John Doe",
			"email": "invalid-email",
		}

		result := e.Validate("contact", data)
		if result.Valid {
			t.Error("Expected invalid due to invalid email")
		}
	})
}

func TestExtractor_Validate_Product(t *testing.T) {
	e := NewExtractor()

	t.Run("valid product", func(t *testing.T) {
		data := map[string]any{
			"name":        "Laptop",
			"price":       "$999.99",
			"description": "High-performance laptop",
			"url":         "https://example.com/laptop",
		}

		result := e.Validate("product", data)
		if !result.Valid {
			t.Errorf("Expected valid, got errors: %v", result.Errors)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		data := map[string]any{
			"price": "$999.99",
		}

		result := e.Validate("product", data)
		if result.Valid {
			t.Error("Expected invalid due to missing required 'name' field")
		}
	})
}

func TestExtractor_Validate_UnknownSchema(t *testing.T) {
	e := NewExtractor()

	data := map[string]any{"field": "value"}
	result := e.Validate("nonexistent", data)

	if result.Valid {
		t.Error("Expected invalid for unknown schema")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected error message for unknown schema")
	}
}

func TestFieldValidation(t *testing.T) {
	e := NewExtractor()

	t.Run("string type", func(t *testing.T) {
		err := e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeString},
			"hello",
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		err = e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeString},
			123,
		)
		if err == nil {
			t.Error("Expected error for non-string value")
		}
	})

	t.Run("number type", func(t *testing.T) {
		err := e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeNumber},
			42,
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		err = e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeNumber},
			3.14,
		)
		if err != nil {
			t.Errorf("Unexpected error for float: %v", err)
		}
	})

	t.Run("boolean type", func(t *testing.T) {
		err := e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeBoolean},
			true,
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("list type", func(t *testing.T) {
		err := e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeList},
			[]any{"a", "b", "c"},
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		err = e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeList},
			[]string{"a", "b"},
		)
		if err != nil {
			t.Errorf("Unexpected error for []string: %v", err)
		}
	})

	t.Run("object type", func(t *testing.T) {
		err := e.validateFieldType(
			ExtractionField{Name: "test", Type: FieldTypeObject},
			map[string]any{"key": "value"},
		)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"test@.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

func TestURLValidation(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://example.com", true},
		{"http://example.com/path", true},
		{"https://example.com/path?query=1", true},
		{"example.com", true},
		{"www.example.com", true},
		{"", false},
		{"not a url", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.valid {
				t.Errorf("isValidURL(%q) = %v, want %v", tt.url, result, tt.valid)
			}
		})
	}
}

func TestPhoneValidation(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"+1-555-123-4567", true},
		{"555-123-4567", true},
		{"(555) 123-4567", true},
		{"5551234567", true},
		{"123", false},
		{"", false},
		{"not a phone", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := isValidPhone(tt.phone)
			if result != tt.valid {
				t.Errorf("isValidPhone(%q) = %v, want %v", tt.phone, result, tt.valid)
			}
		})
	}
}

func TestPatternValidation(t *testing.T) {
	e := NewExtractor()

	t.Run("matching pattern", func(t *testing.T) {
		err := e.validatePattern(`^\d{3}-\d{4}$`, "555-1234")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("non-matching pattern", func(t *testing.T) {
		err := e.validatePattern(`^\d{3}-\d{4}$`, "invalid")
		if err == nil {
			t.Error("Expected error for non-matching pattern")
		}
	})

	t.Run("invalid regex", func(t *testing.T) {
		err := e.validatePattern(`[invalid`, "test")
		if err == nil {
			t.Error("Expected error for invalid regex")
		}
	})
}

func TestCalculateExtractionConfidence(t *testing.T) {
	t.Run("full confidence", func(t *testing.T) {
		result := &ExtractionResult{
			Valid:      true,
			Confidence: 1.0,
		}
		conf := CalculateExtractionConfidence(result, 5, 5)
		if conf != 1.0 {
			t.Errorf("Expected 1.0, got %f", conf)
		}
	})

	t.Run("partial fields", func(t *testing.T) {
		result := &ExtractionResult{
			Valid:      true,
			Confidence: 1.0,
		}
		conf := CalculateExtractionConfidence(result, 10, 5)
		if conf != 0.5 {
			t.Errorf("Expected 0.5, got %f", conf)
		}
	})

	t.Run("with errors", func(t *testing.T) {
		result := &ExtractionResult{
			Valid:      false,
			Confidence: 1.0,
			Errors:     []string{"error1", "error2"},
		}
		conf := CalculateExtractionConfidence(result, 5, 5)
		if conf >= 1.0 {
			t.Error("Confidence should be reduced for errors")
		}
	})

	t.Run("zero fields", func(t *testing.T) {
		result := &ExtractionResult{
			Valid:      true,
			Confidence: 1.0,
		}
		conf := CalculateExtractionConfidence(result, 0, 0)
		if conf != 0.0 {
			t.Errorf("Expected 0.0 for zero fields, got %f", conf)
		}
	})
}

func TestExtractionResult(t *testing.T) {
	result := &ExtractionResult{
		Schema:     "contact",
		Data:       map[string]any{"name": "John"},
		Confidence: 0.85,
		Valid:      true,
		Errors:     nil,
		Warnings:   []string{"optional field missing"},
	}

	if result.Schema != "contact" {
		t.Errorf("Schema = %q", result.Schema)
	}
	if result.Confidence != 0.85 {
		t.Errorf("Confidence = %f", result.Confidence)
	}
	if !result.Valid {
		t.Error("Valid should be true")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

// Benchmark

func BenchmarkValidate_Contact(b *testing.B) {
	e := NewExtractor()
	data := map[string]any{
		"name":    "John Doe",
		"email":   "john@example.com",
		"phone":   "+1-555-123-4567",
		"company": "Acme Inc",
		"title":   "Engineer",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Validate("contact", data)
	}
}

func BenchmarkValidate_Product(b *testing.B) {
	e := NewExtractor()
	data := map[string]any{
		"name":         "Product",
		"price":        "$99.99",
		"description":  "A great product",
		"url":          "https://example.com/product",
		"category":     "Electronics",
		"availability": "In Stock",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Validate("product", data)
	}
}
