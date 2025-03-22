package common

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SuppressEquivalentJSONDiffs suppresses differences between JSON strings that are semantically equivalent
func SuppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	// If the strings are equal, there's nothing to suppress
	if old == new {
		return true
	}

	// If both are empty, they're equivalent
	if old == "" && new == "" {
		return true
	}

	// If one is empty and the other isn't, they're not equivalent
	if (old == "" && new != "") || (old != "" && new == "") {
		return false
	}

	// Unmarshal both strings into interface{} values
	var oldObj, newObj interface{}
	if err := unmarshalJSONString(old, &oldObj); err != nil {
		return false
	}
	if err := unmarshalJSONString(new, &newObj); err != nil {
		return false
	}

	// Compare the unmarshaled objects
	return reflect.DeepEqual(oldObj, newObj)
}

// unmarshalJSONString unmarshals a JSON string into an interface
func unmarshalJSONString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
