package common

import (
	"encoding/json"
)

// FlattenMapToJSON converts a map to a JSON string
func FlattenMapToJSON(m map[string]interface{}) string {
	if m == nil {
		return ""
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(bytes)
}

// ExpandJSONToMap converts a JSON string to a map
func ExpandJSONToMap(s string) (map[string]interface{}, error) {
	if s == "" {
		return nil, nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FlattenGenericList converts a generic slice to a list of interfaces for Terraform state
func FlattenGenericList[T any](items []T, converter func(T) map[string]interface{}) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, converter(item))
	}
	return result
}

// ExpandGenericList converts a list of interfaces from Terraform state to a generic slice
func ExpandGenericList[T any](items []interface{}, converter func(map[string]interface{}) T) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if mapItem, ok := item.(map[string]interface{}); ok {
			result = append(result, converter(mapItem))
		}
	}
	return result
}

// FlattenMapStringInterface creates a flattened version of a map[string]interface{} for Terraform state
func FlattenMapStringInterface(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}
	return result
}

// ExpandMapStringInterface expands a map[string]interface{} from Terraform state
func ExpandMapStringInterface(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}
	return result
}
