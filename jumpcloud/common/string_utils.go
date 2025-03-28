package common

import (
	"fmt"
)

// ExpandStringList converts a []interface{} to a []string
func ExpandStringList(input []interface{}) []string {
	if input == nil {
		return nil
	}

	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.(string)
	}
	return result
}

// FlattenStringList converts a []string to a []interface{}
func FlattenStringList(input []string) []interface{} {
	if input == nil {
		return nil
	}

	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

// ExpandAttributes converts a map[string]interface{} of strings to a map[string]interface{} of actual types
func ExpandAttributes(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	result := make(map[string]interface{}, len(input))
	for k, v := range input {
		result[k] = v
	}
	return result
}

// FlattenAttributes converts a map[string]interface{} to a map[string]string
func FlattenAttributes(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	result := make(map[string]interface{}, len(input))
	for k, v := range input {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
