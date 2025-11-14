package commands

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapKeysToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]bool
		expected []string
	}{
		{
			name:     "empty_map",
			input:    map[string]bool{},
			expected: nil,
		},
		{
			name: "single_entry",
			input: map[string]bool{
				"key1": true,
			},
			expected: []string{"key1"},
		},
		{
			name: "multiple_entries",
			input: map[string]bool{
				"key1": true,
				"key2": false,
				"key3": true,
			},
			expected: []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapKeysToSlice(tt.input)

			// Sort both slices for consistent comparison since map iteration order is not guaranteed
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mapKeysToSlice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandTypes(t *testing.T) {
	types := CommandTypes()

	// Verify we have exactly the expected command types
	expectedTypes := []string{"windows", "linux", "mac"}
	if len(types) != len(expectedTypes) {
		t.Errorf("CommandTypes() returned %d types, want %d", len(types), len(expectedTypes))
	}

	// Check each expected type exists
	for _, expectedType := range expectedTypes {
		if _, exists := types[expectedType]; !exists {
			t.Errorf("CommandTypes() missing expected type: %s", expectedType)
		}
	}

	// Verify all values are true
	for k, v := range types {
		if !v {
			t.Errorf("CommandTypes() has false value for key: %s", k)
		}
	}
}

func TestLaunchTypes(t *testing.T) {
	types := LaunchTypes()

	// Verify we have exactly the expected launch types
	expectedTypes := []string{"manual", "trigger", "schedule", "repeated"}
	if len(types) != len(expectedTypes) {
		t.Errorf("LaunchTypes() returned %d types, want %d", len(types), len(expectedTypes))
	}

	// Check each expected type exists
	for _, expectedType := range expectedTypes {
		if _, exists := types[expectedType]; !exists {
			t.Errorf("LaunchTypes() missing expected type: %s", expectedType)
		}
	}

	// Verify all values are true
	for k, v := range types {
		if !v {
			t.Errorf("LaunchTypes() has false value for key: %s", k)
		}
	}
}

func TestTriggerTypes(t *testing.T) {
	types := TriggerTypes()

	// Verify we have exactly the expected trigger types
	expectedTypes := []string{
		"date", "time", "interval", "session_start", "session_stop",
		"network_join", "network_leave", "filesystem", "pending_reboot",
	}
	if len(types) != len(expectedTypes) {
		t.Errorf("TriggerTypes() returned %d types, want %d", len(types), len(expectedTypes))
	}

	// Check each expected type exists
	for _, expectedType := range expectedTypes {
		if _, exists := types[expectedType]; !exists {
			t.Errorf("TriggerTypes() missing expected type: %s", expectedType)
		}
	}

	// Verify all values are true
	for k, v := range types {
		if !v {
			t.Errorf("TriggerTypes() has false value for key: %s", k)
		}
	}
}

func TestCommonCommandSchema(t *testing.T) {
	schema := CommonCommandSchema()

	// Test essential fields exist
	requiredFields := []string{
		"id", "name", "command", "command_type", "user", "sudo",
		"shell", "timeout", "launch_type", "trigger", "schedule",
		"files", "organization_id", "template_variables",
		"created", "updated",
	}

	for _, field := range requiredFields {
		if _, exists := schema[field]; !exists {
			t.Errorf("CommonCommandSchema() missing required field: %s", field)
		}
	}

	// Test specific field properties
	fieldTests := []struct {
		name          string
		required      bool
		computed      bool
		optional      bool
		hasDefault    bool
		hasValidation bool
	}{
		{"id", false, true, false, false, false},
		{"name", true, false, false, false, false},
		{"command", true, false, false, false, false},
		{"command_type", true, false, false, false, true},
		{"user", false, false, true, true, false},
		{"sudo", false, false, true, true, false},
		{"timeout", false, false, true, true, true},
		{"organization_id", false, true, true, false, false},
	}

	for _, tt := range fieldTests {
		t.Run(tt.name, func(t *testing.T) {
			field, exists := schema[tt.name]
			if !exists {
				t.Fatalf("Field %s not found in schema", tt.name)
			}

			if field.Required != tt.required {
				t.Errorf("Field %s has Required=%v, want %v", tt.name, field.Required, tt.required)
			}

			if field.Computed != tt.computed {
				t.Errorf("Field %s has Computed=%v, want %v", tt.name, field.Computed, tt.computed)
			}

			if field.Optional != tt.optional {
				t.Errorf("Field %s has Optional=%v, want %v", tt.name, field.Optional, tt.optional)
			}

			if (field.Default != nil) != tt.hasDefault {
				t.Errorf("Field %s has Default=%v, expected hasDefault=%v", tt.name, field.Default, tt.hasDefault)
			}

			if (field.ValidateFunc != nil) != tt.hasValidation {
				t.Errorf("Field %s has ValidateFunc=%v, expected hasValidation=%v", tt.name, field.ValidateFunc != nil, tt.hasValidation)
			}
		})
	}
}
